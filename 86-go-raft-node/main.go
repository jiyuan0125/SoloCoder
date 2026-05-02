package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	nodeID = flag.Int("id", 1, "Node ID (1-5)")
)

func main() {
	flag.Parse()

	if *nodeID < 1 || *nodeID > 5 {
		log.Fatalf("Node ID must be between 1 and 5, got %d", *nodeID)
	}

	peers := map[int]string{
		1: "localhost:5001",
		2: "localhost:5002",
		3: "localhost:5003",
		4: "localhost:5004",
		5: "localhost:5005",
	}

	node := NewRaftNode(*nodeID, 5000+*nodeID, peers)
	server := NewRaftServer(node)

	node.Start()
	server.Start()

	fmt.Printf("Raft Node %d started on port %d\n", *nodeID, 5000+*nodeID)
	fmt.Printf("Election timeout: %v\n", node.electionTimeout)
	fmt.Println("Press Ctrl+C to stop...")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	fmt.Println("\nShutting down...")
	node.Stop()
}
