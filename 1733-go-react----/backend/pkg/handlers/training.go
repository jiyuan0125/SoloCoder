package handlers

import (
	"net/http"
	"tdm-system/pkg/models"
	"tdm-system/pkg/storage"

	"github.com/gin-gonic/gin"
)

type CreateTrainingReq struct {
	Name     string               `json:"name"`
	Type     models.TrainingType `json:"type"`
	Form     models.TrainingForm `json:"form"`
	Date     string               `json:"date"`
	Hours    int                  `json:"hours"`
	Lecturer string               `json:"lecturer"`
	Capacity int                  `json:"capacity"`
}

func CreateTraining(c *gin.Context) {
	var req CreateTrainingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := storage.Get()
	t := models.Training{
		ID:       s.NextID(),
		Name:     req.Name,
		Type:     req.Type,
		Form:     req.Form,
		Date:     req.Date,
		Hours:    req.Hours,
		Lecturer: req.Lecturer,
		Capacity: req.Capacity,
	}
	s.SaveTraining(t)
	c.JSON(http.StatusCreated, t)
}

func ListTrainings(c *gin.Context) {
	s := storage.Get()
	c.JSON(http.StatusOK, s.AllTrainings())
}

type RegisterReq struct {
	TeacherID string `json:"teacher_id"`
}

func RegisterTraining(c *gin.Context) {
	tid := c.Param("id")
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := storage.Get()
	t, ok := s.GetTraining(tid)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "培训不存在"})
		return
	}
	regs := s.RegistrationsByTraining(tid)
	if len(regs) >= t.Capacity {
		c.JSON(http.StatusConflict, gin.H{"error": "培训报名已满"})
		return
	}
	for _, r := range regs {
		if r.TeacherID == req.TeacherID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "已报名该培训"})
			return
		}
	}
	reg := models.Registration{
		ID:         s.NextID(),
		TrainingID: tid,
		TeacherID:  req.TeacherID,
	}
	s.SaveRegistration(reg)
	c.JSON(http.StatusCreated, reg)
}

type UpdateAttendanceReq struct {
	Attendance  []models.AttendanceItem `json:"attendance"`
	Score       *float64                `json:"score"`
	StudyReport *bool                   `json:"study_report"`
}

func UpdateRegistration(c *gin.Context) {
	rid := c.Param("rid")
	var req UpdateAttendanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := storage.Get()
	r, ok := s.GetRegistration(rid)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "参训记录不存在"})
		return
	}
	if req.Attendance != nil {
		r.Attendance = req.Attendance
	}
	if req.Score != nil {
		r.Score = req.Score
	}
	if req.StudyReport != nil {
		r.StudyReport = req.StudyReport
	}
	s.SaveRegistration(r)
	c.JSON(http.StatusOK, r)
}

func ListRegistrations(c *gin.Context) {
	teacherID := c.Query("teacher_id")
	s := storage.Get()
	if teacherID != "" {
		c.JSON(http.StatusOK, s.RegistrationsByTeacher(teacherID))
		return
	}
	allRegs := make([]models.Registration, 0)
	for _, t := range s.AllTeachers() {
		regs := s.RegistrationsByTeacher(t.ID)
		allRegs = append(allRegs, regs...)
	}
	c.JSON(http.StatusOK, allRegs)
}

func ListRegistrationsByTeacher(c *gin.Context) {
	tid := c.Param("teacher_id")
	s := storage.Get()
	c.JSON(http.StatusOK, s.RegistrationsByTeacher(tid))
}

func ListRegistrationsByTraining(c *gin.Context) {
	tid := c.Param("id")
	s := storage.Get()
	c.JSON(http.StatusOK, s.RegistrationsByTraining(tid))
}
