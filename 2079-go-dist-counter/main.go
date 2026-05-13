package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Counter struct {
	ID        int64      `json:"-"`
	Namespace string     `json:"namespace"`
	Name      string     `json:"name"`
	Value     int64      `json:"value"`
	Min       *int64     `json:"min,omitempty"`
	Max       *int64     `json:"max,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CounterHistory struct {
	ID        int64     `json:"id"`
	CounterID int64     `json:"-"`
	Operation string    `json:"operation"`
	OldValue  int64     `json:"old_value"`
	NewValue  int64     `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

type BatchOperationRequest struct {
	Items []struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Delta     int64  `json:"delta"`
	} `json:"items"`
}

type BatchItemResult struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Success   bool   `json:"success"`
	Status    int    `json:"status"`
	Value     int64  `json:"value,omitempty"`
	Error     string `json:"error,omitempty"`
}

type CounterStore struct {
	db   *sql.DB
	lock sync.Mutex
}

func NewCounterStore(db *sql.DB) *CounterStore {
	return &CounterStore{db: db}
}

func (s *CounterStore) Init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS counters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			namespace TEXT NOT NULL,
			name TEXT NOT NULL,
			value INTEGER NOT NULL DEFAULT 0,
			min INTEGER,
			max INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(namespace, name)
		);
		CREATE TABLE IF NOT EXISTS counter_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			counter_id INTEGER NOT NULL,
			operation TEXT NOT NULL,
			old_value INTEGER NOT NULL,
			new_value INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(counter_id) REFERENCES counters(id)
		);
	`)
	return err
}

func (s *CounterStore) GetOrCreateCounter(namespace, name string) (*Counter, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var counter Counter
	err = tx.QueryRow(`SELECT id, namespace, name, value, min, max, created_at, updated_at FROM counters WHERE namespace = ? AND name = ?`, namespace, name).Scan(
		&counter.ID, &counter.Namespace, &counter.Name, &counter.Value, &counter.Min, &counter.Max, &counter.CreatedAt, &counter.UpdatedAt)

	if err == sql.ErrNoRows {
		result, err := tx.Exec(`INSERT INTO counters (namespace, name) VALUES (?, ?)`, namespace, name)
		if err != nil {
			return nil, err
		}
		counter.ID, _ = result.LastInsertId()
		counter.Namespace = namespace
		counter.Name = name
		counter.Value = 0
		counter.CreatedAt = time.Now()
		counter.UpdatedAt = time.Now()
	} else if err != nil {
		return nil, err
	}

	return &counter, tx.Commit()
}

func (s *CounterStore) SetValue(namespace, name string, value int64) (int64, int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}
	defer tx.Rollback()

	var counter Counter
	err = tx.QueryRow(`SELECT id, value, min, max FROM counters WHERE namespace = ? AND name = ?`, namespace, name).Scan(
		&counter.ID, &counter.Value, &counter.Min, &counter.Max)

	if err == sql.ErrNoRows {
		result, err := tx.Exec(`INSERT INTO counters (namespace, name, value) VALUES (?, ?, ?)`, namespace, name, value)
		if err != nil {
			return 0, http.StatusInternalServerError, err
		}
		counter.ID, _ = result.LastInsertId()
		oldValue := int64(0)
		_, err = tx.Exec(`INSERT INTO counter_history (counter_id, operation, old_value, new_value) VALUES (?, ?, ?, ?)`, counter.ID, "SET", oldValue, value)
		if err != nil {
			return 0, http.StatusInternalServerError, err
		}
		return value, http.StatusOK, tx.Commit()
	} else if err != nil {
		return 0, http.StatusInternalServerError, err
	}

	if counter.Min != nil && value < *counter.Min {
		return counter.Value, http.StatusConflict, nil
	}
	if counter.Max != nil && value > *counter.Max {
		return counter.Value, http.StatusConflict, nil
	}

	oldValue := counter.Value
	_, err = tx.Exec(`UPDATE counters SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, value, counter.ID)
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}
	_, err = tx.Exec(`INSERT INTO counter_history (counter_id, operation, old_value, new_value) VALUES (?, ?, ?, ?)`, counter.ID, "SET", oldValue, value)
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}
	return value, http.StatusOK, tx.Commit()
}

func (s *CounterStore) AddDelta(namespace, name string, delta int64) (int64, int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}
	defer tx.Rollback()

	var counter Counter
	err = tx.QueryRow(`SELECT id, value, min, max FROM counters WHERE namespace = ? AND name = ?`, namespace, name).Scan(
		&counter.ID, &counter.Value, &counter.Min, &counter.Max)

	if err == sql.ErrNoRows {
		newValue := delta
		if newValue < 0 {
			newValue = 0
		}
		result, err := tx.Exec(`INSERT INTO counters (namespace, name, value) VALUES (?, ?, ?)`, namespace, name, newValue)
		if err != nil {
			return 0, http.StatusInternalServerError, err
		}
		counter.ID, _ = result.LastInsertId()
		_, err = tx.Exec(`INSERT INTO counter_history (counter_id, operation, old_value, new_value) VALUES (?, ?, ?, ?)`, counter.ID, "ADD", int64(0), newValue)
		if err != nil {
			return 0, http.StatusInternalServerError, err
		}
		return newValue, http.StatusOK, tx.Commit()
	} else if err != nil {
		return 0, http.StatusInternalServerError, err
	}

	newValue := counter.Value + delta

	if counter.Min != nil && newValue < *counter.Min {
		return counter.Value, http.StatusConflict, nil
	}
	if counter.Max != nil && newValue > *counter.Max {
		return counter.Value, http.StatusConflict, nil
	}

	oldValue := counter.Value
	_, err = tx.Exec(`UPDATE counters SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, newValue, counter.ID)
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}
	operation := "ADD"
	if delta < 0 {
		operation = "SUB"
	}
	_, err = tx.Exec(`INSERT INTO counter_history (counter_id, operation, old_value, new_value) VALUES (?, ?, ?, ?)`, counter.ID, operation, oldValue, newValue)
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}
	return newValue, http.StatusOK, tx.Commit()
}

func (s *CounterStore) SetBounds(namespace, name string, min, max *int64) (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var counter Counter
	err = tx.QueryRow(`SELECT id, value FROM counters WHERE namespace = ? AND name = ?`, namespace, name).Scan(&counter.ID, &counter.Value)

	if err == sql.ErrNoRows {
		result, err := tx.Exec(`INSERT INTO counters (namespace, name, min, max) VALUES (?, ?, ?, ?)`, namespace, name, min, max)
		if err != nil {
			return 0, err
		}
		counter.ID, _ = result.LastInsertId()
		counter.Value = 0
	} else if err != nil {
		return 0, err
	}

	if min != nil && counter.Value < *min {
		_, err = tx.Exec(`UPDATE counters SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *min, counter.ID)
		if err != nil {
			return 0, err
		}
		counter.Value = *min
	}
	if max != nil && counter.Value > *max {
		_, err = tx.Exec(`UPDATE counters SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, *max, counter.ID)
		if err != nil {
			return 0, err
		}
		counter.Value = *max
	}

	_, err = tx.Exec(`UPDATE counters SET min = ?, max = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, min, max, counter.ID)
	if err != nil {
		return 0, err
	}

	return counter.Value, tx.Commit()
}

func (s *CounterStore) ListNamespaces() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT namespace FROM counters GROUP BY namespace`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var namespaces []string
	for rows.Next() {
		var ns string
		if err := rows.Scan(&ns); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

func (s *CounterStore) GetNamespaceCounters(namespace string) ([]Counter, error) {
	rows, err := s.db.Query(`SELECT id, namespace, name, value, min, max, created_at, updated_at FROM counters WHERE namespace = ? ORDER BY name`, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []Counter
	for rows.Next() {
		var c Counter
		if err := rows.Scan(&c.ID, &c.Namespace, &c.Name, &c.Value, &c.Min, &c.Max, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		counters = append(counters, c)
	}
	return counters, nil
}

func (s *CounterStore) GetHistory(counterID int64) ([]CounterHistory, error) {
	rows, err := s.db.Query(`SELECT id, counter_id, operation, old_value, new_value, created_at FROM counter_history WHERE counter_id = ? ORDER BY created_at DESC`, counterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []CounterHistory
	for rows.Next() {
		var h CounterHistory
		if err := rows.Scan(&h.ID, &h.CounterID, &h.Operation, &h.OldValue, &h.NewValue, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func main() {
	db, err := sql.Open("sqlite", "./counter.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewCounterStore(db)
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/namespaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		namespaces, err := store.ListNamespaces()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusOK, namespaces)
	})

	mux.HandleFunc("/api/namespaces/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		parts := splitPath(path)
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}

		namespace := parts[2]

		if len(parts) == 3 {
			if r.Method != http.MethodGet {
				http.NotFound(w, r)
				return
			}
			counters, err := store.GetNamespaceCounters(namespace)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondWithJSON(w, http.StatusOK, counters)
		}

		if len(parts) >= 4 {
			name := parts[3]

			if len(parts) == 4 {
				switch r.Method {
				case http.MethodGet:
					counter, err := store.GetOrCreateCounter(namespace, name)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					respondWithJSON(w, http.StatusOK, counter)
				case http.MethodPost:
					var req struct {
						Delta int64 `json:"delta"`
					}
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					newValue, status, err := store.AddDelta(namespace, name, req.Delta)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					respondWithJSON(w, status, map[string]int64{"value": newValue})
				case http.MethodPut:
					var req struct {
						Value *int64 `json:"value"`
					}
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					if req.Value == nil {
						http.Error(w, "value is required", http.StatusBadRequest)
						return
					}
					newValue, status, err := store.SetValue(namespace, name, *req.Value)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					respondWithJSON(w, status, map[string]int64{"value": newValue})
				default:
					http.NotFound(w, r)
				}
			}

			if len(parts) == 5 {
				sub := parts[4]
				if sub == "history" {
					if r.Method != http.MethodGet {
						http.NotFound(w, r)
						return
					}
					counter, err := store.GetOrCreateCounter(namespace, name)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					history, err := store.GetHistory(counter.ID)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					respondWithJSON(w, http.StatusOK, history)
				} else if sub == "bounds" {
					if r.Method != http.MethodPut {
						http.NotFound(w, r)
						return
					}
					var req struct {
						Min *int64 `json:"min"`
						Max *int64 `json:"max"`
					}
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					value, err := store.SetBounds(namespace, name, req.Min, req.Max)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					respondWithJSON(w, http.StatusOK, map[string]int64{"value": value})
				} else {
					http.NotFound(w, r)
				}
			}
		}
	})

	mux.HandleFunc("/api/batch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		var req BatchOperationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var results []BatchItemResult
		hasError := false

		for _, item := range req.Items {
			newValue, status, err := store.AddDelta(item.Namespace, item.Name, item.Delta)
			result := BatchItemResult{
				Namespace: item.Namespace,
				Name:      item.Name,
				Value:     newValue,
				Status:    status,
			}
			if err != nil {
				result.Success = false
				result.Error = err.Error()
				hasError = true
			} else {
				result.Success = status == http.StatusOK
				if !result.Success {
					hasError = true
				}
			}
			results = append(results, result)
		}

		statusCode := http.StatusOK
		if hasError {
			statusCode = 207
		}
		respondWithJSON(w, statusCode, results)
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
