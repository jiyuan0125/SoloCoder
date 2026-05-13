package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"approval-flow/pkg/model"
	"approval-flow/pkg/store"
	"approval-flow/pkg/util"
)

var (
	ErrChainEmpty            = errors.New("approval chain has no nodes")
	ErrInvalidCondition      = errors.New("invalid condition expression")
	ErrSelfApproval          = errors.New("cannot approve your own application")
	ErrRejectReasonRequired  = errors.New("reject reason is required")
	ErrInvalidOperation      = errors.New("invalid operation")
	ErrApplicationNotPending = errors.New("application is not pending")
	ErrNoApprover            = errors.New("no current approver available")
	ErrNoManager             = errors.New("no manager found for escalation")
	ErrNotApprover           = errors.New("you are not an approver for this level")
)

type ApprovalService struct{}

func NewApprovalService() *ApprovalService {
	return &ApprovalService{}
}

func (s *ApprovalService) SubmitApplication(app *model.Application) (*model.Application, error) {
	chain, err := store.GetChainByID(app.ChainID)
	if err != nil {
		return nil, err
	}

	if len(chain.Nodes) == 0 {
		return nil, ErrChainEmpty
	}

	sort.Slice(chain.Nodes, func(i, j int) bool {
		return chain.Nodes[i].Level < chain.Nodes[j].Level
	})

	for _, node := range chain.Nodes {
		if node.Condition != "" {
			if err := ValidateExpression(node.Condition); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidCondition, err.Error())
			}
		}
	}

	currentLevel, approvers, err := s.findFirstLevel(chain.Nodes, app.Data)
	if err != nil {
		return nil, err
	}

	app.CurrentLevel = currentLevel
	app.CurrentApproverIDs = approvers
	app.Status = "pending"
	app.SubmissionCount = 1

	if err := store.CreateApplication(app); err != nil {
		return nil, err
	}

	if err := store.CreateTimeoutRecord(app.ID, currentLevel); err != nil {
		return nil, err
	}

	if err := s.sendNotifications(app, "new_application", "新申请待审批"); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApprovalService) ResubmitApplication(appID string, applicantID string, updates *model.Application) (*model.Application, error) {
	app, err := store.GetApplicationByID(appID)
	if err != nil {
		return nil, err
	}

	if app.ApplicantID != applicantID {
		return nil, errors.New("you can only resubmit your own application")
	}

	if app.Status != "rejected" {
		return nil, errors.New("only rejected applications can be resubmitted")
	}

	if updates.Title != "" {
		app.Title = updates.Title
	}
	if updates.Description != "" {
		app.Description = updates.Description
	}
	if updates.Data != nil {
		app.Data = updates.Data
	}

	chain, err := store.GetChainByID(app.ChainID)
	if err != nil {
		return nil, err
	}

	sort.Slice(chain.Nodes, func(i, j int) bool {
		return chain.Nodes[i].Level < chain.Nodes[j].Level
	})

	for _, node := range chain.Nodes {
		if node.Condition != "" {
			if err := ValidateExpression(node.Condition); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidCondition, err.Error())
			}
		}
	}

	currentLevel, approvers, err := s.findFirstLevel(chain.Nodes, app.Data)
	if err != nil {
		return nil, err
	}

	app.CurrentLevel = currentLevel
	app.CurrentApproverIDs = approvers
	app.Status = "pending"
	app.RejectReason = ""
	app.SubmissionCount++

	if err := store.DeleteTimeoutRecords(app.ID); err != nil {
		return nil, err
	}

	if err := store.CreateTimeoutRecord(app.ID, currentLevel); err != nil {
		return nil, err
	}

	if err := store.UpdateApplication(app); err != nil {
		return nil, err
	}

	if err := s.sendNotifications(app, "new_application", "重新提交的申请待审批"); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApprovalService) ApproveApplication(appID string, operatorID string) (*model.Application, error) {
	return s.handleApproval(appID, operatorID, "approve", "")
}

func (s *ApprovalService) RejectApplication(appID string, operatorID string, reason string) (*model.Application, error) {
	if reason == "" {
		return nil, ErrRejectReasonRequired
	}
	return s.handleApproval(appID, operatorID, "reject", reason)
}

func (s *ApprovalService) TransferApplication(appID string, operatorID string, targetUserID string, reason string) (*model.Application, error) {
	app, err := store.GetApplicationByID(appID)
	if err != nil {
		return nil, err
	}

	if app.Status != "pending" {
		return nil, ErrApplicationNotPending
	}

	if app.ApplicantID == operatorID {
		return nil, ErrSelfApproval
	}

	if !util.ContainsString(app.CurrentApproverIDs, operatorID) {
		return nil, ErrNotApprover
	}

	if operatorID == targetUserID {
		return nil, errors.New("cannot transfer to yourself")
	}

	_, err = store.GetUserByID(targetUserID)
	if err != nil {
		return nil, err
	}

	op := &model.ApprovalOperation{
		ApplicationID: appID,
		OperatorID:    operatorID,
		Level:         app.CurrentLevel,
		Operation:     "transfer",
		Reason:        reason,
		TargetUserID:  targetUserID,
	}
	if err := store.CreateOperationLog(op); err != nil {
		return nil, err
	}

	newApprovers := []string{}
	for _, id := range app.CurrentApproverIDs {
		if id == operatorID {
			newApprovers = append(newApprovers, targetUserID)
		} else {
			newApprovers = append(newApprovers, id)
		}
	}
	app.CurrentApproverIDs = newApprovers

	if err := store.UpdateApplication(app); err != nil {
		return nil, err
	}

	notif := &model.Notification{
		UserID:        operatorID,
		ApplicationID: appID,
		Type:          "transfer_origin",
		Message:       "您已将申请转审给他人",
	}
	if err := store.CreateNotification(notif); err != nil {
		return nil, err
	}

	notif2 := &model.Notification{
		UserID:        targetUserID,
		ApplicationID: appID,
		Type:          "transfer_target",
		Message:       "收到转审的申请，需要您审批",
	}
	if err := store.CreateNotification(notif2); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApprovalService) handleApproval(appID string, operatorID string, operation string, reason string) (*model.Application, error) {
	app, err := store.GetApplicationByID(appID)
	if err != nil {
		return nil, err
	}

	if app.Status != "pending" {
		return nil, ErrApplicationNotPending
	}

	if app.ApplicantID == operatorID {
		return nil, ErrSelfApproval
	}

	if !util.ContainsString(app.CurrentApproverIDs, operatorID) {
		return nil, ErrNotApprover
	}

	chain, err := store.GetChainByID(app.ChainID)
	if err != nil {
		return nil, err
	}

	sort.Slice(chain.Nodes, func(i, j int) bool {
		return chain.Nodes[i].Level < chain.Nodes[j].Level
	})

	currentNode := s.findNodeByLevel(chain.Nodes, app.CurrentLevel)

	op := &model.ApprovalOperation{
		ApplicationID: appID,
		OperatorID:    operatorID,
		Level:         app.CurrentLevel,
		Operation:     operation,
		Reason:        reason,
	}
	if err := store.CreateOperationLog(op); err != nil {
		return nil, err
	}

	if operation == "reject" {
		app.Status = "rejected"
		app.RejectReason = reason

		if err := store.UpdateApplication(app); err != nil {
			return nil, err
		}

		notif := &model.Notification{
			UserID:        app.ApplicantID,
			ApplicationID: appID,
			Type:          "rejected",
			Message:       "您的申请已被驳回：" + reason,
		}
		if err := store.CreateNotification(notif); err != nil {
			return nil, err
		}

		if err := s.UpdateReport(app.ChainID); err != nil {
			return nil, err
		}

		return app, nil
	}

	if currentNode != nil && currentNode.IsSignAll {
		operations, err := store.GetOperationsByApplication(appID)
		if err != nil {
			return nil, err
		}

		approvedCount := 0
		for _, prevOp := range operations {
			if prevOp.Level == app.CurrentLevel && prevOp.Operation == "approve" {
				approvedCount++
			}
		}

		if approvedCount < len(currentNode.ApproverIDs) {
			if err := store.UpdateApplication(app); err != nil {
				return nil, err
			}
			return app, nil
		}
	}

	nextLevel, nextApprovers, err := s.findNextLevel(chain.Nodes, app.CurrentLevel, app.Data)
	if err != nil {
		return nil, err
	}

	if nextLevel < 0 {
		app.Status = "approved"

		if err := store.DeleteTimeoutRecords(app.ID); err != nil {
			return nil, err
		}

		if err := store.UpdateApplication(app); err != nil {
			return nil, err
		}

		notif := &model.Notification{
			UserID:        app.ApplicantID,
			ApplicationID: appID,
			Type:          "approved",
			Message:       "您的申请已审批通过",
		}
		if err := store.CreateNotification(notif); err != nil {
			return nil, err
		}

		if err := s.UpdateReport(app.ChainID); err != nil {
			return nil, err
		}

		return app, nil
	}

	app.CurrentLevel = nextLevel
	app.CurrentApproverIDs = nextApprovers

	if err := store.DeleteTimeoutRecords(app.ID); err != nil {
		return nil, err
	}

	if err := store.CreateTimeoutRecord(app.ID, nextLevel); err != nil {
		return nil, err
	}

	if err := store.UpdateApplication(app); err != nil {
		return nil, err
	}

	if err := s.sendNotifications(app, "new_application", "申请进入下一级审批"); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApprovalService) findFirstLevel(nodes []*model.ApprovalNode, data map[string]interface{}) (int, []string, error) {
	for _, node := range nodes {
		if node.Condition != "" {
			result, err := EvaluateExpression(node.Condition, data)
			if err != nil {
				continue
			}
			if !result {
				continue
			}
		}

		if len(node.ApproverIDs) == 0 {
			continue
		}

		return node.Level, node.ApproverIDs, nil
	}

	return -1, nil, ErrNoApprover
}

func (s *ApprovalService) findNextLevel(nodes []*model.ApprovalNode, currentLevel int, data map[string]interface{}) (int, []string, error) {
	for _, node := range nodes {
		if node.Level <= currentLevel {
			continue
		}

		if node.Condition != "" {
			result, err := EvaluateExpression(node.Condition, data)
			if err != nil {
				continue
			}
			if !result {
				continue
			}
		}

		if len(node.ApproverIDs) == 0 {
			continue
		}

		return node.Level, node.ApproverIDs, nil
	}

	return -1, nil, nil
}

func (s *ApprovalService) findNodeByLevel(nodes []*model.ApprovalNode, level int) *model.ApprovalNode {
	for _, node := range nodes {
		if node.Level == level {
			return node
		}
	}
	return nil
}

func (s *ApprovalService) sendNotifications(app *model.Application, notifType string, message string) error {
	for _, approverID := range app.CurrentApproverIDs {
		notif := &model.Notification{
			UserID:        approverID,
			ApplicationID: app.ID,
			Type:          notifType,
			Message:       message,
		}
		if err := store.CreateNotification(notif); err != nil {
			return err
		}
	}
	return nil
}

func (s *ApprovalService) EscalateTimeoutApplication(app *model.Application) error {
	if app.Status != "pending" {
		return nil
	}

	record, err := store.GetTimeoutRecord(app.ID, app.CurrentLevel)
	if err != nil {
		return nil
	}

	now := time.Now()
	elapsed := now.Sub(app.UpdatedAt)

	if elapsed >= 48*time.Hour && !record.Reminded48h {
		for _, approverID := range app.CurrentApproverIDs {
			notif := &model.Notification{
				UserID:        approverID,
				ApplicationID: app.ID,
				Type:          "timeout_reminder",
				Message:       "您的审批任务已超过48小时未处理，请尽快处理",
			}
			if err := store.CreateNotification(notif); err != nil {
				return err
			}
		}

		if err := store.MarkReminded48h(app.ID, app.CurrentLevel); err != nil {
			return err
		}
	}

	if elapsed >= 72*time.Hour && !record.Transferred72h {
		for _, approverID := range app.CurrentApproverIDs {
			manager, err := store.GetManager(approverID)
			if err != nil {
				continue
			}

			notif := &model.Notification{
				UserID:        approverID,
				ApplicationID: app.ID,
				Type:          "timeout_escalation",
				Message:       "您的审批任务已超过72小时未处理，已自动转审给您的上级",
			}
			if err := store.CreateNotification(notif); err != nil {
				return err
			}

			notif2 := &model.Notification{
				UserID:        manager.ID,
				ApplicationID: app.ID,
				Type:          "timeout_escalation_target",
				Message:       "收到因下级超时转审的申请",
			}
			if err := store.CreateNotification(notif2); err != nil {
				return err
			}

			op := &model.ApprovalOperation{
				ApplicationID: app.ID,
				OperatorID:    approverID,
				Level:         app.CurrentLevel,
				Operation:     "transfer",
				Reason:        "审批超时自动转审",
				TargetUserID:  manager.ID,
			}
			if err := store.CreateOperationLog(op); err != nil {
				return err
			}

			chain, err := store.GetChainByID(app.ChainID)
			if err != nil {
				return err
			}

			currentNode := s.findNodeByLevel(chain.Nodes, app.CurrentLevel)
			if currentNode == nil {
				return ErrNodeNotFound
			}

			newApprovers := []string{}
			for _, id := range app.CurrentApproverIDs {
				if id == approverID {
					newApprovers = append(newApprovers, manager.ID)
				} else {
					newApprovers = append(newApprovers, id)
				}
			}
			app.CurrentApproverIDs = newApprovers
		}

		if err := store.MarkTransferred72h(app.ID, app.CurrentLevel); err != nil {
			return err
		}

		if err := store.UpdateApplication(app); err != nil {
			return err
		}
	}

	return nil
}

var ErrNodeNotFound = errors.New("approval node not found")
