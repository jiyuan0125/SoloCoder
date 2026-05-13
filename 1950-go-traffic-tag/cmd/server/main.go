package main

import (
	"log"
	"net/http"
	"os"

	"traffictag/internal/api"
	"traffictag/internal/matcher"
	"traffictag/internal/middleware"
	"traffictag/internal/tagstore"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ruleStore := matcher.NewRuleStore()
	tagStore := tagstore.NewTagStore()

	mux := http.NewServeMux()
	handler := api.NewHandler(ruleStore, tagStore)
	handler.Register(mux)

	var serverHandler http.Handler = mux
	serverHandler = middleware.Tagging(ruleStore, tagStore)(serverHandler)

	log.Printf("Traffic Tag Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, serverHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
