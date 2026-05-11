package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8420", "server port")
	flag.Parse()

	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/convert", server.handleConvert)
	mux.HandleFunc("/api/evaluate", server.handleEvaluate)
	mux.HandleFunc("/api/variable/set", server.handleSetVariable)
	mux.HandleFunc("/api/variable/get", server.handleGetVariable)
	mux.HandleFunc("/api/variables", server.handleListVariables)

	addr := ":" + *port
	fmt.Printf("RPN Evaluator Server starting on %s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/convert      - convert infix to RPN")
	fmt.Println("  POST /api/evaluate     - evaluate expression")
	fmt.Println("  POST /api/variable/set - set variable")
	fmt.Println("  POST /api/variable/get - get variable")
	fmt.Println("  GET  /api/variables    - list all variables")

	log.Fatal(http.ListenAndServe(addr, mux))
}
