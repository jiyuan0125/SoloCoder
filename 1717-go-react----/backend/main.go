package main

import (
	"log"
	"os"
	"strconv"

	"hospital-infection/database"
	"hospital-infection/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDatabase()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")

	departments := api.Group("/departments")
	{
		departments.GET("", handlers.GetDepartments)
		departments.GET("/:id", handlers.GetDepartment)
		departments.POST("", handlers.CreateDepartment)
		departments.PUT("/:id", handlers.UpdateDepartment)
		departments.DELETE("/:id", handlers.DeleteDepartment)
	}

	infection := api.Group("/infections")
	{
		infection.POST("", handlers.CreateInfectionCase)
		infection.GET("", handlers.GetInfectionCases)
		infection.GET("/:id", handlers.GetInfectionCase)
		infection.PUT("/:id", handlers.UpdateInfectionCase)
		infection.DELETE("/:id", handlers.DeleteInfectionCase)
	}

	measures := api.Group("/measures")
	{
		measures.POST("", handlers.CreatePreventionMeasure)
		measures.GET("", handlers.GetPreventionMeasures)
		measures.GET("/recommended", handlers.GetRecommendedMeasures)
		measures.GET("/:id", handlers.GetPreventionMeasure)
		measures.DELETE("/:id", handlers.DeletePreventionMeasure)
	}

	monitoring := api.Group("/monitoring")
	{
		monitoring.POST("", handlers.CreateTargetMonitoring)
		monitoring.GET("", handlers.GetTargetMonitorings)
		monitoring.GET("/departments", handlers.GetTargetDepartments)
		monitoring.GET("/:id", handlers.GetTargetMonitoring)
		monitoring.DELETE("/:id", handlers.DeleteTargetMonitoring)
	}

	stats := api.Group("/statistics")
	{
		stats.GET("/rates", handlers.CalculateMonthlyRates)
		stats.GET("/trend", handlers.Get12MonthTrend)
		stats.GET("/alerts", handlers.GetAlerts)
	}

	reports := api.Group("/reports")
	{
		reports.POST("/generate", handlers.GenerateReport)
		reports.GET("", handlers.GetReports)
		reports.GET("/:id", handlers.GetReport)
		reports.POST("/:id/approve/:action", handlers.ApproveReport)
		reports.GET("/:id/history", handlers.GetReportApprovalHistory)
		reports.DELETE("/:id", handlers.DeleteReport)
	}

	port := getPort()
	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}

	for i, arg := range os.Args {
		if arg == "--port" && i+1 < len(os.Args) {
			port := os.Args[i+1]
			if _, err := strconv.Atoi(port); err == nil {
				return port
			}
		}
	}

	return "8300"
}
