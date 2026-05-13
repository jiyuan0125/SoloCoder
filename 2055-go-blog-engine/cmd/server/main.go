package main

import (
	"log"
	"net/http"
	"strings"

	"blog-engine/internal/article"
	"blog-engine/internal/comment"
	"blog-engine/internal/middleware"
	"blog-engine/internal/user"
	"blog-engine/pkg/db"
)

type Router struct{}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	pathParts := strings.Split(path, "/")

	if method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if len(pathParts) >= 2 {
		switch pathParts[1] {
		case "users":
			handleUsers(w, req, pathParts, method)
			return
		case "articles":
			handleArticles(w, req, pathParts, method)
			return
		case "categories":
			handleCategories(w, req, pathParts, method)
			return
		case "tags":
			handleTags(w, req, pathParts, method)
			return
		case "comments":
			handleComments(w, req, pathParts, method)
			return
		}
	}

	if path == "/" && method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"Blog Engine API"}`))
		return
	}

	http.NotFound(w, req)
}

func handleUsers(w http.ResponseWriter, r *http.Request, parts []string, method string) {
	if len(parts) == 2 {
		switch method {
		case "POST":
			if parts[1] == "users" {
				if len(parts) == 2 {
					user.Register(w, r)
					return
				}
			}
		}
	}

	if len(parts) >= 3 {
		switch parts[2] {
		case "register":
			if method == "POST" {
				user.Register(w, r)
				return
			}
		case "login":
			if method == "POST" {
				user.Login(w, r)
				return
			}
		case "logout":
			if method == "POST" {
				user.Logout(w, r)
				return
			}
		case "me":
			if method == "GET" {
				authRequired(w, r, user.GetCurrentUser)
				return
			}
			if method == "PUT" {
				authRequired(w, r, user.UpdateProfile)
				return
			}
		}
	}

	http.NotFound(w, r)
}

func handleArticles(w http.ResponseWriter, r *http.Request, parts []string, method string) {
	if len(parts) == 2 {
		switch method {
		case "GET":
			article.ListArticles(w, r)
			return
		case "POST":
			authRequired(w, r, article.CreateArticle)
			return
		}
	}

	if len(parts) >= 3 {
		if parts[2] == "hot" && method == "GET" {
			article.GetHotArticles(w, r)
			return
		}
	}

	if len(parts) >= 3 {
		switch method {
		case "GET":
			article.GetArticle(w, r)
			return
		case "PUT", "PATCH":
			authRequired(w, r, article.UpdateArticle)
			return
		case "DELETE":
			authRequired(w, r, article.DeleteArticle)
			return
		}
	}

	http.NotFound(w, r)
}

func handleCategories(w http.ResponseWriter, r *http.Request, parts []string, method string) {
	if len(parts) == 2 {
		switch method {
		case "GET":
			article.ListCategories(w, r)
			return
		case "POST":
			adminRequired(w, r, article.CreateCategory)
			return
		}
	}

	http.NotFound(w, r)
}

func handleTags(w http.ResponseWriter, r *http.Request, parts []string, method string) {
	if len(parts) == 2 && method == "GET" {
		article.ListTags(w, r)
		return
	}

	http.NotFound(w, r)
}

func handleComments(w http.ResponseWriter, r *http.Request, parts []string, method string) {
	if len(parts) == 2 {
		switch method {
		case "GET":
			comment.ListComments(w, r)
			return
		case "POST":
			comment.CreateComment(w, r)
			return
		}
	}

	if len(parts) >= 3 {
		if parts[2] == "pending" && method == "GET" {
			adminRequired(w, r, comment.ListPendingComments)
			return
		}
	}

	if len(parts) >= 3 {
		if len(parts) >= 4 && parts[3] == "review" {
			if method == "POST" {
				adminRequired(w, r, comment.ReviewComment)
				return
			}
		}
	}

	if len(parts) >= 3 {
		switch method {
		case "DELETE":
			authRequired(w, r, comment.DeleteComment)
			return
		}
	}

	http.NotFound(w, r)
}

func authRequired(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	middleware.RequireAuth(next).ServeHTTP(w, r)
}

func adminRequired(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	middleware.RequireAdmin(next).ServeHTTP(w, r)
}

func main() {
	dbPath := db.GetEnv("DB_PATH", "./blog.db")

	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	if err := user.InitAdmin(); err != nil {
		log.Printf("Warning: Failed to initialize admin user: %v", err)
	}

	router := middleware.Auth(&Router{})

	port := db.GetEnv("PORT", "8200")
	addr := ":" + port

	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
