package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"search-service/controllers"
	"search-service/database"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	r := gin.Default()

	api := r.Group("/api")
	{
		documents := api.Group("/documents")
		{
			documents.POST("", controllers.IndexDocument)
			documents.GET("", controllers.ListDocuments)
		}

		search := api.Group("/search")
		{
			search.GET("", controllers.Search)
			search.GET("/hot", controllers.GetHotKeywords)
		}
	}

	log.Println("Server starting on port 8602...")
	if err := r.Run(":8602"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
