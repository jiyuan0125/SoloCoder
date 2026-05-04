package main

import "net/http"

func setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/query", handleQuery)
	mux.HandleFunc("/batch", handleBatch)
	mux.HandleFunc("/same-province", handleSameProvince)
	mux.HandleFunc("/info", handleIPInfo)
	mux.HandleFunc("/health", handleHealth)
}
