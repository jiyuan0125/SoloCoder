package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"epidemic-management/config"
	"epidemic-management/database"
	"epidemic-management/handlers"
	"epidemic-management/services"
)

func main() {
	cfg := config.Load()
	database.Init()
	services.StartScheduler()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.POST("/outbreak/reports", handlers.CreateOutbreakReport)
		api.GET("/outbreak/reports", handlers.ListOutbreakReports)
		api.GET("/outbreak/alerts", handlers.GetActiveAlerts)

		api.POST("/investigations", handlers.CreateInvestigation)
		api.GET("/investigations", handlers.ListInvestigations)

		api.POST("/contacts", handlers.CreateContact)
		api.GET("/contacts", handlers.ListContacts)
		api.PUT("/contacts/:id/status", handlers.UpdateContactStatus)

		api.POST("/vaccines", handlers.CreateVaccine)
		api.GET("/vaccines", handlers.ListVaccines)
		api.POST("/vaccinations", handlers.RecordVaccination)
		api.GET("/vaccinations", handlers.ListVaccinationRecords)

		api.GET("/todos", handlers.ListTodos)
		api.PUT("/todos/:id/status", handlers.UpdateTodoStatus)

		api.GET("/statistics", handlers.GetStatistics)

		api.GET("/weekly-reports", handlers.ListWeeklyReports)
		api.POST("/weekly-reports/generate", handlers.GenerateWeeklyReport)

		api.GET("/diseases", handlers.ListDiseases)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
