package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"release-flow/dao"
	"release-flow/handler"
	"release-flow/model"
)

func main() {
	err := model.InitDB("./releases.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer model.CloseDB()

	releaseDAO := dao.NewReleaseDAO(model.GetDB())
	releaseHandler := handler.NewReleaseHandler(releaseDAO)

	mux := http.NewServeMux()

	releaseIDPattern := regexp.MustCompile(`^/api/releases/(\d+)$`)
	subResourcePattern := regexp.MustCompile(`^/api/releases/(\d+)/([a-z-]+)$`)
	statusActionPattern := regexp.MustCompile(`^/api/releases/(\d+)/actions/([a-z-]+)$`)

	mux.HandleFunc("/api/releases", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			releaseHandler.ListReleases(w, r)
		case http.MethodPost:
			releaseHandler.CreateRelease(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/releases/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if matches := releaseIDPattern.FindStringSubmatch(path); matches != nil {
			if r.Method == http.MethodGet {
				releaseHandler.GetRelease(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}

		if matches := subResourcePattern.FindStringSubmatch(path); matches != nil {
			releaseHandler.HandleReleaseSubResource(w, r)
			return
		}

		if matches := statusActionPattern.FindStringSubmatch(path); matches != nil {
			if r.Method == http.MethodPost {
				releaseHandler.HandleStatusAction(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}

		http.NotFound(w, r)
	})

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
