package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type ConfigData struct {
	UserName    string
	Password    string
	BlogTitle   string
	Description string
	Tags        string
	URI         string
	Author      string
	Email       string
	CommentURI  string
}

// Article 数据结构，用于模板渲染
type Article struct {
	Title    string
	Category string
	URI      string
	Date     string
}

type FTPConfig struct {
	Server   string `json:"server"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Push     bool   `json:"push"`
	RelPath  string `json:"relpath"` // 新增相对路径字段
}
type GitHubConfig struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Token      string `json:"token"`
	Push       bool   `json:"push"`
	Username   string `json:"username"`
	Email      string `json:"email"` // 新增邮箱字段
}

func init() {
	install()
	// 加载 .env 文件
	if err := godotenv.Load("./data/.env"); err != nil {
		log.Fatalf("加载.env文件时出错: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "解析表单错误", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		envUsername := os.Getenv("USER_NAME")
		envPassword := os.Getenv("PASS_WORD")

		if username != envUsername || password != envPassword {
			t := template.Must(template.New("login").Parse(loginForm))
			t.Execute(w, map[string]interface{}{
				"LoginFailed": true,
			})
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:   "logged_in",
				Value:  "true",
				MaxAge: 3600,
			})
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else if r.Method == "GET" {
		t := template.Must(template.New("login").Parse(loginForm))
		t.Execute(w, nil)
	} else {
		http.Error(w, "不受支持的方法", http.StatusMethodNotAllowed)
	}
}

func checkLogin(r *http.Request) bool {
	cookie, err := r.Cookie("logged_in")
	if err != nil {
		return false // 无cookie，未登录
	}
	return cookie.Value == "true"
}

// indexHandler 处理主页请求
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	t := template.Must(template.New("webpage").Parse(BaseTemplate))
	t.Execute(w, map[string]interface{}{"Content": template.HTML(HomePageContent)})
}
func generateHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == "POST" {
		// 执行生成逻辑，并判断成功与否
		success := generateBlogPages() // 假设这个函数返回一个布尔值表示成功与否

		// 从表单获取重定向 URL
		redirectUrl := r.FormValue("redirectUrl")
		if redirectUrl == "" || redirectUrl == "/edit" {
			redirectUrl = "/" // 如果没有提供，则默认回到根目录
		}

		// 附加生成成功或失败的查询参数
		if success {
			http.Redirect(w, r, redirectUrl+"?generateSuccess=true", http.StatusFound)
		} else {
			http.Redirect(w, r, redirectUrl+"?generateSuccess=false", http.StatusFound)
		}
	} else {
		http.Error(w, "不允许", http.StatusMethodNotAllowed)
	}
}

func previewHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fs := http.FileServer(http.Dir("./data/public"))
	http.StripPrefix("/preview/", fs).ServeHTTP(w, r)
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "不允许", http.StatusMethodNotAllowed)
		return
	}

	postMetadatas, err := ReadPostMetadata("./data/posts")
	if err != nil {
		http.Error(w, "服务器内部错误", http.StatusInternalServerError)
		return
	}

	// 按日期排序，由近到远
	sort.Slice(postMetadatas, func(i, j int) bool {
		return postMetadatas[i].Date > postMetadatas[j].Date
	})

	articlesData := struct {
		Content template.HTML
	}{
		Content: template.HTML(renderArticles(postMetadatas)),
	}

	t := template.Must(template.New("webpage").Parse(BaseTemplate))
	t.Execute(w, articlesData)
}

// renderArticles 将文章元数据渲染为HTML字符串
func renderArticles(postMetadatas []PostMetadata) string {
	var articlesHTML strings.Builder
	tmpl := template.Must(template.New("articles").Parse(articlesTemplate))

	tmpl.Execute(&articlesHTML, postMetadatas)
	return articlesHTML.String()
}

func newArticleHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" {
		baseTemplate, err := template.New("base").Parse(BaseTemplate)
		if err != nil {
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

		baseTemplate.Execute(w, map[string]interface{}{
			"Content": template.HTML(newArticle),
		})

	} else if r.Method == "POST" {
		// 处理表单提交
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "解析表单错误", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")
		category := r.FormValue("category")
		tags := r.FormValue("tags")
		date := r.FormValue("date")
		uri := r.FormValue("uri")

		// 创建并写入 Markdown 文件
		filePath := filepath.Join("./data/posts", fmt.Sprintf("%s.md", title))
		file, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "创建文件错误", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		tagsArray := strings.Split(tags, ",")
		for i, tag := range tagsArray {
			tagsArray[i] = fmt.Sprintf("\"%s\"", strings.TrimSpace(tag))
		}
		tagsString := fmt.Sprintf("[%s]", strings.Join(tagsArray, ", "))

		mdContent := strings.ReplaceAll(fmt.Sprintf(
			`---

title: "%s"

description: "%s"

category: "%s"

tags: %s

date: "%s"

uri: "%s"

---`, title, description, category, tagsString, date, uri), "\r\n", "\n")

		_, err = file.WriteString(mdContent)
		if err != nil {
			http.Error(w, "写文件错误", http.StatusInternalServerError)
			return
		}
		// 重定向到编辑页面
		http.Redirect(w, r, fmt.Sprintf("/edit?title=%s", title), http.StatusFound)
	} else {
		http.Error(w, "不允许", http.StatusMethodNotAllowed)
	}
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" {
		// 从 .env 文件读取配置
		env, err := godotenv.Read("./data/.env")
		if err != nil {
			http.Error(w, "Failed to read .env file", http.StatusInternalServerError)
			return
		}

		// 映射 .env 配置到 ConfigData 结构体
		config := ConfigData{
			UserName:    env["USER_NAME"],
			Password:    env["PASS_WORD"], // 注意：出于安全考虑，通常不建议在前端展示密码
			BlogTitle:   env["BLOG_TITLE"],
			Description: env["BLOG_DESCRIPTION"],
			Tags:        env["BLOG_TAGS"],
			URI:         env["BLOG_URI"],
			Author:      env["BLOG_AUTHOR"],
			Email:       env["EMAIL"],
			CommentURI:  env["COMMENT_URI"],
		}

		// 解析并执行模板
		tmpl, err := template.New("settings").Parse(settingsTemplate)
		if err != nil {
			http.Error(w, "解析设置模板失败", http.StatusInternalServerError)
			return
		}

		var settingsContent bytes.Buffer
		err = tmpl.Execute(&settingsContent, config)
		if err != nil {
			http.Error(w, "解析设置模板失败", http.StatusInternalServerError)
			return
		}

		// 解析 BaseTemplate 并插入 settingsTemplate 内容
		baseTmpl, err := template.New("base").Parse(BaseTemplate)
		if err != nil {
			http.Error(w, "解析基本模板失败", http.StatusInternalServerError)
			return
		}

		baseTmpl.Execute(w, map[string]interface{}{
			"Content": template.HTML(settingsContent.String()),
		})

	} else if r.Method == "POST" {
		// 解析表单数据
		if err := r.ParseForm(); err != nil {
			http.Error(w, "解析表单错误", http.StatusBadRequest)
			return
		}

		// 创建映射以存储更新的环境变量
		envMap := make(map[string]string)
		envMap["USER_NAME"] = r.FormValue("username")               // Form 中的 name 应为 "username"
		envMap["PASS_WORD"] = r.FormValue("password")               // Form 中的 name 应为 "password"
		envMap["BLOG_TITLE"] = r.FormValue("blogtitle")             // Form 中的 name 应为 "blogtitle"
		envMap["BLOG_DESCRIPTION"] = r.FormValue("blogdescription") // Form 中的 name 应为 "blogdescription"
		envMap["BLOG_TAGS"] = r.FormValue("blogtags")               // Form 中的 name 应为 "blogtags"
		envMap["BLOG_URI"] = r.FormValue("bloguri")                 // Form 中的 name 应为 "bloguri"
		envMap["BLOG_AUTHOR"] = r.FormValue("blogauthor")           // Form 中的 name 应为 "blogauthor"
		envMap["EMAIL"] = r.FormValue("email")                      // Form 中的 name 应为 "email"
		envMap["COMMENT_URI"] = r.FormValue("commenturi")           // Form 中的 name 应为 "commenturi"
		// 保存更新后的配置
		err := godotenv.Write(envMap, "./data/.env")
		if err != nil {
			http.Error(w, "保存设置错误", http.StatusInternalServerError)
			return
		}

		// 重新加载页面，显示“保存成功”消息
		http.Redirect(w, r, "/settings?saved=true", http.StatusSeeOther)
	} else {
		http.Error(w, "不允许", http.StatusMethodNotAllowed)
	}
}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	title := r.URL.Query().Get("title")
	confirm := r.URL.Query().Get("confirm")

	if title == "" {
		http.Error(w, "缺少标题参数", http.StatusBadRequest)
		return
	}

	if confirm == "true" {
		// 用户已确认删除操作
		filename := filepath.Join("./data/posts", fmt.Sprintf("%s.md", title))
		if err := os.Remove(filename); err != nil {
			// 处理删除过程中可能发生的错误
			fmt.Fprintf(w, "删除文件失败: %v", err)
			return
		}

		// 删除成功，重定向到 /article
		http.Redirect(w, r, "/article", http.StatusSeeOther)
	}
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	t := template.Must(template.New("webpage").Parse(BaseTemplate))
	t.Execute(w, map[string]interface{}{"Content": template.HTML(DeployContent)})
}

var ftpTemplatestr = template.Must(template.New("ftp").Parse(ftpTemplate))

func ftpHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	configPath := "./data/config/ftp.config"
	config := FTPConfig{}

	// 尝试从现有配置文件中加载配置
	if data, err := ioutil.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config)
	}

	if r.Method == "POST" {
		r.ParseForm()
		config.Server = r.FormValue("server")
		config.Port = r.FormValue("port")
		config.Username = r.FormValue("username")
		config.Password = r.FormValue("password")
		config.Push = r.FormValue("push") == "on"
		config.RelPath = r.FormValue("relpath") // 保存表单中的相对路径

		data, _ := json.Marshal(config)
		ioutil.WriteFile(configPath, data, 0644)

		if config.Push {
			if err := pushToFTP(config); err != nil {
				http.Redirect(w, r, "/ftp?success=false", http.StatusFound)
				return
			}
		}

		http.Redirect(w, r, "/ftp?success=true", http.StatusFound)
		return
	} else {
		if r.Method == "GET" {
			// 在处理 GET 请求时渲染页面
			tmpl, err := template.New("ftp").Parse(ftpTemplate)
			if err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}

			var contentBuilder strings.Builder
			if err := tmpl.Execute(&contentBuilder, config); err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}

			baseTmpl, err := template.New("base").Parse(BaseTemplate)
			if err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}

			// 使用匿名结构体来传递 Content 和 FTPConfig
			if err := baseTmpl.Execute(w, struct {
				Content template.HTML
				Config  FTPConfig
			}{
				Content: template.HTML(contentBuilder.String()),
				Config:  config,
			}); err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}
		}
	}
}

func githubHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	configPath := "./data/config/github.config"
	config := GitHubConfig{}

	// 尝试从现有配置文件中加载配置
	if data, err := ioutil.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config)
	}

	if r.Method == "POST" {
		r.ParseForm()
		config.Repository = r.FormValue("repository")
		config.Branch = r.FormValue("branch")
		config.Token = r.FormValue("token")
		config.Username = r.FormValue("username")
		config.Email = r.FormValue("email") // 从表单中获取邮箱并更新配置
		config.Push = r.FormValue("push") == "on"

		data, _ := json.Marshal(config)
		ioutil.WriteFile(configPath, data, 0644)

		if config.Push {
			if err := pushToGitHub(config); err != nil {
				http.Redirect(w, r, "/github?success=false", http.StatusFound)
				return
			}
		}

		http.Redirect(w, r, "/github?success=true", http.StatusFound)
		return
	} else {
		if r.Method == "GET" {
			// 在处理 GET 请求时渲染页面
			tmpl, err := template.New("github").Parse(githubTemplate)
			if err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}

			var contentBuilder strings.Builder
			if err := tmpl.Execute(&contentBuilder, config); err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}

			baseTmpl, err := template.New("base").Parse(BaseTemplate)
			if err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}

			// 使用匿名结构体来传递 Content 和 githubConfig
			if err := baseTmpl.Execute(w, struct {
				Content template.HTML
				Config  GitHubConfig
			}{
				Content: template.HTML(contentBuilder.String()),
				Config:  config,
			}); err != nil {
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
				return
			}
		}
	}
}
func editHandler(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		// 未登录，重定向到登录页
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	postsDir := "./data/posts"
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "缺少标题参数", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(postsDir, title+".md")

	switch r.Method {
	case "GET":
		contentBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			http.Error(w, "读取文章失败", http.StatusInternalServerError)
			return
		}
		content := string(contentBytes)

		// 使用 editTemplate 渲染内容部分
		editTmpl, err := template.New("edit").Parse(editTemplate)
		if err != nil {
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

		var editContentBuilder strings.Builder
		err = editTmpl.Execute(&editContentBuilder, map[string]string{"Content": content})
		if err != nil {
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

		// 将 editTemplate 的渲染结果作为 Content 插入到 BaseTemplate 并渲染整个页面
		baseTmpl, err := template.New("base").Parse(BaseTemplate)
		if err != nil {
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

		err = baseTmpl.Execute(w, map[string]template.HTML{"Content": template.HTML(editContentBuilder.String())})
		if err != nil {
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "解析表单失败", http.StatusInternalServerError)
			return
		}

		editedContent := strings.ReplaceAll(r.FormValue("content"), "\r\n", "\n")
		title := r.URL.Query().Get("title") // 获取标题

		filePath := filepath.Join(postsDir, title+".md")
		if err := ioutil.WriteFile(filePath, []byte(editedContent), 0644); err != nil {
			// 保存失败，重定向时带上失败的标志
			http.Redirect(w, r, "/edit?title="+title+"&save=failed", http.StatusFound)
			return
		}

		// 保存成功，重定向时带上成功的标志
		http.Redirect(w, r, "/edit?title="+title+"&save=success", http.StatusFound)
	}
}

// 鉴权中间件
func authMiddleware(c *gin.Context) {
	// 假设我们通过查询参数 token 来简单实现鉴权
	token := c.Query("token")

	// 这里应该替换为更复杂的逻辑，比如与数据库中的token比对
	if token != "secret-token" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	c.Next()
}

//go:embed res/*
var content embed.FS

// resHandler 返回一个处理嵌入文件请求的 http.HandlerFunc
// editHandler 处理嵌入的文件请求
func resHandler(w http.ResponseWriter, r *http.Request) {
	// 处理请求路径
	path := r.URL.Path[len("/res/"):]

	// 从嵌入的文件系统中读取文件
	data, err := content.ReadFile("res/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 根据文件类型设置 Content-Type
	switch {
	case path[len(path)-4:] == ".html":
		w.Header().Set("Content-Type", "text/html")
	case path[len(path)-4:] == ".css":
		w.Header().Set("Content-Type", "text/css")
	case path[len(path)-4:] == ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case path[len(path)-4:] == ".png":
		w.Header().Set("Content-Type", "image/png")
	case path[len(path)-4:] == ".jpg" || path[len(path)-5:] == ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// 发送文件内容
	w.Write(data)
}
