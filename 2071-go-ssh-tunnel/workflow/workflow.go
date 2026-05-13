package workflow

import (
	"fmt"

	"ssh-tunnel-manager/database"
	"ssh-tunnel-manager/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateRecord(configID int64, operator string) (int64, error) {
	record := &models.TunnelRecord{
		TunnelConfigID: configID,
		Status:         models.ApprovalPending,
	}
	recordID, err := s.db.CreateTunnelRecord(record)
	if err != nil {
		return 0, err
	}

	history := &models.OperationHistory{
		RecordID:    recordID,
		OpType:      models.OpTypeCreate,
		Operator:    operator,
		Description: "创建隧道记录",
	}
	_, err = s.db.CreateOperationHistory(history)
	if err != nil {
		return 0, err
	}

	return recordID, nil
}

func (s *Service) SubmitForReview(recordID int64, operator string) error {
	record, err := s.db.GetTunnelRecord(recordID)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("record not found")
	}
	if record.Status != models.ApprovalPending {
		return fmt.Errorf("invalid status transition: current=%s, expected=%s", record.Status, models.ApprovalPending)
	}

	record.Status = models.ApprovalReviewing
	if err := s.db.UpdateTunnelRecord(record); err != nil {
		return err
	}

	history := &models.OperationHistory{
		RecordID:    recordID,
		OpType:      models.OpTypeSubmit,
		Operator:    operator,
		Description: "提交审核",
	}
	_, err = s.db.CreateOperationHistory(history)
	return err
}

func (s *Service) Review(recordID int64, operator string, approved bool, reason string) error {
	record, err := s.db.GetTunnelRecord(recordID)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("record not found")
	}
	if record.Status != models.ApprovalReviewing {
		return fmt.Errorf("invalid status transition: current=%s, expected=%s", record.Status, models.ApprovalReviewing)
	}

	var opType models.OperationType
	var desc string
	if approved {
		record.Status = models.ApprovalApproved
		opType = models.OpTypeApprove
		desc = "审核通过"
	} else {
		record.Status = models.ApprovalRejected
		opType = models.OpTypeReject
		desc = "审核拒绝: " + reason
	}

	if err := s.db.UpdateTunnelRecord(record); err != nil {
		return err
	}

	history := &models.OperationHistory{
		RecordID:    recordID,
		OpType:      opType,
		Operator:    operator,
		Description: desc,
	}
	_, err = s.db.CreateOperationHistory(history)
	return err
}

func (s *Service) RejectToModify(recordID int64, operator string, reason string) error {
	record, err := s.db.GetTunnelRecord(recordID)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("record not found")
	}
	if record.Status != models.ApprovalReviewing && record.Status != models.ApprovalRejected {
		return fmt.Errorf("invalid status for reject to modify")
	}

	record.Status = models.ApprovalPending
	if err := s.db.UpdateTunnelRecord(record); err != nil {
		return err
	}

	history := &models.OperationHistory{
		RecordID:    recordID,
		OpType:      models.OpTypeReject,
		Operator:    operator,
		Description: "退回修改: " + reason,
	}
	_, err = s.db.CreateOperationHistory(history)
	return err
}

func (s *Service) StartExecution(recordID int64, operator string) error {
	record, err := s.db.GetTunnelRecord(recordID)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("record not found")
	}
	if record.Status != models.ApprovalApproved {
		return fmt.Errorf("invalid status transition: current=%s, expected=%s", record.Status, models.ApprovalApproved)
	}

	record.Status = models.ApprovalExecuting
	if err := s.db.UpdateTunnelRecord(record); err != nil {
		return err
	}

	history := &models.OperationHistory{
		RecordID:    recordID,
		OpType:      models.OpTypeExecute,
		Operator:    operator,
		Description: "开始执行",
	}
	_, err = s.db.CreateOperationHistory(history)
	return err
}

func (s *Service) Complete(recordID int64, operator string) error {
	record, err := s.db.GetTunnelRecord(recordID)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("record not found")
	}
	if record.Status != models.ApprovalExecuting {
		return fmt.Errorf("invalid status transition: current=%s, expected=%s", record.Status, models.ApprovalExecuting)
	}

	record.Status = models.ApprovalCompleted
	if err := s.db.UpdateTunnelRecord(record); err != nil {
		return err
	}

	history := &models.OperationHistory{
		RecordID:    recordID,
		OpType:      models.OpTypeComplete,
		Operator:    operator,
		Description: "执行完成",
	}
	_, err = s.db.CreateOperationHistory(history)
	return err
}

func (s *Service) ListRecords() ([]*models.TunnelRecord, error) {
	return s.db.ListTunnelRecords()
}

func (s *Service) GetRecord(recordID int64) (*models.TunnelRecord, error) {
	return s.db.GetTunnelRecord(recordID)
}

func (s *Service) GetOperationHistories(recordID int64) ([]*models.OperationHistory, error) {
	return s.db.ListOperationHistories(recordID)
}

func (s *Service) AddNote(historyID int64, note string) error {
	history, err := s.db.GetOperationHistory(historyID)
	if err != nil {
		return err
	}
	if history == nil {
		return fmt.Errorf("operation history not found")
	}
	return s.db.UpdateOperationHistoryNote(historyID, note)
}
