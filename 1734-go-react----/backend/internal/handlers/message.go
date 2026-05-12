package handlers

import (
	"net/http"
	"school-connect/internal/models"
	"school-connect/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	store *storage.MemoryStore
}

func NewMessageHandler(store *storage.MemoryStore) *MessageHandler {
	return &MessageHandler{store: store}
}

type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

func (h *MessageHandler) Send(c *gin.Context) {
	userID := c.GetString("userID")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if req.ReceiverID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能给自己发送消息"})
		return
	}

	receiver := h.store.GetUserByID(req.ReceiverID)
	if receiver == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "接收者不存在"})
		return
	}

	msg := h.store.SendMessage(&models.Message{
		SenderID:   userID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})

	c.JSON(http.StatusOK, msg)
}

func (h *MessageHandler) GetConversations(c *gin.Context) {
	userID := c.GetString("userID")
	conversations := h.store.GetConversations(userID)
	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("conversationID")
	userID := c.GetString("userID")

	markAsRead := c.Query("mark_read") != "false"

	messages := h.store.GetMessages(conversationID, userID, markAsRead)
	if messages == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) MarkRead(c *gin.Context) {
	conversationID := c.Param("conversationID")
	userID := c.GetString("userID")

	h.store.MarkConversationRead(conversationID, userID)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *MessageHandler) GetTotalUnread(c *gin.Context) {
	userID := c.GetString("userID")
	count := h.store.GetTotalUnreadCount(userID)
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *MessageHandler) GetContactableUsers(c *gin.Context) {
	userID := c.GetString("userID")
	user, _ := c.Get("user")
	u := user.(*models.User)

	var contacts []gin.H

	switch u.Role {
	case models.RoleTeacher:
		teacher := h.store.GetTeacher(u.ID)
		studentIDs := make(map[string]bool)
		for _, classID := range teacher.ClassIDs {
			for _, s := range h.store.GetStudentsByClass(classID) {
				studentIDs[s.ID] = true
			}
		}

		for _, p := range h.store.GetAllParents() {
			hasChild := false
			for _, cid := range p.ChildIDs {
				if studentIDs[cid] {
					hasChild = true
					break
				}
			}
			if hasChild {
				if parentUser := h.store.GetUserByID(p.UserID); parentUser != nil {
					childNames := []string{}
					for _, cid := range p.ChildIDs {
						if studentIDs[cid] {
							if child := h.store.GetStudent(cid); child != nil {
								childNames = append(childNames, child.Name)
							}
						}
					}
					contacts = append(contacts, gin.H{
						"user":       parentUser,
						"childNames": childNames,
					})
				}
			}
		}

	case models.RoleParent:
		parent := h.store.GetParent(u.ID)
		classTeachers := make(map[string]*models.Teacher)

		for _, childID := range parent.ChildIDs {
			if child := h.store.GetStudent(childID); child != nil {
				for _, t := range h.store.GetAllTeachers() {
					for _, cid := range t.ClassIDs {
						if cid == child.Class {
							classTeachers[t.UserID] = t
							break
						}
					}
				}
			}
		}

		for _, t := range classTeachers {
			if teacherUser := h.store.GetUserByID(t.UserID); teacherUser != nil {
				contacts = append(contacts, gin.H{
					"user":    teacherUser,
					"subject": t.Subject,
				})
			}
		}

	case models.RoleAdmin:
		for _, t := range h.store.GetAllTeachers() {
			if teacherUser := h.store.GetUserByID(t.UserID); teacherUser != nil {
				contacts = append(contacts, gin.H{
					"user":    teacherUser,
					"subject": t.Subject,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"contacts": contacts})
}
