package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"approval-flow/pkg/api"
	"approval-flow/pkg/db"
	"approval-flow/pkg/service"
)

func main() {
	dbPath := "./approval.db"
	if path := os.Getenv("DB_PATH"); path != "" {
		dbPath = path
	}

	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := api.InitTestData(); err != nil {
		log.Printf("Failed to init test data: %v", err)
	}

	bgService := service.NewBackgroundService()
	bgService.Start()
	defer bgService.Stop()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleCreateUser(w, r)
		} else if r.Method == http.MethodGet {
			api.HandleListUsers(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.HandleGetUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/chains", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleCreateChain(w, r)
		} else if r.Method == http.MethodGet {
			api.HandleListChains(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/chains/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			api.HandleUpdateChain(w, r)
		} else if r.Method == http.MethodGet {
			api.HandleGetChain(w, r)
		} else if r.Method == http.MethodDelete {
			api.HandleDeleteChain(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.Handle("/api/applications", api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleSubmitApplication(w, r)
		} else if r.Method == http.MethodGet {
			api.HandleListApplications(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/applications/", api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/api/applications/" {
			if r.Method == http.MethodGet {
				api.HandleListApplications(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/approve") {
			if r.Method == http.MethodPost {
				api.HandleApproveApplication(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/reject") {
			if r.Method == http.MethodPost {
				api.HandleRejectApplication(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/transfer") {
			if r.Method == http.MethodPost {
				api.HandleTransferApplication(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/resubmit") {
			if r.Method == http.MethodPost {
				api.HandleResubmitApplication(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/logs") {
			if r.Method == http.MethodGet {
				api.HandleGetApplicationLogs(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if r.Method == http.MethodGet {
			api.HandleGetApplication(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/notifications", api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.HandleListNotifications(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/notifications/", api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleMarkNotificationRead(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/api/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.HandleListReports(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/reports/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.HandleTriggerReportReconcile(w, r)
		} else if r.Method == http.MethodGet {
			api.HandleListReports(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on :" + port + "...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
}
