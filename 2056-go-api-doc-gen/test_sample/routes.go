package test_sample

import (
	"net/http"
)

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/users", GetUsers)
	mux.HandleFunc("POST /api/users", CreateUser)
	mux.HandleFunc("GET /api/users/{id}", GetUserDetail)
	mux.HandleFunc("PUT /api/users/{id}", UpdateUser)
}
