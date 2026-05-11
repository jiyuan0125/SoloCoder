package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/linearregression/pkg/api"
	"github.com/linearregression/pkg/regression"
)

var (
	currentModel *regression.Model
	currentR2    float64
	mu           sync.RWMutex
)

func fitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.FitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.FitResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	
	var model *regression.Model
	var r2 float64
	var err error
	
	if req.Simple != nil && len(req.Simple) > 0 {
		model, r2, err = regression.FitSimple(req.Simple)
		if err != nil {
			json.NewEncoder(w).Encode(api.FitResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		mu.Lock()
		currentModel = model
		currentR2 = r2
		mu.Unlock()
		
		json.NewEncoder(w).Encode(api.FitResponse{
			Success:  true,
			Slope:    api.Float64(model.Slope()),
			Intercept: api.Float64(model.Intercept()),
			R2:       api.Float64(r2),
		})
	} else if req.Features != nil && req.Target != nil {
		model, r2, err = regression.FitMultiple(req.Features, req.Target)
		if err != nil {
			json.NewEncoder(w).Encode(api.FitResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		mu.Lock()
		currentModel = model
		currentR2 = r2
		mu.Unlock()
		
		json.NewEncoder(w).Encode(api.FitResponse{
			Success:      true,
			Intercept:    api.Float64(model.Intercept()),
			Coefficients: api.ToFloat64Slice(model.Coefficients()),
			R2:           api.Float64(r2),
		})
	} else {
		json.NewEncoder(w).Encode(api.FitResponse{
			Success: false,
			Error:   "no valid input data provided",
		})
	}
}

func predictHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	mu.RLock()
	model := currentModel
	mu.RUnlock()
	
	if model == nil {
		json.NewEncoder(w).Encode(api.PredictResponse{
			Success: false,
			Error:   "no model fitted yet, please call fit first",
		})
		return
	}
	
	var req api.PredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(api.PredictResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	
	var yPred float64
	var err error
	
	if model.IsSimple() {
		if req.SimpleX == nil {
			json.NewEncoder(w).Encode(api.PredictResponse{
				Success: false,
				Error:   "simple model requires simple_x input",
			})
			return
		}
		yPred, err = model.PredictSimple(*req.SimpleX)
	} else {
		if req.Features == nil {
			json.NewEncoder(w).Encode(api.PredictResponse{
				Success: false,
				Error:   "multiple model requires features input",
			})
			return
		}
		yPred, err = model.PredictMultiple(req.Features)
	}
	
	if err != nil {
		json.NewEncoder(w).Encode(api.PredictResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	json.NewEncoder(w).Encode(api.PredictResponse{
		Success: true,
		Value:   api.Float64(yPred),
	})
}

func main() {
	http.HandleFunc("/fit", fitHandler)
	http.HandleFunc("/predict", predictHandler)
	
	http.ListenAndServe(":8080", nil)
}
