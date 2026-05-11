package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"eventbus-demo/eventbus"
)

var bus *eventbus.EventBus

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	cleanup := flag.Duration("cleanup", 30*time.Second, "inactive subscriber cleanup interval")
	timeout := flag.Duration("timeout", 5*time.Minute, "subscriber inactive timeout")
	flag.Parse()

	bus = eventbus.New(&eventbus.Options{
		CleanupInterval: *cleanup,
		InactiveTimeout: *timeout,
	})
	defer bus.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/topic/create", handleCreateTopic)
	mux.HandleFunc("/topic/delete", handleDeleteTopic)
	mux.HandleFunc("/topic/subscribe", handleSubscribe)
	mux.HandleFunc("/topic/unsubscribe", handleUnsubscribe)
	mux.HandleFunc("/topic/subscribers", handleListSubscribers)
	mux.HandleFunc("/event/publish", handlePublish)
	mux.HandleFunc("/event/status", handleEventStatus)

	log.Printf("eventbus server starting on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
