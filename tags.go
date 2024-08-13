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

type TagInfo struct {
	URIs []string // 存储具有相同标签的所有文章的 URI
}

// tagMap 用于存储标签到 TagInfo 的映射
var tagMap = make(map[string]*TagInfo)

func ReadPostMetadataAndFillTagMap(postPath string) ([]PostMetadata, error) {
	var posts []PostMetadata
	// 重置 tagMap
	tagMap = make(map[string]*TagInfo)

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

			// 更新 tagMap
			for _, tag := range metadata.Tags {
				if _, exists := tagMap[tag]; !exists {
					tagMap[tag] = &TagInfo{}
				}
				tagMap[tag].URIs = append(tagMap[tag].URIs, metadata.URI)
			}
		}
	}

	return posts, nil
}

func GenerateTagPages(posts []PostMetadata, blogConfig *BlogConfig, templateDir, outputDir string) {

	funcMap := template.FuncMap{
		"safeHTML": safeHTML,
		"add":      func(x, y int) int { return x + y },
		"sub":      func(x, y int) int { return x - y },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join(templateDir, "header.html"),
		filepath.Join(templateDir, "tags.html"),
		filepath.Join(templateDir, "footer.html"),
	)
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}

	for tag, info := range tagMap {
		var allTaggedPosts []PostMetadata
		for _, uri := range info.URIs {
			for _, post := range posts {
				if post.URI == uri {
					allTaggedPosts = append(allTaggedPosts, post)
				}
			}
		}

		// 对标签下的文章按日期排序
		sort.Slice(allTaggedPosts, func(i, j int) bool {
			return allTaggedPosts[i].Date > allTaggedPosts[j].Date
		})

		// 分页处理
		totalPages := (len(allTaggedPosts) + postsPerPage - 1) / postsPerPage
		for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
			startIndex := pageIndex * postsPerPage
			endIndex := startIndex + postsPerPage
			if endIndex > len(allTaggedPosts) {
				endIndex = len(allTaggedPosts)
			}

			pagePosts := allTaggedPosts[startIndex:endIndex]
			outputPath := filepath.Join(outputDir, "tags", tag, "index.html")
			if pageIndex > 0 {
				outputPath = filepath.Join(outputDir, "tags", tag, fmt.Sprintf("page/%d", pageIndex+1), "index.html")
			}
			os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)

			f, err := os.Create(outputPath)
			if err != nil {
				log.Fatalf("创建文件失败 %s: %v", outputPath, err)
			}
			defer f.Close()

			// 传递数据到模板
			BlogData := map[string]interface{}{
				"BlogTitle":       blogConfig.Title,
				"BlogDescription": blogConfig.Description,
				"BlogURI":         blogConfig.URI,
				"BlogTags":        blogConfig.Tags,
				"BlogAuthor":      blogConfig.Author,
				"Tag":             tag,
				"Posts":           pagePosts,
				"CurrentPage":     pageIndex + 1,
				"TotalPages":      totalPages,
				"PageType":        "tag",
			}

			err = tmpl.ExecuteTemplate(f, "tags.html", BlogData)
			if err != nil {
				log.Fatalf("执行标签模板在页面 %s 失败  %d: %v", tag, pageIndex+1, err)
			}
		}
	}
}
