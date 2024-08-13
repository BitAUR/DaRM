package main

import (
	"encoding/xml"
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

// generateSitemap 更新以包含主页和 changefreq 标签
func generateSitemap(posts []PostMetadata, blogconfigs *BlogConfig, outputPath string) error {
	urlSet := URLSet{}

	// 包含主页
	homeUrl := URL{
		Loc:        blogconfigs.URI,
		LastMod:    time.Now().Format(time.RFC3339), // 或者用你博客的最近更新时间
		ChangeFreq: "always",                        // 或者根据你的更新频率：daily, weekly, monthly 等
	}
	urlSet.Urls = append(urlSet.Urls, homeUrl)

	// 为每篇博客添加 URL 信息
	for _, post := range posts {
		url := URL{
			Loc:        blogconfigs.URI + "/" + post.URI + "/",
			LastMod:    post.Date,
			ChangeFreq: "weekly", // 假设博客文章不经常变化，可以根据实际情况调整
		}
		urlSet.Urls = append(urlSet.Urls, url)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		return err
	}

	return nil
}
