package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"go-cicd-pipeline/internal/api"
	"go-cicd-pipeline/internal/artifact"
	"go-cicd-pipeline/internal/engine"
	"go-cicd-pipeline/internal/logmanager"
	"go-cicd-pipeline/internal/store"
)

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	dataDir := filepath.Join(workDir, "data")
	logDir := filepath.Join(workDir, "logs")
	artifactDir := filepath.Join(workDir, "artifacts")

	s, err := store.New(dataDir)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	lm, err := logmanager.New(logDir)
	if err != nil {
		log.Fatalf("Failed to create log manager: %v", err)
	}

	am, err := artifact.New(artifactDir, s)
	if err != nil {
		log.Fatalf("Failed to create artifact manager: %v", err)
	}

	eng := engine.New(s, lm)

	apiServer := api.New(s, eng, lm, am)

	port := "9101"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("CI/CD Pipeline server starting on port %s...", port)
	log.Printf("Data directory: %s", dataDir)
	log.Printf("Log directory: %s", logDir)
	log.Printf("Artifact directory: %s", artifactDir)

	if err := http.ListenAndServe(":"+port, apiServer.Handler()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
