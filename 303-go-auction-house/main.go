package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	store := NewStore()
	if err := store.Load(); err != nil {
		log.Printf("警告: 加载持久化数据失败: %v", err)
	}

	auctionService := NewAuctionService(store)
	bidService := NewBidService(store, auctionService)
	rulesService := NewRulesService()

	handler := NewHandler(auctionService, bidService, rulesService, store)

	go func() {
		log.Println("服务器启动于 :8080")
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("正在保存数据并关闭服务器...")
	if err := store.Save(); err != nil {
		log.Printf("保存数据失败: %v", err)
	}
	log.Println("服务器已关闭")
}
