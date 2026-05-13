package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"member-points/database"
	"member-points/handlers"
)

func main() {
	err := database.InitDB("./member_points.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	api := r.Group("/api")
	{
		members := api.Group("/members")
		{
			members.POST("", handlers.CreateMember)
			members.GET("/:id", handlers.GetMember)
		}

		consumptions := api.Group("/consumptions")
		{
			consumptions.POST("", handlers.RecordConsumption)
		}

		products := api.Group("/products")
		{
			products.POST("", handlers.CreateProduct)
			products.GET("", handlers.ListProducts)
			products.GET("/:id", handlers.GetProduct)
		}

		exchanges := api.Group("/exchanges")
		{
			exchanges.POST("", handlers.ExchangeProduct)
		}

		yearEnd := api.Group("/year-end")
		{
			yearEnd.POST("/:year", handlers.ProcessYearEnd)
		}
	}

	log.Println("Server starting on port 8103...")
	if err := r.Run(":9600"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
