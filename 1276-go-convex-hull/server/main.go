package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"convex-hull/server/handlers"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("CONVEX_HULL_PORT")
	}
	if port == "" {
		port = "8504"
	}

	server := handlers.NewServer()

	http.HandleFunc("/compute", server.HandleCompute)
	http.HandleFunc("/perimeter-area", server.HandlePerimeterArea)
	http.HandleFunc("/point-in-hull", server.HandlePointInHull)
	http.HandleFunc("/intersection", server.HandleIntersection)
	http.HandleFunc("/filter-boundary", server.HandleFilterBoundary)
	http.HandleFunc("/weighted", server.HandleWeighted)
	http.HandleFunc("/dynamic/add", server.HandleDynamicAdd)
	http.HandleFunc("/dynamic/get", server.HandleDynamicGet)

	log.Printf("Convex Hull server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
