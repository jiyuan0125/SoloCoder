package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "users-service")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service": "users", "method": "%s", "path": "%s", "query": "%s"}`,
			r.Method, r.URL.Path, r.URL.RawQuery)
	})

	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "orders-service")
		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, `{"service": "orders", "method": "%s", "path": "%s", "body": "%s"}`,
			r.Method, r.URL.Path, strings.TrimSpace(string(body)))
	})

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "api-default")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service": "api-default", "method": "%s", "path": "%s"}`,
			r.Method, r.URL.Path)
	})

	mux.HandleFunc("/中文路径", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"service": "chinese", "path": "%s"}`, r.URL.Path)
	})

	log.Println("Test backend starting on :9001")
	log.Fatal(http.ListenAndServe(":9001", mux))
}
