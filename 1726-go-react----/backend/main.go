package main

import (
	"flag"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"paper-management-platform/handlers"
)

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	port := flag.String("port", "8300", "server port")
	flag.Parse()
	return ":" + *port
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api/r")
	{
		api.GET("/users", handlers.GetUsers)
		api.POST("", handlers.CreatePaper)
		api.GET("", handlers.ListPapers)
		api.GET("/:id", handlers.GetPaper)
		api.PUT("/:id", handlers.UpdatePaper)
		api.POST("/:id/submit", handlers.SubmitForReview)
		api.POST("/:id/assign-reviewers", handlers.AssignReviewers)
		api.POST("/:id/submit-review", handlers.SubmitReview)
		api.POST("/:id/publish", handlers.PublishPaper)
		api.GET("/:id/items/reviews", handlers.GetReviews)
		api.GET("/:id/items/history", handlers.GetVersionHistory)
		api.POST("/:id/actions/approve", handlers.ApproveAction)
		api.POST("/:id/actions/reject", handlers.RejectAction)
		api.POST("/:id/actions/cancel", handlers.CancelAction)
	}

	port := getPort()
	r.Run(port)
}
