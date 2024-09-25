package main

import (
	"encoding/xml"
	"os"
	"sort"
	"time"
)

type Feed struct {
	XMLName     xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title       string   `xml:"title"`
	ID          string   `xml:"id"`
	Description string   `xml:"description"`
	Updated     string   `xml:"updated"`
	Author      string   `xml:"author"`
	Uri         string   `xml:"uri"`
	Mail        string   `xml:"mail"`
	Link        string   `xml:"link"`
	Entries     []Entry  `xml:"entry"`
}

type Entry struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	ID          string `xml:"id"`
	Updated     string `xml:"updated"`
	Description string `xml:"description"`
	Summary     string `xml:"summary"`
}

// 文章时间的逻辑
func formatPostDate(dateStr string) string {
	return dateStr + "T20:00:00.000Z"
}

// feed生成时间的逻辑
func formatCurrentTime() string {
	now := time.Now().UTC()                       // 获取当前时间并转换为 UTC
	return now.Format("2006-01-02T15:04:05.000Z") // 格式化为所需格式
}

// 生成 Atom feed
func generateAtomFeed(posts []PostMetadata, config *BlogConfig, outputPath string) error {
	feed := Feed{
		Title:       config.Title,
		Link:        config.URI + "/feed/",
		Updated:     formatCurrentTime(),
		Author:      config.Author,
		ID:          config.URI + "/",
		Description: config.Description,
		Uri:         config.URI,
		Mail:        config.Email,
	}

	// 确保文章按日期排序，最新的在前
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	// 选取最新的10篇文章
	var latestPosts []PostMetadata
	if len(posts) > 10 {
		latestPosts = posts[:10]
	} else {
		latestPosts = posts
	}

	for _, post := range latestPosts {
		entry := Entry{
			Title:       post.Title,
			Link:        config.URI + "/" + post.URI + "/",
			ID:          config.URI + "/" + post.URI + "/",
			Updated:     formatPostDate(post.Date), // 调用新函数处理日期
			Description: post.Description,
			Summary:     convertMarkdownToHTML(post.Content),
		}
		feed.Entries = append(feed.Entries, entry)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(feed); err != nil {
		return err
	}

	return nil
}
