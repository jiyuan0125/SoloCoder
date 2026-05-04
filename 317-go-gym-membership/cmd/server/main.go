package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gym-membership/server"
	"gym-membership/types"
)

const (
	defaultPort      = ":8080"
	dataFilePath     = "gym_data.json"
	autoSaveInterval = 30 * time.Second
)

func main() {
	store := server.NewStore()

	initDefaultCardPrices(store)

	persistence := server.NewPersistence(store, dataFilePath)
	if err := persistence.Load(); err != nil {
		fmt.Printf("警告: 加载数据失败: %v\n", err)
	}

	persistence.StartAutoSave(autoSaveInterval)
	defer persistence.StopAutoSave()

	handler := server.NewHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/members", handler.CreateMember)
	mux.HandleFunc("POST /api/login", handler.Login)
	mux.HandleFunc("POST /api/cards/purchase", handler.PurchaseCard)
	mux.HandleFunc("POST /api/cards/renew", handler.RenewCard)
	mux.HandleFunc("GET /api/members/info", handler.GetMemberInfo)

	mux.HandleFunc("POST /api/classes", handler.CreateClass)
	mux.HandleFunc("GET /api/classes", handler.ListClasses)
	mux.HandleFunc("POST /api/bookings", handler.BookClass)
	mux.HandleFunc("POST /api/bookings/cancel", handler.CancelBooking)
	mux.HandleFunc("POST /api/checkin", handler.CheckIn)

	mux.HandleFunc("GET /api/classes/bookings", handler.GetClassBookings)
	mux.HandleFunc("POST /api/prices", handler.SetCardPrice)
	mux.HandleFunc("GET /api/prices", handler.ListCardPrices)
	mux.HandleFunc("POST /api/classes/instances", handler.GenerateClassInstances)

	server := &http.Server{
		Addr:    defaultPort,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("健身房会员管理系统服务端启动于端口 %s\n", defaultPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器错误: %v\n", err)
			os.Exit(1)
		}
	}()

	go startBackgroundTasks(store)

	<-stop
	fmt.Println("\n正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("服务器强制关闭: %v\n", err)
	}

	if err := persistence.Save(); err != nil {
		fmt.Printf("保存数据失败: %v\n", err)
	}

	fmt.Println("服务器已关闭")
}

func initDefaultCardPrices(store *server.Store) {
	prices := store.GetCardPrices()
	if len(prices) == 0 {
		store.SetCardPrice(types.MonthlyCard, 299.0, 30)
		store.SetCardPrice(types.QuarterlyCard, 799.0, 90)
		store.SetCardPrice(types.YearlyCard, 2999.0, 365)
	}
}

func startBackgroundTasks(store *server.Store) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		store.CheckAndCancelExpiredBookings()
		store.CheckAndCancelLowCapacityClasses()
	}
}
