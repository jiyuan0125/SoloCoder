package handlers

import (
	"net/http"
	"school-connect/internal/models"
	"school-connect/internal/storage"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	store *storage.MemoryStore
}

func NewAuthHandler(store *storage.MemoryStore) *AuthHandler {
	return &AuthHandler{store: store}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	user := h.store.GetUserByUsername(req.Username)
	if user == nil || user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": user.ID,
		"user":  user,
	})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *AuthHandler) GetCurrentUserInfo(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	u := user.(*models.User)
	response := gin.H{"user": u}

	switch u.Role {
	case models.RoleParent:
		parent := h.store.GetParent(u.ID)
		if parent != nil {
			children := []*models.Student{}
			for _, childID := range parent.ChildIDs {
				if child := h.store.GetStudent(childID); child != nil {
					children = append(children, child)
				}
			}
			response["parentInfo"] = parent
			response["children"] = children
		}
	case models.RoleTeacher:
		if teacher := h.store.GetTeacher(u.ID); teacher != nil {
			response["teacherInfo"] = teacher
		}
	}

	totalUnread := h.store.GetTotalUnreadCount(u.ID)
	response["unreadCount"] = totalUnread

	c.JSON(http.StatusOK, response)
}
