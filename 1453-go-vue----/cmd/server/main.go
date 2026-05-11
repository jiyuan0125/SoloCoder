package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"property-management/pkg/core"
	"property-management/pkg/core/announcement"
	"property-management/pkg/core/payment"
	"property-management/pkg/core/repair"
	"property-management/pkg/core/user"
	"property-management/internal/server"
	"time"
)

const defaultPort = "8080"

func getPort() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	port := flag.String("port", defaultPort, "Server port")
	flag.Parse()

	return *port
}

func main() {
	port := getPort()

	store := core.NewStore()

	userSvc := user.NewService(store)
	repairSvc := repair.NewService(store)
	paymentSvc := payment.NewService(store)
	announcementSvc := announcement.NewService(store)

	handler := server.NewHandler(userSvc, repairSvc, paymentSvc, announcementSvc)

	go startBackgroundTasks(repairSvc)

	addr := ":" + port
	fmt.Printf("Starting property management server on %s...\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func startBackgroundTasks(repairSvc *repair.Service) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		repairSvc.CheckAndAutoConfirm()
		repairSvc.CheckAndEscalate()
	}
}
