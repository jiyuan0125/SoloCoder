package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"

	"lsm-storage/internal/lsm"
	"lsm-storage/pkg/api"
)

func main() {
	addr := flag.String("addr", ":8080", "server address")
	dataDir := flag.String("data", "./data", "data directory")
	flag.Parse()

	cfg := lsm.DefaultConfig()
	cfg.Dir = *dataDir

	db, err := lsm.Open(cfg)
	if err != nil {
		fmt.Printf("failed to open db: %v\n", err)
		return
	}
	defer db.Close()

	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(api.PutResponse{Success: false, Error: "method not allowed"})
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.PutResponse{Success: false, Error: err.Error()})
			return
		}
		var req api.PutRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.PutResponse{Success: false, Error: err.Error()})
			return
		}
		if err := db.Put([]byte(req.Key), []byte(req.Value)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.PutResponse{Success: false, Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(api.PutResponse{Success: true})
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(api.GetResponse{Success: false, Error: "method not allowed"})
			return
		}
		var key string
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.GetResponse{Success: false, Error: err.Error()})
				return
			}
			var req api.GetRequest
			if err := json.Unmarshal(body, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.GetResponse{Success: false, Error: err.Error()})
				return
			}
			key = req.Key
		} else {
			key = r.URL.Query().Get("key")
		}
		val, found, err := db.Get([]byte(key))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.GetResponse{Success: false, Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(api.GetResponse{Value: string(val), Found: found, Success: true})
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(api.DeleteResponse{Success: false, Error: "method not allowed"})
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.DeleteResponse{Success: false, Error: err.Error()})
			return
		}
		var req api.DeleteRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.DeleteResponse{Success: false, Error: err.Error()})
			return
		}
		if err := db.Delete([]byte(req.Key)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.DeleteResponse{Success: false, Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(api.DeleteResponse{Success: true})
	})

	http.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(api.ScanResponse{Success: false, Error: "method not allowed"})
			return
		}
		var start, end string
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.ScanResponse{Success: false, Error: err.Error()})
				return
			}
			var req api.ScanRequest
			if err := json.Unmarshal(body, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.ScanResponse{Success: false, Error: err.Error()})
				return
			}
			start = req.Start
			end = req.End
		} else {
			start = r.URL.Query().Get("start")
			end = r.URL.Query().Get("end")
		}
		var startBytes, endBytes []byte
		if start != "" {
			startBytes = []byte(start)
		}
		if end != "" {
			endBytes = []byte(end)
		}
		pairs := db.Scan(startBytes, endBytes)
		var apiPairs []api.KVPair
		for _, p := range pairs {
			apiPairs = append(apiPairs, api.KVPair{Key: string(p.Key), Value: string(p.Value)})
		}
		json.NewEncoder(w).Encode(api.ScanResponse{Pairs: apiPairs, Success: true})
	})

	fmt.Printf("server starting on %s, data dir: %s\n", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
