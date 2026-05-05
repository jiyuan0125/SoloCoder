package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/server/handler"
	"forum/server/store"
)

func main() {
	s := store.NewStore()
	h := handler.NewHandler(s)

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateUser(w, r)
		} else if r.Method == http.MethodGet {
			if r.URL.Query().Get("id") != "" || r.URL.Query().Get("username") != "" {
				h.GetUser(w, r)
			} else {
				h.ListUsers(w, r)
			}
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreatePost(w, r)
		} else if r.Method == http.MethodGet {
			if r.URL.Query().Get("id") != "" {
				h.GetPost(w, r)
			} else {
				h.ListPosts(w, r)
			}
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.UpdatePost(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.DeletePost(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.SearchPosts(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/replies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateReply(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/replies/best", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.SetBestReply(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/likes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Like(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/unlikes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.Unlike(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/admin/top", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.SetTop(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/admin/essence", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.SetEssence(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.CreateReport(w, r)
		} else if r.Method == http.MethodGet {
			h.GetPendingReports(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/reports/review", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.ReviewReport(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetLeaderboard(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetTagCloud(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/hotposts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetHotPosts(w, r)
		} else if r.Method == http.MethodPost {
			h.CalculateHotPosts(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8080"
	fmt.Printf("Forum server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
