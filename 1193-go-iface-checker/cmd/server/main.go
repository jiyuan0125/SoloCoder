package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"iface-checker/pkg/checker"
	"iface-checker/pkg/models"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	registry   = checker.NewTypeRegistry()
	registryMu sync.RWMutex
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		var p int
		if _, err := fmt.Sscanf(envPort, "%d", &p); err == nil {
			port = p
		}
	}

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/check", handleCheck)
	http.HandleFunc("/check-all", handleCheckAll)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	for _, s := range req.Structs {
		registry.RegisterStruct(s)
	}

	for _, iface := range req.Interfaces {
		registry.RegisterInterface(iface)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	registryMu.RLock()
	defer registryMu.RUnlock()

	resp := checker.Check(registry, req.StructName, req.InterfaceName, req.UseValueReceiver)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCheckAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CheckAllRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tempRegistry := checker.NewTypeRegistry()

	for _, s := range req.Structs {
		tempRegistry.RegisterStruct(s)
	}

	for _, iface := range req.Interfaces {
		tempRegistry.RegisterInterface(iface)
	}

	var results []models.InterfaceCheckResult
	for _, iface := range req.Interfaces {
		checkResp := checker.Check(tempRegistry, req.TargetStruct, iface.Name, req.UseValueReceiver)
		results = append(results, models.InterfaceCheckResult{
			InterfaceName: iface.Name,
			CheckResponse: checkResp,
		})
	}

	resp := models.CheckAllResponse{Results: results}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
