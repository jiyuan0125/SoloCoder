package server

import (
	"encoding/json"
	"net/http"
	"property-management/pkg/api"
	"property-management/pkg/core/announcement"
	"property-management/pkg/core/models"
	"property-management/pkg/core/payment"
	"property-management/pkg/core/repair"
	"property-management/pkg/core/user"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	userService        *user.Service
	repairService      *repair.Service
	paymentService     *payment.Service
	announcementService *announcement.Service
}

func NewHandler(userSvc *user.Service, repairSvc *repair.Service, paymentSvc *payment.Service, announcementSvc *announcement.Service) *Handler {
	return &Handler{
		userService:        userSvc,
		repairService:      repairSvc,
		paymentService:     paymentSvc,
		announcementService: announcementSvc,
	}
}

func writeJSON(w http.ResponseWriter, code int, response *api.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, &api.Response{
		Code:    code,
		Message: message,
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, &api.Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string      `json:"username"`
		Name     string      `json:"name"`
		Role     models.Role `json:"role"`
		Building string      `json:"building"`
		Unit     string      `json:"unit"`
		Phone    string      `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.Create(req.Username, req.Name, req.Role, req.Building, req.Unit, req.Phone)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, user)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/users/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, user)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users := h.userService.ListAll()
	writeSuccess(w, users)
}

func (h *Handler) CreateRepair(w http.ResponseWriter, r *http.Request) {
	var req api.CreateRepairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repair, err := h.repairService.Create(req.OwnerID, req.RepairType, req.Description, req.ExpectedTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.repairService.IsPublicFacility(req.RepairType) {
		now := time.Now()
		_, _ = h.paymentService.CreateRepairBill(repair.ID, 0, req.OwnerID, int(now.Month()), now.Year())
	}

	writeSuccess(w, repair)
}

func (h *Handler) AssignRepair(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/repairs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing repair id")
		return
	}

	var req api.AssignRepairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repair, err := h.repairService.Assign(id, req.SupervisorID, req.StaffID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, repair)
}

func (h *Handler) StartRepair(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/repairs/")
	staffID := r.URL.Query().Get("staff_id")
	if id == "" || staffID == "" {
		writeError(w, http.StatusBadRequest, "missing repair id or staff id")
		return
	}

	repair, err := h.repairService.StartProcess(id, staffID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, repair)
}

func (h *Handler) CompleteRepair(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/repairs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing repair id")
		return
	}

	var req api.CompleteRepairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repair, err := h.repairService.Complete(id, req.StaffID, req.Result, req.ResultImageURLs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, repair)
}

func (h *Handler) ConfirmRepair(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/repairs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing repair id")
		return
	}

	var req api.ConfirmRepairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repair, err := h.repairService.Confirm(id, req.OwnerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, repair)
}

func (h *Handler) GetRepair(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/repairs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing repair id")
		return
	}

	repair, err := h.repairService.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, repair)
}

func (h *Handler) ListRepairs(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner_id")
	status := r.URL.Query().Get("status")

	var repairs []*models.Repair
	if ownerID != "" {
		repairs = h.repairService.ListByOwner(ownerID)
	} else if status != "" {
		repairs = h.repairService.ListByStatus(models.RepairStatus(status))
	} else {
		repairs = h.repairService.ListAll()
	}

	writeSuccess(w, &api.ListRepairsResponse{Repairs: repairs})
}

func (h *Handler) CreateBill(w http.ResponseWriter, r *http.Request) {
	var req api.CreateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bill, err := h.paymentService.CreateBill(req.UserID, req.BillType, req.Amount, req.DueDate, req.Month, req.Year)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, bill)
}

func (h *Handler) GenerateMonthlyBills(w http.ResponseWriter, r *http.Request) {
	var req api.GenerateMonthlyBillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bills, err := h.paymentService.GenerateMonthlyBills(req.BillType, req.Amount, req.DueDate, req.Month, req.Year)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, &api.ListBillsResponse{Bills: bills})
}

func (h *Handler) PayBill(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/bills/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing bill id")
		return
	}

	var req api.PayBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payment, err := h.paymentService.PayBill(id, req.UserID, req.Amount, req.PaymentMethod)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, payment)
}

func (h *Handler) GetBill(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/bills/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing bill id")
		return
	}

	bill, err := h.paymentService.GetBillByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, bill)
}

func (h *Handler) ListBills(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	status := r.URL.Query().Get("status")

	var bills []*models.Bill
	if userID != "" {
		bills = h.paymentService.ListBillsByUser(userID)
	} else if status != "" {
		bills = h.paymentService.ListBillsByStatus(models.BillStatus(status))
	} else {
		bills = h.paymentService.ListAllBills()
	}

	writeSuccess(w, &api.ListBillsResponse{Bills: bills})
}

func (h *Handler) CalculatePenalty(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/bills/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing bill id")
		return
	}

	penalty, days, err := h.paymentService.CalculatePenalty(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, &api.CalculatePenaltyResponse{
		PenaltyAmount: penalty,
		OverdueDays:   days,
	})
}

func (h *Handler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var req api.CreateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	announcement, err := h.announcementService.Create(req.Title, req.Content, req.Scope, req.TargetBuilding, req.EffectiveTime, req.ExpiryTime, req.CreatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, announcement)
}

func (h *Handler) UpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/announcements/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing announcement id")
		return
	}

	var req api.UpdateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	announcement, err := h.announcementService.Update(id, req.Title, req.Content, req.Scope, req.TargetBuilding, req.EffectiveTime, req.ExpiryTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, announcement)
}

func (h *Handler) DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/announcements/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing announcement id")
		return
	}

	if err := h.announcementService.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, nil)
}

func (h *Handler) GetAnnouncement(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path, "/announcements/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing announcement id")
		return
	}

	announcement, err := h.announcementService.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, announcement)
}

func (h *Handler) ListAnnouncements(w http.ResponseWriter, r *http.Request) {
	building := r.URL.Query().Get("building")
	validOnly := r.URL.Query().Get("valid_only")

	var announcements []*models.Announcement
	if validOnly == "true" {
		announcements = h.announcementService.ListValid(building)
	} else {
		announcements = h.announcementService.ListAll()
	}

	writeSuccess(w, &api.ListAnnouncementsResponse{Announcements: announcements})
}

func getIDFromPath(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) > 0 {
		id := parts[0]
		if _, err := uuid.Parse(id); err == nil {
			return id
		}
		if _, err := strconv.Atoi(id); err == nil {
			return id
		}
	}
	return ""
}
