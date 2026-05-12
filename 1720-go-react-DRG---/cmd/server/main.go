package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"drg-system/internal/api"
	"drg-system/internal/repository"
	"drg-system/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	var port int
	flag.IntVar(&port, "port", 0, "服务端口")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("DRG_SERVER_PORT")
		if envPort != "" {
			var err error
			port, err = strconv.Atoi(envPort)
			if err != nil {
				port = 8300
			}
		} else {
			port = 8300
		}
	}

	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()

	groupingService := service.NewGroupingService(repo)
	paymentService := service.NewPaymentService(repo)
	settlementService := service.NewSettlementService(repo, groupingService, paymentService)
	todoService := service.NewTodoService(repo)

	router := api.NewRouter(groupingService, paymentService, settlementService, todoService, repo)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("DRG付费管理系统服务启动中，端口: %d", port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
