package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	storage := NewFileStorage()
	
	if err := storage.Load(); err != nil {
		log.Printf("加载数据失败: %v, 使用空数据", err)
	}

	InitMockOrders(storage)

	applyHandler := &ReturnApplyHandler{storage: storage}
	voucherHandler := &VoucherHandler{storage: storage}
	processHandler := &ProcessHandler{storage: storage}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/returns", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			applyHandler.CreateReturn(w, r)
		case http.MethodGet:
			applyHandler.ListApplications(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/exchanges", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			applyHandler.CreateExchange(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/applications/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		applyHandler.GetApplication(w, r)
	})

	mux.HandleFunc("/api/vouchers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			voucherHandler.UploadVoucher(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/vouchers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		voucherHandler.GetVoucher(w, r)
	})

	mux.HandleFunc("/api/review", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		processHandler.ReviewApplication(w, r)
	})

	mux.HandleFunc("/api/refunds/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		processHandler.GetRefund(w, r)
	})

	mux.HandleFunc("/api/shippings/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		processHandler.GetShipping(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务启动在端口 %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func ParseJSONBody(r *http.Request, dest interface{}) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

func GetIDFromPath(r *http.Request, prefix string) (string, bool) {
	path := strings.TrimPrefix(r.URL.Path, prefix)
	if path == "" || strings.Contains(path, "/") {
		return "", false
	}
	return path, true
}

func InitMockOrders(storage *FileStorage) {
	if len(storage.Orders) > 0 {
		return
	}

	orders := []Order{
		{
			OrderID:     "ORD001",
			UserID:      "USER001",
			SKU:         "SKU001",
			Spec:        "红色/M",
			Quantity:    1,
			UnitPrice:   299.0,
			TotalAmount: 319.0,
			PaidAmount:  319.0,
			ShippingFee: 20.0,
			Status:      OrderStatusCompleted,
			CreatedAt:   GetNowTime(),
		},
		{
			OrderID:     "ORD002",
			UserID:      "USER002",
			SKU:         "SKU002",
			Spec:        "蓝色/L",
			Quantity:    2,
			UnitPrice:   159.0,
			TotalAmount: 338.0,
			PaidAmount:  338.0,
			ShippingFee: 15.0,
			Status:      OrderStatusDelivered,
			CreatedAt:   GetNowTime(),
		},
		{
			OrderID:     "ORD003",
			UserID:      "USER003",
			SKU:         "SKU003",
			Spec:        "黑色/XL",
			Quantity:    1,
			UnitPrice:   599.0,
			TotalAmount: 614.0,
			PaidAmount:  614.0,
			ShippingFee: 15.0,
			Status:      OrderStatusCompleted,
			CreatedAt:   GetNowTime(),
		},
	}

	skuPrices := []SKUPrice{
		{SKU: "SKU001", Spec: "红色/M", Price: 299.0},
		{SKU: "SKU001", Spec: "红色/L", Price: 299.0},
		{SKU: "SKU001", Spec: "蓝色/M", Price: 289.0},
		{SKU: "SKU001", Spec: "蓝色/L", Price: 289.0},
		{SKU: "SKU002", Spec: "蓝色/L", Price: 159.0},
		{SKU: "SKU002", Spec: "蓝色/XL", Price: 169.0},
		{SKU: "SKU002", Spec: "绿色/L", Price: 149.0},
		{SKU: "SKU003", Spec: "黑色/XL", Price: 599.0},
		{SKU: "SKU003", Spec: "黑色/XXL", Price: 619.0},
		{SKU: "SKU003", Spec: "灰色/XL", Price: 579.0},
	}

	for i := range orders {
		storage.Orders[orders[i].OrderID] = &orders[i]
	}

	for i := range skuPrices {
		key := fmt.Sprintf("%s_%s", skuPrices[i].SKU, skuPrices[i].Spec)
		storage.SKUPrices[key] = &skuPrices[i]
	}

	storage.Save()
}
