package main

import (
	"log"
	"os"
	"rehab-system/config"
	"rehab-system/handlers"
	"rehab-system/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()
	services.SeedDatabase()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	patientHandler := handlers.NewPatientHandler()
	planHandler := handlers.NewPlanHandler()
	trainingHandler := handlers.NewTrainingHandler()
	assessmentHandler := handlers.NewAssessmentHandler()
	summaryHandler := handlers.NewSummaryHandler()

	api := r.Group("/api")
	{
		patients := api.Group("/patients")
		{
			patients.GET("", patientHandler.List)
			patients.GET("/:id", patientHandler.Get)
			patients.POST("", patientHandler.Create)
			patients.PUT("/:id", patientHandler.Update)
			patients.DELETE("/:id", patientHandler.Delete)
		}

		plans := api.Group("/plans")
		{
			plans.GET("", planHandler.List)
			plans.GET("/:id", planHandler.Get)
			plans.POST("", planHandler.Create)
			plans.GET("/:id/tasks", planHandler.GetPlanTasks)
		}

		training := api.Group("/training")
		{
			training.GET("/patients/:id/tasks", trainingHandler.GetPatientTasks)
			training.GET("/tasks/:id", trainingHandler.GetTask)
			training.POST("/records", trainingHandler.CreateRecord)
			training.GET("/exercises/:id/difficulty-history", trainingHandler.GetDifficultyHistory)
		}

		assessments := api.Group("/assessments")
		{
			assessments.GET("", assessmentHandler.List)
			assessments.GET("/:id", assessmentHandler.Get)
			assessments.POST("/:id/record", assessmentHandler.Record)
		}

		summary := api.Group("/summary")
		{
			summary.GET("/patients/:id", summaryHandler.GetPatientSummary)
		}
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
