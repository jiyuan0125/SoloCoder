package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"epidemic-management/models"
	"epidemic-management/services"
)

type rawJSON map[string]interface{}

func getStr(m rawJSON, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func getUint(m rawJSON, keys ...string) uint {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch val := v.(type) {
			case float64:
				return uint(val)
			case int:
				return uint(val)
			case uint:
				return val
			}
		}
	}
	return 0
}

func getInt(m rawJSON, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch val := v.(type) {
			case float64:
				return int(val)
			case int:
				return val
			}
		}
	}
	return 0
}

func toSnake(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			sb.WriteByte('_')
		}
		sb.WriteByte(byte(r | 32))
	}
	return sb.String()
}

func CreateOutbreakReport(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}

	var raw rawJSON
	if err = json.Unmarshal(body, &raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON格式错误"})
		return
	}

	patientName := getStr(raw, "patientName", "patient_name")
	patientID := getStr(raw, "patientId", "patient_id", "patientID")
	diseaseName := getStr(raw, "diseaseName", "disease_name")
	region := getStr(raw, "region")
	district := getStr(raw, "district")
	reportTimeStr := getStr(raw, "reportTime", "report_time")
	onsetTimeStr := getStr(raw, "onsetTime", "onset_time")

	if patientName == "" || patientID == "" || diseaseName == "" || region == "" || district == "" || reportTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必填字段"})
		return
	}

	report, err := parseTime(reportTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "报告时间格式错误"})
		return
	}

	onset := report
	if onsetTimeStr != "" {
		onset, err = parseTime(onsetTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "发病时间格式错误"})
			return
		}
	}

	r := &models.OutbreakReport{
		PatientName: patientName,
		PatientID:   patientID,
		DiseaseName: diseaseName,
		Region:      region,
		District:    district,
		OnsetTime:   onset,
		ReportTime:  report,
	}

	if err = services.CreateOutbreakReport(r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, r)
}

func ListOutbreakReports(c *gin.Context) {
	reports, err := services.ListOutbreakReports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func GetActiveAlerts(c *gin.Context) {
	alerts, err := services.GetActiveAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

type CreateContactRequest struct {
	InvestigationID  uint   `json:"investigationId"`
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	ContactType      string `json:"contactType"`
	FirstContactDate string `json:"firstContactDate"`
	LastContactDate  string `json:"lastContactDate"`
	DiseaseName      string `json:"diseaseName"`
}

func CreateContact(c *gin.Context) {
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	first, err := parseDate(req.FirstContactDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "首次接触日期格式错误"})
		return
	}
	last, err := parseDate(req.LastContactDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "最后接触日期格式错误"})
		return
	}

	contact := &models.Contact{
		InvestigationID:  req.InvestigationID,
		Name:             req.Name,
		Phone:            req.Phone,
		ContactType:      models.ContactType(req.ContactType),
		FirstContactDate: first,
		LastContactDate:  last,
	}

	if err = services.CreateContact(contact, req.DiseaseName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, contact)
}

func ListContacts(c *gin.Context) {
	invIDStr := c.Query("investigationId")
	if invIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少调查ID"})
		return
	}
	invID, _ := strconv.ParseUint(invIDStr, 10, 64)
	contacts, err := services.ListContactsByInvestigation(uint(invID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

type UpdateContactStatusRequest struct {
	Status string `json:"status"`
}

func UpdateContactStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var req UpdateContactStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateContactStatus(uint(id), models.ContactStatus(req.Status)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}

func CreateInvestigation(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return
	}

	var raw rawJSON
	if err = json.Unmarshal(body, &raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON格式错误"})
		return
	}

	reportID := getUint(raw, "reportId", "report_id", "reportID")
	patientName := getStr(raw, "patientName", "patient_name")
	onsetTimeStr := getStr(raw, "onsetTime", "onset_time")
	visitTimeStr := getStr(raw, "visitTime", "visit_time")
	confirmTimeStr := getStr(raw, "confirmTime", "confirm_time")
	clinical := getStr(raw, "clinical")

	inv := &models.Investigation{
		ReportID:    reportID,
		PatientName: patientName,
		Clinical:    clinical,
	}

	if onsetTimeStr != "" {
		if t, err := parseTime(onsetTimeStr); err == nil {
			inv.OnsetTime = t
		}
	}
	if visitTimeStr != "" {
		if t, err := parseTime(visitTimeStr); err == nil {
			inv.VisitTime = t
		}
	}
	if confirmTimeStr != "" {
		if t, err := parseTime(confirmTimeStr); err == nil {
			inv.ConfirmTime = t
		}
	}

	if err := services.CreateInvestigation(inv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inv)
}

func ListInvestigations(c *gin.Context) {
	invs, err := services.ListInvestigations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invs)
}

type CreateVaccineRequest struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	BatchNumber  string `json:"batchNumber"`
	ExpiryDate   string `json:"expiryDate"`
	Schedule     string `json:"schedule"`
	Doses        int    `json:"doses"`
	IntervalDays string `json:"intervalDays"`
}

func CreateVaccine(c *gin.Context) {
	var req CreateVaccineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	expiry, _ := parseDate(req.ExpiryDate)

	v := &models.Vaccine{
		Name:         req.Name,
		Manufacturer: req.Manufacturer,
		BatchNumber:  req.BatchNumber,
		ExpiryDate:   expiry,
		Schedule:     req.Schedule,
		Doses:        req.Doses,
		IntervalDays: req.IntervalDays,
	}

	if err := services.CreateVaccine(v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, v)
}

func ListVaccines(c *gin.Context) {
	vaccines, err := services.ListVaccines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vaccines)
}

type RecordVaccinationRequest struct {
	RecipientName   string `json:"recipientName"`
	RecipientID     string `json:"recipientId"`
	VaccineName     string `json:"vaccineName"`
	DoseNumber      int    `json:"doseNumber"`
	VaccinationDate string `json:"vaccinationDate"`
	Unit            string `json:"unit"`
	Doctor          string `json:"doctor"`
}

func RecordVaccination(c *gin.Context) {
	var req RecordVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vacDate, err := parseDate(req.VaccinationDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "接种日期格式错误"})
		return
	}

	record := &models.VaccinationRecord{
		RecipientName:   req.RecipientName,
		RecipientID:     req.RecipientID,
		VaccineName:     req.VaccineName,
		DoseNumber:      req.DoseNumber,
		VaccinationDate: vacDate,
		Unit:            req.Unit,
		Doctor:          req.Doctor,
	}

	if err = services.RecordVaccination(record); err != nil {
		if err.Error() == "该剂次已接种" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "超出允许时间范围" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func ListVaccinationRecords(c *gin.Context) {
	recipientID := c.Query("recipientId")
	var records []models.VaccinationRecord
	var err error
	if recipientID != "" {
		records, err = services.ListVaccinationRecords(recipientID)
	} else {
		records, err = services.ListAllVaccinationRecords()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func ListTodos(c *gin.Context) {
	todos, err := services.ListTodos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todos)
}

type UpdateTodoRequest struct {
	Status string `json:"status"`
}

func UpdateTodoStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateTodoStatus(uint(id), models.TodoStatus(req.Status)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}

func GetStatistics(c *gin.Context) {
	stats, err := services.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func ListWeeklyReports(c *gin.Context) {
	reports, err := services.ListWeeklyReports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func GenerateWeeklyReport(c *gin.Context) {
	report, err := services.GenerateWeeklyReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, report)
}

func ListDiseases(c *gin.Context) {
	diseases, err := services.ListDiseases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, diseases)
}
