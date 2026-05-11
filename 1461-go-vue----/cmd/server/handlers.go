package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"quality-trace/pkg/common"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, common.Response{
		Success: false,
		Error:   err.Error(),
	})
}

func (s *server) handleCreateBatch(w http.ResponseWriter, r *http.Request) {
	var req common.CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	batch, err := s.batchSvc.CreateBatch(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    batch,
	})
}

func (s *server) handleGetBatch(w http.ResponseWriter, r *http.Request) {
	batchID := r.URL.Query().Get("batch_id")
	if batchID == "" {
		writeError(w, http.StatusBadRequest, ErrMissingBatchID)
		return
	}

	batch, err := s.batchSvc.GetBatch(batchID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    batch,
	})
}

func (s *server) handleCompleteBatch(w http.ResponseWriter, r *http.Request) {
	var req common.CompleteBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	batch, err := s.batchSvc.CompleteBatch(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    batch,
	})
}

func (s *server) handleListBatches(w http.ResponseWriter, r *http.Request) {
	var filter common.ListBatchesRequest

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = common.BatchStatus(status)
	}
	if productName := r.URL.Query().Get("product_name"); productName != "" {
		filter.ProductName = productName
	}

	batches := s.batchSvc.ListBatches(filter)

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    batches,
	})
}

func (s *server) handleAddProcessFlow(w http.ResponseWriter, r *http.Request) {
	var req common.AddProcessFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	flow, err := s.processSvc.AddProcessFlow(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    flow,
	})
}

func (s *server) handleGetProcessFlow(w http.ResponseWriter, r *http.Request) {
	productName := r.URL.Query().Get("product_name")
	if productName == "" {
		writeError(w, http.StatusBadRequest, ErrMissingProductName)
		return
	}

	flow, err := s.processSvc.GetProcessFlow(productName)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    flow,
	})
}

func (s *server) handleStartProcess(w http.ResponseWriter, r *http.Request) {
	var req common.StartProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	record, err := s.processSvc.StartProcess(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    record,
	})
}

func (s *server) handleCompleteProcess(w http.ResponseWriter, r *http.Request) {
	var req common.CompleteProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	record, err := s.processSvc.CompleteProcess(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    record,
	})
}

func (s *server) handleReworkDecision(w http.ResponseWriter, r *http.Request) {
	var req common.ReworkDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	batch, err := s.processSvc.MakeReworkDecision(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    batch,
	})
}

func (s *server) handleListProcessRecords(w http.ResponseWriter, r *http.Request) {
	batchID := r.URL.Query().Get("batch_id")

	records := s.processSvc.ListProcessRecords(batchID)

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    records,
	})
}

func (s *server) handleAddInspectionSpec(w http.ResponseWriter, r *http.Request) {
	var req common.AddInspectionSpecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	spec, err := s.inspectionSvc.AddInspectionSpec(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    spec,
	})
}

func (s *server) handleAddInspectionRecord(w http.ResponseWriter, r *http.Request) {
	var req common.AddInspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	record, err := s.inspectionSvc.AddInspectionRecord(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    record,
	})
}

func (s *server) handleListInspectionRecords(w http.ResponseWriter, r *http.Request) {
	var filter common.ListInspectionRecordsRequest

	if batchID := r.URL.Query().Get("batch_id"); batchID != "" {
		filter.BatchID = batchID
	}
	if inspectionType := r.URL.Query().Get("inspection_type"); inspectionType != "" {
		filter.InspectionType = common.InspectionType(inspectionType)
	}

	records := s.inspectionSvc.ListInspectionRecords(filter)

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    records,
	})
}

func (s *server) handleExportBatches(w http.ResponseWriter, r *http.Request) {
	var filter common.ListBatchesRequest
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = common.BatchStatus(status)
	}
	if productName := r.URL.Query().Get("product_name"); productName != "" {
		filter.ProductName = productName
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=batches.csv")

	err := s.exportSvc.ExportBatches(w, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleExportProcessRecords(w http.ResponseWriter, r *http.Request) {
	batchID := r.URL.Query().Get("batch_id")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=process_records.csv")

	err := s.exportSvc.ExportProcessRecords(w, batchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleExportInspectionRecords(w http.ResponseWriter, r *http.Request) {
	var filter common.ListInspectionRecordsRequest
	if batchID := r.URL.Query().Get("batch_id"); batchID != "" {
		filter.BatchID = batchID
	}
	if inspectionType := r.URL.Query().Get("inspection_type"); inspectionType != "" {
		filter.InspectionType = common.InspectionType(inspectionType)
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=inspection_records.csv")

	err := s.exportSvc.ExportInspectionRecords(w, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			days = 30
		}
	}

	metrics, err := s.aggregationSvc.GetMetricsSummary(days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    metrics,
	})
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Message: "Quality Trace Server is running",
		Data: map[string]interface{}{
			"time": time.Now().Format(time.RFC3339),
		},
	})
}
