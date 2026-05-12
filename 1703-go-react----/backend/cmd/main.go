package main

import (
	"log"
	"os"
	"strconv"

	"medical-exam-system/internal/handler"
	"medical-exam-system/internal/repository"
	"medical-exam-system/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			return ":" + port
		}
	}
	if len(os.Args) > 1 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			return ":" + os.Args[1]
		}
	}
	return ":8080"
}

func main() {
	repo := repository.New()
	svc := service.New(repo)
	h := handler.New(svc)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")

	items := api.Group("/items")
	{
		items.GET("", h.GetItems)
	}

	packages := api.Group("/packages")
	{
		packages.GET("", h.GetPackages)
		packages.GET("/:id", h.GetPackageByID)
		packages.POST("", h.CreatePackage)
		packages.PUT("/:id", h.UpdatePackage)
		packages.DELETE("/:id", h.DeletePackage)
	}

	api.POST("/calculate-price", h.CalculatePrice)
	api.GET("/availability", h.GetAvailability)

	appointments := api.Group("/appointments")
	{
		appointments.POST("", h.CreateAppointment)
		appointments.GET("", h.GetAppointments)
		appointments.GET("/:id", h.GetAppointmentByID)
		appointments.DELETE("/:id", h.CancelAppointment)
		appointments.GET("/:id/results", h.GetExamResultsByAppointment)
	}

	results := api.Group("/results")
	{
		results.PUT("/:id", h.SubmitExamResult)
	}

	api.GET("/alerts", h.GetCriticalAlerts)

	reports := api.Group("/reports")
	{
		reports.GET("", h.GetReports)
		reports.GET("/:id", h.GetReportByID)
		reports.PUT("/:id", h.UpdateReport)
		reports.POST("/:id/submit-review", h.SubmitReportForReview)
		reports.POST("/:id/publish", h.PublishReport)
	}

	api.GET("/export-records", h.ExportRecords)

	generic := api.Group("/r")
	{
		generic.GET("", h.GetGenericList)
		generic.GET("/:id", h.GetGenericDetail)
		generic.GET("/:id/sub", h.GetGenericSubResources)
	}

	port := getPort()
	log.Printf("Server starting on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
