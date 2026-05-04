package main

import (
	"encoding/json"
	"io"
	"net/http"

	"ip-location/common"
	"ip-location/iplocation"
)

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.QueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := common.QueryResponse{
		IP: req.IP,
	}

	loc, err := db.Query(req.IP)
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else {
		resp.Success = true
		resp.Country = loc.Country
		resp.Province = loc.Province
		resp.City = loc.City
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.BatchQueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := common.BatchQueryResponse{}

	results, err := db.BatchQuery(req.IPs)
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else {
		resp.Success = true
		resp.Results = make([]common.IPQueryResult, len(results))
		for i, r := range results {
			resp.Results[i] = common.IPQueryResult{
				IP:       r.IP,
				Country:  r.Location.Country,
				Province: r.Location.Province,
				City:     r.Location.City,
			}
			if r.Error != nil {
				resp.Results[i].Error = r.Error.Error()
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSameProvince(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.SameProvinceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := common.SameProvinceResponse{}

	same, err := db.SameProvince(req.IP1, req.IP2)
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else {
		resp.Success = true
		resp.Same = same
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleIPInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.IPInfoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := common.IPInfoResponse{
		IP: req.IP,
	}

	isPrivate, err := iplocation.IsPrivateIP(req.IP)
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else {
		ipType, err := iplocation.GetIPType(req.IP)
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
		} else {
			resp.Success = true
			resp.IsPrivate = isPrivate
			resp.Type = ipType.String()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := map[string]interface{}{
		"status":     "ok",
		"rangeCount": db.RangeCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
