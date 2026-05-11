package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"contract-management/core"
	"contract-management/server"
)

const defaultPort = 8080

func main() {
	port := getPort()

	store := core.NewStore()
	contractSvc := core.NewContractService(store)
	milestoneSvc := core.NewMilestoneService(store)
	paymentSvc := core.NewPaymentService(store)
	auditSvc := core.NewAuditService(store)

	handler := server.NewHandler(contractSvc, milestoneSvc, paymentSvc, auditSvc)
	router := server.NewRouter(handler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Contract Management Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func getPort() int {
	portFlag := flag.Int("port", 0, "server port (overrides environment variable)")
	flag.Parse()

	if *portFlag > 0 {
		return *portFlag
	}

	envPort := os.Getenv("SERVER_PORT")
	if envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err == nil && p > 0 {
			return p
		}
	}

	return defaultPort
}
