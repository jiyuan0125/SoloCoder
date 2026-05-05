package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	store := NewStore()
	handler := NewHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handler.CreateArticle(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/articles/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/articles/")
		parts := strings.Split(path, "/")

		if len(parts) == 1 && parts[0] != "" {
			if r.Method == "GET" {
				handler.GetArticle(w, r)
				return
			}
		}

		if len(parts) >= 2 {
			switch parts[1] {
			case "update":
				if r.Method == "POST" {
					handler.UpdateArticle(w, r)
					return
				}
			case "publish":
				if r.Method == "POST" {
					handler.PublishArticle(w, r)
					return
				}
			case "archive":
				if r.Method == "POST" {
					handler.ArchiveArticle(w, r)
					return
				}
			case "rollback":
				if r.Method == "POST" {
					handler.RollbackArticle(w, r)
					return
				}
			case "versions":
				if len(parts) == 2 {
					if r.Method == "GET" {
						handler.GetVersions(w, r)
						return
					}
				} else if len(parts) == 3 {
					if r.Method == "GET" {
						handler.GetVersion(w, r)
						return
					}
				} else if len(parts) >= 5 && parts[3] == "compare" {
					if r.Method == "GET" {
						handler.CompareVersions(w, r)
						return
					}
				}
			}
		}

		http.NotFound(w, r)
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handler.Search(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/favorites/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handler.AddFavorite(w, r)
		case "DELETE":
			handler.RemoveFavorite(w, r)
		case "GET":
			handler.GetFavorites(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/hot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.GetHotArticles(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.GetStats(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	port := ":8080"
	fmt.Printf("知识库服务端启动，监听端口 %s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
