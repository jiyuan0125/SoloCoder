package main

import (
	"errors"
	"net/http"

	"quality-trace/pkg/core/aggregation"
	"quality-trace/pkg/core/batch"
	"quality-trace/pkg/core/export"
	"quality-trace/pkg/core/inspection"
	"quality-trace/pkg/core/store"
	"quality-trace/pkg/core/process"
)

var (
	ErrMissingBatchID      = errors.New("missing batch_id parameter")
	ErrMissingProductName  = errors.New("missing product_name parameter")
)

type server struct {
	store          *store.Store
	batchSvc       *batch.Service
	processSvc     *process.Service
	inspectionSvc  *inspection.Service
	exportSvc      *export.Service
	aggregationSvc *aggregation.Service
}

func newServer() *server {
	s := store.New()
	batchSvc := batch.NewService(s)
	processSvc := process.NewService(s, batchSvc)
	inspectionSvc := inspection.NewService(s, batchSvc)
	exportSvc := export.NewService(batchSvc, processSvc, inspectionSvc)
	aggregationSvc := aggregation.NewService(batchSvc, processSvc, inspectionSvc)

	return &server{
		store:          s,
		batchSvc:       batchSvc,
		processSvc:     processSvc,
		inspectionSvc:  inspectionSvc,
		exportSvc:      exportSvc,
		aggregationSvc: aggregationSvc,
	}
}

func (s *server) router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)

	mux.HandleFunc("/api/batches", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleListBatches(w, r)
		case http.MethodPost:
			s.handleCreateBatch(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/batches/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleGetBatch(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/batches/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleCompleteBatch(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/process-flows", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetProcessFlow(w, r)
		case http.MethodPost:
			s.handleAddProcessFlow(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/processes/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleStartProcess(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/processes/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleCompleteProcess(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/processes/decision", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleReworkDecision(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/processes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleListProcessRecords(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspection-specs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleAddInspectionSpec(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/inspections", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleListInspectionRecords(w, r)
		case http.MethodPost:
			s.handleAddInspectionRecord(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/export/batches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleExportBatches(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/export/processes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleExportProcessRecords(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/export/inspections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleExportInspectionRecords(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleGetMetrics(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
