package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"recycling/api"
	"recycling/core"
)

var service *core.RecyclingService

func main() {
	service = core.NewRecyclingService()

	port := getPort()

	http.HandleFunc("/api/categories", handleCategories)
	http.HandleFunc("/api/categories/update", handleUpdatePrice)
	http.HandleFunc("/api/records", handleRecords)
	http.HandleFunc("/api/records/create", handleCreateRecord)
	http.HandleFunc("/api/export/details", handleExportDetails)
	http.HandleFunc("/api/export/summary", handleExportSummary)

	fmt.Printf("服务端已启动，监听端口: %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func getPort() string {
	var port string
	flag.StringVar(&port, "port", "", "监听端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}

	if port == "" {
		port = "9016"
	}

	return port
}

func handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	categories := service.Categories.GetAll()
	writeJSON(w, categories)
}

func handleUpdatePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req api.UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if success := service.Categories.UpdatePrice(req.CategoryID, req.NewPrice); !success {
		http.Error(w, "分类不存在", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	records := service.Records.GetAll()
	writeJSON(w, records)
}

func handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req api.CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求", http.StatusBadRequest)
		return
	}

	record, err := service.CreateRecord(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, record)
}

func handleExportDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	categoryID := r.URL.Query().Get("category_id")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "无效的开始日期", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "无效的结束日期", http.StatusBadRequest)
		return
	}

	endDate = endDate.Add(24 * time.Hour)
	records := service.Records.GetRecordsByDateRange(startDate, endDate)
	csvData := core.GenerateCSV(records, categoryID)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="records_%s_%s.csv"`, startDateStr, endDateStr))
	w.Write(csvData)
}

func handleExportSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "无效的年份", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		http.Error(w, "无效的月份", http.StatusBadRequest)
		return
	}

	records := service.Records.GetRecordsByMonth(year, month)
	prevYear, prevMonth := core.GetPrevMonth(year, month)
	prevRecords := service.Records.GetRecordsByMonth(prevYear, prevMonth)

	summary := core.CalculateSummary(records, prevRecords)
	csvData := core.GenerateSummaryCSV(summary, year, month)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="summary_%d_%02d.csv"`, year, month))
	w.Write(csvData)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
