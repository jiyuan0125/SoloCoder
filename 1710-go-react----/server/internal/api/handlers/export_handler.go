package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"trial-management-system/internal/models"
	"trial-management-system/internal/services"
	"trial-management-system/internal/utils"
)

func ExportData(c *gin.Context) {
	protocolIDStr := c.Query("protocol_id")
	siteIDStr := c.Query("site_id")
	exportType := c.Query("type")

	if protocolIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要提供protocol_id"})
		return
	}

	protocolID, err := uuid.Parse(protocolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方案ID"})
		return
	}

	var siteID uuid.UUID
	if siteIDStr != "" {
		siteID, err = uuid.Parse(siteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的中心ID"})
			return
		}
	}

	protocol, err := services.GetProtocolByID(protocolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
		return
	}

	subjects, err := services.GetSubjectsWithProtocolSite(protocolID, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	siteMap := make(map[uuid.UUID]string)
	for _, s := range protocol.Sites {
		siteMap[s.ID] = s.SiteCode + " - " + s.SiteName
	}

	visitMap := make(map[uuid.UUID]string)
	for _, v := range protocol.Visits {
		visitMap[v.ID] = v.VisitName
	}

	if exportType == "sae" {
		exportSAE(c, protocol, subjects)
		return
	}

	if exportType == "ae" {
		exportAE(c, protocol, subjects)
		return
	}

	exportSubjects(c, protocol, subjects, siteMap, visitMap)
}

func exportSubjects(c *gin.Context, protocol *models.Protocol, subjects []models.Subject, siteMap map[uuid.UUID]string, visitMap map[uuid.UUID]string) {
	filename := fmt.Sprintf("subjects_%s_%s.csv", protocol.ProtocolNumber, time.Now().Format("20060102"))

	rows := [][]string{
		{"方案编号", protocol.ProtocolNumber},
		{"药物名称", protocol.DrugName},
		{"", ""},
		{"筛选编号", "随机编号", "姓名缩写", "性别", "出生日期", "入组日期", "状态", "中心", "访视记录", "不良事件数"},
	}

	for _, s := range subjects {
		visitCount := len(s.Records)
		aeCount := len(s.AEs)

		enrollmentDate := ""
		if s.EnrollmentDate != nil {
			enrollmentDate = utils.FormatDate(*s.EnrollmentDate)
		}

		rows = append(rows, []string{
			s.ScreeningNumber,
			s.RandomizationID,
			s.NameInitials,
			s.Gender,
			utils.FormatDate(s.BirthDate),
			enrollmentDate,
			string(s.Status),
			siteMap[s.SiteID],
			fmt.Sprintf("%d", visitCount),
			fmt.Sprintf("%d", aeCount),
		})
	}

	utils.WriteCSVResponse(c, filename, rows)
}

func exportAE(c *gin.Context, protocol *models.Protocol, subjects []models.Subject) {
	filename := fmt.Sprintf("ae_%s_%s.csv", protocol.ProtocolNumber, time.Now().Format("20060102"))

	subjectMap := make(map[uuid.UUID]*models.Subject)
	for i := range subjects {
		subjectMap[subjects[i].ID] = &subjects[i]
	}

	aes, err := services.ListAdverseEventsByProtocol(protocol.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows := [][]string{
		{"方案编号", protocol.ProtocolNumber},
		{"", ""},
		{"受试者编号", "事件名称", "开始日期", "结束日期", "严重程度", "与药物关系", "是否SAE", "处理措施", "转归", "超窗"},
	}

	for _, ae := range aes {
		subject := subjectMap[ae.SubjectID]
		if subject == nil {
			continue
		}

		endDate := ""
		if ae.EndDate != nil {
			endDate = utils.FormatDate(*ae.EndDate)
		}

		isSAE := "否"
		if ae.IsSAE {
			isSAE = "是"
		}

		rows = append(rows, []string{
			subject.RandomizationID,
			ae.EventName,
			utils.FormatDate(ae.StartDate),
			endDate,
			string(ae.Severity),
			string(ae.Relationship),
			isSAE,
			ae.Treatment,
			string(ae.Outcome),
			"",
		})
	}

	utils.WriteCSVResponse(c, filename, rows)
}

func exportSAE(c *gin.Context, protocol *models.Protocol, subjects []models.Subject) {
	filename := fmt.Sprintf("sae_%s_%s.csv", protocol.ProtocolNumber, time.Now().Format("20060102"))

	subjectMap := make(map[uuid.UUID]*models.Subject)
	for i := range subjects {
		subjectMap[subjects[i].ID] = &subjects[i]
	}

	aes, err := services.ListSAEByProtocol(protocol.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows := [][]string{
		{"方案编号", protocol.ProtocolNumber},
		{"SAE报告", ""},
		{"受试者编号", "事件名称", "开始日期", "严重程度", "与药物关系", "处理措施", "转归", "报告状态", "报告截止日期"},
	}

	for _, ae := range aes {
		subject := subjectMap[ae.SubjectID]
		if subject == nil {
			continue
		}

		reportStatus := "未提交"
		if ae.ReportSubmitted {
			reportStatus = "已提交"
		}

		deadline := ""
		if ae.ReportDeadline != nil {
			deadline = ae.ReportDeadline.Format("2006-01-02 15:04")
		}

		rows = append(rows, []string{
			subject.RandomizationID,
			ae.EventName,
			utils.FormatDate(ae.StartDate),
			string(ae.Severity),
			string(ae.Relationship),
			ae.Treatment,
			string(ae.Outcome),
			reportStatus,
			deadline,
		})
	}

	utils.WriteCSVResponse(c, filename, rows)
}

func ExportVisitData(c *gin.Context) {
	protocolIDStr := c.Query("protocol_id")
	siteIDStr := c.Query("site_id")

	if protocolIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要提供protocol_id"})
		return
	}

	protocolID, err := uuid.Parse(protocolIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方案ID"})
		return
	}

	var siteID uuid.UUID
	if siteIDStr != "" {
		siteID, err = uuid.Parse(siteIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的中心ID"})
			return
		}
	}

	protocol, err := services.GetProtocolByID(protocolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "方案不存在"})
		return
	}

	subjects, err := services.GetSubjectsWithProtocolSite(protocolID, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("visits_%s_%s.csv", protocol.ProtocolNumber, time.Now().Format("20060102"))

	rows := [][]string{
		{"方案编号", protocol.ProtocolNumber},
		{"访视数据", ""},
		{"受试者编号", "访视名称", "实际日期", "是否超窗", "生命体征", "实验室检查", "其他数据"},
	}

	visitMap := make(map[uuid.UUID]string)
	for _, v := range protocol.Visits {
		visitMap[v.ID] = v.VisitName
	}

	for _, s := range subjects {
		for _, r := range s.Records {
			outOfWindow := "否"
			if r.IsOutOfWindow {
				outOfWindow = "是"
			}

			rows = append(rows, []string{
				s.RandomizationID,
				visitMap[r.VisitID],
				utils.FormatDate(r.ActualDate),
				outOfWindow,
				r.VitalSigns,
				r.LabTests,
				r.OtherData,
			})
		}
	}

	utils.WriteCSVResponse(c, filename, rows)
}
