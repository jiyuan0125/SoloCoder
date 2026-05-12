package main

import (
	"flag"
	"os"
	"tdm-system/pkg/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8406"
	}

	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api")

	api.POST("/teachers", handlers.CreateTeacher)
	api.GET("/teachers", handlers.ListTeachers)
	api.GET("/teachers/:id", handlers.GetTeacher)

	api.POST("/trainings", handlers.CreateTraining)
	api.GET("/trainings", handlers.ListTrainings)
	api.POST("/trainings/:id/register", handlers.RegisterTraining)
	api.GET("/trainings/:id/registrations", handlers.ListRegistrationsByTraining)

	api.PUT("/registrations/:rid", handlers.UpdateRegistration)
	api.GET("/teachers/:teacher_id/registrations", handlers.ListRegistrationsByTeacher)

	api.POST("/evaluations", handlers.CreateEvaluation)
	api.GET("/evaluations", handlers.ListEvaluations)
	api.GET("/teachers/:teacher_id/evaluations", handlers.ListEvaluationsByTeacher)

	api.POST("/applications", handlers.CreateApplication)
	api.GET("/applications", handlers.ListApplications)
	api.GET("/teachers/:teacher_id/applications", handlers.ListApplicationsByTeacher)
	api.POST("/applications/:id/review", handlers.SubmitReview)

	api.GET("/stats", handlers.GetStats)

	r.Run(":" + port)
}
