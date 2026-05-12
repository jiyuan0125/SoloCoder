package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"skill-cert/models"
	"skill-cert/storage"
	"skill-cert/utils"
)

type ExamHandler struct {
	store *storage.Storage
}

func NewExamHandler(store *storage.Storage) *ExamHandler {
	return &ExamHandler{store: store}
}

type CreateBatchRequest struct {
	Name         string             `json:"name" binding:"required"`
	OccupationID string             `json:"occupationId" binding:"required"`
	Level        models.SkillLevel  `json:"level" binding:"required"`
	ExamDate     time.Time          `json:"examDate" binding:"required"`
	ExamRoom     string             `json:"examRoom" binding:"required"`
}

func (h *ExamHandler) CreateBatch(c *gin.Context) {
	var req CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	occ := h.store.GetOccupation(req.OccupationID)
	if occ == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "职业不存在"})
		return
	}

	batch := &models.ExamBatch{
		ID:             uuid.New().String(),
		Name:           req.Name,
		OccupationID:   req.OccupationID,
		OccupationName: occ.Name,
		Level:          req.Level,
		ExamDate:       req.ExamDate,
		ExamRoom:       req.ExamRoom,
		Status:         models.BatchNotStarted,
		Candidates:     []models.ExamCandidate{},
		CreatedAt:      time.Now(),
	}

	h.store.SaveBatch(batch)
	c.JSON(http.StatusCreated, batch)
}

func (h *ExamHandler) GetBatches(c *gin.Context) {
	batches := h.store.GetBatches()
	c.JSON(http.StatusOK, batches)
}

func (h *ExamHandler) GetBatch(c *gin.Context) {
	id := c.Param("id")
	batch := h.store.GetBatch(id)
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "批次不存在"})
		return
	}
	c.JSON(http.StatusOK, batch)
}

func (h *ExamHandler) DeleteBatch(c *gin.Context) {
	id := c.Param("id")
	batch := h.store.GetBatch(id)
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "批次不存在"})
		return
	}

	if len(batch.Candidates) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "已有考生报名的批次不能删除"})
		return
	}

	h.store.DeleteBatch(id)
	c.JSON(http.StatusNoContent, nil)
}

type UpdateBatchStatusRequest struct {
	Action string `json:"action" binding:"required"`
}

func (h *ExamHandler) UpdateBatchStatus(c *gin.Context) {
	id := c.Param("id")
	batch := h.store.GetBatch(id)
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "批次不存在"})
		return
	}

	var req UpdateBatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Action != "next" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的操作"})
		return
	}

	nextStatus, err := utils.NextBatchStatus(batch.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch.Status = nextStatus
	h.store.SaveBatch(batch)
	c.JSON(http.StatusOK, batch)
}

type RegisterCandidateRequest struct {
	Name         string            `json:"name" binding:"required"`
	IDCard       string            `json:"idCard" binding:"required"`
	Phone        string            `json:"phone" binding:"required"`
	AppliedLevel models.SkillLevel `json:"appliedLevel" binding:"required"`
}

func (h *ExamHandler) RegisterCandidate(c *gin.Context) {
	batchID := c.Param("id")
	batch := h.store.GetBatch(batchID)
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "批次不存在"})
		return
	}

	var req RegisterCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !utils.ValidateIDCard(req.IDCard) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "身份证号格式无效"})
		return
	}

	for _, candidate := range batch.Candidates {
		if candidate.IDCard == req.IDCard {
			c.JSON(http.StatusConflict, gin.H{"error": "同批次已存在该身份证号的考生"})
			return
		}
	}

	existingCerts := h.store.GetCertificatesByIdCard(req.IDCard)
	for _, cert := range existingCerts {
		if cert.OccupationID == batch.OccupationID && cert.Level == req.AppliedLevel {
			if cert.Status == models.CertValid {
				c.JSON(http.StatusConflict, gin.H{"error": "该考生已有有效证书，不能重复报考"})
				return
			}
		}
	}

	candidate := models.ExamCandidate{
		ID:           uuid.New().String(),
		BatchID:      batchID,
		Name:         req.Name,
		IDCard:       req.IDCard,
		Phone:        req.Phone,
		AppliedLevel: req.AppliedLevel,
		Scores:       make(map[models.ExamSubject]float64),
		HasTakenExam: false,
		IsPassed:     false,
	}

	batch.Candidates = append(batch.Candidates, candidate)
	h.store.SaveBatch(batch)
	c.JSON(http.StatusCreated, candidate)
}

type ScoreEntry struct {
	Subject models.ExamSubject `json:"subject" binding:"required"`
	Score   float64            `json:"score" binding:"required"`
}

type EnterScoresRequest struct {
	CandidateID string       `json:"candidateId" binding:"required"`
	Scores      []ScoreEntry `json:"scores" binding:"required"`
}

func (h *ExamHandler) EnterScores(c *gin.Context) {
	batchID := c.Param("id")
	batch := h.store.GetBatch(batchID)
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "批次不存在"})
		return
	}

	if batch.Status != models.BatchCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "成绩录入只能在已结束的批次上进行"})
		return
	}

	var req EnterScoresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var candidate *models.ExamCandidate
	for i := range batch.Candidates {
		if batch.Candidates[i].ID == req.CandidateID {
			candidate = &batch.Candidates[i]
			break
		}
	}

	if candidate == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考生不存在"})
		return
	}

	for _, entry := range req.Scores {
		if !utils.ValidateScoreRange(entry.Score) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "成绩必须在0-100分之间"})
			return
		}
	}

	occ := h.store.GetOccupation(batch.OccupationID)
	var levelSubjects []models.ExamSubject
	if occ != nil {
		for _, lc := range occ.Levels {
			if lc.Level == batch.Level {
				levelSubjects = lc.Subjects
				break
			}
		}
	}

	if candidate.Scores == nil {
		candidate.Scores = make(map[models.ExamSubject]float64)
	}
	for _, entry := range req.Scores {
		candidate.Scores[entry.Subject] = entry.Score
	}

	candidate.HasTakenExam = true
	candidate.IsPassed = utils.IsAllSubjectsPassed(candidate.Scores, levelSubjects)

	h.store.SaveBatch(batch)
	c.JSON(http.StatusOK, candidate)
}
