package handlers

import (
	"math"
	"net/http"
	"school-connect/internal/models"
	"school-connect/internal/storage"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	store *storage.MemoryStore
}

func NewScoreHandler(store *storage.MemoryStore) *ScoreHandler {
	return &ScoreHandler{store: store}
}

type BatchRecordScoresRequest struct {
	Class    string                `json:"class" binding:"required"`
	Subject  string                `json:"subject" binding:"required"`
	Semester string                `json:"semester" binding:"required"`
	ExamDate string                `json:"exam_date" binding:"required"`
	Scores   []StudentScoreInput   `json:"scores" binding:"required"`
}

type StudentScoreInput struct {
	StudentID string  `json:"student_id" binding:"required"`
	Score     *float64 `json:"score"`
}

func (h *ScoreHandler) BatchRecord(c *gin.Context) {
	userID := c.GetString("userID")
	user, _ := c.Get("user")
	u := user.(*models.User)

	var req BatchRecordScoresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误: " + err.Error()})
		return
	}

	for _, si := range req.Scores {
		if si.Score != nil {
			if *si.Score < 0 || *si.Score > 150 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "分数范围必须在0-150之间"})
				return
			}
		}
	}

	examDate, err := parseDate(req.ExamDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
		return
	}

	students := h.store.GetStudentsByClass(req.Class)
	studentMap := make(map[string]*models.Student)
	for _, s := range students {
		studentMap[s.ID] = s
	}

	recorded := []*models.ExamScore{}
	for _, si := range req.Scores {
		student := studentMap[si.StudentID]
		if student == nil {
			continue
		}

		score := h.store.RecordScore(&models.ExamScore{
			StudentID: si.StudentID,
			Grade:     student.Grade,
			Class:     student.Class,
			Subject:   req.Subject,
			Semester:  req.Semester,
			Score:     si.Score,
			ExamDate:  examDate,
			RecordedBy: userID,
		})
		recorded = append(recorded, score)
	}

	stats := h.calculateClassStats(req.Class, req.Subject, req.Semester)

	c.JSON(http.StatusOK, gin.H{
		"recorded": len(recorded),
		"stats":    stats,
	})
}

func (h *ScoreHandler) GetStudentScores(c *gin.Context) {
	studentID := c.Param("studentID")
	semester := c.Query("semester")

	user, _ := c.Get("user")
	u := user.(*models.User)

	if u.Role == models.RoleParent {
		parent := h.store.GetParent(u.ID)
		isChild := false
		for _, cid := range parent.ChildIDs {
			if cid == studentID {
				isChild = true
				break
			}
		}
		if !isChild {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此学生成绩"})
			return
		}
	}

	scores := h.store.GetStudentLatestScores(studentID, semester)

	c.JSON(http.StatusOK, gin.H{
		"scores": scores,
	})
}

func (h *ScoreHandler) GetStudentSubjectHistory(c *gin.Context) {
	studentID := c.Param("studentID")
	subject := c.Param("subject")

	user, _ := c.Get("user")
	u := user.(*models.User)

	if u.Role == models.RoleParent {
		parent := h.store.GetParent(u.ID)
		isChild := false
		for _, cid := range parent.ChildIDs {
			if cid == studentID {
				isChild = true
				break
			}
		}
		if !isChild {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此学生成绩"})
			return
		}
	}

	history := h.store.GetStudentSubjectHistory(studentID, subject)

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

func (h *ScoreHandler) GetClassStats(c *gin.Context) {
	class := c.Query("class")
	subject := c.Query("subject")
	semester := c.Query("semester")

	if class == "" || subject == "" || semester == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	stats := h.calculateClassStats(class, subject, semester)

	c.JSON(http.StatusOK, stats)
}

func (h *ScoreHandler) GetClassStudents(c *gin.Context) {
	class := c.Param("class")
	students := h.store.GetStudentsByClass(class)
	c.JSON(http.StatusOK, gin.H{"students": students})
}

func (h *ScoreHandler) calculateClassStats(class, subject, semester string) *models.ClassScoreStats {
	scores := h.store.GetClassScores(class, subject, semester)
	allStudents := h.store.GetStudentsByClass(class)

	if len(scores) == 0 {
		return &models.ClassScoreStats{
			Class:         class,
			Subject:       subject,
			Semester:      semester,
			TotalStudents: len(allStudents),
			ValidScores:   0,
			Segments:      createEmptySegments(),
		}
	}

	var sum float64
	var count int
	max := math.Inf(-1)
	min := math.Inf(1)

	segments := map[string]int{
		"优秀":   0,
		"良好":   0,
		"中等":   0,
		"及格":   0,
		"不及格": 0,
	}

	for _, s := range scores {
		if s.Score != nil {
			score := *s.Score
			sum += score
			count++

			if score > max {
				max = score
			}
			if score < min {
				min = score
			}

			if score >= 90 {
				segments["优秀"]++
			} else if score >= 80 {
				segments["良好"]++
			} else if score >= 70 {
				segments["中等"]++
			} else if score >= 60 {
				segments["及格"]++
			} else {
				segments["不及格"]++
			}
		}
	}

	var avg, maxVal, minVal float64
	if count > 0 {
		avg = sum / float64(count)
		maxVal = max
		minVal = min
	}

	return &models.ClassScoreStats{
		Class:         class,
		Subject:       subject,
		Semester:      semester,
		Average:       avg,
		MaxScore:      maxVal,
		MinScore:      minVal,
		TotalStudents: len(allStudents),
		ValidScores:   count,
		Segments: []models.ScoreSegment{
			{Range: "90-150", Count: segments["优秀"], Label: "优秀"},
			{Range: "80-89", Count: segments["良好"], Label: "良好"},
			{Range: "70-79", Count: segments["中等"], Label: "中等"},
			{Range: "60-69", Count: segments["及格"], Label: "及格"},
			{Range: "0-59", Count: segments["不及格"], Label: "不及格"},
		},
	}
}

func createEmptySegments() []models.ScoreSegment {
	return []models.ScoreSegment{
		{Range: "90-150", Count: 0, Label: "优秀"},
		{Range: "80-89", Count: 0, Label: "良好"},
		{Range: "70-79", Count: 0, Label: "中等"},
		{Range: "60-69", Count: 0, Label: "及格"},
		{Range: "0-59", Count: 0, Label: "不及格"},
	}
}
