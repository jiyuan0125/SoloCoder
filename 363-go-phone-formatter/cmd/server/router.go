package main

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/format", FormatHandler)
	mux.HandleFunc("/validate", ValidateHandler)
	mux.HandleFunc("/extract", ExtractHandler)

	return mux
}
