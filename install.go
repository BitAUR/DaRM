package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func install() {
	// 要检测和创建的目录
	dirs := []string{"./data/config", "./data/public", "./data/templates"}

	// 检测并创建目录
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				fmt.Printf("创建目录 %s 失败: %v\n", dir, err)
				continue
			}
			fmt.Printf("创建目录 %s 完成\n", dir)
		}
	}

	// 检测并创建 posts 目录及文件
	if err := checkAndCreatePostsDir("./data/posts"); err != nil {
		fmt.Println(err)
		return
	}

	// 检测并创建 config 文件
	checkAndCreateFile("./data/config/ftp.config", `{"server":"127.0.0.1","port":"21","username":"test","password":"test","push":false,"relpath":"/"}`)
	checkAndCreateFile("./data/config/github.config", `{"repository":"","branch":"main","token":"","push":false,"username":""}`)
	checkAndCreateFile("./data/config/menu.config", `Frd:./friendlinks/
Feed:./feed/`)

	// 检测并创建 .env 文件
	checkAndCreateFile("./data/.env", `BLOG_AUTHOR="ROYWANG"
BLOG_DESCRIPTION="Hello,DaRM\\!"
BLOG_TAGS="DaRM"
BLOG_TITLE="DaRM"
BLOG_URI="http://localhost:9740/preview"
COMMENT_URI="http://localhost:9740"
EMAIL="admin@darm.bitaur.com"
PASS_WORD="admin"
USER_NAME="admin"`)

	// 检测 templates 目录中的 index.html 文件
	if err := checkAndDownloadTheme("./data/templates/index.html"); err != nil {
		fmt.Println(err)
		return
	}
}

func checkAndCreatePostsDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", dirPath, err)
		}

		// 创建文件路径时应该直接拼接路径
		filePath := filepath.Join(dirPath, "Hello,Darm!.md")
		err = ioutil.WriteFile(filePath, []byte(`---
title: "Hello,DaRM!" 
description: "Hello,DaRM!" 
category: "BLOG" 
tags: ["DaRM"]
date: "2024-08-10" 
uri: "hello-darm"
---

这是我的第一篇博客！`), 0644)
		if err != nil {
			return fmt.Errorf("创建文件 %s 失败: %v", filePath, err)
		}
		fmt.Printf("创建 %s 完成。\n", filePath)
	}
	return nil
}

func checkAndCreateFile(filePath, defaultContent string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := ioutil.WriteFile(filePath, []byte(defaultContent), 0644)
		if err != nil {
			fmt.Printf("创建 %s 失败: %v\n", filePath, err)
			return
		}
		fmt.Printf("创建 %s 完成\n", filePath)
	}
}

func checkAndDownloadTheme(indexPath string) error {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("未找到主题，正在下载并解压主题...")

		// 下载主题文件
		zipURL := "https://darm.bitaur.com/download/theme.zip"
		resp, err := http.Get(zipURL)
		if err != nil {
			return fmt.Errorf("下载主题失败: %v", err)
		}
		defer resp.Body.Close()

		// 读取下载的数据
		zipData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取主题数据失败: %v", err)
		}

		// 解压主题文件
		zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
		if err != nil {
			return fmt.Errorf("创建 zip 读取器失败: %v", err)
		}

		for _, file := range zipReader.File {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("打开 zip 中的文件 %s 失败: %v", file.Name, err)
			}
			defer rc.Close()

			// 使用正确的目标路径
			fpath := filepath.Join("./data/templates", file.Name)
			if file.FileInfo().IsDir() {
				if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
					return fmt.Errorf("创建目录 %s 失败: %v", fpath, err)
				}
			} else {
				f, err := os.Create(fpath)
				if err != nil {
					return fmt.Errorf("创建文件 %s 失败: %v", fpath, err)
				}
				defer f.Close()

				_, err = io.Copy(f, rc)
				if err != nil {
					return fmt.Errorf("写入文件 %s 失败: %v", fpath, err)
				}
			}
		}

		fmt.Println("主题下载并解压完成。")
	}
	return nil
}
