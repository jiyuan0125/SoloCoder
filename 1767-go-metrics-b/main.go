package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"metrics/metrics"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: "Metrics Service",
	})

	app.Use(metrics.Middleware())

	metrics.RegisterRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "metrics",
			"endpoints": fiber.Map{
				"list_metrics":    "/api/metrics",
				"get_metric":      "/api/metrics/:name",
			},
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
