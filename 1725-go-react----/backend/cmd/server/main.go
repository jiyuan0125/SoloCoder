package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"confman/pkg/config"
	"confman/pkg/handler"
	"confman/pkg/repository"
	"confman/pkg/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	portFlag := flag.Int("port", 0, "Server port")
	dbPathFlag := flag.String("db", "", "Database path")
	flag.Parse()

	cfg := config.Load()

	if *portFlag > 0 {
		cfg.Port = *portFlag
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			cfg.Port = p
		}
	}
	if *dbPathFlag != "" {
		cfg.DBPath = *dbPathFlag
	}

	db, err := config.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repo := repository.New(db)

	meetingService := service.NewMeetingService(repo, db)
	paperService := service.NewPaperService(repo, db, meetingService)
	reviewService := service.NewReviewService(repo, db)
	scheduleService := service.NewScheduleService(repo, db)
	userService := service.NewUserService(repo, db)
	notificationService := service.NewNotificationService(repo, db)

	h := handler.New(
		meetingService,
		paperService,
		reviewService,
		scheduleService,
		userService,
		notificationService,
	)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	h.SetupRoutes(r)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			notificationService.CheckAndSendReminders()
			<-ticker.C
		}
	}()

	log.Printf("Conference Management System starting on port %d", cfg.Port)
	log.Printf("Database: %s", cfg.DBPath)

	if err := r.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
