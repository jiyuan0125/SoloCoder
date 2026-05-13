package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db           *sql.DB
	jobQueue     = make(chan deduplicationJob, 100)
	jobResults   = make(map[string]deduplicationResult)
	jobResultsMu sync.RWMutex
)

type deduplicationJob struct {
	ID     string
	Texts  []string
}

type deduplicationResult struct {
	ID     string
	Result []DeduplicatedText
	Error  error
}

type DeduplicatedText struct {
	Text    string   `json:"text"`
	Indices []int    `json:"indices"`
}

type DeduplicationRequest struct {
	Texts []string `json:"texts"`
}

type DeduplicationResponse struct {
	JobID string `json:"job_id"`
}

type QueryResponse struct {
	JobID  string               `json:"job_id"`
	Status string               `json:"status"`
	Result []DeduplicatedText   `json:"result,omitempty"`
	Error  string               `json:"error,omitempty"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./dedup.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	go workerPool(4)

	http.HandleFunc("/dedup", handleDeduplication)
	http.HandleFunc("/result/", handleResult)

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8802", nil))
}

func initDB() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			result TEXT,
			error TEXT
		)
	`)
	return err
}

func workerPool(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
}

func worker() {
	for job := range jobQueue {
		result, err := processDeduplication(job.Texts)
		updateJob(job.ID, result, err)
	}
}

func processDeduplication(texts []string) ([]DeduplicatedText, error) {
	if len(texts) == 0 {
		return []DeduplicatedText{}, nil
	}

	emptyIndices := []int{}
	normalizedMap := make(map[string]*DeduplicatedText)

	for i, text := range texts {
		if isEmptyOrWhitespace(text) {
			emptyIndices = append(emptyIndices, i)
			continue
		}

		normalized := normalize(text)
		if normalized == "" {
			emptyIndices = append(emptyIndices, i)
			continue
		}

		if existing, exists := normalizedMap[normalized]; exists {
			existing.Indices = append(existing.Indices, i)
		} else {
			normalizedMap[normalized] = &DeduplicatedText{
				Text:    text,
				Indices: []int{i},
			}
		}
	}

	result := []DeduplicatedText{}
	for _, dt := range normalizedMap {
		result = append(result, *dt)
	}

	if len(emptyIndices) > 0 {
		result = append(result, DeduplicatedText{
			Text:    "",
			Indices: emptyIndices,
		})
	}

	return result, nil
}

func isEmptyOrWhitespace(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func normalize(text string) string {
	text = removeZeroWidthChars(text)
	text = fullWidthToHalfWidth(text)
	text = removeTrailingPunctuation(text)
	text = strings.ToLower(text)
	text = normalizeSpaces(text)
	return strings.TrimSpace(text)
}

var zeroWidthPattern = regexp.MustCompile(`[\x{200B}-\x{200F}\x{202A}-\x{202E}\x{2060}\x{FEFF}]`)

func removeZeroWidthChars(text string) string {
	return zeroWidthPattern.ReplaceAllString(text, "")
}

func fullWidthToHalfWidth(text string) string {
	var result []rune
	for _, r := range text {
		if r >= 0xFF01 && r <= 0xFF5E {
			r -= 0xFEE0
		} else if r == 0x3000 {
			r = 0x0020
		}
		result = append(result, r)
	}
	return string(result)
}

var trailingPunctuation = regexp.MustCompile(`[。，、；：？！.,;:!?]+$`)

func removeTrailingPunctuation(text string) string {
	return trailingPunctuation.ReplaceAllString(text, "")
}

var multipleSpaces = regexp.MustCompile(`\s+`)

func normalizeSpaces(text string) string {
	return multipleSpaces.ReplaceAllString(text, " ")
}

func handleDeduplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeduplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobID := generateJobID()
	if err := createJob(jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobQueue <- deduplicationJob{ID: jobID, Texts: req.Texts}

	resp := DeduplicationResponse{JobID: jobID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/result/")
	if jobID == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	job, err := getJob(jobID)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

func createJob(jobID string) error {
	_, err := db.Exec("INSERT INTO jobs (id, status) VALUES (?, 'processing')", jobID)
	return err
}

func updateJob(jobID string, result []DeduplicatedText, err error) {
	var resultJSON, errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else {
		resultBytes, _ := json.Marshal(result)
		resultJSON = string(resultBytes)
	}

	status := "completed"
	if err != nil {
		status = "error"
	}

	db.Exec("UPDATE jobs SET status=?, result=?, error=? WHERE id=?", status, resultJSON, errorMsg, jobID)
}

func getJob(jobID string) (*QueryResponse, error) {
	var status, resultJSON, errorMsg sql.NullString
	err := db.QueryRow("SELECT status, result, error FROM jobs WHERE id=?", jobID).Scan(&status, &resultJSON, &errorMsg)
	if err != nil {
		return nil, err
	}

	resp := &QueryResponse{
		JobID:  jobID,
		Status: status.String,
	}

	if resultJSON.Valid && resultJSON.String != "" {
		var result []DeduplicatedText
		if err := json.Unmarshal([]byte(resultJSON.String), &result); err == nil {
			resp.Result = result
		}
	}

	if errorMsg.Valid && errorMsg.String != "" {
		resp.Error = errorMsg.String
	}

	return resp, nil
}
