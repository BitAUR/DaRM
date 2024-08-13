package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
)

// ReadMenuConfig 读取并解析 menu.config 文件
func ReadMenuConfig(filePath string) template.HTML {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("打开文件失败: %v", err)
	}
	defer file.Close()

	var menuItems []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			menuItem := fmt.Sprintf("<li><a target=\"_blank\" rel=\"noopener\" href=\"%s\">%s</a></li>", parts[1], parts[0])
			menuItems = append(menuItems, menuItem)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}
	return template.HTML(strings.Join(menuItems, "\n"))
}
