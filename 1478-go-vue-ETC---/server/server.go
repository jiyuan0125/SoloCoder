package main

import (
	"encoding/json"
	"etc-system/common"
	"etc-system/pkg/etccore"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	storage *etccore.Storage
	mux     *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		storage: etccore.NewStorage(),
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/account/create", s.handleCreateAccount)
	s.mux.HandleFunc("/api/account/get", s.handleGetAccount)
	s.mux.HandleFunc("/api/account/recharge", s.handleRecharge)
	s.mux.HandleFunc("/api/pass", s.handlePass)
	s.mux.HandleFunc("/api/export", s.handleExport)
}

func (s *Server) Start(addr string) {
	fmt.Printf("ETC 管理系统服务端已启动，监听端口 %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, s.mux))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{
			Success: false,
			Message: "仅支持 POST 请求",
		})
		return
	}

	var req common.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "请求参数解析失败: " + err.Error(),
		})
		return
	}

	account, err := s.storage.CreateAccount(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.CreateAccountResponse{
		Success: true,
		Message: "账户创建成功",
		Account: account,
	})
}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{
			Success: false,
			Message: "仅支持 GET 请求",
		})
		return
	}

	licensePlate := r.URL.Query().Get("license_plate")
	if licensePlate == "" {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "缺少车牌号参数",
		})
		return
	}

	account, err := s.storage.GetAccount(licensePlate)
	if err != nil {
		writeJSON(w, http.StatusNotFound, common.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.GetAccountResponse{
		Success: true,
		Account: account,
	})
}

func (s *Server) handleRecharge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{
			Success: false,
			Message: "仅支持 POST 请求",
		})
		return
	}

	var req common.RechargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "请求参数解析失败: " + err.Error(),
		})
		return
	}

	account, err := s.storage.Recharge(req.LicensePlate, req.Amount)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.RechargeResponse{
		Success: true,
		Message: "充值成功",
		Balance: account.Balance,
	})
}

func (s *Server) handlePass(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{
			Success: false,
			Message: "仅支持 POST 请求",
		})
		return
	}

	var req common.PassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "请求参数解析失败: " + err.Error(),
		})
		return
	}

	record, err := s.storage.ProcessPass(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.PassResponse{
		Success:       true,
		Message:       "通行处理完成",
		Record:        record,
		PaymentStatus: record.PaymentStatus,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{
			Success: false,
			Message: "仅支持 GET 请求",
		})
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "缺少开始日期或结束日期参数",
		})
		return
	}

	startDate, err := parseDate(startDateStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "开始日期格式错误，请使用 YYYY-MM-DD 格式",
		})
		return
	}

	endDate, err := parseDate(endDateStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{
			Success: false,
			Message: "结束日期格式错误，请使用 YYYY-MM-DD 格式",
		})
		return
	}

	records := s.storage.GetRecordsInRange(startDate, endDate)
	csvData := etccore.ExportToCSV(records)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=pass_records.csv")
	w.Write([]byte(csvData))
}
