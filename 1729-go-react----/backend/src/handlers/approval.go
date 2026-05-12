package handlers

import (
	"net/http"
	"research-collaboration/src/models"
	"research-collaboration/src/storage"
	"time"

	"github.com/gin-gonic/gin"
)

const smallAmountThreshold = 10000

type ApprovalHandler struct {
	store *storage.Storage
}

func NewApprovalHandler(store *storage.Storage) *ApprovalHandler {
	return &ApprovalHandler{store: store}
}

func (h *ApprovalHandler) CreateApproval(c *gin.Context) {
	var req struct {
		ReferenceID string  `json:"reference_id" binding:"required"`
		Type        string  `json:"type" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	initialStatus := models.ApprovalStatusReview1
	if req.Amount < smallAmountThreshold {
		initialStatus = models.ApprovalStatusFinal
	}

	now := time.Now()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "user-1"
	}

	approval := &models.Approval{
		ReferenceID: req.ReferenceID,
		Type:        req.Type,
		Status:      initialStatus,
		Amount:      req.Amount,
		History: []models.ApprovalStep{
			{
				Status:   models.ApprovalStatusSubmitted,
				ActionBy: userID,
				ActionAt: now,
			},
			{
				Status:   initialStatus,
				ActionBy: userID,
				ActionAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.store.CreateApproval(approval)
	c.JSON(http.StatusCreated, approval)
}

func (h *ApprovalHandler) GetApproval(c *gin.Context) {
	id := c.Param("id")
	approval, ok := h.store.GetApproval(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval not found"})
		return
	}
	c.JSON(http.StatusOK, approval)
}

func (h *ApprovalHandler) ProcessApproval(c *gin.Context) {
	id := c.Param("id")
	approval, ok := h.store.GetApproval(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval not found"})
		return
	}

	var req struct {
		Action   string  `json:"action" binding:"required"`
		Comments string  `json:"comments"`
		NewAmount *float64 `json:"new_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if approval.Status == models.ApprovalStatusApproved || approval.Status == models.ApprovalStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "approval is already finalized"})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "user-1"
	}
	now := time.Now()

	var step models.ApprovalStep
	step.ActionBy = userID
	step.ActionAt = now
	step.Comments = req.Comments

	switch req.Action {
	case "approve":
		switch approval.Status {
		case models.ApprovalStatusReview1:
			if approval.Amount < smallAmountThreshold {
				approval.Status = models.ApprovalStatusFinal
			} else {
				approval.Status = models.ApprovalStatusReview2
			}
		case models.ApprovalStatusReview2:
			approval.Status = models.ApprovalStatusFinal
		case models.ApprovalStatusFinal:
			approval.Status = models.ApprovalStatusApproved
		}
		step.Status = approval.Status
	case "reject":
		approval.Status = models.ApprovalStatusRejected
		step.Status = models.ApprovalStatusRejected
	case "resubmit":
		if approval.Status != models.ApprovalStatusRejected {
			c.JSON(http.StatusBadRequest, gin.H{"error": "can only resubmit rejected approvals"})
			return
		}
		approval.Status = models.ApprovalStatusSubmitted
		step.Status = models.ApprovalStatusSubmitted
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	if req.NewAmount != nil && *req.NewAmount != approval.Amount {
		oldAmount := approval.Amount
		approval.Amount = *req.NewAmount
		step.OldAmount = &oldAmount
		step.NewAmount = req.NewAmount
	}

	approval.History = append(approval.History, step)
	approval.UpdatedAt = now

	h.store.UpdateApproval(approval)
	c.JSON(http.StatusOK, approval)
}
