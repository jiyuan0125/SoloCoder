package main

import (
	"flag"
	"fmt"
	"log"
	"merchant-mgmt-system/internal/core"
	"net/http"
	"os"
	"strings"
)

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "服务端口")
	flag.StringVar(flagPort, "p", "", "服务端口 (简写)")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return port
}

func main() {
	service := core.NewService()
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateCustomer(w, r)
		} else if r.Method == http.MethodGet {
			handler.ListCustomers(w, r)
		} else {
			handler.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/api/customers/", handler.GetCustomer)
	mux.HandleFunc("/api/follows", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateFollow(w, r)
		} else if r.Method == http.MethodGet {
			handler.ListFollows(w, r)
		} else {
			handler.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})
	mux.HandleFunc("/api/follows/close", handler.CloseFollow)
	mux.HandleFunc("/api/contracts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateContract(w, r)
		} else if r.Method == http.MethodGet {
			handler.ListContracts(w, r)
		} else {
			handler.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})
	mux.HandleFunc("/api/dashboard", handler.GetDashboard)

	port := getPort()
	fmt.Printf("招商管理系统服务端启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
