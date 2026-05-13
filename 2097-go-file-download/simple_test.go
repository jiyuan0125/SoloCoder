//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting simple test server on :8082")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	log.Fatal(http.ListenAndServe(":8082", nil))
}
