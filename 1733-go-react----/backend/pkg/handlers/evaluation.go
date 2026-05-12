package handlers

import (
	"math"
	"net/http"
	"tdm-system/pkg/models"
	"tdm-system/pkg/storage"

	"github.com/gin-gonic/gin"
)

type CreateEvaluationReq struct {
	TeacherID string       `json:"teacher_id"`
	Semester  string       `json:"semester"`
	Attitude   *models.Score `json:"attitude"`
	Content    *models.Score `json:"content"`
	Method     *models.Score `json:"method"`
	Effect     *models.Score `json:"effect"`
}

func isValidScore(v float64) bool {
	if v < 1 || v > 5 {
		return false
	}
	scaled := v * 10
	if math.Abs(scaled-math.Round(scaled)) > 0.0001 {
		return false
	}
	return true
}

func calcDimension(s *models.Score) float64 {
	if s == nil {
		return 0
	}
	count := 0.0
	sum := 0.0
	if s.Student != nil {
		sum += *s.Student
		count++
	}
	if s.Supervisor != nil {
		sum += *s.Supervisor
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

func validateScore(s *models.Score) bool {
	if s == nil {
		return true
	}
	if s.Student != nil && !isValidScore(*s.Student) {
		return false
	}
	if s.Supervisor != nil && !isValidScore(*s.Supervisor) {
		return false
	}
	return true
}

func CreateEvaluation(c *gin.Context) {
	var req CreateEvaluationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validateScore(req.Attitude) || !validateScore(req.Content) ||
		!validateScore(req.Method) || !validateScore(req.Effect) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评教分数必须在1-5之间，且为整数或一位小数"})
		return
	}
	s := storage.Get()
	dims := []float64{
		calcDimension(req.Attitude),
		calcDimension(req.Content),
		calcDimension(req.Method),
		calcDimension(req.Effect),
	}
	final := 0.0
	count := 0.0
	for _, d := range dims {
		if d > 0 {
			final += d
			count++
		}
	}
	if count > 0 {
		final = final / count
	}
	e := models.TeachingEvaluation{
		ID:         s.NextID(),
		TeacherID:  req.TeacherID,
		Semester:   req.Semester,
		Attitude:   req.Attitude,
		Content:    req.Content,
		Method:     req.Method,
		Effect:     req.Effect,
		FinalScore: final,
		Qualified:  final >= 3.0,
	}
	s.SaveEvaluation(e)
	c.JSON(http.StatusCreated, e)
}

func ListEvaluations(c *gin.Context) {
	s := storage.Get()
	c.JSON(http.StatusOK, s.AllEvaluations())
}

func ListEvaluationsByTeacher(c *gin.Context) {
	tid := c.Param("teacher_id")
	s := storage.Get()
	c.JSON(http.StatusOK, s.EvaluationsByTeacher(tid))
}
