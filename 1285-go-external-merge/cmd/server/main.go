package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/example/externalsort/pkg/externalsort"
)

const defaultPort = "8512"

type server struct {
	sorter *externalsort.ExternalSorter
}

func main() {
	port := flag.String("port", "", "server port (default: 8080, can also be set via PORT env var)")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = defaultPort
		}
	}

	s := &server{
		sorter: externalsort.New(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", s.handleAddData)
	mux.HandleFunc("/api/config/memory", s.handleSetMemoryLimit)
	mux.HandleFunc("/api/config/merge-ways", s.handleSetMergeWays)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/sort", s.handleSort)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/result", s.handleResult)
	mux.HandleFunc("/api/reset", s.handleReset)

	addr := ":" + *port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
