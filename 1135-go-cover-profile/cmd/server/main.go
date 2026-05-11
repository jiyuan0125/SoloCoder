package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	store := NewAnalysisStore()
	server := NewServer(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/upload", server.UploadHandler)
	mux.HandleFunc("/coverage/function", server.FunctionCoverageHandler)
	mux.HandleFunc("/coverage/package", server.PackageCoverageHandler)
	mux.HandleFunc("/coverage/line", server.LineCoverageHandler)
	mux.HandleFunc("/summary", server.SummaryHandler)
	mux.HandleFunc("/diff", server.DiffHandler)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Coverage analysis server starting on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  POST /upload                 - Upload coverprofile file")
	log.Printf("  GET  /coverage/function?id=  - Get function-level coverage")
	log.Printf("  GET  /coverage/package?id=   - Get package-level coverage")
	log.Printf("  GET  /coverage/line?id=      - Get line-level coverage")
	log.Printf("  GET  /summary?id=            - Get overall summary")
	log.Printf("  GET  /diff?old=&new=         - Compare two analyses")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
