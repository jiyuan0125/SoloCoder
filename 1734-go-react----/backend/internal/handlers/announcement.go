package handlers

import (
	"net/http"
	"school-connect/internal/models"
	"school-connect/internal/storage"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnnouncementHandler struct {
	store *storage.MemoryStore
}

func NewAnnouncementHandler(store *storage.MemoryStore) *AnnouncementHandler {
	return &AnnouncementHandler{store: store}
}

type CreateAnnouncementRequest struct {
	Title      string         `json:"title" binding:"required"`
	Content    string         `json:"content" binding:"required"`
	ScopeType  models.ScopeType `json:"scope_type" binding:"required"`
	ScopeValue string        `json:"scope_value"`
	Urgency    models.Urgency   `json:"urgency" binding:"required"`
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(*models.User)

	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误: " + err.Error()})
		return
	}

	if req.ScopeType != models.ScopeAll && req.ScopeValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "发布范围不能为空"})
		return
	}

	ann := h.store.CreateAnnouncement(&models.Announcement{
		Title:       req.Title,
		Content:     req.Content,
		ScopeType:   req.ScopeType,
		ScopeValue:  req.ScopeValue,
		Urgency:     req.Urgency,
		PublishedBy: u.ID,
	})

	c.JSON(http.StatusOK, ann)
}

func (h *AnnouncementHandler) ListForParent(c *gin.Context) {
	userID := c.GetString("userID")
	announcements := h.store.GetAnnouncementsForParent(userID)

	result := []gin.H{}
	for _, ann := range announcements {
		stats := h.store.GetAnnouncementStats(ann.ID)
		isRead := h.store.IsAnnouncementRead(ann.ID, userID)

		author := h.store.GetUserByID(ann.PublishedBy)
		authorName := ""
		if author != nil {
			authorName = author.Name
		}

		result = append(result, gin.H{
			"announcement": ann,
			"stats":        stats,
			"isRead":       isRead,
			"authorName":   authorName,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnnouncementHandler) ListAll(c *gin.Context) {
	announcements := h.store.GetAllAnnouncements()

	result := []gin.H{}
	for _, ann := range announcements {
		stats := h.store.GetAnnouncementStats(ann.ID)

		author := h.store.GetUserByID(ann.PublishedBy)
		authorName := ""
		if author != nil {
			authorName = author.Name
		}

		result = append(result, gin.H{
			"announcement": ann,
			"stats":        stats,
			"authorName":   authorName,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnnouncementHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	ann := h.store.GetAnnouncement(id)
	if ann == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	userID := c.GetString("userID")
	user, _ := c.Get("user")
	u := user.(*models.User)

	if u.Role == models.RoleParent {
		parentAnns := h.store.GetAnnouncementsForParent(userID)
		accessible := false
		for _, a := range parentAnns {
			if a.ID == ann.ID {
				accessible = true
				break
			}
		}
		if !accessible {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此通知"})
			return
		}
	}

	stats := h.store.GetAnnouncementStats(id)
	author := h.store.GetUserByID(ann.PublishedBy)
	authorName := ""
	if author != nil {
		authorName = author.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"announcement": ann,
		"stats":        stats,
		"authorName":   authorName,
	})
}

func (h *AnnouncementHandler) MarkRead(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	ann := h.store.GetAnnouncement(id)
	if ann == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	h.store.MarkAnnouncementRead(id, userID)
	stats := h.store.GetAnnouncementStats(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

func (h *AnnouncementHandler) GetStats(c *gin.Context) {
	id := c.Param("id")
	stats := h.store.GetAnnouncementStats(id)
	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *AnnouncementHandler) GetStudentList(c *gin.Context) {
	class := c.Query("class")
	grade := c.Query("grade")

	var students []*models.Student
	if class != "" {
		students = h.store.GetStudentsByClass(class)
	} else if grade != "" {
		students = h.store.GetStudentsByGrade(grade)
	}

	result := []gin.H{}
	for _, s := range students {
		parentName := ""
		for _, p := range h.store.GetAllParents() {
			for _, cid := range p.ChildIDs {
				if cid == s.ID {
					if user := h.store.GetUserByID(p.UserID); user != nil {
						parentName = user.Name
					}
					break
				}
			}
		}
		result = append(result, gin.H{
			"student":    s,
			"parentName": parentName,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnnouncementHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("userID")
	announcements := h.store.GetAnnouncementsForParent(userID)

	unreadCount := 0
	for _, ann := range announcements {
		if !h.store.IsAnnouncementRead(ann.ID, userID) {
			unreadCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"count": unreadCount,
	})
}

func (h *AnnouncementHandler) GetClassesAndGrades(c *gin.Context) {
	gradesMap := make(map[string]bool)
	classesMap := make(map[string]bool)
	
	for _, s := range h.store.GetStudentsByGrade("") {
	}
	
	allStudents := append(h.store.GetStudentsByClass(""), h.store.GetStudentsByGrade("")...)
	
	for _, s := range allStudents {
		if s.Grade != "" {
			gradesMap[s.Grade] = true
		}
		if s.Class != "" {
			classesMap[s.Class] = true
		}
	}

	grades := []string{}
	for g := range gradesMap {
		grades = append(grades, g)
	}

	classes := []string{}
	for c := range classesMap {
		classes = append(classes, c)
	}

	c.JSON(http.StatusOK, gin.H{
		"grades":  grades,
		"classes": classes,
	})
}
