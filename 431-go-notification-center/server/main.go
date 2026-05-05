package main

import (
	"flag"
	"fmt"
	"net/http"
	"notification-center/common"
)

func main() {
	port := flag.Int("port", common.DefaultPort, "server port")
	flag.Parse()
	
	store := NewStore()
	service := NewService(store)
	handler := NewHandler(service)
	
	go service.RunBackgroundTasks()
	
	http.HandleFunc("/api/notification/send", handler.HandleSend)
	http.HandleFunc("/api/notification/list", handler.HandleList)
	http.HandleFunc("/api/notification/mark-read", handler.HandleMarkRead)
	http.HandleFunc("/api/notification/unread-count", handler.HandleUnreadCount)
	
	http.HandleFunc("/api/template/create", handler.HandleCreateTemplate)
	http.HandleFunc("/api/template/update", handler.HandleUpdateTemplate)
	http.HandleFunc("/api/template/list", handler.HandleListTemplates)
	http.HandleFunc("/api/template/delete", handler.HandleDeleteTemplate)
	
	http.HandleFunc("/api/user/activity", handler.HandleRecordActivity)
	
	http.HandleFunc("/api/log/failed", handler.HandleListFailedLogs)
	
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Notification Center Server starting on %s...\n", addr)
	fmt.Printf("API Endpoints:\n")
	fmt.Printf("  POST   /api/notification/send      - Send notification\n")
	fmt.Printf("  GET    /api/notification/list      - List notifications\n")
	fmt.Printf("  POST   /api/notification/mark-read - Mark as read (batch)\n")
	fmt.Printf("  GET    /api/notification/unread-count - Get unread count\n")
	fmt.Printf("  POST   /api/template/create        - Create template\n")
	fmt.Printf("  PUT    /api/template/update        - Update template\n")
	fmt.Printf("  GET    /api/template/list          - List templates\n")
	fmt.Printf("  DELETE /api/template/delete        - Delete template\n")
	fmt.Printf("  POST   /api/user/activity          - Record user activity\n")
	fmt.Printf("  GET    /api/log/failed             - List failed logs\n")
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
