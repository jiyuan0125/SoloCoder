package controllers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"telemedicine/config"
	"telemedicine/models"
	"time"

	"github.com/gin-gonic/gin"
)

type StatisticsResponse struct {
	TotalConsultations      int64   `json:"total_consultations"`
	PendingConsultations    int64   `json:"pending_consultations"`
	InProgressConsultations int64   `json:"in_progress_consultations"`
	CompletedConsultations  int64   `json:"completed_consultations"`
	TotalDoctors            int64   `json:"total_doctors"`
	TotalExperts            int64   `json:"total_experts"`
	TotalPatients           int64   `json:"total_patients"`
	TotalPrescriptions      int64   `json:"total_prescriptions"`
	AverageResponseTime     float64 `json:"average_response_time_hours"`
}

func GetStatistics(c *gin.Context) {
	var stats StatisticsResponse

	config.DB.Model(&models.Consultation{}).Count(&stats.TotalConsultations)
	config.DB.Model(&models.Consultation{}).Where("status = ?", "pending").Count(&stats.PendingConsultations)
	config.DB.Model(&models.Consultation{}).Where("status = ?", "in_progress").Count(&stats.InProgressConsultations)
	config.DB.Model(&models.Consultation{}).Where("status = ?", "completed").Count(&stats.CompletedConsultations)
	config.DB.Model(&models.User{}).Where("role = ?", "grassroot").Count(&stats.TotalDoctors)
	config.DB.Model(&models.User{}).Where("role = ?", "expert").Count(&stats.TotalExperts)
	config.DB.Model(&models.Patient{}).Count(&stats.TotalPatients)
	config.DB.Model(&models.Prescription{}).Count(&stats.TotalPrescriptions)

	type ConsultationTime struct {
		CreatedAt  time.Time
		AcceptedAt *time.Time
	}

	var times []ConsultationTime
	config.DB.Model(&models.Consultation{}).
		Where("accepted_at IS NOT NULL").
		Select("created_at, accepted_at").
		Scan(&times)

	if len(times) > 0 {
		var totalDiff float64
		count := 0
		for _, t := range times {
			if t.AcceptedAt != nil {
				diff := t.AcceptedAt.Sub(t.CreatedAt).Hours()
				if diff > 0 {
					totalDiff += diff
					count++
				}
			}
		}
		if count > 0 {
			stats.AverageResponseTime = totalDiff / float64(count)
		}
	}

	c.JSON(http.StatusOK, stats)
}

func ExportByDateRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供开始和结束日期"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
		return
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
		return
	}
	end = end.AddDate(0, 0, 1).Add(-time.Second)

	var consultations []models.Consultation
	config.DB.Preload("Patient").Preload("GrassrootDoctor").Preload("ExpertDoctor").
		Preload("Prescriptions.Items").
		Where("created_at BETWEEN ? AND ?", start, end).
		Find(&consultations)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=consultations.csv")

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"问诊编号", "患者姓名", "性别", "年龄", "身份证号", "电话", "基层医生", "接诊专家", "状态", "诊断结果", "处方内容", "创建时间"})

	for _, cons := range consultations {
		expertName := "-"
		if cons.ExpertDoctor != nil {
			expertName = cons.ExpertDoctor.Name
		}

		prescriptions := []string{}
		for _, p := range cons.Prescriptions {
			items := []string{}
			for _, item := range p.Items {
				items = append(items, fmt.Sprintf("%s(%s)", item.DrugName, item.Specification))
			}
			if len(items) > 0 {
				prescriptions = append(prescriptions, strings.Join(items, "; "))
			}
		}

		row := []string{
			cons.ConsultationNo,
			cons.Patient.Name,
			cons.Patient.Gender,
			strconv.Itoa(cons.Patient.Age),
			cons.Patient.IDCard,
			cons.Patient.Phone,
			cons.GrassrootDoctor.Name,
			expertName,
			getStatusText(cons.Status),
			cons.Diagnosis,
			strings.Join(prescriptions, " | "),
			cons.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		w.Write(row)
	}

	w.Flush()
}

type ExpertStatistics struct {
	ExpertID           uint    `json:"expert_id"`
	ExpertName         string  `json:"expert_name"`
	TotalAccepted      int64   `json:"total_accepted"`
	CompletedCount     int64   `json:"completed_count"`
	CompletionRate     float64 `json:"completion_rate"`
	AvgResponseHours   float64 `json:"avg_response_hours"`
}

func ExportByExpert(c *gin.Context) {
	var experts []models.User
	config.DB.Where("role = ?", "expert").Find(&experts)

	var results []ExpertStatistics

	for _, expert := range experts {
		var consultations []models.Consultation
		config.DB.Where("expert_doctor_id = ?", expert.ID).Find(&consultations)

		accepted := int64(len(consultations))
		completed := int64(0)
		var totalResponseTime float64
		responseCount := 0

		for _, cons := range consultations {
			if cons.Status == "completed" {
				completed++
			}
			if cons.AcceptedAt != nil {
				diff := cons.AcceptedAt.Sub(cons.CreatedAt).Hours()
				if diff > 0 {
					totalResponseTime += diff
					responseCount++
				}
			}
		}

		rate := 0.0
		if accepted > 0 {
			rate = float64(completed) / float64(accepted) * 100
		}

		avgResponse := 0.0
		if responseCount > 0 {
			avgResponse = totalResponseTime / float64(responseCount)
		}

		results = append(results, ExpertStatistics{
			ExpertID:         expert.ID,
			ExpertName:       expert.Name,
			TotalAccepted:    accepted,
			CompletedCount:   completed,
			CompletionRate:   rate,
			AvgResponseHours: avgResponse,
		})
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=expert_statistics.csv")

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"专家ID", "专家姓名", "接诊总数", "完成数量", "完成率(%)", "平均响应时间(小时)"})

	for _, r := range results {
		row := []string{
			strconv.Itoa(int(r.ExpertID)),
			r.ExpertName,
			strconv.Itoa(int(r.TotalAccepted)),
			strconv.Itoa(int(r.CompletedCount)),
			fmt.Sprintf("%.2f", r.CompletionRate),
			fmt.Sprintf("%.2f", r.AvgResponseHours),
		}
		w.Write(row)
	}

	w.Flush()
}

func getStatusText(status string) string {
	switch status {
	case "pending":
		return "待接诊"
	case "in_progress":
		return "问诊中"
	case "pending_supplement":
		return "待补充"
	case "completed":
		return "已完成"
	default:
		return status
	}
}
