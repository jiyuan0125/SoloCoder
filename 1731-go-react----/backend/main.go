package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"skill-cert/handlers"
	"skill-cert/storage"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8402"
	}

	store := storage.NewStorage()
	occupationHandler := handlers.NewOccupationHandler(store)
	examHandler := handlers.NewExamHandler(store)
	certHandler := handlers.NewCertificateHandler(store)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	api := r.Group("/api")

	occs := api.Group("/occupations")
	{
		occs.POST("", occupationHandler.CreateOccupation)
		occs.GET("", occupationHandler.GetOccupations)
		occs.GET("/:id", occupationHandler.GetOccupation)
		occs.PUT("/:id", occupationHandler.UpdateOccupation)
		occs.DELETE("/:id", occupationHandler.DeleteOccupation)
	}

	batches := api.Group("/batches")
	{
		batches.POST("", examHandler.CreateBatch)
		batches.GET("", examHandler.GetBatches)
		batches.GET("/:id", examHandler.GetBatch)
		batches.DELETE("/:id", examHandler.DeleteBatch)
		batches.POST("/:id/status", examHandler.UpdateBatchStatus)
		batches.POST("/:id/register", examHandler.RegisterCandidate)
		batches.POST("/:id/scores", examHandler.EnterScores)
	}

	certs := api.Group("/certificates")
	{
		certs.POST("", certHandler.IssueCertificate)
		certs.GET("", certHandler.GetCertificates)
		certs.GET("/:id", certHandler.GetCertificate)
		certs.PUT("/:id/status", certHandler.UpdateCertificateStatus)
		certs.GET("/expiring-soon", certHandler.GetExpiringSoon)
	}

	api.GET("/stats/pass-rate", certHandler.GetPassRateStats)

	r.Run(":" + port)
}
