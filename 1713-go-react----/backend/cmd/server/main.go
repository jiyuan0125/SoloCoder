package main

import (
	"fmt"
	"log"
	"net/http"
	"ohims/internal/config"
	"ohims/internal/handlers"
	"ohims/internal/middleware"
	"ohims/pkg/db"
	"strings"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg.DBPath); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	enterpriseHandler := &handlers.EnterpriseHandler{}
	hazardHandler := &handlers.HazardHandler{}
	workerHandler := &handlers.WorkerHandler{}
	examHandler := &handlers.ExaminationHandler{}
	reportHandler := &handlers.ReportHandler{}
	todoHandler := &handlers.TodoHandler{}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/enterprises", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			enterpriseHandler.List(w, r)
		case http.MethodPost:
			enterpriseHandler.Create(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/enterprises/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.Contains(path, "/sub") {
			if r.Method == http.MethodGet {
				enterpriseHandler.GetSub(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			enterpriseHandler.Get(w, r)
		case http.MethodPut:
			enterpriseHandler.Update(w, r)
		case http.MethodDelete:
			enterpriseHandler.Delete(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/hazard-factors", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hazardHandler.List(w, r)
		case http.MethodPost:
			hazardHandler.Create(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/hazard-factors/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/monitor") {
			if r.Method == http.MethodPut {
				hazardHandler.UpdateMonitor(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			hazardHandler.Get(w, r)
		case http.MethodPut:
			hazardHandler.Update(w, r)
		case http.MethodDelete:
			hazardHandler.Delete(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/workers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			workerHandler.List(w, r)
		case http.MethodPost:
			workerHandler.Create(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/workers/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/exposed-factors") {
			if r.Method == http.MethodPut {
				workerHandler.AddExposedFactors(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			workerHandler.Get(w, r)
		case http.MethodPut:
			workerHandler.Update(w, r)
		case http.MethodDelete:
			workerHandler.Delete(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/examinations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			examHandler.List(w, r)
		case http.MethodPost:
			examHandler.CreateSchedule(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/examinations/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/results") {
			if r.Method == http.MethodPut {
				examHandler.RecordResult(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}
		if strings.HasSuffix(path, "/flow") {
			if r.Method == http.MethodPut {
				examHandler.UpdateFlow(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			examHandler.Get(w, r)
		case http.MethodPut:
			examHandler.Update(w, r)
		case http.MethodDelete:
			examHandler.Delete(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/reports", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			reportHandler.List(w, r)
		case http.MethodPost:
			reportHandler.GenerateEnterpriseReport(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/reports/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/export") {
			if r.Method == http.MethodGet {
				reportHandler.ExportReport(w, r)
			} else {
				http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			reportHandler.Get(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/stats/regulatory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			reportHandler.GetRegulatoryStats(w, r)
		} else {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/stats/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			reportHandler.GetAggregatedMetrics(w, r)
		} else {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoHandler.List(w, r)
		case http.MethodPost:
			todoHandler.Create(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/todos/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoHandler.Get(w, r)
		case http.MethodPut:
			todoHandler.Update(w, r)
		case http.MethodDelete:
			todoHandler.Delete(w, r)
		default:
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		}
	})

	handler := middleware.CORS(mux)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("职业病防治管理系统服务器启动在端口 %d...", cfg.Port)
	log.Fatal(http.ListenAndServe(addr, handler))
}
