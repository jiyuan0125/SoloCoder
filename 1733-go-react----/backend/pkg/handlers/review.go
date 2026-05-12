package handlers

import (
	"net/http"
	"tdm-system/pkg/models"
	"tdm-system/pkg/storage"

	"github.com/gin-gonic/gin"
)

type CreateApplicationReq struct {
	TeacherID          string       `json:"teacher_id"`
	ApplyTitle         models.Title `json:"apply_title"`
	Materials          string       `json:"materials"`
	AchievementSummary string       `json:"achievement_summary"`
}

func CreateApplication(c *gin.Context) {
	var req CreateApplicationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := storage.Get()
	teacher, ok := s.GetTeacher(req.TeacherID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "教师不存在"})
		return
	}
	curOrder := models.TitleOrder[teacher.CurrentTitle]
	appOrder := models.TitleOrder[req.ApplyTitle]
	if appOrder <= curOrder {
		c.JSON(http.StatusBadRequest, gin.H{"error": "申报职称必须高于当前职称"})
		return
	}
	if appOrder-curOrder > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能跨级申报"})
		return
	}
	openYear := s.GetReviewOpenYear()
	apps := s.ApplicationsByTeacher(req.TeacherID)
	for _, a := range apps {
		if a.Year == openYear {
			c.JSON(http.StatusBadRequest, gin.H{"error": "本年度已申报"})
			return
		}
	}
	a := models.TitleApplication{
		ID:                 s.NextID(),
		TeacherID:          req.TeacherID,
		Year:               openYear,
		ApplyTitle:         req.ApplyTitle,
		Materials:          req.Materials,
		AchievementSummary: req.AchievementSummary,
		CurrentStage:       models.RSInitial,
	}
	s.SaveApplication(a)
	c.JSON(http.StatusCreated, a)
}

type ReviewReq struct {
	Stage   models.ReviewStage  `json:"stage"`
	Status  models.ReviewStatus `json:"status"`
	Comment string              `json:"comment"`
}

func stageToIndex(stage models.ReviewStage) int {
	switch stage {
	case models.RSInitial:
		return 0
	case models.RSExternal:
		return 1
	case models.RSFinal:
		return 2
	default:
		return -1
	}
}

func SubmitReview(c *gin.Context) {
	aid := c.Param("id")
	var req ReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := storage.Get()
	a, ok := s.GetApplication(aid)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}
	reqStage := stageToIndex(req.Stage)
	curStage := stageToIndex(a.CurrentStage)
	if reqStage != curStage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评审只能按顺序推进，不能跳步"})
		return
	}
	result := &models.ReviewResult{
		Status:  req.Status,
		Comment: req.Comment,
	}
	if req.Stage == models.RSInitial {
		a.InitialReview = result
		if req.Status == models.RSPassed {
			a.CurrentStage = models.RSExternal
		}
	} else if req.Stage == models.RSExternal {
		a.ExternalReview = result
		if req.Status == models.RSPassed {
			a.CurrentStage = models.RSFinal
		}
	} else if req.Stage == models.RSFinal {
		a.FinalReview = result
		allPassed := a.InitialReview != nil && a.InitialReview.Status == models.RSPassed &&
			a.ExternalReview != nil && a.ExternalReview.Status == models.RSPassed &&
			req.Status == models.RSPassed
		a.FinalStatus = &allPassed
	}
	s.SaveApplication(a)
	c.JSON(http.StatusOK, a)
}

func ListApplications(c *gin.Context) {
	s := storage.Get()
	c.JSON(http.StatusOK, s.AllApplications())
}

func ListApplicationsByTeacher(c *gin.Context) {
	tid := c.Param("teacher_id")
	s := storage.Get()
	c.JSON(http.StatusOK, s.ApplicationsByTeacher(tid))
}
