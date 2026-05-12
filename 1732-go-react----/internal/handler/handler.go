package handler

import (
	"credit-system/internal/model"
	"credit-system/internal/store"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		courses := api.Group("/courses")
		{
			courses.GET("", h.listCourses)
			courses.GET("/:id", h.getCourse)
			courses.POST("", h.createCourse)
			courses.PUT("/:id", h.updateCourse)
			courses.DELETE("/:id", h.deleteCourse)
			courses.POST("/:id/enroll", h.enrollCourse)
			courses.POST("/:id/withdraw", h.withdrawCourse)
			courses.GET("/:id/enrollments", h.getCourseEnrollments)
			courses.POST("/:id/enrollments/:enrollmentId/score", h.recordScore)
		}

		teachers := api.Group("/teachers")
		{
			teachers.GET("", h.listTeachers)
			teachers.GET("/:id", h.getTeacher)
			teachers.POST("", h.createTeacher)
			teachers.PUT("/:id", h.updateTeacher)
			teachers.POST("/:id/transfer", h.transferTeacher)
			teachers.GET("/:id/enrollments", h.getTeacherEnrollments)
			teachers.GET("/:id/report", h.getCreditReport)
		}

		todos := api.Group("/todos")
		{
			todos.GET("", h.listTodos)
			todos.POST("/:id/read", h.markTodoRead)
		}
	}
}

type CreateCourseRequest struct {
	Name         string `json:"name" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Hours        int    `json:"hours" binding:"required"`
	StartDate    string `json:"start_date" binding:"required"`
	EndDate      string `json:"end_date" binding:"required"`
	Instructor   string `json:"instructor" binding:"required"`
	Description  string `json:"description"`
	Capacity     int    `json:"capacity" binding:"required"`
}

func (h *Handler) createCourse(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	if req.Hours <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "学时数必须为正整数"})
		return
	}

	if req.Capacity <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "课程容量必须大于0"})
		return
	}

	courseType := model.CourseType(strings.ToLower(req.Type))
	if courseType != model.CourseTypeOnline && courseType != model.CourseTypeOffline && courseType != model.CourseTypeHybrid {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "课程类型必须是online、offline或hybrid"})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "开始日期格式错误，请使用YYYY-MM-DD"})
		return
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "结束日期格式错误，请使用YYYY-MM-DD"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "结束日期不能早于开始日期"})
		return
	}

	course := &model.Course{
		Name:        req.Name,
		Type:        courseType,
		Hours:       req.Hours,
		Credits:     model.CalculateCredits(req.Hours),
		StartDate:   startDate,
		EndDate:     endDate,
		Instructor:  req.Instructor,
		Description: req.Description,
		Capacity:    req.Capacity,
	}

	created := h.store.CreateCourse(course)
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) listCourses(c *gin.Context) {
	courses := h.store.GetAllCourses()
	c.JSON(http.StatusOK, courses)
}

func (h *Handler) getCourse(c *gin.Context) {
	id := c.Param("id")
	course := h.store.GetCourse(id)
	if course == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "课程不存在"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *Handler) updateCourse(c *gin.Context) {
	id := c.Param("id")
	existing := h.store.GetCourse(id)
	if existing == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "课程不存在"})
		return
	}

	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	if req.Hours <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "学时数必须为正整数"})
		return
	}

	if req.Capacity < existing.EnrolledCount {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "课程容量不能小于当前已报名人数"})
		return
	}

	courseType := model.CourseType(strings.ToLower(req.Type))
	if courseType != model.CourseTypeOnline && courseType != model.CourseTypeOffline && courseType != model.CourseTypeHybrid {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "课程类型必须是online、offline或hybrid"})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "开始日期格式错误，请使用YYYY-MM-DD"})
		return
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "结束日期格式错误，请使用YYYY-MM-DD"})
		return
	}

	existing.Name = req.Name
	existing.Type = courseType
	existing.Hours = req.Hours
	existing.Credits = model.CalculateCredits(req.Hours)
	existing.StartDate = startDate
	existing.EndDate = endDate
	existing.Instructor = req.Instructor
	existing.Description = req.Description
	existing.Capacity = req.Capacity

	updated := h.store.UpdateCourse(existing)
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) deleteCourse(c *gin.Context) {
	id := c.Param("id")
	existing := h.store.GetCourse(id)
	if existing == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "课程不存在"})
		return
	}

	if h.store.HasEnrollmentsForCourse(id) {
		c.JSON(http.StatusConflict, model.ErrorResponse{Error: "已有教师报名的课程不能删除"})
		return
	}

	h.store.DeleteCourse(id)
	c.JSON(http.StatusOK, model.SuccessResponse{Success: true})
}

type EnrollRequest struct {
	TeacherID string `json:"teacher_id" binding:"required"`
}

func (h *Handler) enrollCourse(c *gin.Context) {
	courseID := c.Param("id")
	var req EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	course := h.store.GetCourse(courseID)
	if course == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "课程不存在"})
		return
	}

	teacher := h.store.GetTeacher(req.TeacherID)
	if teacher == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}

	if existing := h.store.FindEnrollment(courseID, req.TeacherID); existing != nil {
		c.JSON(http.StatusConflict, model.ErrorResponse{Error: "同一课程同一教师只能报名一次"})
		return
	}

	if course.EnrolledCount >= course.Capacity {
		c.JSON(http.StatusConflict, model.ErrorResponse{Error: "课程已满"})
		return
	}

	enrollment := &model.Enrollment{
		CourseID:  courseID,
		TeacherID: req.TeacherID,
		Status:    model.EnrollmentStatusEnrolled,
		College:   teacher.College,
	}

	h.store.CreateEnrollment(enrollment)

	todo := &model.Todo{
		TeacherID: req.TeacherID,
		CourseID:  courseID,
		Type:      model.TodoTypeCourseStart,
		Message:   fmt.Sprintf("您已成功报名课程：%s，%s开始上课", course.Name, formatDate(course.StartDate)),
	}
	h.store.CreateTodo(todo)

	c.JSON(http.StatusCreated, enrollment)
}

type WithdrawRequest struct {
	TeacherID string `json:"teacher_id" binding:"required"`
}

func (h *Handler) withdrawCourse(c *gin.Context) {
	courseID := c.Param("id")
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	enrollment := h.store.FindEnrollment(courseID, req.TeacherID)
	if enrollment == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "未找到报名记录"})
		return
	}

	if enrollment.Status != model.EnrollmentStatusEnrolled {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "只有已报名状态才能退课"})
		return
	}

	enrollment.Status = model.EnrollmentStatusWithdrawn
	h.store.UpdateEnrollment(enrollment)

	course := h.store.GetCourse(courseID)
	if course != nil {
		course.EnrolledCount--
		h.store.UpdateCourse(course)
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Success: true})
}

func (h *Handler) getCourseEnrollments(c *gin.Context) {
	courseID := c.Param("id")
	course := h.store.GetCourse(courseID)
	if course == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "课程不存在"})
		return
	}

	enrollments := h.store.GetEnrollmentsByCourse(courseID)
	c.JSON(http.StatusOK, enrollments)
}

type RecordScoreRequest struct {
	Score int `json:"score" binding:"required"`
}

func (h *Handler) recordScore(c *gin.Context) {
	courseID := c.Param("id")
	enrollmentID := c.Param("enrollmentId")

	course := h.store.GetCourse(courseID)
	if course == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "课程不存在"})
		return
	}

	var req RecordScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	if req.Score < 0 || req.Score > 100 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "成绩必须在0到100之间"})
		return
	}

	enrollment := h.store.GetEnrollment(enrollmentID)
	if enrollment == nil || enrollment.CourseID != courseID {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "报名记录不存在"})
		return
	}

	if enrollment.Status == model.EnrollmentStatusWithdrawn {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "已退课的报名不能录入成绩"})
		return
	}

	enrollment.Score = &req.Score
	if req.Score >= 60 {
		enrollment.Status = model.EnrollmentStatusPassed
	} else {
		enrollment.Status = model.EnrollmentStatusFailed
	}
	h.store.UpdateEnrollment(enrollment)

	var resultMsg string
	if req.Score >= 60 {
		resultMsg = fmt.Sprintf("恭喜！您在课程《%s》中取得了%d分的成绩，获得%d学分", course.Name, req.Score, course.Credits)
	} else {
		resultMsg = fmt.Sprintf("您在课程《%s》中取得了%d分的成绩，未达到合格标准", course.Name, req.Score)
	}
	todo := &model.Todo{
		TeacherID: enrollment.TeacherID,
		CourseID:  courseID,
		Type:      model.TodoTypeResultAvailable,
		Message:   resultMsg,
	}
	h.store.CreateTodo(todo)

	c.JSON(http.StatusOK, enrollment)
}

type CreateTeacherRequest struct {
	Name       string `json:"name" binding:"required"`
	EmployeeID string `json:"employee_id" binding:"required"`
	College    string `json:"college" binding:"required"`
	Position   string `json:"position" binding:"required"`
}

func (h *Handler) createTeacher(c *gin.Context) {
	var req CreateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	position := model.Position(req.Position)
	if position != model.PositionAssistant && position != model.PositionLecturer &&
		position != model.PositionAssociate && position != model.PositionProfessor {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "职称必须是助教、讲师、副教授或教授"})
		return
	}

	teacher := &model.Teacher{
		Name:       req.Name,
		EmployeeID: req.EmployeeID,
		College:    req.College,
		Position:   position,
	}

	created := h.store.CreateTeacher(teacher)
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) listTeachers(c *gin.Context) {
	teachers := h.store.GetAllTeachers()
	c.JSON(http.StatusOK, teachers)
}

func (h *Handler) getTeacher(c *gin.Context) {
	id := c.Param("id")
	teacher := h.store.GetTeacher(id)
	if teacher == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}
	c.JSON(http.StatusOK, teacher)
}

func (h *Handler) updateTeacher(c *gin.Context) {
	id := c.Param("id")
	existing := h.store.GetTeacher(id)
	if existing == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}

	var req CreateTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	position := model.Position(req.Position)
	if position != model.PositionAssistant && position != model.PositionLecturer &&
		position != model.PositionAssociate && position != model.PositionProfessor {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "职称必须是助教、讲师、副教授或教授"})
		return
	}

	existing.Name = req.Name
	existing.EmployeeID = req.EmployeeID
	existing.Position = position

	updated := h.store.UpdateTeacher(existing)
	c.JSON(http.StatusOK, updated)
}

type TransferRequest struct {
	NewCollege string `json:"new_college" binding:"required"`
	IsAdmin    bool   `json:"is_admin"`
}

func (h *Handler) transferTeacher(c *gin.Context) {
	id := c.Param("id")
	teacher := h.store.GetTeacher(id)
	if teacher == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	if !req.IsAdmin {
		c.JSON(http.StatusForbidden, model.ErrorResponse{Error: "只有管理员才能进行调岗操作"})
		return
	}

	if teacher.College == req.NewCollege {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "新学院与原学院相同"})
		return
	}

	h.store.TransferTeacher(id, req.NewCollege)
	updated := h.store.GetTeacher(id)
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) getTeacherEnrollments(c *gin.Context) {
	teacherID := c.Param("id")
	teacher := h.store.GetTeacher(teacherID)
	if teacher == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}

	enrollments := h.store.GetEnrollmentsByTeacher(teacherID)
	c.JSON(http.StatusOK, enrollments)
}

func (h *Handler) getCreditReport(c *gin.Context) {
	teacherID := c.Param("id")
	teacher := h.store.GetTeacher(teacherID)
	if teacher == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}

	now := time.Now()
	currentYear := now.Year()
	threshold := model.GetCreditThreshold(teacher.Position)

	enrollments := h.store.GetEnrollmentsByTeacher(teacherID)
	collegeHistories := h.store.GetCollegeHistories(teacherID)

	currentYearCredits := 0
	yearlyStats := make(map[int]int)

	for _, e := range enrollments {
		if e.Status != model.EnrollmentStatusPassed {
			continue
		}

		course := h.store.GetCourse(e.CourseID)
		if course == nil {
			continue
		}

		enrollmentYear := e.EnrolledAt.Year()
		_, exists := yearlyStats[enrollmentYear]
		if !exists {
			yearlyStats[enrollmentYear] = 0
		}

		creditsForCourse := course.Credits
		if enrollmentYear == currentYear {
			useCredits := false
			if len(collegeHistories) == 0 {
				useCredits = true
			} else {
				for _, ch := range collegeHistories {
					if ch.College == e.College {
						enrollmentTime := e.EnrolledAt
						if ch.EndDate == nil {
							if enrollmentTime.After(ch.StartDate) || enrollmentTime.Equal(ch.StartDate) {
								useCredits = true
							}
						} else {
							if (enrollmentTime.After(ch.StartDate) || enrollmentTime.Equal(ch.StartDate)) &&
								enrollmentTime.Before(*ch.EndDate) {
								useCredits = true
							}
						}
					}
				}
			}
			if useCredits {
				currentYearCredits += creditsForCourse
			}
		}

		yearlyStats[enrollmentYear] += creditsForCourse
	}

	diff := threshold - currentYearCredits
	if diff < 0 {
		diff = 0
	}

	yearlyTrend := []map[string]interface{}{}
	for year, credits := range yearlyStats {
		yearlyTrend = append(yearlyTrend, map[string]interface{}{
			"year":    year,
			"credits": credits,
		})
	}

	report := map[string]interface{}{
		"teacher_id":           teacher.ID,
		"teacher_name":         teacher.Name,
		"position":             teacher.Position,
		"college":              teacher.College,
		"threshold":            threshold,
		"current_year":         currentYear,
		"current_year_credits": currentYearCredits,
		"is_qualified":         currentYearCredits >= threshold,
		"diff":                 diff,
		"yearly_trend":         yearlyTrend,
	}

	c.JSON(http.StatusOK, report)
}

func (h *Handler) listTodos(c *gin.Context) {
	teacherID := c.Query("teacher_id")
	if teacherID == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "缺少teacher_id参数"})
		return
	}

	teacher := h.store.GetTeacher(teacherID)
	if teacher == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "教师不存在"})
		return
	}

	todos := h.store.GetTodosByTeacher(teacherID)
	c.JSON(http.StatusOK, todos)
}

func (h *Handler) markTodoRead(c *gin.Context) {
	id := c.Param("id")
	h.store.MarkTodoRead(id)
	c.JSON(http.StatusOK, model.SuccessResponse{Success: true})
}

func (h *Handler) CheckCourseStartReminders() {
	now := time.Now()
	threeDaysLater := now.AddDate(0, 0, 3)

	courses := h.store.GetAllCourses()
	for _, course := range courses {
		startDate := course.StartDate
		if startDate.Year() == threeDaysLater.Year() &&
			startDate.YearDay() == threeDaysLater.YearDay() {
			enrollments := h.store.GetEnrollmentsByCourse(course.ID)
			for _, e := range enrollments {
				if e.Status != model.EnrollmentStatusEnrolled {
					continue
				}
				existingTodos := h.store.GetTodosByTeacher(e.TeacherID)
				alreadyHas := false
				for _, t := range existingTodos {
					if t.CourseID == course.ID && t.Type == model.TodoTypeCourseStart {
						alreadyHas = true
						break
					}
				}
				if !alreadyHas {
					todo := &model.Todo{
						TeacherID: e.TeacherID,
						CourseID:  course.ID,
						Type:      model.TodoTypeCourseStart,
						Message:   fmt.Sprintf("提醒：课程《%s》将于3天后（%s）开始上课", course.Name, formatDate(startDate)),
					}
					h.store.CreateTodo(todo)
				}
			}
		}
	}
}

func init() {
	_ = strconv.Itoa
}
