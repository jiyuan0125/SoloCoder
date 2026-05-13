package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"course-enroll/api"
	"course-enroll/modules/course"
	"course-enroll/modules/enrollment"
	"course-enroll/modules/queue"
	"course-enroll/modules/resource"
	"course-enroll/modules/schedule"
	"course-enroll/storage"
)

func main() {
	dbPath := "./course_enroll.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	courseMgr := course.NewManager(store)
	queueSys := queue.NewSystem(store)
	scheduleChk := schedule.NewChecker(store)
	resourceMgr := resource.NewManager(store)

	engine := enrollment.NewEngine(store, courseMgr, queueSys, scheduleChk)
	engine.SetNotificationFunc(func(studentID int64, message string) {
		log.Printf("Notification to student %d: %s", studentID, message)
	})

	handler := api.NewHandler(courseMgr, engine, queueSys, resourceMgr, store)

	mux := http.NewServeMux()
	handler.Register(mux)

	server := &http.Server{
		Addr:         ":8400",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
