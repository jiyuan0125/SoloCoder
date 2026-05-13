package main

import (
	"course-schedule/database"
	"course-schedule/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.InitDB("./courses.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	r.POST("/api/courses", handlers.CreateCourse)
	r.GET("/api/courses", handlers.ListCourses)
	r.GET("/api/courses/:id", handlers.GetCourse)
	r.PUT("/api/courses/:id", handlers.UpdateCourse)

	r.POST("/api/schedules", handlers.CreateSchedule)
	r.GET("/api/schedules", handlers.ListSchedules)

	r.POST("/api/enrollments", handlers.EnrollCourse)
	r.DELETE("/api/enrollments", handlers.DropCourse)
	r.GET("/api/enrollments", handlers.ListEnrollments)

	r.GET("/api/notifications", handlers.ListNotifications)

	log.Println("Server starting on port 8080...")
	if err := r.Run(":8200"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
