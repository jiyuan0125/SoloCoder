package main

import (
	"log"
	"net/http"
	"time"
)

const ServerPort = ":8080"

func main() {
	initStore()

	startScheduleTicker()

	mux := setupRouter()

	log.Printf("公告管理服务启动，监听端口 %s", ServerPort)
	log.Printf("可用用户ID: user_admin(管理员), user_manager_1(部门经理), user_emp_1(员工), user_emp_2(员工)")

	if err := http.ListenAndServe(ServerPort, mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func startScheduleTicker() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			ProcessScheduledTasks()
		}
	}()
}
