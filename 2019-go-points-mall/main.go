package main

import (
	"log"
	"net/http"
	"strings"

	"points-mall/db"
	"points-mall/handlers"
	"points-mall/scheduler"
)

func main() {
	if err := db.Init("./data/points_mall.db"); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	scheduler.StartExpiredLockCleanup()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		handlers.ListProducts(w, r)
	})

	mux.HandleFunc("/api/products/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/products/")
		parts := strings.Split(path, "/")
		if len(parts) == 1 {
			handlers.GetProduct(w, r)
		} else if len(parts) == 2 && parts[1] == "exchanges" {
			handlers.GetProductExchanges(w, r)
		} else {
			handlers.RespondError(w, http.StatusNotFound, "路由不存在")
		}
	})

	mux.HandleFunc("/api/admin/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateProduct(w, r)
		} else if r.Method == http.MethodGet {
			handlers.ListAdminProducts(w, r)
		} else {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/api/admin/products/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
		parts := strings.Split(path, "/")
		if len(parts) == 2 && parts[1] == "status" {
			if r.Method == http.MethodPut {
				handlers.UpdateProductStatus(w, r)
			} else {
				handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
		} else if len(parts) == 2 && parts[1] == "points" {
			if r.Method == http.MethodPut {
				handlers.UpdateProductPoints(w, r)
			} else {
				handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
		} else {
			handlers.RespondError(w, http.StatusNotFound, "路由不存在")
		}
	})

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		handlers.CreateUser(w, r)
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		if strings.HasSuffix(path, "/exchanges") {
			handlers.GetUserExchangeHistory(w, r)
		} else if strings.HasSuffix(path, "/codes") {
			handlers.GetUserCodes(w, r)
		} else {
			handlers.GetUser(w, r)
		}
	})

	mux.HandleFunc("/api/exchanges/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		handlers.StartExchange(w, r)
	})

	mux.HandleFunc("/api/exchanges/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/exchanges/")
		parts := strings.Split(path, "/")
		if len(parts) == 2 && parts[1] == "confirm" {
			if r.Method == http.MethodPost {
				handlers.ConfirmExchange(w, r)
			} else {
				handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
		} else if len(parts) == 2 && parts[1] == "cancel" {
			if r.Method == http.MethodPost {
				handlers.CancelExchange(w, r)
			} else {
				handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
		} else if len(parts) == 2 && parts[1] == "details" {
			if r.Method == http.MethodGet {
				handlers.GetExchangeDetails(w, r)
			} else {
				handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
		} else if len(parts) == 1 {
			if r.Method == http.MethodGet {
				handlers.GetExchange(w, r)
			} else {
				handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			}
		} else {
			handlers.RespondError(w, http.StatusNotFound, "路由不存在")
		}
	})

	mux.HandleFunc("/api/codes/redeem", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlers.RespondError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		handlers.RedeemCode(w, r)
	})

	log.Println("服务器启动在 :8402")
	if err := http.ListenAndServe(":8402", mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
