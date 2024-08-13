package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type CategoryInfo struct {
	URIs []string // 存储具有相同分类的所有文章的 URI
}

// categoryMap 用于存储分类到 CategoryInfo 的映射
var categoryMap = make(map[string]*CategoryInfo)

func ReadPostMetadataAndFillCategoryMap(postPath string) ([]PostMetadata, error) {
	var posts []PostMetadata
	// 重置 categoryMap
	categoryMap = make(map[string]*CategoryInfo)

	files, err := ioutil.ReadDir(postPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			content, err := ioutil.ReadFile(filepath.Join(postPath, file.Name()))
			if err != nil {
				continue
			}

			sections := strings.SplitN(string(content), "---", 3)
			if len(sections) < 3 {
				continue
			}

			var metadata PostMetadata
			err = yaml.Unmarshal([]byte(sections[1]), &metadata)
			if err != nil {
				continue
			}
			metadata.Content = sections[2]
			posts = append(posts, metadata)

			// 更新 categoryMap
			category := metadata.Category // 假设 PostMetadata 有一个 Category 字段
			if _, exists := categoryMap[category]; !exists {
				categoryMap[category] = &CategoryInfo{}
			}
			categoryMap[category].URIs = append(categoryMap[category].URIs, metadata.URI)
		}
	}

	return posts, nil
}

func GenerateCategoryPages(posts []PostMetadata, blogConfig *BlogConfig, templateDir, outputDir string) {

	funcMap := template.FuncMap{
		"safeHTML": safeHTML,
		"add":      func(x, y int) int { return x + y },
		"sub":      func(x, y int) int { return x - y },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join(templateDir, "header.html"),
		filepath.Join(templateDir, "categories.html"), // 确保有一个 categories.html 模板
		filepath.Join(templateDir, "footer.html"),
	)
	if err != nil {
		log.Fatalf("无法解析模板： %v", err)
	}

	for category, info := range categoryMap {
		var allCategorizedPosts []PostMetadata
		for _, uri := range info.URIs {
			for _, post := range posts {
				if post.URI == uri {
					allCategorizedPosts = append(allCategorizedPosts, post)
				}
			}
		}

		// 对分类下的文章按日期排序
		sort.Slice(allCategorizedPosts, func(i, j int) bool {
			return allCategorizedPosts[i].Date > allCategorizedPosts[j].Date
		})

		// 分页处理
		totalPages := (len(allCategorizedPosts) + postsPerPage - 1) / postsPerPage
		for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
			startIndex := pageIndex * postsPerPage
			endIndex := startIndex + postsPerPage
			if endIndex > len(allCategorizedPosts) {
				endIndex = len(allCategorizedPosts)
			}

			pagePosts := allCategorizedPosts[startIndex:endIndex]
			outputPath := filepath.Join(outputDir, "categories", category, "index.html")
			if pageIndex > 0 {
				outputPath = filepath.Join(outputDir, "categories", category, fmt.Sprintf("page/%d", pageIndex+1), "index.html")
			}
			os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)

			f, err := os.Create(outputPath)
			if err != nil {
				log.Fatalf("无法创建文件 %s: %v", outputPath, err)
			}
			defer f.Close()

			BlogData := map[string]interface{}{
				"BlogTitle":       blogConfig.Title,
				"BlogDescription": blogConfig.Description,
				"BlogURI":         blogConfig.URI,
				"BlogAuthor":      blogConfig.Author,
				"Category":        category,
				"Posts":           pagePosts,
				"CurrentPage":     pageIndex + 1,
				"TotalPages":      totalPages,
				"PageType":        "category",
			}

			err = tmpl.ExecuteTemplate(f, "categories.html", BlogData)
			if err != nil {
				log.Fatalf("无法执行类别的模板 %s at page %d: %v", category, pageIndex+1, err)
			}
		}
	}
}
