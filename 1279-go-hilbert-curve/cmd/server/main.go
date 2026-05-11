package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var portFlag string
	flag.StringVar(&portFlag, "port", "", "Port to listen on")
	flag.Parse()

	port := getPort()
	if portFlag != "" {
		if _, err := strconv.Atoi(portFlag); err != nil {
			fmt.Printf("Invalid port number: %s. Using default %s.\n", portFlag, port)
		} else {
			port = portFlag
		}
	}

	http.HandleFunc("/point-to-index", pointToIndexHandler)
	http.HandleFunc("/index-to-point", indexToPointHandler)
	http.HandleFunc("/range-query", rangeQueryHandler)
	http.HandleFunc("/estimate-distance", estimateDistanceHandler)
	http.HandleFunc("/batch-point-to-index", batchPointToIndexHandler)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Hilbert curve service starting on %s\n", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed to start: %v\n", err)
		os.Exit(1)
	}
}
