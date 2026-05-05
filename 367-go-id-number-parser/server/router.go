package main

import (
	"net/http"
)

func setupRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/api/parse", parseHandler)
	router.HandleFunc("/api/validate", validateHandler)

	return router
}
