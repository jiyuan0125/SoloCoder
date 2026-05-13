package main

import (
	"go-mq-admin/internal/api"
	"go-mq-admin/internal/deadletter"
	"go-mq-admin/internal/dispatcher"
	"go-mq-admin/internal/storage"
	"go-mq-admin/internal/subscription"
	"go-mq-admin/internal/topic"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./mqadmin.db"
	}

	store, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	topicMgr := topic.NewManager(store)
	subMgr := subscription.NewManager(store)
	disp := dispatcher.NewDispatcher(store)
	dlMgr := deadletter.NewManager(store)

	handler := api.NewHandler(topicMgr, subMgr, disp, dlMgr, store)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handler.ServeUI)

	mux.HandleFunc("GET /api/dashboard", handler.Dashboard)

	mux.HandleFunc("GET /api/topics", handler.ListTopics)
	mux.HandleFunc("POST /api/topics", handler.CreateTopic)
	mux.HandleFunc("GET /api/topics/{topic}", handler.GetTopic)
	mux.HandleFunc("GET /api/topics/{topic}/stats", handler.GetTopicStats)
	mux.HandleFunc("DELETE /api/topics/{topic}", handler.DeleteTopic)

	mux.HandleFunc("GET /api/topics/{topic}/subscriptions", handler.ListSubscriptions)
	mux.HandleFunc("POST /api/topics/{topic}/subscriptions", handler.CreateSubscription)
	mux.HandleFunc("GET /api/topics/{topic}/subscriptions/{subscription}", handler.GetSubscription)
	mux.HandleFunc("GET /api/topics/{topic}/subscriptions/{subscription}/stats", handler.GetSubscriptionStats)
	mux.HandleFunc("DELETE /api/topics/{topic}/subscriptions/{subscription}", handler.DeleteSubscription)

	mux.HandleFunc("POST /api/topics/{topic}/publish", handler.PublishMessage)
	mux.HandleFunc("GET /api/topics/{topic}/subscriptions/{subscription}/pull", handler.PullMessages)
	mux.HandleFunc("POST /api/topics/{topic}/subscriptions/{subscription}/messages/{message_id}/ack", handler.AckMessage)
	mux.HandleFunc("POST /api/topics/{topic}/subscriptions/{subscription}/messages/{message_id}/nack", handler.NackMessage)

	mux.HandleFunc("POST /api/consumers/{consumer_id}/heartbeat", handler.Heartbeat)

	mux.HandleFunc("GET /api/topics/{topic}/dead-letters", handler.ListDeadLetters)
	mux.HandleFunc("GET /api/topics/{topic}/dead-letters/{dead_letter_id}", handler.GetDeadLetter)
	mux.HandleFunc("POST /api/topics/{topic}/dead-letters/{dead_letter_id}/requeue", handler.RequeueDeadLetter)
	mux.HandleFunc("DELETE /api/topics/{topic}/dead-letters/{dead_letter_id}", handler.DropDeadLetter)

	go startBackgroundTasks(dlMgr)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9203"
	}
	log.Println("Message Queue Admin Server starting on :" + port + "...")
	log.Println("Dashboard available at: http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func startBackgroundTasks(dlMgr *deadletter.Manager) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := dlMgr.ProcessExpiredAndOverflow(); err != nil {
			log.Printf("[Background] Error processing expired messages: %v", err)
		}

		if err := dlMgr.RequeueTimedOutMessages(); err != nil {
			log.Printf("[Background] Error requeuing timed out messages: %v", err)
		}

		if err := dlMgr.CleanupInactiveConsumers(); err != nil {
			log.Printf("[Background] Error cleaning up consumers: %v", err)
		}
	}
}
