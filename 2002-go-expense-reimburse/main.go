package main

import (
	"log"

	"expense-reimburse/database"
	"expense-reimburse/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/reimbursements", handlers.SubmitReimbursement)
		api.GET("/reimbursements", handlers.ListReimbursements)
		api.GET("/reimbursements/:id", handlers.GetReimbursement)
		api.POST("/reimbursements/:id/approve", handlers.ApproveReimbursement)
		api.PUT("/reimbursements", handlers.UpdateReimbursement)
		api.GET("/reimbursements/:id/history", handlers.GetStatusHistory)
	}

	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
