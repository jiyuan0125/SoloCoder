package main

import (
	"flag"
	"log"
	"os"
	"school-connect/internal/handlers"
	"school-connect/internal/middleware"
	"school-connect/internal/models"
	"school-connect/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flagPort := flag.String("port", "8080", "服务端口")
		flag.Parse()
		port = *flagPort
	}
	return ":" + port
}

func main() {
	store := storage.NewMemoryStore()
	store.SeedDemoData()

	authHandler := handlers.NewAuthHandler(store)
	announcementHandler := handlers.NewAnnouncementHandler(store)
	scoreHandler := handlers.NewScoreHandler(store)
	messageHandler := handlers.NewMessageHandler(store)

	authMiddleware := middleware.NewAuthMiddleware(store)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")

	api.POST("/auth/login", authHandler.Login)

	auth := api.Group("")
	auth.Use(authMiddleware.AuthRequired())
	{
		auth.GET("/auth/me", authHandler.GetCurrentUserInfo)

		announcements := auth.Group("/announcements")
		{
			announcements.POST("", authMiddleware.RoleRequired(models.RoleAdmin, models.RoleTeacher), announcementHandler.Create)
			announcements.GET("", announcementHandler.ListForParent)
			announcements.GET("/all", authMiddleware.RoleRequired(models.RoleAdmin, models.RoleTeacher), announcementHandler.ListAll)
			announcements.GET("/classes-grades", announcementHandler.GetClassesAndGrades)
			announcements.GET("/unread-count", announcementHandler.GetUnreadCount)
			announcements.GET("/:id", announcementHandler.GetByID)
			announcements.POST("/:id/read", announcementHandler.MarkRead)
			announcements.GET("/:id/stats", announcementHandler.GetStats)
		}

		scores := auth.Group("/scores")
		{
			scores.POST("/batch", authMiddleware.RoleRequired(models.RoleAdmin, models.RoleTeacher), scoreHandler.BatchRecord)
			scores.GET("/student/:studentID", scoreHandler.GetStudentScores)
			scores.GET("/student/:studentID/subject/:subject", scoreHandler.GetStudentSubjectHistory)
			scores.GET("/class-stats", scoreHandler.GetClassStats)
			scores.GET("/class/:class/students", scoreHandler.GetClassStudents)
		}

		messages := auth.Group("/messages")
		{
			messages.POST("/send", messageHandler.Send)
			messages.GET("/conversations", messageHandler.GetConversations)
			messages.GET("/conversation/:conversationID", messageHandler.GetMessages)
			messages.POST("/conversation/:conversationID/read", messageHandler.MarkRead)
			messages.GET("/unread-count", messageHandler.GetTotalUnread)
			messages.GET("/contacts", messageHandler.GetContactableUsers)
		}
	}

	port := getPort()
	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
