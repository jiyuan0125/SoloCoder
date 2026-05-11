package main

import (
	"flag"
	"log"
	"marketplace/core"
	"net/http"
	"os"
	"strings"
	"time"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flagPort := flag.String("port", "9017", "服务监听端口")
		flag.Parse()
		port = *flagPort
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return port
}

func main() {
	port := getPort()
	service := core.NewService()
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/users/register", handler.RegisterUser)
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.SearchItems(w, r)
		} else if r.Method == http.MethodPost {
			handler.CreateItem(w, r)
		} else {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/items/")
		if strings.Contains(path, "/approve") {
			handler.ApproveItem(w, r)
		} else if strings.Contains(path, "/offer") {
			handler.MakeOffer(w, r)
		} else {
			if r.Method == http.MethodGet {
				handler.GetItem(w, r)
			} else if r.Method == http.MethodDelete {
				handler.RemoveItem(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
		}
	})

	mux.HandleFunc("/negotiations/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/seller-respond") {
			handler.SellerRespond(w, r)
		} else if strings.Contains(r.URL.Path, "/buyer-respond") {
			handler.BuyerRespond(w, r)
		} else {
			http.Error(w, "路径错误", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/pay") {
			handler.PayOrder(w, r)
		} else if strings.Contains(r.URL.Path, "/ship") {
			handler.ShipOrder(w, r)
		} else if strings.Contains(r.URL.Path, "/confirm") {
			handler.ConfirmReceiveOrder(w, r)
		} else {
			if r.Method == http.MethodGet {
				handler.GetOrder(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
		}
	})

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			service.ProcessTimeouts()
		}
	}()

	log.Printf("服务端启动，监听端口 %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
