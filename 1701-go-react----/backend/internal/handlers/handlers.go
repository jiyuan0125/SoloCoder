package handlers

import (
	"cmms/internal/models"
	"cmms/internal/storage"
	"cmms/internal/ws"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	Store     *storage.Storage
	WSManager *ws.Manager
}

func NewHandler(store *storage.Storage, wsManager *ws.Manager) *Handler {
	return &Handler{
		Store:     store,
		WSManager: wsManager,
	}
}

type RegisterRequest struct {
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone" binding:"required"`
	ChiefComplaint string `json:"chief_complaint" binding:"required"`
}

func (h *Handler) RegisterPatient(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patientID := uuid.New().String()
	patient := &models.Patient{
		ID:             patientID,
		Name:           req.Name,
		Phone:          req.Phone,
		ChiefComplaint: req.ChiefComplaint,
		CreatedAt:      time.Now(),
	}
	h.Store.SavePatient(patient)

	serialNum, err := h.Store.GetSerialNum()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成挂号序号失败"})
		return
	}

	reg := &models.Registration{
		ID:        uuid.New().String(),
		PatientID: patientID,
		SerialNum: serialNum,
		Date:      time.Now().Format("20060102"),
		Status:    models.StatusWaiting,
		CreatedAt: time.Now(),
	}
	h.Store.SaveRegistration(reg)

	c.JSON(http.StatusCreated, gin.H{
		"registration": reg,
		"patient":      patient,
	})
}

func (h *Handler) GetTodayRegistrations(c *gin.Context) {
	regs := h.Store.GetTodayRegistrations()

	sort.Slice(regs, func(i, j int) bool {
		return regs[i].CreatedAt.Before(regs[j].CreatedAt)
	})

	result := make([]gin.H, 0)
	for _, reg := range regs {
		patient, _ := h.Store.GetPatient(reg.PatientID)
		result = append(result, gin.H{
			"id":              reg.ID,
			"serial_num":      reg.SerialNum,
			"patient_name":    patient.Name,
			"patient_phone":   patient.Phone,
			"chief_complaint": patient.ChiefComplaint,
			"status":          reg.Status,
			"created_at":      reg.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"registrations": result})
}

type CallRequest struct {
	RegistrationID string `json:"registration_id" binding:"required"`
	DoctorName     string `json:"doctor_name" binding:"required"`
}

func (h *Handler) CallPatient(c *gin.Context) {
	var req CallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reg, ok := h.Store.GetRegistration(req.RegistrationID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "挂号记录不存在"})
		return
	}

	if reg.Status != models.StatusWaiting {
		c.JSON(http.StatusBadRequest, gin.H{"error": "患者不在候诊中"})
		return
	}

	h.Store.UpdateRegistrationStatus(reg.ID, models.StatusInVisit)

	ws.BroadcastCallNumber(h.WSManager, reg.SerialNum, req.DoctorName)

	c.JSON(http.StatusOK, gin.H{
		"serial_num":  reg.SerialNum,
		"doctor_name": req.DoctorName,
	})
}

type CreatePrescriptionRequest struct {
	RegistrationID string                    `json:"registration_id" binding:"required"`
	Diagnosis      string                    `json:"diagnosis" binding:"required"`
	Syndrome       string                    `json:"syndrome" binding:"required"`
	Items          []models.PrescriptionItem `json:"items" binding:"required"`
}

func (h *Handler) CreatePrescription(c *gin.Context) {
	var req CreatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reg, ok := h.Store.GetRegistration(req.RegistrationID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "挂号记录不存在"})
		return
	}

	patient, _ := h.Store.GetPatient(reg.PatientID)

	if err := h.validatePrescriptionItems(req.Items); err != nil {
		c.JSON(err.Code, gin.H{"error": err.Message})
		return
	}

	adjustedItems, warnings := h.adjustDosages(req.Items)

	var totalDosage float64
	for _, item := range adjustedItems {
		if item.Unit == "克" || item.Unit == "g" {
			totalDosage += item.Dosage
		}
	}

	if totalDosage > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "总剂量不能超过500克"})
		return
	}

	pr := &models.Prescription{
		ID:             uuid.New().String(),
		RegistrationID: req.RegistrationID,
		PatientID:      reg.PatientID,
		PatientName:    patient.Name,
		Diagnosis:      req.Diagnosis,
		Syndrome:       req.Syndrome,
		Items:          adjustedItems,
		TotalDosage:    totalDosage,
		Status:         models.PrescriptionDraft,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	h.Store.SavePrescription(pr)

	response := gin.H{
		"prescription": pr,
	}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	c.JSON(http.StatusCreated, response)
}

type UpdatePrescriptionRequest struct {
	Diagnosis string                    `json:"diagnosis"`
	Syndrome  string                    `json:"syndrome"`
	Items     []models.PrescriptionItem `json:"items"`
}

func (h *Handler) UpdatePrescription(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePrescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, ok := h.Store.GetPrescription(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
		return
	}

	if pr.Status != models.PrescriptionDraft {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "非草稿状态不能修改"})
		return
	}

	if req.Items != nil {
		if err := h.validatePrescriptionItems(req.Items); err != nil {
			c.JSON(err.Code, gin.H{"error": err.Message})
			return
		}
	}

	adjustedItems, warnings := h.adjustDosages(req.Items)
	var totalDosage float64
	for _, item := range adjustedItems {
		if item.Unit == "克" || item.Unit == "g" {
			totalDosage += item.Dosage
		}
	}
	if totalDosage > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "总剂量不能超过500克"})
		return
	}

	if req.Diagnosis != "" {
		pr.Diagnosis = req.Diagnosis
	}
	if req.Syndrome != "" {
		pr.Syndrome = req.Syndrome
	}
	if req.Items != nil {
		pr.Items = adjustedItems
		pr.TotalDosage = totalDosage
	}
	pr.UpdatedAt = time.Now()

	h.Store.SavePrescription(&pr)

	response := gin.H{
		"prescription": pr,
	}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ConfirmPrescription(c *gin.Context) {
	id := c.Param("id")
	pr, ok := h.Store.GetPrescription(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
		return
	}

	if pr.Status != models.PrescriptionDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态流转非法"})
		return
	}

	pr.Status = models.PrescriptionConfirmed
	pr.UpdatedAt = time.Now()
	h.Store.SavePrescription(&pr)

	c.JSON(http.StatusOK, pr)
}

func (h *Handler) VoidPrescription(c *gin.Context) {
	id := c.Param("id")
	pr, ok := h.Store.GetPrescription(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
		return
	}

	pr.StatusReason = "已作废"
	pr.UpdatedAt = time.Now()
	h.Store.SavePrescription(&pr)

	c.JSON(http.StatusOK, gin.H{
		"status":         pr.Status,
		"status_reason":  pr.StatusReason,
		"message":        "已作废",
	})
}

type DispenseRequest struct {
	DispensingMode     string `json:"dispensing_mode" binding:"required"`
	DecoctionMachineID string `json:"decoction_machine_id,omitempty"`
	PotCount           int    `json:"pot_count,omitempty"`
	MinutesFromNow     int    `json:"minutes_from_now,omitempty"`
}

func (h *Handler) DispensePrescription(c *gin.Context) {
	id := c.Param("id")
	var req DispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pr, ok := h.Store.GetPrescription(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "处方不存在"})
		return
	}

	if pr.StatusReason == "已作废" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "处方已作废"})
		return
	}

	if pr.Status != models.PrescriptionConfirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态流转非法"})
		return
	}

	if req.DispensingMode != "self" && req.DispensingMode != "decoction" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "发放模式只能是self或decoction"})
		return
	}

	if req.DispensingMode == "decoction" {
		if req.DecoctionMachineID == "" || req.PotCount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "代煎需要提供煎药机编号和锅数"})
			return
		}
	}

	success, missing := h.Store.DeductStock(pr.Items)
	if !success {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "药材库存不足",
			"missing_herbs": missing,
		})
		return
	}

	pr.Status = models.PrescriptionDispensed
	pr.DispensingMode = req.DispensingMode
	if req.DispensingMode == "decoction" {
		pr.DecoctionMachineID = req.DecoctionMachineID
		pr.PotCount = req.PotCount
		minutes := 60
		if req.MinutesFromNow > 0 {
			minutes = req.MinutesFromNow
		}
		finishTime := time.Now().Add(time.Duration(minutes) * time.Minute)
		pr.EstimatedFinish = &finishTime
	}
	pr.UpdatedAt = time.Now()
	h.Store.SavePrescription(&pr)

	c.JSON(http.StatusOK, pr)
}

func (h *Handler) GetPrescriptions(c *gin.Context) {
	prescriptions := h.Store.GetAllPrescriptions()
	sort.Slice(prescriptions, func(i, j int) bool {
		return prescriptions[i].CreatedAt.After(prescriptions[j].CreatedAt)
	})
	c.JSON(http.StatusOK, gin.H{"prescriptions": prescriptions})
}

func (h *Handler) GetPrescriptionsByRegistration(c *gin.Context) {
	regID := c.Param("registration_id")
	prescriptions := h.Store.GetPrescriptionsByRegistration(regID)
	sort.Slice(prescriptions, func(i, j int) bool {
		return prescriptions[i].CreatedAt.After(prescriptions[j].CreatedAt)
	})
	c.JSON(http.StatusOK, gin.H{"prescriptions": prescriptions})
}

func (h *Handler) GetHerbs(c *gin.Context) {
	herbs := h.Store.GetAllHerbs()
	now := time.Now()

	result := make([]gin.H, 0)
	for _, herb := range herbs {
		daysUntilExpiry := int(herb.ExpiryDate.Sub(now).Hours() / 24)
		status := "normal"
		if daysUntilExpiry < 0 {
			status = "expired"
		} else if daysUntilExpiry <= 30 {
			status = "warning"
		}

		result = append(result, gin.H{
			"name":             herb.Name,
			"stock":            herb.Stock,
			"expiry_date":      herb.ExpiryDate.Format("2006-01-02"),
			"days_until_expiry": daysUntilExpiry,
			"status":           status,
			"price":            herb.Price,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})

	c.JSON(http.StatusOK, gin.H{"herbs": result})
}

func (h *Handler) WebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		ID:   uuid.New().String(),
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	h.WSManager.Register <- client

	go client.Write()
	client.Read(h.WSManager)
}

type ValidationError struct {
	Code    int
	Message string
}

func (h *Handler) validatePrescriptionItems(items []models.PrescriptionItem) *ValidationError {
	if len(items) == 0 {
		return &ValidationError{Code: http.StatusBadRequest, Message: "处方不能为空"}
	}

	if len(items) > 20 {
		return &ValidationError{Code: http.StatusBadRequest, Message: "处方最多20味药"}
	}

	seen := make(map[string]bool)
	for _, item := range items {
		if seen[item.HerbName] {
			return &ValidationError{Code: http.StatusConflict, Message: fmt.Sprintf("药材%s重复", item.HerbName)}
		}
		seen[item.HerbName] = true

		if _, ok := h.Store.GetHerb(item.HerbName); !ok {
			return &ValidationError{Code: http.StatusNotFound, Message: fmt.Sprintf("药材%s不存在", item.HerbName)}
		}

		herb, _ := h.Store.GetHerb(item.HerbName)
		if time.Now().After(herb.ExpiryDate) {
			return &ValidationError{Code: http.StatusBadRequest, Message: fmt.Sprintf("药材%s已过期", item.HerbName)}
		}
	}

	return nil
}

func (h *Handler) adjustDosages(items []models.PrescriptionItem) ([]models.PrescriptionItem, []string) {
	var warnings []string
	result := make([]models.PrescriptionItem, len(items))

	for i, item := range items {
		result[i] = item

		if (item.Unit == "克" || item.Unit == "g") && item.Dosage > 0 {
			if math.Mod(item.Dosage, 5) != 0 {
				adjusted := math.Ceil(item.Dosage/5) * 5
				result[i].Dosage = adjusted
				warnings = append(warnings, fmt.Sprintf("%s已自动调整为%d克", item.HerbName, int(adjusted)))
			}
		}
	}

	return result, warnings
}

type UpdateHerbRequest struct {
	Name  string  `json:"name" binding:"required"`
	Stock float64 `json:"stock"`
	Price float64 `json:"price"`
}

func (h *Handler) UpdateHerb(c *gin.Context) {
	var req UpdateHerbRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	herb, ok := h.Store.GetHerb(req.Name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "药材不存在"})
		return
	}

	oldPrice := herb.Price

	if req.Stock >= 0 {
		herb.Stock = req.Stock
	}
	if req.Price >= 0 {
		herb.Price = req.Price
	}

	h.Store.SaveHerb(herb)

	if req.Price >= 0 && math.Abs(oldPrice-req.Price) > 0.0001 {
		h.updatePendingPrescriptionsPrice(herb.Name, req.Price)
	}

	c.JSON(http.StatusOK, gin.H{
		"herb": herb,
		"message": "药材信息已更新",
	})
}

func (h *Handler) updatePendingPrescriptionsPrice(herbName string, newPrice float64) {
	allPrescriptions := h.Store.GetAllPrescriptions()
	for _, pr := range allPrescriptions {
		if pr.Status != models.PrescriptionDispensed && pr.StatusReason != "已作废" {
			updated := false
			for _, item := range pr.Items {
				if item.HerbName == herbName {
					updated = true
				}
			}
			if updated {
				pr.UpdatedAt = time.Now()
				h.Store.SavePrescription(&pr)
			}
		}
	}
}

func (h *Handler) FinishVisit(c *gin.Context) {
	id := c.Param("id")
	reg, ok := h.Store.GetRegistration(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "挂号记录不存在"})
		return
	}

	h.Store.UpdateRegistrationStatus(reg.ID, models.StatusDone)

	c.JSON(http.StatusOK, gin.H{"message": "就诊完成"})
}
