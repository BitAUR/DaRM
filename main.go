package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	install()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/preview/", previewHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/article", articleHandler)
	http.HandleFunc("/new", newArticleHandler)
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/deploy", deployHandler)
	http.HandleFunc("/ftp", ftpHandler)
	http.HandleFunc("/github", githubHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/res/", resHandler)

	http.HandleFunc("/delete-success", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "文件删除成功。")
	})

	log.Println("服务地址： http://localhost:9740/")
	log.Fatal(http.ListenAndServe(":9740", nil))

}
