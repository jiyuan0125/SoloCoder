package main

import (
	"log"
	"net/http"
	"os"

	"gateway/auth"
	"gateway/config"
	"gateway/logger"
	"gateway/proxy"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfgManager, err := config.NewManager(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	go cfgManager.Watch()

	authMiddleware := auth.NewMiddleware(cfgManager)
	logMiddleware := logger.NewLogger(cfgManager)
	router := proxy.NewRouter(cfgManager)

	var handler http.Handler = router.Handler()
	handler = authMiddleware.Handler(handler)
	handler = logMiddleware.Middleware(handler)

	addr := ":" + port
	log.Printf("[gateway] 服务启动, 监听 %s, 配置文件: %s", addr, configPath)
	log.Printf("[gateway] 认证模式: %s", cfgManager.Get().Auth.Mode)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
