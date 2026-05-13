package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"cron-scheduler/handlers"
	"cron-scheduler/scheduler"

	"github.com/gofiber/fiber/v2"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	sched := scheduler.New()
	handler := handlers.NewTaskHandler(sched)

	app := fiber.New(fiber.Config{
		AppName: "Cron Scheduler",
	})

	app.Post("/tasks", handler.CreateTask)
	app.Get("/tasks", handler.ListTasks)
	app.Get("/tasks/:id", handler.GetTask)
	app.Put("/tasks/:id/pause", handler.PauseTask)
	app.Put("/tasks/:id/resume", handler.ResumeTask)
	app.Delete("/tasks/:id", handler.DeleteTask)
	app.Get("/tasks/:id/logs", handler.GetTaskLogs)

	sched.Start()
	log.Printf("Cron scheduler started, listening on port %s", port)

	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	sched.Stop()
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("gracefully stopped")
}
