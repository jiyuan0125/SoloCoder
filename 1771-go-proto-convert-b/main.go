package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dataconverter/engine"
	"dataconverter/router"
)

func main() {
	ruleEngine := engine.NewRuleEngine()
	loader := engine.NewRuleLoader(ruleEngine)
	
	rulesFile := os.Getenv("RULES_FILE")
	rulesDir := os.Getenv("RULES_DIR")
	
	if rulesFile != "" {
		loader.SetRulesFile(rulesFile)
		if err := loader.LoadFromFile(rulesFile); err != nil {
			log.Printf("Warning: failed to load rules from file %s: %v", rulesFile, err)
		} else {
			log.Printf("Loaded rules from file: %s", rulesFile)
		}
	} else if rulesDir != "" {
		loader.SetRulesDir(rulesDir)
		if err := loader.LoadFromDir(rulesDir); err != nil {
			log.Printf("Warning: failed to load rules from directory %s: %v", rulesDir, err)
		} else {
			log.Printf("Loaded rules from directory: %s", rulesDir)
		}
	}
	
	if rulesFile != "" || rulesDir != "" {
		if err := loader.StartWatching(); err != nil {
			log.Printf("Warning: failed to start rules watcher: %v", err)
		} else {
			log.Println("Started rules file watcher")
		}
	}
	
	r := router.SetupRouter(loader)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	go func() {
		addr := fmt.Sprintf(":%s", port)
		log.Printf("Data converter service starting on port %s", port)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-stopChan
	log.Println("Shutting down...")
	loader.StopWatching()
	log.Println("Service stopped")
}
