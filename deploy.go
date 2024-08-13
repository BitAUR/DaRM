package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

func pushToFTP(config FTPConfig) error {
	c, err := ftp.Dial(config.Server+":"+config.Port, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Printf("无法连接到FTP服务器： %v\n", err)
		return err
	}

	defer c.Quit()

	if err := c.Login(config.Username, config.Password); err != nil {
		log.Printf("无法连接到FTP服务器： %v\n", err)
		return err
	}

	// Process and clean the relative path from config
	configRelPath := strings.Replace(config.RelPath, string(os.PathSeparator), "/", -1)
	if !strings.HasPrefix(configRelPath, "/") {
		configRelPath = "/" + configRelPath
	}
	if !strings.HasSuffix(configRelPath, "/") {
		configRelPath += "/"
	}

	err = filepath.Walk("./data/public", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("遍历 public 目录时出错： %v\n", err)
			return err
		}

		if !info.IsDir() {
			localPath := strings.Replace(path, string(os.PathSeparator), "/", -1)
			relPath, err := filepath.Rel("./data/public", localPath)
			if err != nil {
				log.Printf("无法获取文件的相对路径： %s, 错误: %v\n", localPath, err)
				return err
			}
			relPath = strings.Replace(relPath, string(os.PathSeparator), "/", -1)

			// Prepend the user-specified relative path
			ftpPath := configRelPath + relPath

			// Attempt to create the directory structure
			dirPath, _ := filepath.Split(ftpPath)
			parts := strings.Split(dirPath, "/")
			for i := range parts {
				partPath := strings.Join(parts[:i+1], "/")
				if partPath != "" && partPath != "/" {
					if err := c.MakeDir(partPath); err != nil {
					}
				}
			}

			file, err := os.Open(path)
			if err != nil {
				log.Printf("无法打开文件 %s, 错误: %v\n", path, err)
				return err
			}
			defer file.Close()

			if err = c.Stor(ftpPath, file); err != nil {
				log.Printf("无法上传文件 %s, 错误: %v\n", ftpPath, err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("FTP上载过程中出错： %v\n", err)
		return err
	}

	log.Println("FTP上传成功。")
	return nil
}

func pushToGitHub(config GitHubConfig) error {
	publicDir := "./data/public"
	// 验证令牌和仓库 URL 是否存在
	if config.Token == "" {
		log.Println("GitHub令牌为空。请提供有效的令牌。")
		return errors.New("GitHub token is empty")
	}

	if !strings.HasPrefix(config.Repository, "https://") {
		log.Println("无效的存储库URL。URL应以“https://”开头。")
		return errors.New("无效的存储库URL")
	}

	// 构建带有令牌的 URL
	repoURLWithToken := strings.Replace(config.Repository, "https://", "https://"+strings.TrimSpace(config.Token)+"@", 1)

	// 初始化 git 仓库的命令
	gitInitCmd := exec.Command("git", "init")
	gitInitCmd.Dir = publicDir

	// 检查 .git 目录是否存在
	if _, err := os.Stat(filepath.Join(publicDir, ".git")); os.IsNotExist(err) {
		// .git 目录不存在，初始化 git 仓库
		if output, err := gitInitCmd.CombinedOutput(); err != nil {
			log.Printf("初始化 git 仓库失败: %v, 错误: %s", err, string(output))
			return err
		}
		log.Println("Git 仓库初始化成功")
	}

	// Check if the remote already exists and update it if necessary
	checkRemoteCmd := exec.Command("git", "remote", "get-url", "origin")
	checkRemoteCmd.Dir = publicDir

	if _, err := checkRemoteCmd.Output(); err == nil {
		// Remote exists, set the URL
		setRemoteCmd := exec.Command("git", "remote", "set-url", "origin", repoURLWithToken)
		setRemoteCmd.Dir = publicDir
		if output, err := setRemoteCmd.CombinedOutput(); err != nil {
			log.Printf("设置远程存储库URL失败: %v, 错误: %s", err, string(output))
			return err
		}
	} else {
		// Remote does not exist, add it
		addRemoteCmd := exec.Command("git", "remote", "add", "origin", repoURLWithToken)
		addRemoteCmd.Dir = publicDir
		if output, err := addRemoteCmd.CombinedOutput(); err != nil {
			log.Printf("添加远程存储库失败: %v, 错误: %s", err, string(output))
			return err
		}
	}

	// Set user config
	setUserCmd := exec.Command("git", "config", "user.name", config.Username)
	setUserCmd.Dir = publicDir
	if output, err := setUserCmd.CombinedOutput(); err != nil {
		log.Printf("设置git user.name失败 %v, 错误: %s", err, string(output))
		return err
	}

	setEmailCmd := exec.Command("git", "config", "user.email", config.Email)
	setEmailCmd.Dir = publicDir
	if output, err := setEmailCmd.CombinedOutput(); err != nil {
		log.Printf("设置git user.email失败: %v, 错误: %s", err, string(output))
		return err
	}

	// Add changes
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = publicDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		log.Printf("添加更改失败: %v, 错误: %s", err, string(output))
		return err
	}

	// Commit changes, check if there's anything to commit
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = publicDir
	output, err := statusCmd.Output()
	if err != nil {
		log.Printf("检查git状态失败: %v", err)
		return err
	}

	if len(output) > 0 {
		commitCmd := exec.Command("git", "commit", "-m", "Update configurations")
		commitCmd.Dir = publicDir
		if output, err := commitCmd.CombinedOutput(); err != nil {
			log.Printf("提交更改失败: %v, 错误: %s", err, string(output))
			// Don't return error here to allow push attempt
		}
	} else {
		log.Println("没有要提交的更改。")
	}

	// Push changes
	branchCmd := exec.Command("git", "branch", "-M", config.Branch)
	branchCmd.Dir = publicDir
	if output, err := branchCmd.CombinedOutput(); err != nil {
		log.Printf("未能推送到GitHub: %v, 错误: %s", err, string(output))
		return err
	}

	// Push changes
	pushCmd := exec.Command("git", "push", "-u", "origin", config.Branch)
	pushCmd.Dir = publicDir
	if output, err := pushCmd.CombinedOutput(); err != nil {
		log.Printf("未能推送到GitHub: %v, 错误: %s", err, string(output))
		return err
	}

	return nil
}
