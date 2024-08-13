package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// 首页文章数量
const postsPerPage = 10

var success bool

func generateBlogPages() bool {

	// 加载博客配置
	BlogConfig, err := LoadBlogConfig("./data/.env")
	if err != nil {
		success = false

		panic(err)
	}

	// 读取所有文章的元数据并排序
	posts, err := ReadPostMetadata("./data/posts")
	if err != nil {
		success = false

		panic(err)
	}

	// 按日期排序文章
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	funcMap := template.FuncMap{
		"safeHTML": safeHTML,
		"add":      func(x, y int) int { return x + y },
		"sub":      func(x, y int) int { return x - y },
		"indexOf": func(slice []string, item string) int {
			for i, v := range slice {
				if v == item {
					return i
				}
			}
			return -1
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"./data/templates/header.html",
		"./data/templates/footer.html",
		"./data/templates/index.html",
		"./data/templates/post.html",
		"./data/templates/comment.html",
	)
	if err != nil {
		success = false
		panic(err)
	}

	//清空public目录
	if err := clearDir("./data/public"); err != nil {
		log.Printf("清空目录失败: %v", err)
	}

	//菜单生成
	menuHTML := ReadMenuConfig("./data/config/menu.config")

	// 生成主页页面
	totalPages := (len(posts) + postsPerPage - 1) / postsPerPage
	for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
		startIndex := pageIndex * postsPerPage
		endIndex := startIndex + postsPerPage
		if endIndex > len(posts) {
			endIndex = len(posts)
		}

		pagePosts := posts[startIndex:endIndex]
		indexPath := "./data/public/index.html"
		if pageIndex > 0 {
			indexPath = filepath.Join("./data/public", "page", strconv.Itoa(pageIndex+1), "index.html")
			os.MkdirAll(filepath.Dir(indexPath), os.ModePerm)
		}

		indexFile, err := os.Create(indexPath)
		if err != nil {
			success = false
			panic(err)
		}
		defer indexFile.Close()

		BlogData := map[string]interface{}{
			"BlogTitle":       BlogConfig.Title,
			"BlogDescription": BlogConfig.Description,
			"BlogURI":         BlogConfig.URI,
			"BlogTags":        BlogConfig.Tags,
			"BlogAuthor":      BlogConfig.Author,
			"Menu":            menuHTML,
			"Posts":           pagePosts,
			"CurrentPage":     pageIndex + 1,
			"TotalPages":      totalPages,
			"PageType":        "index",
		}

		err = tmpl.ExecuteTemplate(indexFile, "index.html", BlogData)
		if err != nil {
			success = false
			panic(err)
		}
	}

	// 生成每篇文章的页面
	for _, post := range posts {
		postDir := "./data/public/" + post.URI
		os.MkdirAll(postDir, os.ModePerm)
		postPath := postDir + "/index.html"
		postFile, err := os.Create(postPath)
		if err != nil {
			continue
		}
		defer postFile.Close()

		post.Content = convertMarkdownToHTML(post.Content)
		err = tmpl.ExecuteTemplate(postFile, "post.html", map[string]interface{}{
			"Title":           post.Title,
			"Content":         post.Content,
			"URI":             post.URI,
			"Description":     post.Description,
			"Category":        post.Category,
			"Date":            post.Date,
			"TagsArray":       post.Tags,
			"Tags":            post.TagsStr,
			"Menu":            menuHTML,
			"BlogTitle":       BlogConfig.Title,
			"BlogDescription": BlogConfig.Description,
			"BlogURI":         BlogConfig.URI,
			"BlogTags":        BlogConfig.Tags,
			"BlogAuthor":      BlogConfig.Author,
			"BlogCommentUri":  BlogConfig.CommentUri,
			"PageType":        "post",
		})
		if err != nil {
			continue
		}
		success = true // 假设大多数情况下都成功
	}

	//复制主题模板下的res静态文件文件夹
	resSrcPath := "./data/templates/res"
	resDstPath := "./data/public/res"
	if err := copyDir(resSrcPath, resDstPath); err != nil {
		success = false
		log.Fatalf("复制资源失败: %v", err)
	}

	//生成feed订阅文件。
	// 确保 /public/feed 目录存在
	feedDir := "./data/public/feed"
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		success = false
		log.Fatalf("为feed 创建目录失败: %v", err)
	}

	// 生成 Atom feed，输出到 /public/feed/index.xml
	atomFeedPath := filepath.Join(feedDir, "index.xml")
	if err := generateAtomFeed(posts, BlogConfig, atomFeedPath); err != nil {
		success = false
		log.Fatalf("生成Atom Feed失败:  %v", err)
	}

	// 生成站点地图
	sitemapPath := "./data/public/sitemap.xml"
	if err := generateSitemap(posts, BlogConfig, sitemapPath); err != nil {
		success = false
		log.Fatalf("生成站点地图失败: %v", err)
	}

	//生成 tag 页面
	posts, err = ReadPostMetadataAndFillTagMap("./data/posts")
	if err != nil {
		success = false
		log.Fatalf("读取文章元数据失败: %v", err)
	}

	GenerateTagPages(posts, BlogConfig, "./data/templates", "./data/public")

	//生成分类页面
	posts, err = ReadPostMetadataAndFillCategoryMap("./data/posts")
	if err != nil {
		success = false
		log.Fatalf("读取文章元数据失败 %v", err)
	}
	GenerateCategoryPages(posts, BlogConfig, "./data/templates", "./data/public")
	//生成搜索页面
	tagsData, err := ReadTags("./data/posts")
	if err != nil {
		log.Fatalf("读取标签错误: %v", err)
		success = false
	}
	GenerateSearchPage(tagsData, BlogConfig, posts, "./data/public")
	//生成robot.txt
	robotTxtPath := "./data/public/robots.txt"
	if err := generateRobotsTxt(posts, BlogConfig, robotTxtPath); err != nil {
		success = false
		log.Fatalf("生成robots.txt失败: %v", err)
	}

	// 根据上述操作的结果返回布尔值
	return success
}

// copyDir 递归地复制一个目录及其子目录和文件。
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	directory, err := os.Open(src)
	if err != nil {
		return err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		srcFilePath := filepath.Join(src, obj.Name())
		dstFilePath := filepath.Join(dst, obj.Name())

		if obj.IsDir() {
			// 递归复制子目录
			err = copyDir(srcFilePath, dstFilePath)
			if err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := copyFile(srcFilePath, dstFilePath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile 复制单个文件。
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func clearDir(dir string) error {
	// 打开目录
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()

	// 读取目录中的文件和子目录
	entries, err := f.Readdir(-1)
	if err != nil {
		return err
	}

	// 遍历所有条目
	for _, entry := range entries {
		// 跳过 .git 目录
		if entry.IsDir() && entry.Name() == ".git" {
			continue
		}

		// 构造完整路径
		fullPath := filepath.Join(dir, entry.Name())

		// 删除文件或目录
		if entry.IsDir() {
			err := os.RemoveAll(fullPath)
			if err != nil {
				return fmt.Errorf("无法删除目录 %s: %v", fullPath, err)
			}
		} else {
			err := os.Remove(fullPath)
			if err != nil {
				return fmt.Errorf("无法删除文件 %s: %v", fullPath, err)
			}
		}
	}

	return nil
}

func generateRobotsTxt(posts []PostMetadata, blogconfigs *BlogConfig, outputPath string) error {
	// 定义 robots.txt 的内容
	robotsContent := `User-agent: *
Disallow: 
Sitemap:` + blogconfigs.URI + `/sitemap.xml`

	// 创建并写入文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(robotsContent)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}
