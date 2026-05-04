package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	StorageFile = "feedbacks.json"
	Port        = 8080
)

func main() {
	storage := NewStorage(StorageFile)

	submissionHandler := NewSubmissionHandler(storage)
	queryHandler := NewQueryHandler(storage)
	adminHandler := NewAdminHandler(storage)

	http.HandleFunc("/api/feedbacks/submit", submissionHandler.HandleSubmit)
	http.HandleFunc("/api/feedbacks", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/feedbacks/") && r.URL.Path != "/api/feedbacks" {
			queryHandler.HandleGetByID(w, r)
		} else {
			queryHandler.HandleList(w, r)
		}
	})
	http.HandleFunc("/api/statistics", queryHandler.HandleStatistics)
	http.HandleFunc("/api/admin/login", adminHandler.HandleLogin)
	http.HandleFunc("/api/admin/feedbacks/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/status") {
			adminHandler.HandleUpdateStatus(w, r)
		} else if strings.Contains(r.URL.Path, "/note") {
			adminHandler.HandleUpdateNote(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	fmt.Printf("Server starting on port %d...\n", Port)
	fmt.Printf("Storage file: %s\n", getStoragePath())
	fmt.Printf("Admin credentials: username=%s, password=%s\n", AdminUsername, AdminPassword)
	fmt.Println("API Endpoints:")
	fmt.Println("  POST   /api/feedbacks/submit      - 提交反馈")
	fmt.Println("  GET    /api/feedbacks             - 获取反馈列表（分页、筛选）")
	fmt.Println("  GET    /api/feedbacks/{id}        - 获取单条反馈详情")
	fmt.Println("  GET    /api/statistics             - 获取统计概览")
	fmt.Println("  POST   /api/admin/login             - 管理员登录")
	fmt.Println("  PUT    /api/admin/feedbacks/{id}/status - 更新反馈状态")
	fmt.Println("  PUT    /api/admin/feedbacks/{id}/note   - 更新内部备注")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func getStoragePath() string {
	path, err := os.Getwd()
	if err != nil {
		return StorageFile
	}
	return path + string(os.PathSeparator) + StorageFile
}
