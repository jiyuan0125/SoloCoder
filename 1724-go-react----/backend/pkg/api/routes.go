package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teaching-evaluation-system/pkg/models"
	"teaching-evaluation-system/pkg/services"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		tasks := api.Group("/tasks")
		{
			tasks.GET("", listTasks)
			tasks.POST("", createTask)
			tasks.GET("/:id", getTask)
			tasks.GET("/:id/items", getTaskItems)
			tasks.POST("/:id/actions/:action", taskAction)
		}

		api.GET("/questions/:course_type", getQuestions)

		api.POST("/evaluations", submitEvaluation)

		results := api.Group("/results")
		{
			results.GET("/task/:task_id", getResults)
			results.POST("/task/:task_id/calculate", calculateResults)
			results.GET("/export/:task_id", exportResults)
			results.GET("/:result_id/comments", getComments)
		}

		feedback := api.Group("/feedback")
		{
			feedback.GET("", listFeedback)
			feedback.GET("/:id", getFeedback)
			feedback.PUT("/:id/plan", updateImprovementPlan)
			feedback.PUT("/:id/status", updateFeedbackStatus)
		}
	}
}

func listTasks(c *gin.Context) {
	tasks, err := services.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var req services.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := services.CreateTask(req)
	if err != nil {
		if err.Error() == "该学期评估任务已存在" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func getTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := services.GetEvaluationTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func getTaskItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := services.GetEvaluationTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task.Courses)
}

func taskAction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	action := c.Param("action")
	if action != "approve" && action != "reject" && action != "cancel" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的操作"})
		return
	}

	userRole := c.GetHeader("X-User-Role")
	if userRole == "" {
		userRole = "admin"
	}

	if err := services.TaskAction(uint(id), action, userRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

func getQuestions(c *gin.Context) {
	courseType := models.CourseType(c.Param("course_type"))
	if courseType != models.CourseTypeTheory && courseType != models.CourseTypeExperiment && courseType != models.CourseTypeSport {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的课程类型"})
		return
	}

	questions, err := services.GetQuestionsForCourse(courseType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, questions)
}

func submitEvaluation(c *gin.Context) {
	var req services.EvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submission, err := services.SubmitEvaluation(req)
	if err != nil {
		switch err.Error() {
		case "您已对该课程完成评估":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case "评分必须在1到5之间":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case "评估尚未开始", "评估已结束", "评估任务未激活":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, submission)
}

func getResults(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	results, err := services.GetCourseResults(uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	c.JSON(http.StatusOK, results)
}

func calculateResults(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	results, err := services.CalculateStatistics(uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func exportResults(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	results, err := services.GetCourseResults(uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=statistics.json")
	c.JSON(http.StatusOK, results)
}

func getComments(c *gin.Context) {
	resultID, err := strconv.ParseUint(c.Param("result_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结果ID"})
		return
	}

	comments, err := services.GetComments(uint(resultID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func listFeedback(c *gin.Context) {
	var teacherID *uint
	if tid := c.Query("teacher_id"); tid != "" {
		id, err := strconv.ParseUint(tid, 10, 64)
		if err == nil {
			uid := uint(id)
			teacherID = &uid
		}
	}

	items, err := services.GetFeedbackItems(teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func getFeedback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反馈ID"})
		return
	}

	item, err := services.GetFeedbackItem(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "反馈不存在"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func updateImprovementPlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反馈ID"})
		return
	}

	var req services.UpdateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateImprovementPlan(uint(id), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func updateFeedbackStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的反馈ID"})
		return
	}

	var body struct {
		Status models.FeedbackStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateFeedbackStatus(uint(id), body.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}
