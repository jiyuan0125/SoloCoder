package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"smart-park/core/access"
	"smart-park/core/patrol"
	"smart-park/core/visitor"
)

func main() {
	port := flag.String("port", "", "Server port (e.g., 8080)")
	flag.Parse()

	if *port == "" {
		*port = os.Getenv("SERVER_PORT")
	}
	if *port == "" {
		*port = "8080"
	}

	accessStore := access.NewStore()
	accessService := access.NewService(accessStore)

	visitorStore := visitor.NewStore()
	visitorService := visitor.NewService(visitorStore, accessService)

	patrolStore := patrol.NewStore()
	patrolService := patrol.NewService(patrolStore, accessService)

	h := NewHandler(accessService, visitorService, patrolService)

	mux := http.NewServeMux()

	mux.HandleFunc("/access/point", h.handleAccessPoint)
	mux.HandleFunc("/access/points", h.handleAccessPointsList)
	mux.HandleFunc("/access/rule", h.handleAccessRule)
	mux.HandleFunc("/access/area-rules", h.handleBatchAreaRules)
	mux.HandleFunc("/access/verify", h.handleAccessVerify)
	mux.HandleFunc("/access/records", h.handleAccessRecords)
	mux.HandleFunc("/access/fix", h.handleFixAccessPoint)

	mux.HandleFunc("/visitor/reserve", h.handleVisitorReserve)
	mux.HandleFunc("/visitor/review", h.handleVisitorReview)
	mux.HandleFunc("/visitor/checkin", h.handleVisitorCheckIn)
	mux.HandleFunc("/visitor/list", h.handleVisitorList)

	mux.HandleFunc("/patrol/point", h.handlePatrolPoint)
	mux.HandleFunc("/patrol/route", h.handlePatrolRoute)
	mux.HandleFunc("/patrol/routes", h.handlePatrolRoutes)
	mux.HandleFunc("/patrol/task", h.handlePatrolTask)
	mux.HandleFunc("/patrol/tasks", h.handlePatrolTasks)
	mux.HandleFunc("/patrol/start", h.handlePatrolStart)
	mux.HandleFunc("/patrol/checkin", h.handlePatrolCheckIn)
	mux.HandleFunc("/patrol/checkins", h.handlePatrolCheckIns)

	fmt.Printf("Server starting on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
