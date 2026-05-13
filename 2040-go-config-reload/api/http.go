package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"config-reload/config"
	"config-reload/storage"
)

type API struct {
	configManager *config.ConfigManager
	storage     *storage.Storage
	mu          sync.RWMutex
}

type Response struct {
	Success bool        `json:"success" xml:"success"`
	Message string      `json:"message,omitempty" xml:"message,omitempty"`
	Data    interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

type ConfigStatusResponse struct {
	CurrentHash string         `json:"currentHash" xml:"currentHash"`
	Services   []ServiceInfo `json:"services" xml:"services>service"`
}

type ServiceInfo struct {
	Name    string `json:"name" xml:"name"`
	Host    string `json:"host" xml:"host"`
	Port    int    `json:"port" xml:"port"`
	Enabled bool   `json:"enabled" xml:"enabled"`
}

type ReloadResponse struct {
	Changed  bool           `json:"changed" xml:"changed"`
	OldHash string        `json:"oldHash" xml:"oldHash"`
	NewHash string        `json:"newHash" xml:"newHash"`
	Added   []string      `json:"added" xml:"added>item"`
	Removed []string      `json:"removed" xml:"removed>item"`
	Modified []string     `json:"modified" xml:"modified>item"`
	Timestamp string     `json:"timestamp" xml:"timestamp"`
}

type HistoryResponse struct {
	History []storage.ConfigHistory `json:"history" xml:"history>item"`
}

func NewAPI(cm *config.ConfigManager, s *storage.Storage) *API {
	return &API{
		configManager: cm,
		storage:       s,
	}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/status", a.handleStatus)
	mux.HandleFunc("/reload", a.handleReload)
	mux.HandleFunc("/history", a.handleHistory)
	mux.HandleFunc("/services", a.handleServices)
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	a.writeResponse(w, r, Response{Success: true, Message: "ok"})
}

func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	cfg := a.configManager.GetConfig()
	hash := a.configManager.GetCurrentHash()

	resp := ConfigStatusResponse{
		CurrentHash: hash,
	}

	if cfg != nil {
		for _, svc := range cfg.Services {
			resp.Services = append(resp.Services, ServiceInfo{
				Name:    svc.Name,
				Host:    svc.Host,
				Port:    svc.Port,
				Enabled: svc.Enabled,
			})
		}
	}

	a.writeResponse(w, r, Response{Success: true, Data: resp})
}

func (a *API) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	oldHash := a.configManager.GetCurrentHash()

	change, err := a.configManager.Reload()
	if err != nil {
		a.storage.LogError(fmt.Sprintf("Manual reload failed: %v", err))
		a.storage.RecordFailure("", oldHash, err.Error())
		a.writeError(w, r, http.StatusBadRequest, fmt.Sprintf("Reload failed: %v", err))
		return
	}

	if change == nil {
		a.writeResponse(w, r, Response{
			Success: true,
			Message: "No changes detected",
			Data: ReloadResponse{
				Changed: false,
				OldHash: oldHash,
				NewHash: oldHash,
				Timestamp: time.Now().Format(time.RFC3339),
			},
		})
		return
	}

	newHash := a.configManager.GetCurrentHash()
	message := fmt.Sprintf("Config reloaded successfully")
	a.storage.RecordChange(change, storage.ConfigToJSON(a.configManager.GetConfig()), message, true)
	a.storage.LogInfo(fmt.Sprintf("Manual reload succeeded. New hash: %s", newHash))

	a.writeResponse(w, r, Response{
		Success: true,
		Message: message,
		Data: ReloadResponse{
			Changed:  true,
			OldHash:  oldHash,
			NewHash:  newHash,
			Added:   change.Added,
			Removed: change.Removed,
			Modified: change.Modified,
			Timestamp: change.Timestamp.Format(time.RFC3339),
		},
	})
}

func (a *API) handleHistory(w http.ResponseWriter, r *http.Request) {
	history, err := a.storage.GetRecentHistory(50)
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get history: %v", err))
		return
	}

	a.writeResponse(w, r, Response{
		Success: true,
		Data: HistoryResponse{History: history},
	})
}

func (a *API) handleServices(w http.ResponseWriter, r *http.Request) {
	cfg := a.configManager.GetConfig()
	if cfg == nil {
		a.writeResponse(w, r, Response{Success: true, Data: []ServiceInfo{}})
		return
	}

	a.writeResponse(w, r, Response{Success: true, Data: cfg.Services})
}

func (a *API) writeResponse(w http.ResponseWriter, r *http.Request, resp Response) {
	a.writeJSON(w, http.StatusOK, resp)
}

func (a *API) writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	a.writeJSON(w, status, Response{Success: false, Message: message})
}

func (a *API) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
