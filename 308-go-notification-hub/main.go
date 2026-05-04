package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 初始化存储
	storage, err := NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// 初始化各模块
	dispatcher := NewNotificationDispatcher(storage)
	statusManager := NewDeliveryStatusManager(storage)
	userHandler := NewUserNotificationHandler(storage)

	// 注册路由
	http.HandleFunc("/notifications", dispatcher.CreateNotificationHandler)
	http.HandleFunc("/notifications/", dispatcher.GetNotificationHandler)
	http.HandleFunc("/notifications/status", statusManager.UpdateStatusHandler)
	http.HandleFunc("/notifications/stats", statusManager.GetStatsHandler)
	http.HandleFunc("/user/notifications", userHandler.GetUserNotificationsHandler)
	http.HandleFunc("/user/notifications/read", userHandler.MarkAsReadHandler)

	// 启动服务器
	fmt.Println("Notification Hub server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
