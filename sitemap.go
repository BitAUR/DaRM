package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"
)

type URLSet struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Urls    []URL    `xml:"url"`
}

type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
}

func generateSitemap(posts []PostMetadata, blogconfigs *BlogConfig, outputPath string) error {
	urlSet := URLSet{}

	// 包含主页
	homeUrl := URL{
		Loc:        blogconfigs.URI,
		LastMod:    time.Now().Format(time.RFC3339),
		ChangeFreq: "always",
	}
	urlSet.Urls = append(urlSet.Urls, homeUrl)

	// 为每篇博客添加 URL 信息
	for _, post := range posts {
		url := URL{
			Loc:        blogconfigs.URI + "/" + post.URI + "/",
			LastMod:    post.Date,
			ChangeFreq: "weekly",
		}
		urlSet.Urls = append(urlSet.Urls, url)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// 写入 XML 声明
	if _, err := file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"); err != nil {
		return fmt.Errorf("failed to write XML declaration: %w", err)
	}

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		return fmt.Errorf("failed to encode sitemap: %w", err)
	}

	return nil
}
