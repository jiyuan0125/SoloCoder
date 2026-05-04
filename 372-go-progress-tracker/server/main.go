package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	store := NewProgressStore()
	handler := NewHandler(store)

	http.HandleFunc("/progress", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetProgress(w, r)
		case http.MethodPost:
			handler.CreateProgress(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/progress/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/progress/update" || path == "/progress/update/" {
			handler.UpdateProgress(w, r)
			return
		}
		if path == "/progress/set" || path == "/progress/set/" {
			handler.SetProgress(w, r)
			return
		}
		if path == "/progress/cancel" || path == "/progress/cancel/" {
			handler.CancelProgress(w, r)
			return
		}
		if path == "/progress/close" || path == "/progress/close/" {
			handler.CloseProgress(w, r)
			return
		}
		if path == "/progress/subtask" || path == "/progress/subtask/" {
			handler.AddSubTask(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handler.GetProgress(w, r)
		case http.MethodPost:
			if strings.Contains(path, "/update") {
				handler.UpdateProgress(w, r)
			} else if strings.Contains(path, "/set") {
				handler.SetProgress(w, r)
			} else if strings.Contains(path, "/cancel") {
				handler.CancelProgress(w, r)
			} else if strings.Contains(path, "/close") {
				handler.CloseProgress(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Progress tracker server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  POST /progress          - Create new progress")
	log.Println("  GET  /progress          - List all progress")
	log.Println("  GET  /progress/{id}     - Get specific progress")
	log.Println("  POST /progress/update   - Update progress (delta)")
	log.Println("  POST /progress/set      - Set progress (absolute)")
	log.Println("  POST /progress/cancel   - Cancel progress")
	log.Println("  POST /progress/close    - Close progress")
	log.Println("  POST /progress/subtask  - Add subtask")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
