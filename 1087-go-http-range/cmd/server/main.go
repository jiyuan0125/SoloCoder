package main

import (
	"fmt"
	"net/http"
	"os"

	"example.com/httprange/pkg/server"
)

func main() {
	content := []byte("This is a test file for HTTP Range requests. It contains multiple lines of text to demonstrate various Range header scenarios. The file is long enough to test different range combinations including single ranges, suffix ranges, and multiple ranges.")

	srv := server.NewServer(content)
	port := ":8102"

	fmt.Printf("Server starting on %s\n", port)
	fmt.Printf("ETag: %s\n", srv.GetETag())
	fmt.Printf("Content length: %d bytes\n", len(content))

	http.Handle("/", srv)

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
