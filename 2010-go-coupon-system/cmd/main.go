package main

import (
	"log"
	"time"

	"coupon-system/internal/database"
	"coupon-system/internal/handler"
	"coupon-system/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.Init("coupons.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	go startExpireScheduler()

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/coupons", handler.CreateCoupon)
		api.POST("/coupons/claim", handler.ClaimCoupon)
		api.POST("/coupons/redeem", handler.RedeemCoupon)
		api.POST("/coupons/refund", handler.PartialRefund)
		api.GET("/coupons/stats", handler.GetStats)
		api.GET("/coupons/redemption-records", handler.GetRedemptionRecords)
	}

	log.Println("Server starting on port 8080...")
	if err := r.Run(":9604"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startExpireScheduler() {
	log.Println("Expire scheduler started")
	
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	durationUntilMidnight := nextMidnight.Sub(now)
	
	time.Sleep(durationUntilMidnight)
	
	for {
		log.Println("Running daily expiration check...")
		if err := service.ExpireCoupons(); err != nil {
			log.Printf("Error during expiration check: %v", err)
		} else {
			log.Println("Expiration check completed successfully")
		}
		time.Sleep(24 * time.Hour)
	}
}
