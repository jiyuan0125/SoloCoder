package handlers

import (
	"net/http"
	"research-collaboration/src/models"
	"research-collaboration/src/storage"
	"time"

	"github.com/gin-gonic/gin"
)

type AchievementHandler struct {
	store *storage.Storage
}

func NewAchievementHandler(store *storage.Storage) *AchievementHandler {
	return &AchievementHandler{store: store}
}

func (h *AchievementHandler) CreateAchievement(c *gin.Context) {
	var req struct {
		Type          models.AchievementType       `json:"type" binding:"required"`
		Title         string                       `json:"title" binding:"required"`
		OutputDate    time.Time                    `json:"output_date" binding:"required"`
		Participants  []string                     `json:"participants"`
		Status        models.AchievementStatus     `json:"status"`
		Contributions []models.ProjectContribution `json:"contributions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validateContributions(h.store, req.Contributions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contribution ratios must sum to exactly 100%"})
		return
	}

	if req.Status == "" {
		req.Status = models.AchievementStatusSubmitted
	}

	achievement := &models.Achievement{
		Type:          req.Type,
		Title:         req.Title,
		OutputDate:    req.OutputDate,
		Participants:  req.Participants,
		Status:        req.Status,
		Contributions: req.Contributions,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	h.store.CreateAchievement(achievement)
	c.JSON(http.StatusCreated, achievement)
}

func (h *AchievementHandler) GetAchievement(c *gin.Context) {
	id := c.Param("id")
	ach, ok := h.store.GetAchievement(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "achievement not found"})
		return
	}
	c.JSON(http.StatusOK, ach)
}

func (h *AchievementHandler) ListAchievements(c *gin.Context) {
	achievements := h.store.GetAllAchievements()
	c.JSON(http.StatusOK, achievements)
}

func (h *AchievementHandler) UpdateAchievement(c *gin.Context) {
	id := c.Param("id")
	ach, ok := h.store.GetAchievement(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "achievement not found"})
		return
	}

	var req struct {
		Type          *models.AchievementType       `json:"type"`
		Title         *string                       `json:"title"`
		OutputDate    *time.Time                    `json:"output_date"`
		Participants  *[]string                     `json:"participants"`
		Status        *models.AchievementStatus     `json:"status"`
		Contributions *[]models.ProjectContribution `json:"contributions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type != nil {
		ach.Type = *req.Type
	}
	if req.Title != nil {
		ach.Title = *req.Title
	}
	if req.OutputDate != nil {
		ach.OutputDate = *req.OutputDate
	}
	if req.Participants != nil {
		ach.Participants = *req.Participants
	}
	if req.Status != nil {
		ach.Status = *req.Status
	}
	if req.Contributions != nil {
		if !validateContributions(h.store, *req.Contributions) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "contribution ratios must sum to exactly 100%"})
			return
		}
		ach.Contributions = *req.Contributions
	}
	ach.UpdatedAt = time.Now()

	h.store.UpdateAchievement(ach)
	c.JSON(http.StatusOK, ach)
}

func validateContributions(store *storage.Storage, contribs []models.ProjectContribution) bool {
	if len(contribs) == 0 {
		return true
	}
	sum := 0
	for _, c := range contribs {
		if c.Ratio <= 0 {
			return false
		}
		if _, ok := store.GetProject(c.ProjectID); !ok {
			return false
		}
		sum += c.Ratio
	}
	return sum == 100
}

func (h *AchievementHandler) ExportAchievements(c *gin.Context) {
	achievements := h.store.GetAllAchievements()

	type ExportItem struct {
		ID           string    `json:"id"`
		Type         string    `json:"type"`
		Title        string    `json:"title"`
		OutputDate   time.Time `json:"output_date"`
		Status       string    `json:"status"`
		Participants string    `json:"participants"`
		Projects     string    `json:"projects"`
	}

	exportList := make([]ExportItem, 0, len(achievements))
	for _, a := range achievements {
		projNames := ""
		for i, c := range a.Contributions {
			if i > 0 {
				projNames += "; "
			}
			if proj, ok := h.store.GetProject(c.ProjectID); ok {
				projNames += proj.Name + "(" + string(rune('0'+c.Ratio/10)) + "%)"
			}
		}
		exportList = append(exportList, ExportItem{
			ID:           a.ID,
			Type:         string(a.Type),
			Title:        a.Title,
			OutputDate:   a.OutputDate,
			Status:       string(a.Status),
			Participants: joinStrings(a.Participants, "; "),
			Projects:     projNames,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       exportList,
		"exported_at": time.Now(),
	})
}

func joinStrings(arr []string, sep string) string {
	result := ""
	for i, s := range arr {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
