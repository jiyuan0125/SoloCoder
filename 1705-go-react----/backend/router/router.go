package router

import (
	"telemedicine/controllers"
	"telemedicine/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")

	api.POST("/login", controllers.Login)

	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/users/me", controllers.GetCurrentUser)
		auth.GET("/users", controllers.ListUsers)

		consultations := auth.Group("/consultations")
		{
			consultations.GET("", controllers.ListConsultations)
			consultations.POST("", controllers.CreateConsultation)
			consultations.GET("/pool", controllers.GetConsultationPool)
			consultations.GET("/:id", controllers.GetConsultation)
			consultations.POST("/:id/accept", controllers.AcceptConsultation)
			consultations.POST("/:id/messages", controllers.SendMessage)
			consultations.GET("/:id/messages", controllers.ListMessages)
			consultations.POST("/:id/request-supplement", controllers.RequestSupplement)
			consultations.POST("/:id/supplement", controllers.SubmitSupplement)
			consultations.POST("/:id/complete", controllers.CompleteConsultation)
			consultations.POST("/:id/exams", controllers.AddExam)
			consultations.POST("/:id/images", controllers.AddImageData)
			consultations.DELETE("/:id", controllers.DeleteConsultation)
		}

		prescriptions := auth.Group("/prescriptions")
		{
			prescriptions.POST("", controllers.CreatePrescription)
			prescriptions.GET("/:id", controllers.GetPrescription)
			prescriptions.GET("/consultation/:consultationID", controllers.GetPrescriptionsByConsultation)
		}

		auth.GET("/forbidden-drugs", controllers.ListForbiddenDrugs)
		auth.GET("/image-templates", controllers.ListImageTemplates)

		admin := auth.Group("/admin")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.GET("/statistics", controllers.GetStatistics)
			admin.GET("/export/date-range", controllers.ExportByDateRange)
			admin.GET("/export/by-expert", controllers.ExportByExpert)
		}
	}

	return r
}
