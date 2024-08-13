package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/russross/blackfriday/v2"
)

// convertMarkdownToHTML 使用 blackfriday 库将 Markdown 转换为 HTML 并添加特定格式的锚点
func convertMarkdownToHTML(markdown string) string {
	// 首先将 Markdown 转换为 HTML
	output := blackfriday.Run([]byte(markdown))

	// 解析 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(output))
	if err != nil {
		fmt.Println("解析HTML出错:", err)
		return ""
	}

	// 为每个 <h1> - <h6> 标签添加特定格式的锚点和链接
	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		// 提取标题文本并将其转换为 URL 编码，以便用作锚点
		text := s.Text()
		encodedText := url.QueryEscape(text)

		// 创建并设置锚点及链接，链接不包含文本
		anchor := fmt.Sprintf(`<a style="padding:0px" href="#%s" target="_blank"></a>%s`, encodedText, text)
		s.SetHtml(anchor) // 将标题内容设置为锚点链接加上原标题文本
		s.SetAttr("id", encodedText)
	})

	// 输出修改后的 HTML
	htmlString, err := doc.Html()
	if err != nil {
		fmt.Println("生成HTML时出错: ", err)
		return ""
	}

	// goquery.Html() 会将整个文档序列化，包括<html>和<body>标签，我们需要的只是<body>内部的内容
	htmlOutput, err := goquery.NewDocumentFromReader(strings.NewReader(htmlString))
	if err != nil {
		fmt.Println("解析最后的HTML错误: ", err)
		return ""
	}
	bodyContent, err := htmlOutput.Find("body").Html()
	if err != nil {
		fmt.Println("提取正文内容错误: ", err)
		return ""
	}

	return bodyContent
}

// safeHTML 是一个自定义模板函数，用来确保 HTML 内容不会被转义
func safeHTML(html string) template.HTML {
	return template.HTML(html)
}
