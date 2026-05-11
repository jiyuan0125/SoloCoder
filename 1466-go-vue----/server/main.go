package main

import (
	"encoding/json"
	"flag"
	"laboratory/common"
	"laboratory/core"
	"log"
	"net/http"
	"os"
	"strings"
)

var store = core.NewStore()

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("LAB_PORT")
	}
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/samples", handleSamples)
	http.HandleFunc("/samples/", handleSampleByID)
	http.HandleFunc("/samples/assign", handleAssignCabinet)
	http.HandleFunc("/samples/status", handleUpdateStatus)
	http.HandleFunc("/samples/claim", handleClaimSample)
	http.HandleFunc("/samples/overdue", handleOverdue)

	http.HandleFunc("/cabinets", handleCabinets)
	http.HandleFunc("/test-items", handleTestItems)
	http.HandleFunc("/test-items/assign", handleAssignTestItems)
	http.HandleFunc("/test-results", handleTestResults)
	http.HandleFunc("/test-results/", handleSampleTests)

	http.HandleFunc("/retest", handleRetest)
	http.HandleFunc("/retest/approve", handleApproveRetest)

	http.HandleFunc("/reports", handleReports)
	http.HandleFunc("/reports/issue", handleIssueReport)
	http.HandleFunc("/reports/void", handleVoidReport)

	log.Printf("服务端启动，监听端口: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, code int, resp common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func handleSamples(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateSampleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
			return
		}
		result, err := store.CreateSample(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	case http.MethodGet:
		status := common.SampleStatus(r.URL.Query().Get("status"))
		result := store.ListSamples(status)
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
	}
}

func handleSampleByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "路径错误"})
		return
	}

	sampleID := parts[2]
	result, err := store.GetSample(sampleID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.Response{Code: 404, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})
}

func handleAssignCabinet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.AssignCabinetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	if err := store.AssignCabinet(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "分配成功"})
}

func handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.UpdateSampleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	if err := store.UpdateSampleStatus(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "状态更新成功"})
}

func handleClaimSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.ClaimSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	if err := store.ClaimSample(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "领取成功"})
}

func handleOverdue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	result := store.GetOverdueSamples()
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})
}

func handleCabinets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateCabinetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
			return
		}
		if err := store.CreateCabinet(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "创建成功"})

	case http.MethodGet:
		result := store.ListCabinets()
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
	}
}

func handleTestItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.CreateTestItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
			return
		}
		result, err := store.CreateTestItem(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	case http.MethodGet:
		result := store.ListTestItems()
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
	}
}

func handleAssignTestItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.AssignTestItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	if err := store.AssignTestItems(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "分配成功"})
}

func handleTestResults(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.RecordTestResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
			return
		}

		if err := store.RecordTestResult(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "录入成功"})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
	}
}

func handleSampleTests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "路径错误"})
		return
	}

	sampleID := parts[3]
	result, err := store.GetSampleTests(sampleID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.Response{Code: 404, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})
}

func handleRetest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.ApplyRetestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
			return
		}
		result, err := store.ApplyRetest(&req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	case http.MethodGet:
		result := store.ListRetestRequests()
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
	}
}

func handleApproveRetest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.ApproveRetestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	if err := store.ApproveRetest(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "审批完成"})
}

func handleReports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req common.GenerateReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
			return
		}
		result, err := store.GenerateReport(req.SampleID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})

	case http.MethodGet:
		reportID := r.URL.Query().Get("id")
		if reportID != "" {
			result, err := store.GetReport(reportID)
			if err != nil {
				writeJSON(w, http.StatusNotFound, common.Response{Code: 404, Message: err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})
		} else {
			result := store.ListReports()
			writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: result})
		}

	default:
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
	}
}

func handleIssueReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.IssueReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	if err := store.IssueReport(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Message: "签发成功"})
}

func handleVoidReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Code: 405, Message: "方法不支持"})
		return
	}

	var req common.VoidReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: "参数错误"})
		return
	}

	newReport, err := store.VoidReport(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Code: 400, Message: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.Response{Code: 0, Data: newReport, Message: "作废成功，已生成新报告草稿"})
}
