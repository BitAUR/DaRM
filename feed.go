package main

import (
	"encoding/xml"
	"os"
	"sort"
	"time"
)

type Feed struct {
	XMLName   xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	ID        string   `xml:"id"`
	Title     string   `xml:"title"`
	Updated   string   `xml:"updated"`
	Generator string   `xml:"generator"`
	Author    []Author `xml:"author"` // 支持多个作者
	Links     []Link   `xml:"link"`   // 支持多个链接
	Subtitle  string   `xml:"subtitle"`
	Logo      string   `xml:"logo"`
	Icon      string   `xml:"icon,omitempty"` // 添加图标
	Rights    string   `xml:"rights"`         // 添加版权声明
	Entries   []Entry  `xml:"entry"`
}

type Author struct {
	Name string `xml:"name"`
	URI  string `xml:"uri"`
}

type Link struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"` // 可选的 rel 属性
}

type Entry struct {
	Title     Content  `xml:"title"` // 使用 Content 以支持 CDATA
	ID        string   `xml:"id"`
	Link      Link     `xml:"link"`
	Updated   string   `xml:"updated"`
	Summary   Content  `xml:"summary"` // 这里使用 Content 结构体
	Content   Content  `xml:"content"`
	Category  Category `xml:"category"` // 类别
	Published string   `xml:"published"`
	Rights    string   `xml:"rights,omitempty"` // 可选的版权声明
}

type Content struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"` // 添加类型属性
}

type Category struct {
	Label string `xml:"label,attr"`
	Term  string `xml:"term,attr,omitempty"` // 可选的 term 属性
}

// 文章时间的逻辑
func formatPostDate(dateStr string) string {
	return dateStr + "T20:00:00.000Z"
}

// feed生成时间的逻辑
func formatCurrentTime() string {
	now := time.Now()                             // 获取当前本地时间
	return now.Format("2006-01-02T15:04:05.000Z") // 格式化为所需格式
}

// 生成 Atom feed
func generateAtomFeed(posts []PostMetadata, config *BlogConfig, outputPath string) error {
	feed := Feed{
		ID:        config.URI + "/",
		Title:     config.Title,
		Updated:   formatCurrentTime(),
		Generator: "DaRM",
		Author:    []Author{{Name: config.Author, URI: config.URI}},
		Links: []Link{
			{Href: config.URI, Rel: "alternate"},
			{Href: config.URI + "/feed/", Rel: "self"},
		},
		Subtitle: config.Description,
		Logo:     config.URI + "/res/image/logo.png",
		Icon:     config.URI + "/res/image/logo.png", // 添加 favicon 链接
		Rights:   "Copyright © 2019 - Now " + config.Title,
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
			Title:     Content{Text: "<![CDATA[" + post.Title + "]]>", Type: "html"},
			ID:        config.URI + "/" + post.URI + "/",
			Link:      Link{Href: config.URI + "/" + post.URI + "/"},
			Updated:   formatPostDate(post.Date),
			Summary:   Content{Text: "<![CDATA[" + post.Description + "]]>", Type: "html"},
			Content:   Content{Text: "<![CDATA[" + convertMarkdownToHTML(post.Content) + "]]>", Type: "html"},
			Category:  Category{Label: post.Category, Term: post.Category},
			Published: formatPostDate(post.Date),
			Rights:    "Copyright © 2019 - Now " + config.Title,
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
