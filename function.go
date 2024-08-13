package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// PostMetadata 用于存储文章头部的元数据
type PostMetadata struct {
	Title       string
	Description string
	Category    string
	Tags        []string `yaml:"tags"`
	TagsStr     string
	Date        string
	URI         string
	Content     string // 新增字段用于存储 Markdown 正文
}

// BlogConfig 用于存储从.env文件中读取的博客配置
type BlogConfig struct {
	Title       string
	Description string
	Tags        string
	URI         string
	Author      string
	Email       string
	CommentUri  string
}

type TagsData struct {
	AllTag []string
}

// 读取并解析 Markdown 文件中的头部信息及正文内容
func ReadPostMetadata(postPath string) ([]PostMetadata, error) {
	install()

	var posts []PostMetadata
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

			// 分割头部信息和正文内容
			sections := strings.SplitN(string(content), "---", 3)
			if len(sections) < 3 {
				continue // 如果没有头部和正文的分隔，则跳过这个文件
			}

			var metadata PostMetadata
			err = yaml.Unmarshal([]byte(sections[1]), &metadata)
			if err != nil {
				continue
			}

			// 将标签数组转换为逗号分隔的字符串
			if len(metadata.Tags) > 0 {
				metadata.TagsStr = strings.Join(metadata.Tags, ",")
			}
			metadata.Content = sections[2] // 存储正文内容
			posts = append(posts, metadata)
		}
	}

	return posts, nil
}

// 读取 .env 文件并解析博客配置信息
func LoadBlogConfig(envPath string) (*BlogConfig, error) {
	var config BlogConfig

	// 使用 godotenv 库加载和解析 .env 文件
	err := godotenv.Overload("./data/.env")
	if err != nil {
		return nil, err
	}

	config.Title = os.Getenv("BLOG_TITLE")
	config.Description = os.Getenv("BLOG_DESCRIPTION")
	config.Tags = os.Getenv("BLOG_TAGS")
	config.URI = os.Getenv("BLOG_URI")
	config.Author = os.Getenv("BLOG_AUTHOR")
	config.Email = os.Getenv("EMAIL")
	config.CommentUri = os.Getenv("COMMENT_URI")

	return &config, nil
}

func ReadTags(postPath string) (*TagsData, error) {
	tagsData := &TagsData{}
	tagSet := make(map[string]bool) // 使用 set 结构去重

	files, err := ioutil.ReadDir(postPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			content, err := ioutil.ReadFile(filepath.Join(postPath, file.Name()))
			if err != nil {
				log.Printf("无法读取文件: %s, 跳过。错误: %v", file.Name(), err)
				continue
			}

			// 分割头部信息和正文内容
			sections := strings.SplitN(string(content), "---", 3)
			if len(sections) < 3 {
				log.Printf("无效的文件格式: %s, 跳过.", file.Name())
				continue
			}

			var metadata PostMetadata
			err = yaml.Unmarshal([]byte(sections[1]), &metadata)
			if err != nil {
				log.Printf("无法解析文件中的YAML: %s, 跳过。 错误: %v", file.Name(), err)
				continue
			}

			for _, tag := range metadata.Tags {
				if _, exists := tagSet[tag]; !exists {
					tagSet[tag] = true
					tagsData.AllTag = append(tagsData.AllTag, tag)
				}
			}
		}
	}

	// 将所有标签转换为逗号分隔的字符串

	return tagsData, nil
}
