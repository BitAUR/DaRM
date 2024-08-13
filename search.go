package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

// GenerateSearchPage 生成一个包含所有标签的 JavaScript 数组的搜索页面
func GenerateSearchPage(allTags *TagsData, blogConfig *BlogConfig, posts []PostMetadata, outputDir string) {
	searchDir := filepath.Join(outputDir, "search")
	os.MkdirAll(searchDir, os.ModePerm)

	// 创建并写入 index.txt 文件
	indexPath := filepath.Join(searchDir, "index.txt")
	file, err := os.Create(indexPath)
	if err != nil {
		log.Fatalf("创建index.txt失败: %v", err)
	}

	for _, tag := range allTags.AllTag {
		_, err := file.WriteString(tag + "\n")
		if err != nil {
			log.Printf("向index.txt写入标签失败: %v", err)
			continue
		}
	}
	file.Close()

	// 准备模板数据
	data := map[string]interface{}{
		"BlogTitle":       blogConfig.Title,
		"BlogDescription": blogConfig.Description,
		"BlogURI":         blogConfig.URI,
		"BlogTags":        blogConfig.Tags,
		"BlogAuthor":      blogConfig.Author,
		"PageType":        "search",
	}

	// 解析模板
	tmpl, err := template.ParseFiles(
		filepath.Join("./data/templates", "header.html"),
		filepath.Join("./data/templates", "search.html"),
		filepath.Join("./data/templates", "footer.html"),
	)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}

	// 生成 index.html
	outputPath := filepath.Join(searchDir, "index.html")
	searchFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("创建search.html失败: %v", err)
	}
	defer searchFile.Close()

	// 执行模板
	err = tmpl.ExecuteTemplate(searchFile, "search.html", data)
	if err != nil {
		log.Fatalf("执行搜索模板失败: %v", err)
	}
}
