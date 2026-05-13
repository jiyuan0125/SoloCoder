package qc

import (
	"crypto/rand"
	"encoding/hex"

	"aftersale-ticket/db"
	"aftersale-ticket/models"
)

type Service struct {
	store *db.Store
}

func NewService(store *db.Store) *Service {
	return &Service{store: store}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "QC" + hex.EncodeToString(b)
}

type QCRequest struct {
	TicketID  string
	Passed    bool
	Reason    string
	Inspector string
}

func (s *Service) PerformQC(req *QCRequest) (*models.QCResult, error) {
	qc := &models.QCResult{
		ID:        generateID(),
		TicketID:  req.TicketID,
		Passed:    req.Passed,
		Reason:    req.Reason,
		Inspector: req.Inspector,
	}

	if err := s.store.CreateQCResult(qc); err != nil {
		return nil, err
	}

	return qc, nil
}

type RefundRequest struct {
	OrderID string
	Amount  float64
	Method  string
}

func (s *Service) ProcessRefund(req *RefundRequest) error {
	return nil
}
