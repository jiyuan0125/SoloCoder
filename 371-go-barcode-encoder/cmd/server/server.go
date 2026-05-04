package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	port    int
	handler *Handler
}

func NewServer(port int) *Server {
	return &Server{
		port:    port,
		handler: NewHandler(),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/encode", s.handler.Encode)

	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, mux)
}
