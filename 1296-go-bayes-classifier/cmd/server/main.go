package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bayes-classifier/bayes"
	"github.com/bayes-classifier/common"
)

type Server struct {
	classifier *bayes.Classifier
	port       string
	modelPath  string
}

func NewServer(port string, useTFIDF bool, modelPath string) *Server {
	featureType := bayes.FeatureTypeBagOfWords
	if useTFIDF {
		featureType = bayes.FeatureTypeTFIDF
	}

	return &Server{
		classifier: bayes.NewClassifier(featureType),
		port:       port,
		modelPath:  modelPath,
	}
}

func (s *Server) handleTrain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.TrainRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if len(req.Samples) == 0 {
		sendError(w, "No training samples provided", http.StatusBadRequest)
		return
	}

	samples := make([]bayes.TrainingSample, len(req.Samples))
	for i, sample := range req.Samples {
		samples[i] = bayes.TrainingSample{
			Text:  sample.Text,
			Label: sample.Label,
		}
	}

	if err := s.classifier.Train(samples); err != nil {
		sendError(w, fmt.Sprintf("Training failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp := common.TrainResponse{
		Success:   true,
		Message:   "Training completed successfully",
		TotalDocs: len(req.Samples),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleClassify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.ClassifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		sendError(w, "No text provided for classification", http.StatusBadRequest)
		return
	}

	predictions, bestLabel, err := s.classifier.Predict(req.Text)
	if err != nil {
		sendError(w, fmt.Sprintf("Classification failed: %v", err), http.StatusInternalServerError)
		return
	}

	results := make([]common.ClassifyResult, len(predictions))
	for i, pred := range predictions {
		results[i] = common.ClassifyResult{
			Label: pred.Label,
			Prob:  pred.Prob,
		}
	}

	resp := common.ClassifyResponse{
		Success:   true,
		BestLabel: bestLabel,
		Results:   results,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleExportModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modelPath := s.modelPath
	if modelPath == "" {
		modelPath = "bayes_model.json"
	}

	if err := s.classifier.Save(modelPath); err != nil {
		sendError(w, fmt.Sprintf("Failed to export model: %v", err), http.StatusInternalServerError)
		return
	}

	resp := common.ExportModelResponse{
		Success:   true,
		Message:   "Model exported successfully",
		ModelFile: modelPath,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleImportModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req common.ImportModelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	modelPath := req.ModelFile
	if modelPath == "" {
		modelPath = s.modelPath
	}
	if modelPath == "" {
		modelPath = "bayes_model.json"
	}

	if err := s.classifier.Load(modelPath); err != nil {
		sendError(w, fmt.Sprintf("Failed to import model: %v", err), http.StatusInternalServerError)
		return
	}

	resp := common.ImportModelResponse{
		Success: true,
		Message: "Model imported successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(common.ErrorResponse{
		Success: false,
		Message: message,
	})
}

func getPort() string {
	port := flag.String("port", "", "Server port (default: 8080)")
	flag.Parse()

	if *port != "" {
		return *port
	}

	envPort := os.Getenv("BAYES_PORT")
	if envPort != "" {
		return envPort
	}

	return "8105"
}

func main() {
	port := getPort()
	useTFIDF := flag.Bool("tfidf", false, "Use TF-IDF feature extraction (default: Bag of Words)")
	modelPath := flag.String("model", "", "Default model file path for import/export")
	flag.Parse()

	server := NewServer(port, *useTFIDF, *modelPath)

	http.HandleFunc("/train", server.handleTrain)
	http.HandleFunc("/classify", server.handleClassify)
	http.HandleFunc("/export", server.handleExportModel)
	http.HandleFunc("/import", server.handleImportModel)

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
