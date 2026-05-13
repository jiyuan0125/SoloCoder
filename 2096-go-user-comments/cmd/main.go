package main

import (
	"log"
	"net/http"
	"strings"

	"usercomments/internal/database"
	"usercomments/internal/handler"
)

func main() {
	if err := database.InitDB("./comments.db"); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.DB.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			return
		}
		handler.GetArticlesHandler(w, r)
	})

	mux.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateCommentHandler(w, r)
		case http.MethodGet:
			handler.GetCommentsHandler(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/comments/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		if len(parts) >= 3 && parts[2] == "like" {
			if r.Method != http.MethodPost {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
				return
			}
			handler.LikeCommentHandler(w, r)
			return
		}

		if len(parts) >= 2 {
			if r.Method != http.MethodDelete {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
				return
			}
			handler.DeleteCommentHandler(w, r)
			return
		}

		http.Error(w, "无效的请求路径", http.StatusBadRequest)
	})

	mux.HandleFunc("/search/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			return
		}
		handler.SearchCommentsHandler(w, r)
	})

	mux.HandleFunc("/search/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			return
		}
		handler.SearchCommentsByNicknameHandler(w, r)
	})

	log.Println("评论系统服务启动在 :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
