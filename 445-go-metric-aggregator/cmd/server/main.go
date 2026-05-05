package main

import (
	"context"
	"flag"
	"log"
	"metric-aggregator/internal/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP服务地址")
	flag.Parse()
	
	store := server.NewStore()
	defer store.Stop()
	
	handler := server.NewHandler(store)
	
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	
	srv := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}
	
	go func() {
		log.Printf("指标聚合服务启动，监听地址: %s", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("正在关闭服务...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败: %v", err)
	}
	
	log.Println("服务已关闭")
}
