package main

import (
	"os"
	"sort"
	"strings"
	"time"
)

type Feed struct {
	ID        string
	Title     string
	Updated   string
	Generator string
	Author    []Author
	Links     []Link
	Subtitle  string
	Logo      string
	Rights    string
	Entries   []Entry
}

type Author struct {
	Name string
	URI  string
}

type Link struct {
	Href string
	Rel  string
}

type Entry struct {
	Title     string
	ID        string
	Link      Link
	Updated   string
	Summary   string
	Content   string
	Category  Category
	Published string
	Rights    string
}

type Category struct {
	Label string
	Term  string
}

// 文章时间的逻辑
func formatPostDate(dateStr string) string {
	return dateStr + "T20:00:00.000Z"
}

// feed生成时间的逻辑
func formatCurrentTime() string {
	now := time.Now()
	return now.Format("2006-01-02T15:04:05.000Z")
}

// 生成 Atom feed
func generateAtomFeed(posts []PostMetadata, config *BlogConfig, outputPath string) error {
	var builder strings.Builder

	builder.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n")
	builder.WriteString("<feed xmlns=\"http://www.w3.org/2005/Atom\">\n")
	builder.WriteString("<id>" + config.URI + "/</id>\n")
	builder.WriteString("<title>" + config.Title + "</title>\n")
	builder.WriteString("<updated>" + formatCurrentTime() + "</updated>\n")
	builder.WriteString("<generator>DaRM</generator>\n")
	builder.WriteString("<author><name>" + config.Author + "</name><uri>" + config.URI + "</uri></author>\n")
	builder.WriteString("<link href=\"" + config.URI + "\" rel=\"alternate\"/>\n")
	builder.WriteString("<link href=\"" + config.URI + "/feed/\" rel=\"self\"/>\n")
	builder.WriteString("<subtitle>" + config.Description + "</subtitle>\n")
	builder.WriteString("<logo>" + config.URI + "/res/image/logo.png</logo>\n")
	builder.WriteString("<rights>Copyright © 2019 - Now " + config.Title + "</rights>\n")

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	latestPosts := posts
	if len(posts) > 10 {
		latestPosts = posts[:10]
	}

	for _, post := range latestPosts {
		builder.WriteString("<entry>\n")
		builder.WriteString("<title><![CDATA[" + post.Title + "]]></title>\n")
		builder.WriteString("<id>" + config.URI + "/" + post.URI + "/</id>\n")
		builder.WriteString("<link href=\"" + config.URI + "/" + post.URI + "/\"/>\n")
		builder.WriteString("<updated>" + formatPostDate(post.Date) + "</updated>\n")
		builder.WriteString("<summary type=\"html\"><![CDATA[" + post.Description + "]]></summary>\n")
		builder.WriteString("<content type=\"html\"><![CDATA[" + convertMarkdownToHTML(post.Content) + "]]></content>\n")
		builder.WriteString("<category label=\"" + post.Category + "\" term=\"" + post.Category + "\"/>\n")
		builder.WriteString("<published>" + formatPostDate(post.Date) + "</published>\n")
		builder.WriteString("<rights>Copyright © 2019 - Now " + config.Title + "</rights>\n")
		builder.WriteString("</entry>\n")
	}

	builder.WriteString("</feed>")

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(builder.String())
	return err
}
