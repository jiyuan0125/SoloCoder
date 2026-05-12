package handlers

import (
	"net/http"
	"strconv"

	"learning-platform/internal/models"
	"learning-platform/internal/services"
	"learning-platform/internal/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	courseService     *services.CourseService
	learningService   *services.LearningService
	achievementService *services.AchievementService
	quizService       *services.QuizService
	orderService      *services.OrderService
}

func NewHandler(cs *services.CourseService, ls *services.LearningService, as *services.AchievementService, qs *services.QuizService, os *services.OrderService) *Handler {
	return &Handler{
		courseService:     cs,
		learningService:   ls,
		achievementService: as,
		quizService:       qs,
		orderService:      os,
	}
}

func (h *Handler) CreateCourse(c *gin.Context) {
	var course models.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	created, err := h.courseService.CreateCourse(&course)
	if err != nil {
		if _, ok := err.(*storage.DuplicateError); ok {
			c.JSON(http.StatusConflict, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: created})
}

func (h *Handler) AddPrerequisite(c *gin.Context) {
	var req struct {
		CourseID   string `json:"course_id" binding:"required"`
		PrereqID   string `json:"prerequisite_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	err := h.courseService.AddPrerequisite(req.CourseID, req.PrereqID)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*storage.CycleError); ok {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "cycle dependency detected"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *Handler) GetAllCourses(c *gin.Context) {
	courses := h.courseService.GetAllCourses()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: courses})
}

func (h *Handler) GetCourse(c *gin.Context) {
	id := c.Param("id")
	course, exists := h.courseService.GetCourse(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "course not found"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: course})
}

func (h *Handler) GetCourseGraph(c *gin.Context) {
	graph := h.courseService.GetCourseGraph()
	nodes := graph.GetNodes()
	edges := graph.GetEdges()
	levels, _ := graph.GetLevels()

	result := map[string]interface{}{
		"nodes":  nodes,
		"edges":  edges,
		"levels": levels,
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: result})
}

func (h *Handler) CreateStudent(c *gin.Context) {
	var student models.Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	created, err := h.learningService.CreateStudent(&student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: created})
}

func (h *Handler) StartLearning(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		CourseID  string `json:"course_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	err := h.learningService.StartLearning(req.StudentID, req.CourseID)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*services.MaxCoursesError); ok {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "maximum concurrent courses exceeded"})
			return
		}
		if _, ok := err.(*services.PrerequisiteNotMetError); ok {
			c.JSON(http.StatusForbidden, models.APIResponse{Success: false, Error: "prerequisites not completed"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *Handler) CompleteUnit(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		CourseID  string `json:"course_id" binding:"required"`
		UnitID    string `json:"unit_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	err := h.learningService.CompleteUnit(req.StudentID, req.CourseID, req.UnitID)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *Handler) GetProgress(c *gin.Context) {
	studentID := c.Query("student_id")
	courseID := c.Query("course_id")

	if studentID == "" || courseID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "student_id and course_id required"})
		return
	}

	progress, exists := h.learningService.GetProgress(studentID, courseID)
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "progress not found"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: progress})
}

func (h *Handler) GenerateLearningPath(c *gin.Context) {
	var req struct {
		StudentID      string `json:"student_id" binding:"required"`
		TargetCourseID string `json:"target_course_id" binding:"required"`
		Goal           string `json:"goal"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	path, err := h.learningService.GenerateLearningPath(req.StudentID, req.TargetCourseID, req.Goal)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: path})
}

func (h *Handler) GetQuiz(c *gin.Context) {
	studentID := c.Query("student_id")
	courseID := c.Query("course_id")

	if studentID == "" || courseID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "student_id and course_id required"})
		return
	}

	questions, err := h.quizService.GenerateQuiz(studentID, courseID)
	if err != nil {
		if _, ok := err.(*services.QuizLockedError); ok {
			c.JSON(http.StatusForbidden, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	for i := range questions {
		questions[i].CorrectIndex = -1
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: questions})
}

func (h *Handler) SubmitQuiz(c *gin.Context) {
	var req struct {
		StudentID string   `json:"student_id" binding:"required"`
		CourseID  string   `json:"course_id" binding:"required"`
		Answers   []int    `json:"answers" binding:"required"`
		Questions []models.Question `json:"questions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	attempt, err := h.quizService.SubmitQuiz(req.StudentID, req.CourseID, req.Answers, req.Questions)
	if err != nil {
		if _, ok := err.(*services.QuizLockedError); ok {
			c.JSON(http.StatusForbidden, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: attempt})
}

func (h *Handler) GetAnalytics(c *gin.Context) {
	studentID := c.Query("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "student_id required"})
		return
	}

	analytics := h.achievementService.GetLearningAnalytics(studentID)
	achievements := h.achievementService.GetStudentAchievements(studentID)

	result := map[string]interface{}{
		"analytics":    analytics,
		"achievements": achievements,
		"recommendations": h.getRecommendations(analytics),
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: result})
}

func (h *Handler) getRecommendations(analytics *models.LearningAnalytics) []string {
	recommendations := []string{}
	if analytics.LearningSpeed < 0.5 {
		recommendations = append(recommendations, "建议减少同时学习的课程数量")
	}
	if analytics.MasteryLevel < 60 {
		recommendations = append(recommendations, "建议复习先修课程以加强基础")
	}
	return recommendations
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	created, err := h.orderService.CreateProduct(&product)
	if err != nil {
		if _, ok := err.(*storage.DuplicateError); ok {
			c.JSON(http.StatusConflict, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: created})
}

func (h *Handler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	products, total := h.orderService.GetProductsPaginated(page, size)
	result := map[string]interface{}{
		"items": products,
		"page":  page,
		"size":  size,
		"total": total,
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: result})
}

func (h *Handler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	product, exists := h.orderService.GetProduct(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "product not found"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: product})
}

func (h *Handler) GetProductSubResources(c *gin.Context) {
	id := c.Param("id")
	history := h.orderService.GetPriceHistory(id)
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: history})
}

func (h *Handler) UpdateProductPrice(c *gin.Context) {
	var req struct {
		NewPrice  float64 `json:"new_price" binding:"required"`
		ChangedBy string  `json:"changed_by"`
	}

	id := c.Param("id")
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	err := h.orderService.UpdateProductPrice(id, req.NewPrice, req.ChangedBy)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		Items     []struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int    `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	items := []struct {
		ProductID string
		Quantity  int
	}{}
	for _, item := range req.Items {
		items = append(items, struct {
			ProductID string
			Quantity  int
		}{ProductID: item.ProductID, Quantity: item.Quantity})
	}

	order, err := h.orderService.CreateOrder(req.StudentID, items)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*services.InsufficientStockError); ok {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: order})
}

func (h *Handler) ConfirmOrder(c *gin.Context) {
	id := c.Param("id")
	err := h.orderService.ConfirmOrder(id)
	if err != nil {
		if _, ok := err.(*storage.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*services.InsufficientStockError); ok {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		if _, ok := err.(*services.InvalidOrderStatusError); ok {
			c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *Handler) GetOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	orders, total := h.orderService.GetOrdersPaginated(page, size)
	result := map[string]interface{}{
		"items": orders,
		"page":  page,
		"size":  size,
		"total": total,
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: result})
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, exists := h.orderService.GetOrder(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "order not found"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: order})
}

func (h *Handler) GetPurchaseRequests(c *gin.Context) {
	requests := h.orderService.GetAllPurchaseRequests()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: requests})
}
