package main

import (
	"encoding/json"
	"net/http"

	"repair-platform/common"
	"repair-platform/core"
)

type App struct {
	service *core.Service
}

func newApp() *App {
	return &App{
		service: core.NewService(),
	}
}

func (a *App) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (a *App) writeError(w http.ResponseWriter, status int, msg string) {
	a.writeJSON(w, status, common.ErrorResponse{Error: msg})
}

func (a *App) handleCreateRepair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.CreateRepairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	repairReq, diagnosis, err := a.service.CreateRepairRequest(req.UserID, req.Address, req.Area, req.Appliance, req.FaultDesc)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.writeJSON(w, http.StatusOK, common.CreateRepairResponse{
		Request:   *repairReq,
		Diagnosis: diagnosis,
	})
}

func (a *App) handleListRepairs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	a.writeJSON(w, http.StatusOK, common.ListRepairRequestsResponse{Requests: a.service.ListRepairRequests()})
}

func (a *App) handleAssignTech(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.AssignTechnicianRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	techID, err := a.service.AssignTechnician(req.RequestID)
	if err != nil {
		a.writeJSON(w, http.StatusOK, common.AssignTechnicianResponse{Success: false, Message: err.Error()})
		return
	}
	a.writeJSON(w, http.StatusOK, common.AssignTechnicianResponse{Success: true, TechID: techID})
}

func (a *App) handleApplySpare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.ApplySpareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	result, err := a.service.ApplySpareParts(req.RequestID, req.TechID, req.SpareParts)
	if err != nil {
		a.writeJSON(w, http.StatusOK, common.ApplySpareResponse{Success: false, Message: err.Error()})
		return
	}

	appliedRecords := make([]common.SparePartOutboundRecord, 0, len(result.AppliedRecords))
	for _, r := range result.AppliedRecords {
		appliedRecords = append(appliedRecords, *r)
	}

	a.writeJSON(w, http.StatusOK, common.ApplySpareResponse{
		Success:        len(appliedRecords) > 0 || len(result.PurchaseOrders) > 0,
		AppliedParts:   appliedRecords,
		FailedItems:    result.FailedCodes,
		PurchaseOrders: result.PurchaseOrders,
	})
}

func (a *App) handleCompleteRepair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.CompleteRepairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	result, err := a.service.CompleteRepair(req.RequestID, req.TechID, req.FaultCause, req.PartsUsed, req.PartsReturn)
	if err != nil {
		a.writeJSON(w, http.StatusOK, common.CompleteRepairResponse{Success: false, Message: err.Error()})
		return
	}

	a.writeJSON(w, http.StatusOK, common.CompleteRepairResponse{Success: true, Report: result.Report})
}

func (a *App) handleAddSpare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.AddSparePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}
	if err := a.service.AddSparePart(req.Part); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleListSpares(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	a.writeJSON(w, http.StatusOK, common.GetSparePartsResponse{Parts: a.service.ListSpareParts()})
}

func (a *App) handleUpdateSparePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.UpdateSparePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}
	if err := a.service.UpdateSparePrice(req.Code, req.NewPrice); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleListPOs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	a.writeJSON(w, http.StatusOK, common.GetPurchaseOrdersResponse{Orders: a.service.ListPurchaseOrders()})
}

func (a *App) handleAddTech(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	var req common.AddTechnicianRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}
	if err := a.service.AddTechnician(req.Tech); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleListTechs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}
	a.writeJSON(w, http.StatusOK, common.GetTechniciansResponse{Techs: a.service.ListTechnicians()})
}
