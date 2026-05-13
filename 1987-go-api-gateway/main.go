package main

import (
	"fmt"
	"net/http"
	"os"

	"api-gateway/admin"
	"api-gateway/backend"
	"api-gateway/logger"
	"api-gateway/proxy"
	"api-gateway/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.NewRouter()
	bm := backend.NewBackendManager()
	al := logger.NewAccessLogger()
	p := proxy.NewProxy(r, bm, al)
	as := admin.NewAdminServer(r, bm, al)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" || len(req.URL.Path) < 7 || req.URL.Path[:7] != "/admin/" {
			p.ServeHTTP(w, req)
		} else {
			as.ServeHTTP(w, req)
		}
	})

	addr := ":" + port
	fmt.Printf("API Gateway listening on %s\n", addr)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
