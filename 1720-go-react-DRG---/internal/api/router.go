package api

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"drg-system/internal/model"
	"drg-system/internal/repository"
	"drg-system/internal/service"
)

type Router struct {
	*mux.Router
	groupingService   *service.GroupingService
	paymentService    *service.PaymentService
	settlementService *service.SettlementService
	todoService       *service.TodoService
	repo              repository.Repository
}

func NewRouter(
	groupingService *service.GroupingService,
	paymentService *service.PaymentService,
	settlementService *service.SettlementService,
	todoService *service.TodoService,
	repo repository.Repository,
) *Router {
	r := &Router{
		Router:            mux.NewRouter(),
		groupingService:   groupingService,
		paymentService:    paymentService,
		settlementService: settlementService,
		todoService:       todoService,
		repo:              repo,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	r.Use(corsMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/health", r.handleHealth).Methods("GET")

	drg := api.PathPrefix("/drg-groups").Subrouter()
	drg.HandleFunc("", r.handleListDRGGroups).Methods("GET")
	drg.HandleFunc("", r.handleCreateDRGGroup).Methods("POST")
	drg.HandleFunc("/{code}", r.handleGetDRGGroup).Methods("GET")
	drg.HandleFunc("/{code}", r.handleUpdateDRGGroup).Methods("PUT")
	drg.HandleFunc("/{code}", r.handleDeleteDRGGroup).Methods("DELETE")

	api.HandleFunc("/grouping-rules", r.handleListGroupingRules).Methods("GET")
	api.HandleFunc("/grouping-rules", r.handleCreateGroupingRule).Methods("POST")

	settlements := api.PathPrefix("/settlements").Subrouter()
	settlements.HandleFunc("", r.handleListSettlements).Methods("GET")
	settlements.HandleFunc("", r.handleCreateSettlement).Methods("POST")
	settlements.HandleFunc("/{id}", r.handleGetSettlement).Methods("GET")
	settlements.HandleFunc("/{id}/submit", r.handleSubmitSettlement).Methods("POST")
	settlements.HandleFunc("/{id}/approve", r.handleApproveSettlement).Methods("POST")
	settlements.HandleFunc("/{id}/reject", r.handleRejectSettlement).Methods("POST")
	settlements.HandleFunc("/{id}/publish", r.handlePublishSettlement).Methods("POST")
	settlements.HandleFunc("/{id}/withdraw", r.handleWithdrawSettlement).Methods("POST")
	settlements.HandleFunc("/{id}/recalculate", r.handleRecalculateSettlement).Methods("POST")
	settlements.HandleFunc("/{id}/export", r.handleExportSettlement).Methods("GET")

	api.HandleFunc("/payment-params", r.handleListPaymentParams).Methods("GET")
	api.HandleFunc("/payment-params", r.handleCreatePaymentParam).Methods("POST")
	api.HandleFunc("/payment-params/active", r.handleGetActivePaymentParam).Methods("GET")

	api.HandleFunc("/todos", r.handleListTodos).Methods("GET")
	api.HandleFunc("/todos/{id}/complete", r.handleCompleteTodo).Methods("POST")

	api.HandleFunc("/statistics", r.handleGetStatistics).Methods("GET")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string, details string) {
	writeJSON(w, status, model.APIError{
		Code:    status,
		Message: message,
		Details: details,
	})
}

func decodeBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleListDRGGroups(w http.ResponseWriter, req *http.Request) {
	groups := r.groupingService.ListDRGGroups()
	writeJSON(w, http.StatusOK, groups)
}

func (r *Router) handleCreateDRGGroup(w http.ResponseWriter, req *http.Request) {
	var group model.DRGGroup
	if err := decodeBody(req, &group); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	created, err := r.groupingService.CreateDRGGroup(group)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, "DRG组编号已存在", group.Code)
			return
		}
		writeError(w, http.StatusBadRequest, "创建失败", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *Router) handleGetDRGGroup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	code := vars["code"]
	group, err := r.groupingService.GetDRGGroup(code)
	if err != nil {
		writeError(w, http.StatusNotFound, "DRG组不存在", code)
		return
	}
	writeJSON(w, http.StatusOK, group)
}

func (r *Router) handleUpdateDRGGroup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	code := vars["code"]

	var group model.DRGGroup
	if err := decodeBody(req, &group); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	updated, err := r.groupingService.UpdateDRGGroup(code, group)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "DRG组不存在", code)
			return
		}
		writeError(w, http.StatusBadRequest, "更新失败", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (r *Router) handleDeleteDRGGroup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	code := vars["code"]

	if err := r.groupingService.DeleteDRGGroup(code); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "DRG组不存在", code)
			return
		}
		writeError(w, http.StatusInternalServerError, "删除失败", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) handleListGroupingRules(w http.ResponseWriter, req *http.Request) {
	rules := r.groupingService.ListGroupingRules()
	writeJSON(w, http.StatusOK, rules)
}

func (r *Router) handleCreateGroupingRule(w http.ResponseWriter, req *http.Request) {
	var rule model.GroupingRule
	if err := decodeBody(req, &rule); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}
	rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	created, err := r.groupingService.CreateGroupingRule(rule)
	if err != nil {
		writeError(w, http.StatusBadRequest, "创建失败", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *Router) handleListSettlements(w http.ResponseWriter, req *http.Request) {
	settlements := r.settlementService.ListSettlements()
	writeJSON(w, http.StatusOK, settlements)
}

type CreateSettlementRequest struct {
	HospitalID string               `json:"hospital_id"`
	Period     string               `json:"period"`
	Records    []model.MedicalRecord `json:"records"`
}

func (r *Router) handleCreateSettlement(w http.ResponseWriter, req *http.Request) {
	var reqBody CreateSettlementRequest
	if err := decodeBody(req, &reqBody); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	settlement, err := r.settlementService.CreateSettlement(reqBody.HospitalID, reqBody.Period, reqBody.Records)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			writeError(w, http.StatusConflict, "病案号重复", "")
			return
		}
		writeError(w, http.StatusBadRequest, "创建结算失败", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, settlement)
}

func (r *Router) handleGetSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.GetSettlement(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "结算不存在", id)
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

func (r *Router) handleSubmitSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.SubmitSettlement(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "提交失败", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

func (r *Router) handleApproveSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.ApproveSettlement(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "审核失败", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

type RejectRequest struct {
	Comment string `json:"comment"`
}

func (r *Router) handleRejectSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]

	var reqBody RejectRequest
	if err := decodeBody(req, &reqBody); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	settlement, err := r.settlementService.RejectSettlement(id, reqBody.Comment)
	if err != nil {
		writeError(w, http.StatusBadRequest, "审核失败", err.Error())
		return
	}

	r.todoService.CreateModifyTodo(settlement.HospitalID, id, reqBody.Comment)
	writeError(w, http.StatusBadRequest, "审核不通过", reqBody.Comment)
}

func (r *Router) handlePublishSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.PublishSettlement(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "发布失败", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

func (r *Router) handleWithdrawSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.WithdrawSettlement(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "撤回失败", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

func (r *Router) handleRecalculateSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.RecalculateSettlement(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "重新计算失败", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settlement)
}

func centsToYuan(cents int64) string {
	yuan := float64(cents) / 100.0
	return fmt.Sprintf("%.2f", yuan)
}

func (r *Router) handleExportSettlement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	settlement, err := r.settlementService.GetSettlement(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "结算不存在", id)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=settlement_%s.csv", settlement.Period))
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"DRG组编号", "DRG组名称", "病案数量", "实际总费用(元)", "效率指数", "标准支付(元)", "调整后支付(元)"})
	for _, summary := range settlement.GroupSummaries {
		writer.Write([]string{
			summary.DRGGroupCode,
			summary.DRGGroupName,
			strconv.Itoa(summary.CaseCount),
			centsToYuan(summary.TotalActualCost),
			fmt.Sprintf("%.4f", summary.EfficiencyIndex),
			centsToYuan(summary.StandardPayment),
			centsToYuan(summary.AdjustedPayment),
		})
	}
	writer.Write([]string{})
	writer.Write([]string{"汇总"})
	writer.Write([]string{"实际总费用(元)", centsToYuan(settlement.TotalActualCost)})
	writer.Write([]string{"应支付总额(元)", centsToYuan(settlement.TotalPayment)})
	writer.Write([]string{"差额(元)", centsToYuan(settlement.Difference)})
}

func (r *Router) handleListPaymentParams(w http.ResponseWriter, req *http.Request) {
	params := r.paymentService.ListParams()
	writeJSON(w, http.StatusOK, params)
}

func (r *Router) handleCreatePaymentParam(w http.ResponseWriter, req *http.Request) {
	var param model.PaymentParam
	if err := decodeBody(req, &param); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}
	param.ID = fmt.Sprintf("p_%d", time.Now().UnixNano())
	param.EffectiveAt = time.Now()
	created, err := r.paymentService.CreateParam(param)
	if err != nil {
		writeError(w, http.StatusBadRequest, "创建失败", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (r *Router) handleGetActivePaymentParam(w http.ResponseWriter, req *http.Request) {
	param, err := r.paymentService.GetActiveParam()
	if err != nil {
		writeError(w, http.StatusNotFound, "未找到活跃的支付参数", "")
		return
	}
	writeJSON(w, http.StatusOK, param)
}

func (r *Router) handleListTodos(w http.ResponseWriter, req *http.Request) {
	todos := r.todoService.ListTodos()
	writeJSON(w, http.StatusOK, todos)
}

func (r *Router) handleCompleteTodo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	todo, err := r.todoService.CompleteTodo(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "待办不存在", id)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (r *Router) handleGetStatistics(w http.ResponseWriter, req *http.Request) {
	stats := r.settlementService.GetStatistics()
	writeJSON(w, http.StatusOK, stats)
}
