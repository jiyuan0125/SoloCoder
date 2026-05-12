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

type CertificateHandler struct {
	store *storage.Storage
}

func NewCertificateHandler(store *storage.Storage) *CertificateHandler {
	return &CertificateHandler{store: store}
}

type IssueCertificateRequest struct {
	CandidateID string `json:"candidateId" binding:"required"`
	BatchID     string `json:"batchId" binding:"required"`
}

func (h *CertificateHandler) IssueCertificate(c *gin.Context) {
	var req IssueCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch := h.store.GetBatch(req.BatchID)
	if batch == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "批次不存在"})
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

	if !candidate.IsPassed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "考生未通过鉴定，不能颁发证书"})
		return
	}

	occ := h.store.GetOccupation(batch.OccupationID)
	if occ == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "职业不存在"})
		return
	}

	year := time.Now().Year()
	key := generateSerialKey(year, occ.Code, candidate.AppliedLevel)
	serial, ok := h.store.GetNextSerialNumber(key)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "证书编号流水号溢出"})
		return
	}

	certNo, err := utils.GenerateCertificateNo(year, occ.Code, candidate.AppliedLevel, serial)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	issueDate := time.Now()
	expiryDate := issueDate.AddDate(5, 0, 0)

	cert := &models.Certificate{
		ID:             uuid.New().String(),
		CertificateNo:  certNo,
		CandidateID:    candidate.ID,
		Name:           candidate.Name,
		IDCard:         candidate.IDCard,
		OccupationID:   batch.OccupationID,
		OccupationName: batch.OccupationName,
		Level:          candidate.AppliedLevel,
		Status:         models.CertValid,
		IssueDate:      issueDate,
		ExpiryDate:     expiryDate,
	}

	h.store.SaveCertificate(cert)
	c.JSON(http.StatusCreated, cert)
}

func generateSerialKey(year int, occupationCode string, level models.SkillLevel) string {
	levelCode := utils.GetLevelCode(level)
	if len(occupationCode) < 3 {
		return ""
	}
	return string(rune(year)) + occupationCode[:3] + levelCode
}

func (h *CertificateHandler) GetCertificates(c *gin.Context) {
	certs := h.store.GetCertificates()
	c.JSON(http.StatusOK, certs)
}

func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	id := c.Param("id")
	cert := h.store.GetCertificate(id)
	if cert == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
		return
	}
	c.JSON(http.StatusOK, cert)
}

type UpdateCertificateStatusRequest struct {
	Status models.CertificateStatus `json:"status" binding:"required"`
}

func (h *CertificateHandler) UpdateCertificateStatus(c *gin.Context) {
	id := c.Param("id")
	cert := h.store.GetCertificate(id)
	if cert == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "证书不存在"})
		return
	}

	var req UpdateCertificateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cert.Status = req.Status
	h.store.SaveCertificate(cert)
	c.JSON(http.StatusOK, cert)
}

func (h *CertificateHandler) GetExpiringSoon(c *gin.Context) {
	allCerts := h.store.GetCertificates()
	expiring := []*models.Certificate{}

	for _, cert := range allCerts {
		if cert.Status == models.CertValid && utils.IsExpiringSoon(cert.ExpiryDate) {
			expiring = append(expiring, cert)
		}
	}

	c.JSON(http.StatusOK, expiring)
}

func (h *CertificateHandler) GetPassRateStats(c *gin.Context) {
	occs := h.store.GetOccupations()
	batches := h.store.GetBatches()

	statsMap := make(map[string]*models.PassRateStats)

	for _, batch := range batches {
		if batch.Status != models.BatchCompleted {
			continue
		}

		key := batch.OccupationID + "-" + string(batch.Level)
		if _, exists := statsMap[key]; !exists {
			statsMap[key] = &models.PassRateStats{
				OccupationID:   batch.OccupationID,
				OccupationName: batch.OccupationName,
				Level:          batch.Level,
				TotalTaken:     0,
				TotalPassed:    0,
				PassRate:       0,
			}
		}

		for _, candidate := range batch.Candidates {
			if candidate.HasTakenExam {
				statsMap[key].TotalTaken++
				if candidate.IsPassed {
					statsMap[key].TotalPassed++
				}
			}
		}
	}

	for _, occ := range occs {
		for _, levelConfig := range occ.Levels {
			key := occ.ID + "-" + string(levelConfig.Level)
			if _, exists := statsMap[key]; !exists {
				statsMap[key] = &models.PassRateStats{
					OccupationID:   occ.ID,
					OccupationName: occ.Name,
					Level:          levelConfig.Level,
					TotalTaken:     0,
					TotalPassed:    0,
					PassRate:       0,
				}
			}
		}
	}

	result := make([]models.PassRateStats, 0, len(statsMap))
	for _, stat := range statsMap {
		if stat.TotalTaken > 0 {
			stat.PassRate = float64(stat.TotalPassed) / float64(stat.TotalTaken) * 100
		}
		result = append(result, *stat)
	}

	c.JSON(http.StatusOK, result)
}
