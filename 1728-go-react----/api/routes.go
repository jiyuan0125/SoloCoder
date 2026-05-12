package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lab-safety/config"
	"lab-safety/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
	}))

	api := r.Group("/api")
	{
		labs := api.Group("/labs")
		{
			labs.GET("", listLabs)
			labs.POST("", createLab)
			labs.GET("/:id", getLab)
			labs.PUT("/:id", updateLab)
			labs.DELETE("/:id", deleteLab)
		}

		chemicals := api.Group("/chemicals")
		{
			chemicals.GET("", listChemicals)
			chemicals.POST("", createChemical)
			chemicals.GET("/:id", getChemical)
			chemicals.PUT("/:id", updateChemical)
			chemicals.POST("/:id/use", useChemical)
		}

		trainings := api.Group("/trainings")
		{
			trainings.GET("", listTrainings)
			trainings.POST("", createTraining)
			trainings.GET("/:id", getTraining)
			trainings.PUT("/:id", updateTraining)
			trainings.POST("/:id/participants", addTrainingParticipant)
		}

		api.GET("/users/:id/qualification", checkUserQualification)

		checks := api.Group("/checks")
		{
			checks.GET("", listChecks)
			checks.POST("", createCheck)
			checks.GET("/:id", getCheck)
		}

		todos := api.Group("/todos")
		{
			todos.GET("", listTodos)
			todos.PUT("/:id", updateTodo)
		}

		audit := api.Group("/audit")
		{
			audit.GET("", listAuditLogs)
			audit.DELETE("/:id", deleteAuditLog)
		}

		departments := api.Group("/departments")
		{
			departments.GET("", listDepartments)
			departments.POST("", createDepartment)
		}

		reqs := api.Group("/r")
		{
			reqs.GET("", listRequests)
			reqs.POST("", createRequest)
			reqs.GET("/:id", getRequest)
			reqs.PUT("/:id", updateRequest)

			reqs.GET("/:id/items", listRequestItems)
			reqs.POST("/:id/items", addRequestItem)

			reqs.POST("/:id/actions/approve", approveRequest)
			reqs.POST("/:id/actions/reject", rejectRequest)
			reqs.POST("/:id/actions/cancel", cancelRequest)
		}

		alerts := api.Group("/alerts")
		{
			alerts.GET("", listAlerts)
		}

		api.GET("/users/:id/expense", getUserMonthExpense)
	}
}

func generateID() string {
	id := config.Storage.NextID
	config.Storage.NextID++
	return fmt.Sprintf("%d", id)
}

func getCurrentUser(ctx *gin.Context) string {
	user := ctx.GetHeader("X-User-Name")
	if user == "" {
		return "anonymous"
	}
	return user
}

func listLabs(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	result := make([]*models.Lab, 0, len(config.Storage.Labs))
	for _, lab := range config.Storage.Labs {
		result = append(result, lab)
	}

	ctx.JSON(http.StatusOK, result)
}

func createLab(ctx *gin.Context) {
	var lab models.Lab
	if err := ctx.ShouldBindJSON(&lab); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !config.ValidDangerLevel(lab.DangerLevel) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "危险等级必须是甲、乙、丙、丁之一"})
		return
	}

	config.Storage.Lock()
	lab.ID = generateID()
	config.Storage.Labs[lab.ID] = &lab
	config.Storage.Unlock()

	config.AddAuditLog("创建实验室", getCurrentUser(ctx), "实验室 ID: %s, 名称: %s", lab.ID, lab.Name)

	ctx.JSON(http.StatusCreated, lab)
}

func getLab(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.RLock()
	lab, exists := config.Storage.Labs[id]
	config.Storage.RUnlock()

	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "实验室不存在"})
		return
	}

	ctx.JSON(http.StatusOK, lab)
}

func updateLab(ctx *gin.Context) {
	id := ctx.Param("id")

	var lab models.Lab
	if err := ctx.ShouldBindJSON(&lab); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !config.ValidDangerLevel(lab.DangerLevel) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "危险等级必须是甲、乙、丙、丁之一"})
		return
	}

	config.Storage.Lock()
	_, exists := config.Storage.Labs[id]
	if !exists {
		config.Storage.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "实验室不存在"})
		return
	}

	lab.ID = id
	config.Storage.Labs[id] = &lab
	config.Storage.Unlock()

	config.AddAuditLog("更新实验室", getCurrentUser(ctx), "实验室 ID: %s", id)

	ctx.JSON(http.StatusOK, lab)
}

func deleteLab(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.Lock()
	_, exists := config.Storage.Labs[id]
	if !exists {
		config.Storage.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "实验室不存在"})
		return
	}

	delete(config.Storage.Labs, id)
	config.Storage.Unlock()

	config.AddAuditLog("删除实验室", getCurrentUser(ctx), "实验室 ID: %s", id)

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func listChemicals(ctx *gin.Context) {
	category := ctx.Query("category")
	labID := ctx.Query("lab_id")

	config.Storage.RLock()
	defer config.Storage.RUnlock()

	result := make([]*models.Chemical, 0)
	for _, c := range config.Storage.Chemicals {
		if category != "" && c.HazardCategory != category {
			continue
		}
		if labID != "" && c.StorageLabID != labID {
			continue
		}
		result = append(result, c)
	}

	ctx.JSON(http.StatusOK, result)
}

func createChemical(ctx *gin.Context) {
	var input struct {
		Name           string  `json:"name" binding:"required"`
		CAS            string  `json:"cas"`
		HazardCategory string  `json:"hazard_category" binding:"required"`
		Quantity       float64 `json:"quantity" binding:"required"`
		StorageLabID   string  `json:"storage_lab_id" binding:"required"`
		StorageCabinet string  `json:"storage_cabinet" binding:"required"`
		ExpiryDays     *int    `json:"expiry_days"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !config.ValidHazardCategory(input.HazardCategory) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "危险类别必须是：易燃、易爆、有毒、腐蚀、氧化、其他"})
		return
	}

	config.Storage.RLock()
	_, labExists := config.Storage.Labs[input.StorageLabID]
	config.Storage.RUnlock()

	if !labExists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "指定的实验室不存在"})
		return
	}

	now := time.Now()
	var expiry *time.Time
	if input.ExpiryDays != nil {
		e := now.AddDate(0, 0, *input.ExpiryDays)
		expiry = &e
	}

	chemical := &models.Chemical{
		ID:             generateID(),
		Name:           input.Name,
		CAS:            input.CAS,
		HazardCategory: input.HazardCategory,
		Quantity:       input.Quantity,
		StorageLabID:   input.StorageLabID,
		StorageCabinet: input.StorageCabinet,
		InboundDate:    now,
		ExpiryDate:     expiry,
		Status:         "在库",
	}

	config.Storage.Lock()
	config.Storage.Chemicals[chemical.ID] = chemical
	config.Storage.Unlock()

	config.AddAuditLog("危化品入库", getCurrentUser(ctx), "ID: %s, 名称: %s, 数量: %.2f", chemical.ID, chemical.Name, input.Quantity)

	ctx.JSON(http.StatusCreated, chemical)
}

func getChemical(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.RLock()
	c, exists := config.Storage.Chemicals[id]
	config.Storage.RUnlock()

	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "危化品不存在"})
		return
	}

	ctx.JSON(http.StatusOK, c)
}

func updateChemical(ctx *gin.Context) {
	id := ctx.Param("id")

	var chemical models.Chemical
	if err := ctx.ShouldBindJSON(&chemical); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	_, exists := config.Storage.Chemicals[id]
	if !exists {
		config.Storage.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "危化品不存在"})
		return
	}

	chemical.ID = id
	config.Storage.Chemicals[id] = &chemical
	config.Storage.Unlock()

	config.AddAuditLog("更新危化品", getCurrentUser(ctx), "ID: %s", id)

	ctx.JSON(http.StatusOK, chemical)
}

func useChemical(ctx *gin.Context) {
	id := ctx.Param("id")
	userName := ctx.GetHeader("X-User-Name")
	userID := ctx.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}
	if userName == "" {
		userName = "default-user"
	}

	var input struct {
		Quantity float64 `json:"quantity" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isUserQualified(userID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "培训证已过期或无有效培训记录，不能操作危化品"})
		return
	}

	config.Storage.Lock()
	defer config.Storage.Unlock()

	chemical, exists := config.Storage.Chemicals[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "危化品不存在"})
		return
	}

	if chemical.Quantity <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "库存为零，无法继续领用"})
		return
	}

	if input.Quantity > chemical.Quantity {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "领用数量超过库存"})
		return
	}

	chemical.Quantity -= input.Quantity

	if chemical.Quantity <= 0 {
		chemical.Status = "已用完"
	} else {
		chemical.Status = "领用中"
	}

	usage := &models.UsageRecord{
		ID:         generateID(),
		ChemicalID: id,
		UserID:     userID,
		UserName:   userName,
		Quantity:   input.Quantity,
		UsedAt:     time.Now(),
	}
	config.Storage.Usages[usage.ID] = usage

	config.AddAuditLog("危化品领用", userName, "危化品 ID: %s, 领用数量: %.2f", id, input.Quantity)

	ctx.JSON(http.StatusOK, gin.H{
		"chemical": chemical,
		"usage":    usage,
	})
}

func isUserQualified(userID string) bool {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	var latestDate time.Time
	found := false

	for _, t := range config.Storage.Trainings {
		for _, p := range t.Participants {
			if p.UserID == userID && p.Passed {
				if !found || t.Date.After(latestDate) {
					latestDate = t.Date
					found = true
				}
			}
		}
	}

	if !found {
		return false
	}

	validUntil := latestDate.AddDate(1, 0, 0)
	return validUntil.After(time.Now())
}

func listTrainings(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	result := make([]*models.Training, 0, len(config.Storage.Trainings))
	for _, t := range config.Storage.Trainings {
		result = append(result, t)
	}

	ctx.JSON(http.StatusOK, result)
}

func createTraining(ctx *gin.Context) {
	var input struct {
		Topic         string  `json:"topic" binding:"required"`
		Date          string  `json:"date" binding:"required"`
		DurationHours float64 `json:"duration_hours" binding:"required"`
		Instructor    string  `json:"instructor" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误，应为 YYYY-MM-DD"})
		return
	}

	training := &models.Training{
		ID:            generateID(),
		Topic:         input.Topic,
		Date:          date,
		DurationHours: input.DurationHours,
		Instructor:    input.Instructor,
		Participants:  []models.TrainingParticipant{},
	}

	config.Storage.Lock()
	config.Storage.Trainings[training.ID] = training
	config.Storage.Unlock()

	config.AddAuditLog("创建培训", getCurrentUser(ctx), "培训 ID: %s, 主题: %s", training.ID, training.Topic)

	ctx.JSON(http.StatusCreated, training)
}

func getTraining(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.RLock()
	t, exists := config.Storage.Trainings[id]
	config.Storage.RUnlock()

	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "培训不存在"})
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func updateTraining(ctx *gin.Context) {
	id := ctx.Param("id")

	var training models.Training
	if err := ctx.ShouldBindJSON(&training); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	_, exists := config.Storage.Trainings[id]
	if !exists {
		config.Storage.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "培训不存在"})
		return
	}

	training.ID = id
	config.Storage.Trainings[id] = &training
	config.Storage.Unlock()

	config.AddAuditLog("更新培训", getCurrentUser(ctx), "培训 ID: %s", id)

	ctx.JSON(http.StatusOK, training)
}

func addTrainingParticipant(ctx *gin.Context) {
	id := ctx.Param("id")

	var input struct {
		UserID   string  `json:"user_id" binding:"required"`
		UserName string  `json:"user_name" binding:"required"`
		Score    float64 `json:"score" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	defer config.Storage.Unlock()

	training, exists := config.Storage.Trainings[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "培训不存在"})
		return
	}

	for _, p := range training.Participants {
		if p.UserID == input.UserID {
			ctx.JSON(http.StatusConflict, gin.H{"error": "同一人同一场培训只能一条记录"})
			return
		}
	}

	participant := models.TrainingParticipant{
		UserID:   input.UserID,
		UserName: input.UserName,
		Score:    input.Score,
		Passed:   input.Score >= 60,
	}

	training.Participants = append(training.Participants, participant)

	config.AddAuditLog("添加培训参训人员", getCurrentUser(ctx), "培训 ID: %s, 用户: %s", id, input.UserName)

	ctx.JSON(http.StatusCreated, participant)
}

func checkUserQualification(ctx *gin.Context) {
	userID := ctx.Param("id")

	config.Storage.RLock()
	defer config.Storage.RUnlock()

	var latestTraining *models.Training
	found := false

	for _, t := range config.Storage.Trainings {
		for _, p := range t.Participants {
			if p.UserID == userID && p.Passed {
				if !found || t.Date.After(latestTraining.Date) {
					latestTraining = t
					found = true
				}
			}
		}
	}

	if !found {
		ctx.JSON(http.StatusOK, gin.H{
			"qualified": false,
			"message":   "无有效培训记录",
		})
		return
	}

	validUntil := latestTraining.Date.AddDate(1, 0, 0)
	qualified := validUntil.After(time.Now())

	ctx.JSON(http.StatusOK, gin.H{
		"qualified":      qualified,
		"last_training":  latestTraining.Date.Format("2006-01-02"),
		"valid_until":    validUntil.Format("2006-01-02"),
		"topic":          latestTraining.Topic,
	})
}

func listChecks(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	result := make([]*models.SafetyCheck, 0, len(config.Storage.Checks))
	for _, c := range config.Storage.Checks {
		result = append(result, c)
	}

	ctx.JSON(http.StatusOK, result)
}

func createCheck(ctx *gin.Context) {
	var input struct {
		LabID     string `json:"lab_id" binding:"required"`
		Inspector string `json:"inspector" binding:"required"`
		Items     []struct {
			Item   string `json:"item" binding:"required"`
			Result string `json:"result" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validItems := map[string]bool{
		"消防设施": true, "通风系统": true, "防护装备": true, "危化品存放": true, "应急预案": true,
	}
	validResults := map[string]bool{
		"合格": true, "不合格": true, "待整改": true,
	}

	for _, item := range input.Items {
		if !validItems[item.Item] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "检查项目必须是：消防设施、通风系统、防护装备、危化品存放、应急预案"})
			return
		}
		if !validResults[item.Result] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "检查结果必须是：合格、不合格、待整改"})
			return
		}
	}

	config.Storage.RLock()
	lab, labExists := config.Storage.Labs[input.LabID]
	config.Storage.RUnlock()

	if !labExists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "指定的实验室不存在"})
		return
	}

	items := make([]models.SafetyCheckItem, len(input.Items))
	for i, item := range input.Items {
		items[i] = models.SafetyCheckItem{
			Item:   item.Item,
			Result: item.Result,
		}
	}

	check := &models.SafetyCheck{
		ID:        generateID(),
		LabID:     input.LabID,
		Inspector: input.Inspector,
		CheckDate: time.Now(),
		Items:     items,
	}

	config.Storage.Lock()
	config.Storage.Checks[check.ID] = check

	hasUnqualified := false
	for _, item := range check.Items {
		if item.Result == "不合格" {
			hasUnqualified = true
		}
	}

	if hasUnqualified {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "检查不合格项自动生成待办，请指定责任人和截止日期"})
		config.Storage.Unlock()
		return
	}

	config.Storage.Unlock()

	config.AddAuditLog("创建安全检查", getCurrentUser(ctx), "检查 ID: %s, 实验室: %s", check.ID, lab.Name)

	ctx.JSON(http.StatusCreated, check)
}

func getCheck(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.RLock()
	c, exists := config.Storage.Checks[id]
	config.Storage.RUnlock()

	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "检查记录不存在"})
		return
	}

	ctx.JSON(http.StatusOK, c)
}

func listTodos(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	now := time.Now()
	result := make([]*models.Todo, 0, len(config.Storage.Todos))
	for _, t := range config.Storage.Todos {
		if t.DueDate != nil && t.Status != "已完成" && t.DueDate.Before(now) {
			t.Overdue = true
		}
		result = append(result, t)
	}

	ctx.JSON(http.StatusOK, result)
}

func updateTodo(ctx *gin.Context) {
	id := ctx.Param("id")

	var input struct {
		Status      string `json:"status"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	defer config.Storage.Unlock()

	todo, exists := config.Storage.Todos[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "待办不存在"})
		return
	}

	if input.Status != "" {
		todo.Status = input.Status
	}
	if input.Description != "" {
		todo.Description = input.Description
	}

	config.AddAuditLog("更新待办事项", getCurrentUser(ctx), "待办 ID: %s, 新状态: %s", id, input.Status)

	ctx.JSON(http.StatusOK, todo)
}

func listAuditLogs(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	result := make([]*models.AuditLog, 0, len(config.Storage.AuditLogs))
	for _, log := range config.Storage.AuditLogs {
		result = append(result, log)
	}

	ctx.JSON(http.StatusOK, result)
}

func deleteAuditLog(ctx *gin.Context) {
	ctx.JSON(http.StatusMethodNotAllowed, gin.H{"error": "审计日志不可删改"})
}

func listDepartments(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	result := make([]*models.Department, 0, len(config.Storage.Departments))
	for _, d := range config.Storage.Departments {
		result = append(result, d)
	}

	ctx.JSON(http.StatusOK, result)
}

func createDepartment(ctx *gin.Context) {
	var dept models.Department
	if err := ctx.ShouldBindJSON(&dept); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	dept.ID = generateID()
	config.Storage.Departments[dept.ID] = &dept
	config.Storage.Unlock()

	config.AddAuditLog("创建部门", getCurrentUser(ctx), "部门 ID: %s, 名称: %s", dept.ID, dept.Name)

	ctx.JSON(http.StatusCreated, dept)
}

var alerts []*models.Alert

func listRequests(ctx *gin.Context) {
	config.Storage.RLock()
	defer config.Storage.RUnlock()

	now := time.Now()
	for _, req := range config.Storage.Requests {
		if req.Status == "待接单" && req.CreatedAt.Add(48*time.Hour).Before(now) {
			req.Status = "已关闭"
			closedAt := now
			req.ClosedAt = &closedAt
		}
	}

	result := make([]*models.Request, 0, len(config.Storage.Requests))
	for _, req := range config.Storage.Requests {
		result = append(result, req)
	}

	ctx.JSON(http.StatusOK, result)
}

func createRequest(ctx *gin.Context) {
	var input struct {
		Title        string `json:"title" binding:"required"`
		DepartmentID string `json:"department_id" binding:"required"`
		Applicant    string `json:"applicant" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.RLock()
	_, deptExists := config.Storage.Departments[input.DepartmentID]
	config.Storage.RUnlock()

	if !deptExists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "部门不存在"})
		return
	}

	req := &models.Request{
		ID:           generateID(),
		Title:        input.Title,
		DepartmentID: input.DepartmentID,
		Applicant:    input.Applicant,
		Status:       "待接单",
		CreatedAt:    time.Now(),
		Items:        []models.RequestItem{},
	}

	config.Storage.Lock()
	config.Storage.Requests[req.ID] = req
	config.Storage.Unlock()

	config.AddAuditLog("创建申请", getCurrentUser(ctx), "申请 ID: %s", req.ID)

	ctx.JSON(http.StatusCreated, req)
}

func getRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.RLock()
	req, exists := config.Storage.Requests[id]
	config.Storage.RUnlock()

	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	ctx.JSON(http.StatusOK, req)
}

func updateRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	var input struct {
		Status string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	defer config.Storage.Unlock()

	req, exists := config.Storage.Requests[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	now := time.Now()
	allowed := false

	switch req.Status {
	case "待接单":
		if input.Status == "已接单" {
			req.AssignedAt = &now
			allowed = true
		}
	case "已接单":
		if input.Status == "处理中" {
			allowed = true
		}
	case "处理中":
		if input.Status == "待验收" {
			req.ProcessedAt = &now
			allowed = true
		}
	case "待验收":
		if input.Status == "已完成" {
			req.CompletedAt = &now
			allowed = true
		} else if input.Status == "处理中" {
			allowed = true
		}
	}

	if !allowed {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "状态转换不允许"})
		return
	}

	req.Status = input.Status
	config.AddAuditLog("更新申请状态", getCurrentUser(ctx), "申请 ID: %s, 新状态: %s", id, input.Status)

	ctx.JSON(http.StatusOK, req)
}

func listRequestItems(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.RLock()
	req, exists := config.Storage.Requests[id]
	config.Storage.RUnlock()

	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	ctx.JSON(http.StatusOK, req.Items)
}

func addRequestItem(ctx *gin.Context) {
	id := ctx.Param("id")

	var item models.RequestItem
	if err := ctx.ShouldBindJSON(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Storage.Lock()
	defer config.Storage.Unlock()

	req, exists := config.Storage.Requests[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	item.ID = generateID()
	req.Items = append(req.Items, item)
	req.TotalAmount += item.Quantity * item.UnitPrice

	config.AddAuditLog("添加申请项目", getCurrentUser(ctx), "申请 ID: %s, 项目: %s", id, item.Name)

	ctx.JSON(http.StatusCreated, item)
}

func approveRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.Lock()
	defer config.Storage.Unlock()

	req, exists := config.Storage.Requests[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	dept, deptExists := config.Storage.Departments[req.DepartmentID]
	if !deptExists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "部门不存在"})
		return
	}

	now := time.Now()
	year, month, _ := now.Date()
	monthKey := fmt.Sprintf("%d-%02d", year, month)

	monthExpense := 0.0
	for _, r := range config.Storage.Requests {
		if r.DepartmentID == req.DepartmentID && r.Status == "已完成" {
			if r.CompletedAt != nil {
				ry, rm, _ := r.CompletedAt.Date()
				rKey := fmt.Sprintf("%d-%02d", ry, rm)
				if rKey == monthKey {
					monthExpense += r.TotalAmount
				}
			}
		}
	}

	newTotal := monthExpense + req.TotalAmount
	threshold := dept.MonthlyBudget * 0.8

	if newTotal > threshold {
		alert := &models.Alert{
			ID:           generateID(),
			DepartmentID: dept.ID,
			Message:      fmt.Sprintf("本月累计支出已超过预算80%%，预算: %.2f, 当前: %.2f", dept.MonthlyBudget, newTotal),
			Month:        monthKey,
			CreatedAt:    now,
		}
		alerts = append(alerts, alert)
	}

	now2 := now
	req.CompletedAt = &now2
	req.Status = "已完成"

	config.AddAuditLog("审批通过", getCurrentUser(ctx), "申请 ID: %s, 金额: %.2f", id, req.TotalAmount)

	ctx.JSON(http.StatusOK, req)
}

func rejectRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.Lock()
	defer config.Storage.Unlock()

	req, exists := config.Storage.Requests[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	req.Status = "已拒绝"

	config.AddAuditLog("审批拒绝", getCurrentUser(ctx), "申请 ID: %s", id)

	ctx.JSON(http.StatusOK, req)
}

func cancelRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	config.Storage.Lock()
	defer config.Storage.Unlock()

	req, exists := config.Storage.Requests[id]
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "申请不存在"})
		return
	}

	req.Status = "已取消"
	now := time.Now()
	req.ClosedAt = &now

	config.AddAuditLog("取消申请", getCurrentUser(ctx), "申请 ID: %s", id)

	ctx.JSON(http.StatusOK, req)
}

func listAlerts(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, alerts)
}

func getUserMonthExpense(ctx *gin.Context) {
	userID := ctx.Param("id")
	month := ctx.DefaultQuery("month", time.Now().Format("2006-01"))

	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "月份格式应为 YYYY-MM"})
		return
	}
	year, _ := strconv.Atoi(parts[0])
	monthNum, _ := strconv.Atoi(parts[1])

	config.Storage.RLock()
	defer config.Storage.RUnlock()

	total := 0.0
	for _, r := range config.Storage.Requests {
		if r.Applicant == userID && r.Status == "已完成" && r.CompletedAt != nil {
			ry, rm, _ := r.CompletedAt.Date()
			if ry == year && int(rm) == monthNum {
				total += r.TotalAmount
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"month":   month,
		"total":   total,
	})
}
