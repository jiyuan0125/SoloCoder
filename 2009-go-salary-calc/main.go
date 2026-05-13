package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	port := 9603

	db, err := InitDB("salary.db")
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	if err := InitTables(db); err != nil {
		log.Fatalf("表初始化失败: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateEmployeeHandler(w, r, db)
		case http.MethodGet:
			ListEmployeesHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/employees/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetEmployeeHandler(w, r, db)
		case http.MethodPut:
			UpdateEmployeeHandler(w, r, db)
		case http.MethodDelete:
			DeleteEmployeeHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/resources/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}
		ResourceSummaryHandler(w, r, db)
	})

	mux.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateResourceHandler(w, r, db)
		case http.MethodGet:
			ListResourcesHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/resources/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetResourceHandler(w, r, db)
		case http.MethodPut:
			UpdateResourceHandler(w, r, db)
		case http.MethodDelete:
			DeleteResourceHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/resource-relations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateResourceRelationHandler(w, r, db)
		case http.MethodGet:
			ListResourceRelationsHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/resource-relations/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			DeleteResourceRelationHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/salaries", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CalculateSalaryHandler(w, r, db)
		case http.MethodGet:
			ListSalariesHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	mux.HandleFunc("/salaries/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/next-status") {
			if r.Method == http.MethodPost {
				UpdateSalaryStatusHandler(w, r, db)
				return
			}
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}

		switch r.Method {
		case http.MethodGet:
			GetSalaryHandler(w, r, db)
		default:
			WriteError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
	})

	log.Printf("服务器启动在端口 %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
