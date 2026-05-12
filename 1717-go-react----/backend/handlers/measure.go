package handlers

import (
	"net/http"
	"time"

	"hospital-infection/database"
	"hospital-infection/models"

	"github.com/gin-gonic/gin"
)

type RecommendedMeasure struct {
	MeasureType models.MeasureType `json:"MeasureType"`
	Description string             `json:"Description"`
}

func GetRecommendedMeasures(c *gin.Context) {
	site := models.InfectionSite(c.Query("site"))
	if site == "" {
		site = models.InfectionSite(c.Query("Site"))
	}
	if site == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供感染部位"})
		return
	}

	recommendations := getRecommendationsBySite(site)
	c.JSON(http.StatusOK, recommendations)
}

func getRecommendationsBySite(site models.InfectionSite) []RecommendedMeasure {
	switch site {
	case models.SiteRespiratory:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureIsolation, Description: "飞沫隔离"},
			{MeasureType: models.MeasureHandHygiene, Description: "强化手卫生"},
			{MeasureType: models.MeasureEnvironmentClean, Description: "环境消毒"},
		}
	case models.SiteSurgicalIncision:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureIsolation, Description: "接触隔离"},
			{MeasureType: models.MeasureEquipmentSterile, Description: "术中无菌操作核查"},
			{MeasureType: models.MeasureEnvironmentClean, Description: "术后伤口护理规范"},
		}
	case models.SiteUrinaryTract:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureHandHygiene, Description: "强化手卫生"},
			{MeasureType: models.MeasureEquipmentSterile, Description: "导尿管护理规范"},
			{MeasureType: models.MeasureAntibioticAdjust, Description: "抗生素调整评估"},
		}
	case models.SiteBloodstream:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureIsolation, Description: "接触隔离"},
			{MeasureType: models.MeasureEquipmentSterile, Description: "中心静脉导管护理"},
			{MeasureType: models.MeasureAntibioticAdjust, Description: "根据药敏调整抗生素"},
		}
	case models.SiteDigestive:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureIsolation, Description: "肠道隔离"},
			{MeasureType: models.MeasureHandHygiene, Description: "强化手卫生"},
			{MeasureType: models.MeasureEnvironmentClean, Description: "环境消毒"},
		}
	case models.SiteSkinSoftTissue:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureIsolation, Description: "接触隔离"},
			{MeasureType: models.MeasureHandHygiene, Description: "强化手卫生"},
			{MeasureType: models.MeasureEnvironmentClean, Description: "伤口换药规范"},
		}
	default:
		return []RecommendedMeasure{
			{MeasureType: models.MeasureHandHygiene, Description: "强化手卫生"},
		}
	}
}

type CreatePreventionMeasureRequest struct {
	InfectionCaseID uint              `json:"InfectionCaseID" binding:"required"`
	MeasureType     models.MeasureType `json:"MeasureType" binding:"required"`
	DepartmentID    uint              `json:"DepartmentID" binding:"required"`
	Executor        string            `json:"Executor" binding:"required"`
	ExecuteDate     time.Time         `json:"ExecuteDate"`
}

func CreatePreventionMeasure(c *gin.Context) {
	var req CreatePreventionMeasureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var caseItem models.InfectionCase
	if err := database.DB.First(&caseItem, req.InfectionCaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "感染病例不存在"})
		return
	}

	var department models.Department
	if err := database.DB.First(&department, req.DepartmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "执行科室不存在"})
		return
	}

	executeDate := req.ExecuteDate
	if executeDate.IsZero() {
		executeDate = time.Now()
	}

	measure := models.PreventionMeasure{
		InfectionCaseID: req.InfectionCaseID,
		MeasureType:     req.MeasureType,
		DepartmentID:    req.DepartmentID,
		Executor:        req.Executor,
		ExecuteDate:     executeDate,
		Status:          "executed",
	}

	if err := database.DB.Create(&measure).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建防控措施失败"})
		return
	}

	database.DB.Preload("InfectionCase").Preload("Department").First(&measure, measure.ID)
	c.JSON(http.StatusCreated, measure)
}

func GetPreventionMeasures(c *gin.Context) {
	var measures []models.PreventionMeasure
	query := database.DB.Preload("InfectionCase").Preload("Department")

	caseID := c.Query("case_id")
	if caseID == "" {
		caseID = c.Query("CaseID")
	}
	if caseID != "" {
		query = query.Where("infection_case_id = ?", caseID)
	}

	departmentID := c.Query("department_id")
	if departmentID == "" {
		departmentID = c.Query("DepartmentID")
	}
	if departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}

	query.Order("created_at desc").Find(&measures)
	c.JSON(http.StatusOK, measures)
}

func GetPreventionMeasure(c *gin.Context) {
	id := c.Param("id")
	var measure models.PreventionMeasure
	if err := database.DB.Preload("InfectionCase").Preload("Department").First(&measure, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "防控措施不存在"})
		return
	}
	c.JSON(http.StatusOK, measure)
}

func DeletePreventionMeasure(c *gin.Context) {
	id := c.Param("id")
	var measure models.PreventionMeasure
	if err := database.DB.First(&measure, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "防控措施不存在"})
		return
	}

	if err := database.DB.Delete(&measure).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除防控措施失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
