package main

import (
	"github.com/gin-gonic/gin"
	"github.com/performance-review/database"
	"github.com/performance-review/handlers"
)

func main() {
	database.InitDB()
	defer database.DB.Close()

	r := gin.Default()

	api := r.Group("/api")
	{
		employees := api.Group("/employees")
		{
			employees.POST("", handlers.CreateEmployee)
			employees.GET("", handlers.ListEmployees)
			employees.GET("/:id", handlers.GetEmployee)
			employees.PUT("/:id", handlers.UpdateEmployee)
			employees.DELETE("/:id", handlers.DeleteEmployee)
		}

		quarters := api.Group("/quarters")
		{
			quarters.POST("", handlers.CreateQuarter)
			quarters.GET("", handlers.ListQuarters)
			quarters.GET("/:id", handlers.GetQuarter)
		}

		reviews := api.Group("/reviews")
		{
			reviews.POST("", handlers.CreateOrUpdateReview)
			reviews.POST("/confirm", handlers.ConfirmReviews)
			reviews.GET("", handlers.ListReviews)
			reviews.GET("/:id", handlers.GetReview)
		}

		annual := api.Group("/annual")
		{
			annual.POST("/generate", handlers.GenerateAnnualReviews)
			annual.GET("", handlers.ListAnnualReviews)
			annual.GET("/:employee_id/:year", handlers.GetEmployeeAnnualReview)
		}
	}

	r.Run(":9602")
}
