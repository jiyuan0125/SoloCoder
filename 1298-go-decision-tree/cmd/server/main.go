package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"decisiontree/pkg/common"
	"decisiontree/pkg/decisiontree"
)

type trainedTree struct {
	Tree         *decisiontree.Node
	FeatureTypes []decisiontree.FeatureType
	Features     []string
}

var (
	trees     = make(map[string]*trainedTree)
	treesLock sync.RWMutex
	nextID    = 1
	idLock    sync.Mutex
)

func getPort() string {
	port := flag.String("port", "", "Port to listen on (default: 8080 or PORT env)")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	return "8080"
}

func generateTreeID() string {
	idLock.Lock()
	defer idLock.Unlock()
	id := nextID
	nextID++
	return fmt.Sprintf("tree_%d", id)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, common.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func trainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.TrainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.CSVData == "" {
		writeError(w, http.StatusBadRequest, "csv_data is required")
		return
	}

	ds, err := decisiontree.ParseCSV(req.CSVData)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse CSV: %v", err))
		return
	}

	handleMissing := "skip"
	if req.HandleMissing != "" {
		handleMissing = req.HandleMissing
	}

	ds, err = ds.HandleMissingValues(handleMissing)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to handle missing values: %v", err))
		return
	}

	maxDepth := 10
	if req.MaxDepth > 0 {
		maxDepth = req.MaxDepth
	}

	minSamples := 1
	if req.MinSamples > 0 {
		minSamples = req.MinSamples
	}

	params := &decisiontree.Parameters{
		MaxDepth:      maxDepth,
		MinSamples:    minSamples,
		HandleMissing: handleMissing,
	}

	tree, err := decisiontree.Train(ds, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to train tree: %v", err))
		return
	}

	treeID := generateTreeID()

	treesLock.Lock()
	trees[treeID] = &trainedTree{
		Tree:         tree,
		FeatureTypes: ds.FeatureTypes,
		Features:     ds.Features,
	}
	treesLock.Unlock()

	writeJSON(w, http.StatusOK, common.TrainResponse{
		Success: true,
		TreeID:  treeID,
		Message: "tree trained successfully",
	})
}

func predictHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.PredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.TreeID == "" {
		writeError(w, http.StatusBadRequest, "tree_id is required")
		return
	}

	treesLock.RLock()
	tt, exists := trees[req.TreeID]
	treesLock.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, fmt.Sprintf("tree %s not found", req.TreeID))
		return
	}

	if len(req.Data) == 0 {
		writeJSON(w, http.StatusOK, common.PredictResponse{
			Success: true,
			Labels:  []string{},
			Message: "no data to predict",
		})
		return
	}

	labels := make([]string, 0, len(req.Data))
	for i, features := range req.Data {
		if len(features) != len(tt.Features) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("row %d has %d features, expected %d", i+1, len(features), len(tt.Features)))
			return
		}

		label, err := tt.Tree.Predict(features, tt.FeatureTypes)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("prediction failed for row %d: %v", i+1, err))
			return
		}
		labels = append(labels, label)
	}

	writeJSON(w, http.StatusOK, common.PredictResponse{
		Success: true,
		Labels:  labels,
		Message: "prediction successful",
	})
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.TreeExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.TreeID == "" {
		writeError(w, http.StatusBadRequest, "tree_id is required")
		return
	}

	treesLock.RLock()
	tt, exists := trees[req.TreeID]
	treesLock.RUnlock()

	if !exists {
		writeError(w, http.StatusNotFound, fmt.Sprintf("tree %s not found", req.TreeID))
		return
	}

	writeJSON(w, http.StatusOK, common.TreeExportResponse{
		Success:  true,
		JSONTree: tt.Tree.ExportJSON(),
		TextTree: tt.Tree.ExportText(),
		Message:  "export successful",
	})
}

func main() {
	port := getPort()

	http.HandleFunc("/train", trainHandler)
	http.HandleFunc("/predict", predictHandler)
	http.HandleFunc("/export", exportHandler)

	fmt.Printf("Decision Tree Server listening on :%s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /train    - Train a new decision tree\n")
	fmt.Printf("  POST /predict  - Make predictions with a trained tree\n")
	fmt.Printf("  POST /export   - Export a trained tree structure\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
