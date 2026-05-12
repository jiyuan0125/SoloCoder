package main

import (
	"conn-pool/internal/config"
	"conn-pool/internal/manager"
	"conn-pool/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	
	mgr := manager.NewManager()
	defer mgr.Close()
	
	backends := []string{
		"localhost:5432",
		"localhost:3306",
	}
	
	for _, addr := range backends {
		cfg := config.DefaultPoolConfig(addr)
		mgr.AddPool(cfg)
	}
	
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})
	
	server.SetupRoutes(app, mgr)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	
	addr := ":" + port
	
	go func() {
		log.Printf("[INFO] 服务器启动中，监听端口: %s", port)
		if err := app.Listen(addr); err != nil {
			log.Printf("[ERROR] 服务器启动失败: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("[INFO] 正在关闭服务器...")
	if err := app.Shutdown(); err != nil {
		log.Printf("[ERROR] 服务器关闭异常: %v", err)
	}
	
	log.Println("[INFO] 服务器已关闭")
}
