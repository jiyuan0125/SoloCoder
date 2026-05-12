package handlers

import (
	"net/http"
	"research-collaboration/src/models"
	"research-collaboration/src/storage"
	"time"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	store *storage.Storage
}

func NewProjectHandler(store *storage.Storage) *ProjectHandler {
	return &ProjectHandler{store: store}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name              string    `json:"name" binding:"required"`
		LeadDepartment    string    `json:"lead_department" binding:"required"`
		ParticipatingDepts []string `json:"participating_depts"`
		LeaderID          string    `json:"leader_id" binding:"required"`
		StartDate         time.Time `json:"start_date" binding:"required"`
		EndDate           time.Time `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := &models.Project{
		Name:              req.Name,
		LeadDepartment:    req.LeadDepartment,
		ParticipatingDepts: req.ParticipatingDepts,
		LeaderID:          req.LeaderID,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		Status:            models.ProjectStatusPreparing,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	h.store.CreateProject(project)
	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	project, ok := h.store.GetProject(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects := h.store.GetAllProjects()
	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	project, ok := h.store.GetProject(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var req struct {
		Name              string     `json:"name"`
		LeadDepartment    string     `json:"lead_department"`
		ParticipatingDepts []string  `json:"participating_depts"`
		LeaderID          string     `json:"leader_id"`
		StartDate         *time.Time `json:"start_date"`
		EndDate           *time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.LeadDepartment != "" {
		project.LeadDepartment = req.LeadDepartment
	}
	if req.ParticipatingDepts != nil {
		project.ParticipatingDepts = req.ParticipatingDepts
	}
	if req.LeaderID != "" {
		project.LeaderID = req.LeaderID
	}
	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		project.EndDate = *req.EndDate
	}
	project.UpdatedAt = time.Now()

	h.store.UpdateProject(project)
	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) UpdateProjectStatus(c *gin.Context) {
	id := c.Param("id")
	project, ok := h.store.GetProject(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var req struct {
		Status models.ProjectStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status == models.ProjectStatusCompleted {
		tasks := h.store.GetTasksByProject(id)
		for _, t := range tasks {
			if t.Status != models.TaskStatusCompleted && t.Status != models.TaskStatusCancelled {
				c.JSON(http.StatusBadRequest, gin.H{"error": "project has incomplete tasks"})
				return
			}
		}
	}

	project.Status = req.Status
	project.UpdatedAt = time.Now()

	h.store.UpdateProject(project)
	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	_, ok := h.store.GetProject(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	h.store.DeleteTasksByProject(id)
	h.store.DeleteAchievementContribution(id)
	h.store.DeleteProject(id)

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}
