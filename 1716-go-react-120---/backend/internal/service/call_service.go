package service

import (
	"errors"
	"time"

	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/store"
)

type CallService struct {
	store *store.Store
}

func NewCallService(store *store.Store) *CallService {
	return &CallService{store: store}
}

type CreateCallRequest struct {
	CallTime       time.Time         `json:"call_time"`
	CallerName     string            `json:"caller_name"`
	CallerPhone    string            `json:"caller_phone"`
	Location       string            `json:"location"`
	PatientCount   int               `json:"patient_count"`
	PatientGender  string            `json:"patient_gender"`
	PatientAgeGroup models.AgeGroup  `json:"patient_age_group"`
	ChiefComplaint string            `json:"chief_complaint"`
	SeverityLevel  models.SeverityLevel `json:"severity_level"`
	BillID         string            `json:"bill_id,omitempty"`
	Amount         float64           `json:"amount,omitempty"`
}

func (s *CallService) CreateCall(req *CreateCallRequest) (*models.EmergencyCall, error) {
	if req.ChiefComplaint == "" {
		return nil, errors.New("主诉症状不能为空")
	}

	parsedSeverity, ok := models.ParseSeverity(string(req.SeverityLevel))
	if !ok {
		return nil, errors.New("病情严重程度不在有效范围内")
	}

	call := models.NewEmergencyCall()
	call.CallTime = req.CallTime
	call.CallerName = req.CallerName
	call.CallerPhone = req.CallerPhone
	call.Location = req.Location
	call.PatientCount = req.PatientCount
	call.PatientGender = req.PatientGender
	call.PatientAgeGroup = req.PatientAgeGroup
	call.ChiefComplaint = req.ChiefComplaint
	call.SeverityLevel = parsedSeverity
	call.BillID = req.BillID
	call.Amount = req.Amount

	s.store.CreateCall(call)
	return call, nil
}

func (s *CallService) GetCall(id string) (*models.EmergencyCall, bool) {
	return s.store.GetCall(id)
}

func (s *CallService) ListCalls() []*models.EmergencyCall {
	return s.store.ListCalls()
}

func (s *CallService) AcceptCall(callID string) (*models.EmergencyCall, error) {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	if call.Status != models.CallStatusPendingAccept {
		return nil, errors.New("只能接单待接单状态的求救")
	}

	call.Status = models.CallStatusAccepted
	call.UpdateTime = time.Now()
	s.store.UpdateCall(call)
	return call, nil
}

func (s *CallService) ProcessCall(callID string) (*models.EmergencyCall, error) {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	if call.Status != models.CallStatusAccepted && call.Status != models.CallStatusPendingCheck {
		return nil, errors.New("只能从已接单或待验收状态进入处理中")
	}

	call.Status = models.CallStatusProcessing
	call.UpdateTime = time.Now()
	s.store.UpdateCall(call)
	return call, nil
}

func (s *CallService) SubmitForReview(callID string) (*models.EmergencyCall, error) {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	if call.Status != models.CallStatusProcessing {
		return nil, errors.New("只能从处理中状态提交验收")
	}

	call.Status = models.CallStatusPendingCheck
	call.UpdateTime = time.Now()
	s.store.UpdateCall(call)
	return call, nil
}

func (s *CallService) CompleteCall(callID string) (*models.EmergencyCall, error) {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	if call.Status != models.CallStatusPendingCheck {
		return nil, errors.New("只能验收待验收状态的求救")
	}

	call.Status = models.CallStatusCompleted
	call.UpdateTime = time.Now()
	s.store.UpdateCall(call)

	if call.BillID != "" {
		if bill, ok := s.store.GetBill(call.BillID); ok {
			bill.Amount -= call.Amount
			if bill.Amount <= 0 {
				bill.Paid = true
			}
			s.store.UpdateBill(bill)
		}
	}

	return call, nil
}

func (s *CallService) RejectCall(callID string) (*models.EmergencyCall, error) {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	if call.Status != models.CallStatusPendingCheck {
		return nil, errors.New("只能拒绝待验收状态的求救")
	}

	call.Status = models.CallStatusProcessing
	call.UpdateTime = time.Now()
	s.store.UpdateCall(call)
	return call, nil
}

func (s *CallService) CloseExpiredCalls() {
	now := time.Now()
	calls := s.store.ListCalls()
	for _, call := range calls {
		if call.Status == models.CallStatusPendingAccept {
			if now.Sub(call.CreateTime) > 48*time.Hour {
				call.Status = models.CallStatusClosed
				call.UpdateTime = now
				s.store.UpdateCall(call)
			}
		}
	}
}
