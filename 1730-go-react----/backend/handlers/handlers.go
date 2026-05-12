package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"diploma-auth-system/middleware"
	"diploma-auth-system/models"
	"diploma-auth-system/store"
	"diploma-auth-system/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数错误",
		})
		return
	}
	
	user, ok := h.store.Authenticate(req.Username, req.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "用户名或密码错误",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"token": user.ID,
	})
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	c.JSON(http.StatusOK, user)
}

func (h *Handler) ListDiplomas(c *gin.Context) {
	diplomas := h.store.ListDiplomas()
	c.JSON(http.StatusOK, diplomas)
}

func (h *Handler) GetDiploma(c *gin.Context) {
	id := c.Param("id")
	diploma, exists := h.store.GetDiploma(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "学历信息不存在",
		})
		return
	}
	c.JSON(http.StatusOK, diploma)
}

func (h *Handler) CreateDiploma(c *gin.Context) {
	var req models.CreateDiplomaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数错误",
		})
		return
	}
	
	if err := utils.ValidateIDCard(req.IDCard); err != nil {
		if idErr, ok := err.(*utils.IDCardError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_id_card",
				Message: idErr.Message,
			})
			return
		}
	}
	
	if h.store.CheckDuplicate(req.Name, req.IDCard) {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "duplicate",
			Message: "可能重复录入：姓名和身份证号已存在",
		})
		return
	}
	
	diploma, err := h.store.CreateDiploma(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "创建失败",
		})
		return
	}
	
	user := middleware.GetCurrentUser(c)
	content, _ := json.Marshal(diploma)
	h.store.AddLog(user, models.OpCreateDiploma, string(content), nil)
	
	c.JSON(http.StatusCreated, diploma)
}

func (h *Handler) UpdateDiploma(c *gin.Context) {
	id := c.Param("id")
	
	oldDiploma, exists := h.store.GetDiploma(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "学历信息不存在",
		})
		return
	}
	
	var req models.UpdateDiplomaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数错误",
		})
		return
	}
	
	if req.IDCard != nil {
		if err := utils.ValidateIDCard(*req.IDCard); err != nil {
			if idErr, ok := err.(*utils.IDCardError); ok {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   "invalid_id_card",
					Message: idErr.Message,
				})
				return
			}
		}
	}
	
	checkName := oldDiploma.Name
	if req.Name != nil {
		checkName = *req.Name
	}
	checkIDCard := oldDiploma.IDCard
	if req.IDCard != nil {
		checkIDCard = *req.IDCard
	}
	
	if h.store.CheckDuplicateExclude(id, checkName, checkIDCard) {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "duplicate",
			Message: "可能重复录入：姓名和身份证号已存在",
		})
		return
	}
	
	diploma, err := h.store.UpdateDiploma(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "更新失败",
		})
		return
	}
	
	user := middleware.GetCurrentUser(c)
	changes := h.store.LogDiplomaChanges(oldDiploma, diploma)
	h.store.AddLog(user, models.OpUpdateDiploma, changes, nil)
	
	c.JSON(http.StatusOK, diploma)
}

func (h *Handler) DeleteDiploma(c *gin.Context) {
	id := c.Param("id")
	
	diploma, exists := h.store.GetDiploma(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "学历信息不存在",
		})
		return
	}
	
	if err := h.store.DeleteDiploma(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "删除失败",
		})
		return
	}
	
	user := middleware.GetCurrentUser(c)
	content, _ := json.Marshal(diploma)
	h.store.AddLog(user, models.OpDeleteDiploma, string(content), nil)
	
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) VerifyDiploma(c *gin.Context) {
	var req models.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数错误",
		})
		return
	}
	
	if err := utils.ValidateIDCard(req.IDCard); err != nil {
		if idErr, ok := err.(*utils.IDCardError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_id_card",
				Message: idErr.Message,
			})
			return
		}
	}
	
	diplomas := h.store.VerifyDiploma(req.Name, req.IDCard)
	
	user := middleware.GetCurrentUser(c)
	content, _ := json.Marshal(req)
	
	var result models.VerificationResult
	if len(diplomas) > 0 {
		result = models.ResultMatched
	} else {
		result = models.ResultUnmatched
	}
	
	h.store.AddLog(user, models.OpVerifyDiploma, string(content), &result)
	
	if len(diplomas) > 0 {
		c.JSON(http.StatusOK, models.VerifyResponse{
			Diplomas: diplomas,
			Result:   "一致",
		})
	} else {
		c.JSON(http.StatusOK, models.VerifyResponse{
			Diplomas: []models.Diploma{},
			Result:   "未查到",
		})
	}
}

func (h *Handler) ListUsers(c *gin.Context) {
	users := h.store.ListUsers()
	c.JSON(http.StatusOK, users)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数错误",
		})
		return
	}
	
	if req.Role != models.RoleAdmin && req.Role != models.RoleVerifier && req.Role != models.RoleViewer {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role",
			Message: "角色类型错误",
		})
		return
	}
	
	user, err := h.store.CreateUser(req)
	if err != nil {
		if _, ok := err.(*store.DuplicateError); ok {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "duplicate",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "创建失败",
		})
		return
	}
	
	currentUser := middleware.GetCurrentUser(c)
	content, _ := json.Marshal(user)
	h.store.AddLog(currentUser, models.OpCreateUser, string(content), nil)
	
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数错误",
		})
		return
	}
	
	if req.Role != nil {
		if *req.Role != models.RoleAdmin && *req.Role != models.RoleVerifier && *req.Role != models.RoleViewer {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_role",
				Message: "角色类型错误",
			})
			return
		}
	}
	
	user, err := h.store.UpdateUser(id, req)
	if err != nil {
		if _, ok := err.(*store.NotFoundError); ok {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "更新失败",
		})
		return
	}
	
	currentUser := middleware.GetCurrentUser(c)
	content, _ := json.Marshal(user)
	h.store.AddLog(currentUser, models.OpUpdateUser, string(content), nil)
	
	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	
	user, exists := h.store.GetUser(id)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "用户不存在",
		})
		return
	}
	
	if err := h.store.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "删除失败",
		})
		return
	}
	
	currentUser := middleware.GetCurrentUser(c)
	content, _ := json.Marshal(user)
	h.store.AddLog(currentUser, models.OpDeleteUser, string(content), nil)
	
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (h *Handler) ListLogs(c *gin.Context) {
	filter := models.LogFilter{
		PageSize: 20,
	}
	
	pageStr := c.Query("page")
	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_page",
				Message: "页码必须为大于等于1的整数",
			})
			return
		}
		filter.Page = page
	} else {
		filter.Page = 1
	}
	
	pageSizeStr := c.Query("page_size")
	if pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize != 20 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_page_size",
				Message: "每页条数必须为20",
			})
			return
		}
		filter.PageSize = pageSize
	}
	
	opTypeStr := c.Query("operation_type")
	if opTypeStr != "" {
		opType := models.OperationType(opTypeStr)
		filter.OperationType = &opType
	}
	
	startDateStr := c.Query("start_date")
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_start_date",
				Message: "开始日期格式错误，应为YYYY-MM-DD",
			})
			return
		}
		startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		filter.StartDate = &startOfDay
	}
	
	endDateStr := c.Query("end_date")
	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_end_date",
				Message: "结束日期格式错误，应为YYYY-MM-DD",
			})
			return
		}
		endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
		filter.EndDate = &endOfDay
	}
	
	response, err := h.store.ListLogs(filter)
	if err != nil {
		if _, ok := err.(*store.ValidationError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_filter",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "查询失败",
		})
		return
	}
	
	c.JSON(http.StatusOK, response)
}
