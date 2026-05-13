package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ColumnType int

const (
	TypeString ColumnType = iota
	TypeDate
	TypeNumber
)

type ImportTask struct {
	ID           string
	Status       string
	TotalRows    int
	ProcessedRows int
	SuccessRows  int
	FailedRows   int
	Errors       []RowError
	CreatedAt    time.Time
	mu           sync.Mutex
}

type RowError struct {
	RowNum  int
	Reason  string
	Content string
}

type TaskManager struct {
	tasks map[string]*ImportTask
	mu    sync.RWMutex
}

var taskManager = NewTaskManager()

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*ImportTask),
	}
}

func (tm *TaskManager) Create() *ImportTask {
	id := generateTaskID()
	task := &ImportTask{
		ID:        id,
		Status:    "pending",
		CreatedAt: time.Now(),
		Errors:    make([]RowError, 0),
	}
	tm.mu.Lock()
	tm.tasks[id] = task
	tm.mu.Unlock()
	return task
}

func (tm *TaskManager) Get(id string) (*ImportTask, bool) {
	tm.mu.RLock()
	task, exists := tm.tasks[id]
	tm.mu.RUnlock()
	return task, exists
}

func (t *ImportTask) Update(processed, success, failed int, err *RowError) {
	t.mu.Lock()
	t.ProcessedRows += processed
	t.SuccessRows += success
	t.FailedRows += failed
	if err != nil {
		t.Errors = append(t.Errors, *err)
	}
	t.mu.Unlock()
}

func (t *ImportTask) SetStatus(status string) {
	t.mu.Lock()
	t.Status = status
	t.mu.Unlock()
}

func (t *ImportTask) SetTotal(total int) {
	t.mu.Lock()
	t.TotalRows = total
	t.mu.Unlock()
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func detectEncoding(data []byte) string {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "UTF-8"
	}
	
	validUTF8 := true
	gbkCount := 0
	
	for i := 0; i < len(data); {
		if data[i] < 0x80 {
			i++
			continue
		}
		
		if data[i] < 0xC0 {
			validUTF8 = false
			break
		}
		
		seqLen := 0
		switch {
		case data[i] < 0xE0:
			seqLen = 2
		case data[i] < 0xF0:
			seqLen = 3
		case data[i] < 0xF8:
			seqLen = 4
		default:
			validUTF8 = false
			break
		}
		
		if validUTF8 && seqLen > 0 {
			if i+seqLen > len(data) {
				validUTF8 = false
				break
			}
			for j := 1; j < seqLen; j++ {
				if data[i+j]&0xC0 != 0x80 {
					validUTF8 = false
					break
				}
			}
		}
		
		if !validUTF8 {
			break
		}
		i += seqLen
	}
	
	if !validUTF8 {
		for i := 0; i < len(data)-1; {
			if data[i] >= 0x81 && data[i] <= 0xFE {
				if data[i+1] >= 0x40 && data[i+1] <= 0xFE && data[i+1] != 0x7F {
					gbkCount++
					i += 2
					continue
				}
			}
			i++
		}
	}
	
	if gbkCount > 0 {
		return "GBK"
	}
	return "UTF-8"
}

func gbkToUtf8(gbk []byte) ([]byte, error) {
	if len(gbk) == 0 {
		return []byte{}, nil
	}
	
	var result []byte
	i := 0
	for i < len(gbk) {
		if gbk[i] < 0x80 {
			result = append(result, gbk[i])
			i++
			continue
		}
		
		if i+1 >= len(gbk) {
			result = append(result, '?')
			i++
			continue
		}
		
		high := uint16(gbk[i])
		low := uint16(gbk[i+1])
		code := high<<8 | low
		
		if uni, ok := gbkToUnicode[code]; ok {
			if uni < 0x80 {
				result = append(result, byte(uni))
			} else if uni < 0x800 {
				result = append(result, 0xC0|byte(uni>>6), 0x80|byte(uni&0x3F))
			} else if uni < 0x10000 {
				result = append(result, 0xE0|byte(uni>>12), 0x80|byte((uni>>6)&0x3F), 0x80|byte(uni&0x3F))
			} else {
				result = append(result, '?')
			}
		} else {
			result = append(result, '?')
		}
		i += 2
	}
	return result, nil
}

func detectDelimiter(sample string) rune {
	commaCount := strings.Count(sample, ",")
	tabCount := strings.Count(sample, "\t")
	
	if tabCount > commaCount && tabCount > 0 {
		return '\t'
	}
	return ','
}

func normalizeNewlines(data []byte) []byte {
	result := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == '\r' {
			if i+1 < len(data) && data[i+1] == '\n' {
				result = append(result, '\n')
				i += 2
			} else {
				result = append(result, '\n')
				i++
			}
		} else {
			result = append(result, data[i])
			i++
		}
	}
	return result
}

func cleanField(field string) string {
	field = strings.TrimSpace(field)
	
	for len(field) >= 2 && 
		((field[0] == '"' && field[len(field)-1] == '"') ||
		 (field[0] == '\'' && field[len(field)-1] == '\'')) {
		field = field[1 : len(field)-1]
		field = strings.TrimSpace(field)
	}
	
	return field
}

var dateFormats = []string{
	"2006-01-02",
	"2006/01/02",
	"2006-1-2",
	"2006/1/2",
	"01/02/2006",
	"1/2/2006",
	"Jan 2 2006",
	"January 2 2006",
	"2 Jan 2006",
	"2 January 2006",
	"Jan-2-2006",
	"Jan/2/2006",
}

func parseDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	
	for _, format := range dateFormats {
		if t, err := time.Parse(format, value); err == nil {
			return t, true
		}
		if t, err := time.Parse(strings.ToLower(format), strings.ToLower(value)); err == nil {
			return t, true
		}
	}
	
	return time.Time{}, false
}

func isNumeric(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	
	if strings.Contains(value, ".") {
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			return true
		}
	} else {
		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			return true
		}
	}
	return false
}

func inferColumnType(samples []string) ColumnType {
	dateCount := 0
	numberCount := 0
	validCount := 0
	
	for _, s := range samples {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		validCount++
		if _, ok := parseDate(s); ok {
			dateCount++
		} else if isNumeric(s) {
			numberCount++
		}
	}
	
	if validCount == 0 {
		return TypeString
	}
	
	dateRatio := float64(dateCount) / float64(validCount)
	numberRatio := float64(numberCount) / float64(validCount)
	
	if dateRatio > 0.7 {
		return TypeDate
	} else if numberRatio > 0.7 {
		return TypeNumber
	}
	return TypeString
}

func adjustRow(row []string, expectedCols int) []string {
	if len(row) == expectedCols {
		return row
	}
	
	result := make([]string, expectedCols)
	for i := 0; i < expectedCols; i++ {
		if i < len(row) {
			result[i] = row[i]
		} else {
			result[i] = ""
		}
	}
	return result
}

func countLines(data []byte) int {
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		count++
	}
	return count
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	
	if len(data) == 0 {
		http.Error(w, "文件为空", http.StatusBadRequest)
		return
	}
	
	encoding := detectEncoding(data)
	if encoding == "GBK" {
		data, err = gbkToUtf8(data)
		if err != nil {
			http.Error(w, "Failed to convert encoding", http.StatusInternalServerError)
			return
		}
	}
	
	data = normalizeNewlines(data)
	
	totalLines := countLines(data)
	if totalLines <= 1 {
		http.Error(w, "文件为空", http.StatusBadRequest)
		return
	}
	
	sampleSize := 2000
	if len(data) < sampleSize {
		sampleSize = len(data)
	}
	sample := string(data[:sampleSize])
	delimiter := detectDelimiter(sample)
	
	task := taskManager.Create()
	task.SetStatus("processing")
	task.SetTotal(totalLines - 1)
	
	go s.processImport(task, data, delimiter)
	
	response := map[string]interface{}{
		"task_id": task.ID,
		"message": "Import started",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) processImport(task *ImportTask, data []byte, delimiter rune) {
	defer func() {
		if r := recover(); r != nil {
			task.SetStatus("failed")
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()
	
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	
	headers, err := reader.Read()
	if err != nil {
		task.SetStatus("failed")
		task.Update(0, 0, 1, &RowError{RowNum: 1, Reason: "Failed to read headers: " + err.Error()})
		return
	}
	
	for i := range headers {
		headers[i] = cleanField(headers[i])
	}
	
	numCols := len(headers)
	if numCols == 0 {
		task.SetStatus("failed")
		task.Update(0, 0, 1, &RowError{RowNum: 1, Reason: "No columns found in header"})
		return
	}
	
	rowNum := 1
	sampleRows := make([][]string, 0, 50)
	rowsToProcess := make([][]string, 0, 50)
	
	for len(sampleRows) < 50 {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rowNum++
		cleaned := make([]string, len(record))
		for i, v := range record {
			cleaned[i] = cleanField(v)
		}
		sampleRows = append(sampleRows, cleaned)
		rowsToProcess = append(rowsToProcess, cleaned)
	}
	
	columnTypes := make([]ColumnType, numCols)
	for col := 0; col < numCols; col++ {
		samples := make([]string, 0, len(sampleRows))
		for _, row := range sampleRows {
			if col < len(row) {
				samples = append(samples, row[col])
			}
		}
		columnTypes[col] = inferColumnType(samples)
	}
	
	tableName := fmt.Sprintf("import_%s", task.ID)
	
	createSQL := s.generateCreateTableSQL(tableName, headers, columnTypes)
	if err := s.executeSQL(createSQL); err != nil {
		task.SetStatus("failed")
		task.Update(0, 0, 1, &RowError{RowNum: 0, Reason: "Failed to create table: " + err.Error()})
		return
	}
	
	insertSQL := s.generateInsertSQL(tableName, headers)
	
	for _, record := range rowsToProcess {
		rowNum--
		processRow(task, record, rowNum, numCols, columnTypes, insertSQL, s)
	}
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		
		if err != nil {
			task.Update(1, 0, 1, &RowError{
				RowNum:  rowNum,
				Reason:  "Parse error: " + err.Error(),
				Content: fmt.Sprintf("%v", record),
			})
			continue
		}
		
		cleaned := make([]string, len(record))
		for i, v := range record {
			cleaned[i] = cleanField(v)
		}
		
		processRow(task, cleaned, rowNum, numCols, columnTypes, insertSQL, s)
	}
	
	task.SetStatus("completed")
}

func processRow(task *ImportTask, record []string, rowNum, numCols int, columnTypes []ColumnType, insertSQL string, s *Server) {
	row := adjustRow(record, numCols)
	
	values := make([]interface{}, numCols)
	valid := true
	
	for col := 0; col < numCols; col++ {
		value := row[col]
		
		switch columnTypes[col] {
		case TypeDate:
			if value == "" {
				values[col] = nil
			} else if t, ok := parseDate(value); ok {
				values[col] = t.Format("2006-01-02")
			} else {
				task.Update(1, 0, 1, &RowError{
					RowNum:  rowNum,
					Reason:  fmt.Sprintf("Column %d: Invalid date format: %s", col+1, value),
					Content: strings.Join(row, ","),
				})
				valid = false
				break
			}
		case TypeNumber:
			if value == "" {
				values[col] = nil
			} else if isNumeric(value) {
				values[col] = value
			} else {
				task.Update(1, 0, 1, &RowError{
					RowNum:  rowNum,
					Reason:  fmt.Sprintf("Column %d: Invalid numeric value: %s", col+1, value),
					Content: strings.Join(row, ","),
				})
				valid = false
				break
			}
		default:
			values[col] = value
		}
	}
	
	if !valid {
		return
	}
	
	if err := s.executeInsert(insertSQL, values...); err != nil {
		task.Update(1, 0, 1, &RowError{
			RowNum:  rowNum,
			Reason:  "Database error: " + err.Error(),
			Content: strings.Join(row, ","),
		})
		return
	}
	
	task.Update(1, 1, 0, nil)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	taskID := strings.TrimPrefix(r.URL.Path, "/api/status/")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}
	
	task, exists := taskManager.Get(taskID)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	task.mu.Lock()
	response := map[string]interface{}{
		"task_id":        task.ID,
		"status":         task.Status,
		"total_rows":     task.TotalRows,
		"processed_rows": task.ProcessedRows,
		"success_rows":   task.SuccessRows,
		"failed_rows":    task.FailedRows,
		"errors":         task.Errors,
		"created_at":     task.CreatedAt.Format(time.RFC3339),
	}
	task.mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := ":8800"
	server, err := NewServer("./data.db")
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		return
	}
	defer server.Close()
	
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/upload", server.handleUpload)
	http.HandleFunc("/api/status/", server.handleStatus)
	
	fmt.Printf("Server starting on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
