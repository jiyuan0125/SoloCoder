package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fuzzysearch/server"
)

func main() {
	var portFlag string
	flag.StringVar(&portFlag, "port", "", "Port to listen on (default: 8303 or FUZZY_PORT env variable)")
	flag.StringVar(&portFlag, "p", "", "Short alias for -port")
	flag.Parse()

	port, err := server.ParsePort(portFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fs := server.NewFuzzyServer()
	if err := fs.Start(port); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
