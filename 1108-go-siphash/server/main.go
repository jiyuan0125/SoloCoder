package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/siphash-service/siphash"
)

var globalKey *siphash.Key
var globalMap *siphash.HashMap

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	key, err := siphash.NewKeyFromBytes([]byte("siphash_test_key_!"))
	if err != nil {
		log.Fatalf("Failed to create key: %v", err)
	}
	globalKey = key
	globalMap = siphash.NewHashMap24(globalKey)

	http.HandleFunc("/hash", handleHash)
	http.HandleFunc("/put", handlePut)
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/batch/put", handleBatchPut)
	http.HandleFunc("/batch/get", handleBatchGet)
	http.HandleFunc("/batch/delete", handleBatchDelete)
	http.HandleFunc("/rotate-key", handleRotateKey)
	http.HandleFunc("/list", handleList)

	log.Printf("SipHash server starting on port %d...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
