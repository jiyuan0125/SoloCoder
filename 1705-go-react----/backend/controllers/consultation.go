package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"telemedicine/config"
	"telemedicine/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateConsultationRequest struct {
	PatientName    string     `json:"patient_name" binding:"required"`
	PatientGender  string     `json:"patient_gender" binding:"required"`
	PatientAge     int        `json:"patient_age" binding:"required"`
	PatientIDCard  string     `json:"patient_id_card" binding:"required"`
	PatientPhone   string     `json:"patient_phone" binding:"required"`
	ChiefComplaint string     `json:"chief_complaint" binding:"required"`
	PastHistory    string     `json:"past_history"`
	Exams          []ExamItem `json:"exams"`
}

type ExamItem struct {
	ExamType   string     `json:"exam_type" binding:"required"`
	ExamResult string     `json:"exam_result" binding:"required"`
	ExamDate   *time.Time `json:"exam_date"`
}

func generateConsultationNo() string {
	now := time.Now()
	return fmt.Sprintf("CS%s%s", now.Format("20060102150405"), fmt.Sprintf("%04d", now.Nanosecond()/1000000))
}

func CreateConsultation(c *gin.Context) {
	userID := c.GetUint("userID")

	var req CreateConsultationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len([]rune(req.ChiefComplaint)) < 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "主诉症状描述至少50字"})
		return
	}

	tx := config.DB.Begin()

	var patient models.Patient
	err := tx.Where("id_card = ?", req.PatientIDCard).First(&patient).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			patient = models.Patient{
				Name:   req.PatientName,
				Gender: req.PatientGender,
				Age:    req.PatientAge,
				IDCard: req.PatientIDCard,
				Phone:  req.PatientPhone,
			}
			if err := tx.Create(&patient).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建患者信息失败"})
				return
			}
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询患者信息失败"})
			return
		}
	}

	var existingCount int64
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := tx.Model(&models.Consultation{}).
		Where("patient_id = ? AND created_at >= ? AND status != ?", patient.ID, startOfDay, "completed").
		Count(&existingCount).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询问诊记录失败"})
		return
	}

	if existingCount > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "该患者今日已有问诊记录"})
		return
	}

	consultation := models.Consultation{
		ConsultationNo:    generateConsultationNo(),
		PatientID:         patient.ID,
		GrassrootDoctorID: userID,
		ChiefComplaint:    req.ChiefComplaint,
		PastHistory:       req.PastHistory,
		Status:            "pending",
		Version:           0,
	}

	if err := tx.Create(&consultation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建问诊失败"})
		return
	}

	for _, exam := range req.Exams {
		e := models.Exam{
			ConsultationID: consultation.ID,
			ExamType:       exam.ExamType,
			ExamResult:     exam.ExamResult,
			ExamDate:       exam.ExamDate,
		}
		if err := tx.Create(&e).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加检查记录失败"})
			return
		}
	}

	tx.Commit()

	var result models.Consultation
	config.DB.Preload("Patient").Preload("GrassrootDoctor").First(&result, consultation.ID)
	c.JSON(http.StatusCreated, result)
}

func ListConsultations(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")
	status := c.Query("status")

	query := config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor")

	switch userRole {
	case "grassroot":
		query = query.Where("grassroot_doctor_id = ?", userID)
	case "expert":
		query = query.Where("expert_doctor_id = ?", userID)
	}

	if status != "" {
		statusList := strings.Split(status, ",")
		query = query.Where("status IN ?", statusList)
	}

	var consultations []models.Consultation
	query.Order("created_at DESC").Find(&consultations)
	c.JSON(http.StatusOK, consultations)
}

func GetConsultationPool(c *gin.Context) {
	userRole := c.GetString("userRole")
	if userRole != "expert" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有专家可以查看候诊池"})
		return
	}

	var consultations []models.Consultation
	config.DB.Preload("Patient").Preload("GrassrootDoctor").
		Where("status = ?", "pending").
		Order("created_at ASC").
		Find(&consultations)
	c.JSON(http.StatusOK, consultations)
}

func GetConsultation(c *gin.Context) {
	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor").
		Preload("Exams").Preload("ImageDatas.Template").
		First(&consultation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	if userRole == "grassroot" && consultation.GrassrootDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看"})
		return
	}
	if userRole == "expert" && consultation.ExpertDoctorID != nil && *consultation.ExpertDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看"})
		return
	}

	c.JSON(http.StatusOK, consultation)
}

func AcceptConsultation(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")
	userName := c.GetString("userName")

	if userRole != "expert" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有专家可以接诊"})
		return
	}

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	tx := config.DB.Begin()

	var consultation models.Consultation
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&consultation, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if consultation.Status != "pending" {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "该问诊已被其他专家接诊"})
		return
	}

	now := time.Now()
	result := tx.Model(&consultation).
		Where("id = ? AND status = ? AND version = ?", id, "pending", consultation.Version).
		Updates(map[string]interface{}{
			"status":           "in_progress",
			"expert_doctor_id": userID,
			"accepted_at":      &now,
			"version":          consultation.Version + 1,
		})

	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "接诊失败"})
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "该问诊已被其他专家接诊"})
		return
	}

	tx.Commit()

	message := models.Message{
		ConsultationID: consultation.ID,
		SenderID:       userID,
		SenderRole:     "expert",
		SenderName:     userName,
		Content:        "专家已接诊，问诊开始。",
	}
	config.DB.Create(&message)

	config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor").First(&consultation, id)
	c.JSON(http.StatusOK, consultation)
}

type CompleteRequest struct {
	Diagnosis          string `json:"diagnosis" binding:"required"`
	PrescriptionAdvice string `json:"prescription_advice" binding:"required"`
	TreatmentAdvice    string `json:"treatment_advice"`
}

func CompleteConsultation(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	if userRole != "expert" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有专家可以完成问诊"})
		return
	}

	var req CompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	tx := config.DB.Begin()

	var consultation models.Consultation
	if err := tx.First(&consultation, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if consultation.ExpertDoctorID == nil || *consultation.ExpertDoctorID != userID {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
		return
	}

	if consultation.Status == "completed" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "问诊已完成，不能重复提交"})
		return
	}

	if consultation.Status != "in_progress" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许完成问诊"})
		return
	}

	now := time.Now()
	if err := tx.Model(&consultation).Updates(map[string]interface{}{
		"status":             "completed",
		"diagnosis":          req.Diagnosis,
		"prescription_advice": req.PrescriptionAdvice,
		"treatment_advice":   req.TreatmentAdvice,
		"completed_at":       &now,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	tx.Commit()

	config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor").First(&consultation, id)
	c.JSON(http.StatusOK, consultation)
}

func RequestSupplement(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	if userRole != "expert" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有专家可以要求补充资料"})
		return
	}

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.First(&consultation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if consultation.ExpertDoctorID == nil || *consultation.ExpertDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
		return
	}

	if consultation.Status != "in_progress" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许要求补充资料"})
		return
	}

	config.DB.Model(&consultation).Update("status", "pending_supplement")

	config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor").First(&consultation, id)
	c.JSON(http.StatusOK, consultation)
}

type SupplementRequest struct {
	Exams []ExamItem `json:"exams"`
}

func SubmitSupplement(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	if userRole != "grassroot" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有基层医生可以补充资料"})
		return
	}

	var req SupplementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if len(req.Exams) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须补充至少一项资料"})
		return
	}

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	tx := config.DB.Begin()

	var consultation models.Consultation
	if err := tx.First(&consultation, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if consultation.GrassrootDoctorID != userID {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
		return
	}

	if consultation.Status != "pending_supplement" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前状态不允许补充资料"})
		return
	}

	for _, exam := range req.Exams {
		e := models.Exam{
			ConsultationID: consultation.ID,
			ExamType:       exam.ExamType,
			ExamResult:     exam.ExamResult,
			ExamDate:       exam.ExamDate,
		}
		if err := tx.Create(&e).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加检查记录失败"})
			return
		}
	}

	if err := tx.Model(&consultation).Update("status", "in_progress").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}

	tx.Commit()

	config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor").First(&consultation, id)
	c.JSON(http.StatusOK, consultation)
}

func DeleteConsultation(c *gin.Context) {
	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	var consultation models.Consultation
	if err := config.DB.First(&consultation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNoContent, gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if userRole == "grassroot" && consultation.GrassrootDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此问诊"})
		return
	}
	if userRole == "expert" && consultation.ExpertDoctorID != nil && *consultation.ExpertDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此问诊"})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Where("consultation_id = ?", id).Delete(&models.Message{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除消息失败"})
		return
	}

	if err := tx.Where("consultation_id = ?", id).Delete(&models.Exam{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除检查记录失败"})
		return
	}

	if err := tx.Where("consultation_id = ?", id).Delete(&models.ImageData{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除影像数据失败"})
		return
	}

	var prescriptions []models.Prescription
	if err := tx.Where("consultation_id = ?", id).Find(&prescriptions).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询处方失败"})
		return
	}

	for _, p := range prescriptions {
		if err := tx.Where("prescription_id = ?", p.ID).Delete(&models.PrescriptionItem{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除处方项目失败"})
			return
		}
	}

	if err := tx.Where("consultation_id = ?", id).Delete(&models.Prescription{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除处方失败"})
		return
	}

	if err := tx.Delete(&consultation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除问诊失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusNoContent, gin.H{})
}

func AddExam(c *gin.Context) {
	userID := c.GetUint("userID")
	userRole := c.GetString("userRole")

	id, err := parseInt(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问诊ID"})
		return
	}

	var req ExamItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var consultation models.Consultation
	if err := config.DB.First(&consultation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "问诊记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if userRole == "grassroot" && consultation.GrassrootDoctorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此问诊"})
		return
	}

	exam := models.Exam{
		ConsultationID: consultation.ID,
		ExamType:       req.ExamType,
		ExamResult:     req.ExamResult,
		ExamDate:       req.ExamDate,
	}

	if err := config.DB.Create(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加检查记录失败"})
		return
	}

	c.JSON(http.StatusCreated, exam)
}
