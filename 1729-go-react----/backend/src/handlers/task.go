package handlers

import (
	"net/http"
	"research-collaboration/src/models"
	"research-collaboration/src/storage"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	store *storage.Storage
}

func NewTaskHandler(store *storage.Storage) *TaskHandler {
	return &TaskHandler{store: store}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		ProjectID  string             `json:"project_id" binding:"required"`
		Title      string             `json:"title" binding:"required"`
		AssigneeID string             `json:"assignee_id" binding:"required"`
		DueDate    time.Time          `json:"due_date" binding:"required"`
		Priority   models.TaskPriority `json:"priority" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, ok := h.store.GetProject(req.ProjectID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if project.Status != models.ProjectStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only in-progress projects can have tasks"})
		return
	}

	if !h.store.IsInProjectDepartment(req.AssigneeID, req.ProjectID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assignee must be in project department"})
		return
	}

	task := &models.Task{
		ProjectID:  req.ProjectID,
		Title:      req.Title,
		AssigneeID: req.AssigneeID,
		DueDate:    req.DueDate,
		Priority:   req.Priority,
		Status:     models.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	h.store.CreateTask(task)
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, ok := h.store.GetTask(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) GetTasksByProject(c *gin.Context) {
	projectID := c.Param("projectId")
	_, ok := h.store.GetProject(projectID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	tasks := h.store.GetTasksByProject(projectID)
	sortTasks(tasks)
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks := h.store.GetAllTasks()
	sortTasks(tasks)
	c.JSON(http.StatusOK, tasks)
}

func sortTasks(tasks []*models.Task) {
	priorityOrder := map[models.TaskPriority]int{
		models.TaskPriorityHigh:   3,
		models.TaskPriorityMedium: 2,
		models.TaskPriorityLow:    1,
	}

	sort.Slice(tasks, func(i, j int) bool {
		if priorityOrder[tasks[i].Priority] != priorityOrder[tasks[j].Priority] {
			return priorityOrder[tasks[i].Priority] > priorityOrder[tasks[j].Priority]
		}
		return tasks[i].DueDate.Before(tasks[j].DueDate)
	})
}

func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	task, ok := h.store.GetTask(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var req struct {
		Status models.TaskStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidStatusTransition(task.Status, req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
		return
	}

	task.Status = req.Status
	task.UpdatedAt = time.Now()

	h.store.UpdateTask(task)
	c.JSON(http.StatusOK, task)
}

func isValidStatusTransition(from, to models.TaskStatus) bool {
	switch from {
	case models.TaskStatusPending:
		return to == models.TaskStatusInProgress
	case models.TaskStatusInProgress:
		return to == models.TaskStatusCompleted || to == models.TaskStatusCancelled
	case models.TaskStatusCompleted, models.TaskStatusCancelled:
		return false
	default:
		return false
	}
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	task, ok := h.store.GetTask(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var req struct {
		Title      string              `json:"title"`
		AssigneeID string              `json:"assignee_id"`
		DueDate    *time.Time          `json:"due_date"`
		Priority   *models.TaskPriority `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.AssigneeID != "" {
		if !h.store.IsInProjectDepartment(req.AssigneeID, task.ProjectID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "assignee must be in project department"})
			return
		}
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	task.UpdatedAt = time.Now()

	h.store.UpdateTask(task)
	c.JSON(http.StatusOK, task)
}
