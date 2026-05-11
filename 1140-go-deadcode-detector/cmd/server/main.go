package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"deadcode-detector/internal/common"
	"deadcode-detector/internal/deadcode"
)

type ReportStore struct {
	mu       sync.RWMutex
	reports  map[string]common.AnalysisReport
	refMaps  map[string]*deadcode.ReferenceMap
	nextID   int
}

func NewReportStore() *ReportStore {
	return &ReportStore{
		reports:  make(map[string]common.AnalysisReport),
		refMaps:  make(map[string]*deadcode.ReferenceMap),
		nextID:   1,
	}
}

func (s *ReportStore) Add(report common.AnalysisReport, refMap *deadcode.ReferenceMap) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("report_%d", s.nextID)
	s.nextID++
	s.reports[id] = report
	s.refMaps[id] = refMap
	return id
}

func (s *ReportStore) Get(id string) (common.AnalysisReport, *deadcode.ReferenceMap, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, ok := s.reports[id]
	if !ok {
		return common.AnalysisReport{}, nil, false
	}
	refMap := s.refMaps[id]
	return report, refMap, true
}

var store = NewReportStore()

func main() {
	http.HandleFunc("/analyze", handleAnalyze)
	http.HandleFunc("/report", handleGetReport)
	http.HandleFunc("/filter", handleFilter)
	http.HandleFunc("/references", handleReferences)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("Dead Code Detector Server starting on :8440")
	log.Fatal(http.ListenAndServe(":8440", nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		sendError(w, "No files provided", http.StatusBadRequest)
		return
	}

	filesMap := make(map[string]string)
	for _, f := range req.Files {
		filesMap[f.Path] = f.Content
	}

	analyzer, err := deadcode.NewAnalyzer(filesMap, req.PackageName)
	if err != nil {
		sendError(w, "Failed to parse files: "+err.Error(), http.StatusBadRequest)
		return
	}

	result := analyzer.Analyze()

	reportID := store.Add(result.Report, result.ReferenceMap)

	resp := common.AnalyzeResponse{
		Report:   result.Report,
		ReportID: reportID,
		Success:  true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reportID string
	if r.Method == http.MethodPost {
		var req common.GetReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		reportID = req.ReportID
	} else {
		reportID = r.URL.Query().Get("id")
	}

	if reportID == "" {
		sendError(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	report, _, ok := store.Get(reportID)
	if !ok {
		sendError(w, "Report not found", http.StatusNotFound)
		return
	}

	resp := common.GetReportResponse{
		Report:  report,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	report, _, ok := store.Get(req.ReportID)
	if !ok {
		sendError(w, "Report not found", http.StatusNotFound)
		return
	}

	filteredReport := deadcode.FilterReportByType(report, req.Filter)

	resp := common.FilterResponse{
		Report:  filteredReport,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleReferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reportID string
	var name string

	if r.Method == http.MethodPost {
		var req common.ReferencePathRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		reportID = req.ReportID
		name = req.Name
	} else {
		reportID = r.URL.Query().Get("report_id")
		name = r.URL.Query().Get("name")
	}

	if reportID == "" || name == "" {
		sendError(w, "Report ID and name are required", http.StatusBadRequest)
		return
	}

	report, refMap, ok := store.Get(reportID)
	if !ok {
		sendError(w, "Report not found", http.StatusNotFound)
		return
	}

	var declaration common.Location
	for _, entry := range report.Entries {
		if entry.Declaration.Name == name {
			declaration = entry.Declaration.Location
			break
		}
	}

	if declaration.File == "" {
		if loc, exists := deadcode.GetDeclarationLocation(name, refMap); exists {
			declaration = loc
		}
	}

	references := deadcode.GetReferences(name, refMap)

	resp := common.ReferencePathResponse{
		Name:        name,
		References:  references,
		Declaration: declaration,
		Success:     true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, status int) {
	resp := common.AnalyzeResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
