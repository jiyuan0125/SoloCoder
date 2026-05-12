package main

import (
	"flag"
	"net/http"
	"os"

	"learning-platform/internal/handlers"
	"learning-platform/internal/services"
	"learning-platform/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	port := flag.String("port", "8300", "server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = &envPort
	}

	store := storage.NewStorage()

	courseService := services.NewCourseService(store)
	achievementService := services.NewAchievementService(store)
	learningService := services.NewLearningService(store, courseService, achievementService)
	quizService := services.NewQuizService(store, achievementService)
	orderService := services.NewOrderService(store)

	handler := handlers.NewHandler(courseService, learningService, achievementService, quizService, orderService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		courses := api.Group("/courses")
		{
			courses.POST("", handler.CreateCourse)
			courses.GET("", handler.GetAllCourses)
			courses.GET("/graph", handler.GetCourseGraph)
			courses.GET("/:id", handler.GetCourse)
			courses.POST("/prerequisites", handler.AddPrerequisite)
		}

		students := api.Group("/students")
		{
			students.POST("", handler.CreateStudent)
			students.POST("/start-learning", handler.StartLearning)
			students.POST("/complete-unit", handler.CompleteUnit)
			students.GET("/progress", handler.GetProgress)
			students.POST("/learning-path", handler.GenerateLearningPath)
			students.GET("/analytics", handler.GetAnalytics)
		}

		quiz := api.Group("/quiz")
		{
			quiz.GET("", handler.GetQuiz)
			quiz.POST("/submit", handler.SubmitQuiz)
		}

		products := api.Group("/r")
		{
			products.POST("", handler.CreateProduct)
			products.GET("", handler.GetProducts)
			products.GET("/:id", handler.GetProduct)
			products.GET("/:id/sub", handler.GetProductSubResources)
			products.PUT("/:id/price", handler.UpdateProductPrice)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", handler.CreateOrder)
			orders.GET("", handler.GetOrders)
			orders.GET("/:id", handler.GetOrder)
			orders.POST("/:id/confirm", handler.ConfirmOrder)
		}

		api.GET("/purchase-requests", handler.GetPurchaseRequests)
	}

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: r,
	}

	server.ListenAndServe()
}
