package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"workorder-flow/internal/database"
	"workorder-flow/internal/duty"
	"workorder-flow/internal/models"
)

type FlowEngine struct {
	db        *database.DB
	scheduler *duty.Scheduler
}

func NewFlowEngine(db *database.DB, scheduler *duty.Scheduler) *FlowEngine {
	return &FlowEngine{db: db, scheduler: scheduler}
}

var (
	ErrInvalidStateTransition   = errors.New("invalid state transition")
	ErrAlreadyAssigned          = errors.New("work order already assigned to someone else")
	ErrWorkOrderClosed          = errors.New("work order is already closed")
	ErrReasonRequired           = errors.New("reason is required for this operation")
	ErrAlreadyEscalated         = errors.New("work order is already escalated")
	ErrNotCurrentHandler        = errors.New("only current handler can perform this operation")
	ErrInvalidHandler           = errors.New("invalid handler")
	ErrEscalatedCannotOperate   = errors.New("cannot operate on escalated work order")
	ErrNotSubmitter             = errors.New("only submitter can perform this operation")
	ErrCommentRequired          = errors.New("comment is required for this operation")
)

type CreateWorkOrderRequest struct {
	Type        models.WorkOrderType
	Title       string
	Description string
	Content     string
	SubmitterID int64
	ResourceIDs []int64
}

func (e *FlowEngine) CreateWorkOrder(ctx context.Context, req *CreateWorkOrderRequest) (*models.WorkOrder, error) {
	for _, rid := range req.ResourceIDs {
		_, err := e.db.GetResourceByID(ctx, rid)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return nil, fmt.Errorf("resource %d not found", rid)
			}
			return nil, err
		}
	}

	user, team, err := e.scheduler.AssignDuty(ctx, req.Type)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	deadline := now.Add(24 * time.Hour)

	wo := &models.WorkOrder{
		Type:             req.Type,
		Title:            req.Title,
		Description:      req.Description,
		Content:          req.Content,
		CurrentHandlerID: &user.ID,
		CurrentTeamID:    team.ID,
		ProcessStatus:    models.ProcessStatusSubmitted,
		WorkOrderStatus:  models.WorkOrderStatusPending,
		SubmitterID:      req.SubmitterID,
		Escalated:        false,
		DeadlineAt:       &deadline,
	}

	created, err := e.db.CreateWorkOrder(ctx, wo, req.ResourceIDs)
	if err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        created.ID,
		OperationType:      "创建",
		OperatorID:         req.SubmitterID,
		OldProcessStatus:   "",
		NewProcessStatus:   created.ProcessStatus,
		OldWorkOrderStatus: "",
		NewWorkOrderStatus: created.WorkOrderStatus,
		AssignedTo:         created.CurrentHandlerID,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return created, nil
}

func (e *FlowEngine) ClaimWorkOrder(ctx context.Context, workOrderID int64, handlerID int64) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.CurrentHandlerID != nil {
		_, err := e.db.GetUserByID(ctx, *wo.CurrentHandlerID)
		if err == nil {
			return nil, fmt.Errorf("%w: current handler is %d", ErrAlreadyAssigned, *wo.CurrentHandlerID)
		}
	}

	oldHandler := wo.CurrentHandlerID
	wo.CurrentHandlerID = &handlerID

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "认领",
		OperatorID:         handlerID,
		OldProcessStatus:   wo.ProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
		AssignedFrom:       oldHandler,
		AssignedTo:         &handlerID,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) AdvanceWorkOrderStatus(ctx context.Context, workOrderID int64, operatorID int64) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.Escalated && wo.CurrentHandlerID != nil && *wo.CurrentHandlerID == operatorID {
		return nil, ErrEscalatedCannotOperate
	}

	if wo.CurrentHandlerID == nil || *wo.CurrentHandlerID != operatorID {
		return nil, ErrNotCurrentHandler
	}

	oldStatus := wo.WorkOrderStatus
	var newStatus models.WorkOrderStatus

	switch wo.WorkOrderStatus {
	case models.WorkOrderStatusPending:
		newStatus = models.WorkOrderStatusProcessing
		now := time.Now()
		wo.ProcessStartAt = &now
	case models.WorkOrderStatusProcessing:
		newStatus = models.WorkOrderStatusConfirming
	case models.WorkOrderStatusConfirming:
		return nil, fmt.Errorf("%w: cannot advance from %s to next status, submitter must confirm", ErrInvalidStateTransition, wo.WorkOrderStatus)
	case models.WorkOrderStatusClosed:
		return nil, ErrWorkOrderClosed
	default:
		return nil, fmt.Errorf("%w: unknown status %s", ErrInvalidStateTransition, wo.WorkOrderStatus)
	}

	wo.WorkOrderStatus = newStatus

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "状态流转",
		OperatorID:         operatorID,
		OldProcessStatus:   wo.ProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: oldStatus,
		NewWorkOrderStatus: newStatus,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) ConfirmAndClose(ctx context.Context, workOrderID int64, submitterID int64, rating *int) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.WorkOrderStatus != models.WorkOrderStatusConfirming {
		return nil, fmt.Errorf("%w: current status is %s, must be 待确认 to confirm", ErrInvalidStateTransition, wo.WorkOrderStatus)
	}

	if wo.SubmitterID != submitterID {
		return nil, errors.New("only submitter can confirm the work order")
	}

	oldStatus := wo.WorkOrderStatus
	wo.WorkOrderStatus = models.WorkOrderStatusClosed
	wo.ProcessStatus = models.ProcessStatusCompleted

	if rating != nil {
		wo.Rating = rating
	}

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	comment := ""
	if rating != nil {
		comment = fmt.Sprintf("评分: %d", *rating)
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "确认关闭",
		OperatorID:         submitterID,
		OldProcessStatus:   wo.ProcessStatus,
		NewProcessStatus:   models.ProcessStatusCompleted,
		OldWorkOrderStatus: oldStatus,
		NewWorkOrderStatus: models.WorkOrderStatusClosed,
		Comment:            comment,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) ReassignWorkOrder(ctx context.Context, workOrderID int64, currentHandlerID int64, newHandlerID int64, reason string) (*models.WorkOrder, error) {
	if reason == "" {
		return nil, ErrReasonRequired
	}

	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.Escalated && wo.CurrentHandlerID != nil && *wo.CurrentHandlerID == currentHandlerID {
		return nil, ErrEscalatedCannotOperate
	}

	if wo.CurrentHandlerID == nil || *wo.CurrentHandlerID != currentHandlerID {
		return nil, ErrNotCurrentHandler
	}

	newHandler, err := e.db.GetUserByID(ctx, newHandlerID)
	if err != nil {
		return nil, ErrInvalidHandler
	}

	if newHandler.TeamID == nil {
		return nil, ErrInvalidHandler
	}

	oldHandlerID := wo.CurrentHandlerID
	wo.CurrentHandlerID = &newHandlerID
	wo.CurrentTeamID = *newHandler.TeamID

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "转派",
		OperatorID:         currentHandlerID,
		OldProcessStatus:   wo.ProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
		AssignedFrom:       oldHandlerID,
		AssignedTo:         &newHandlerID,
		Reason:             reason,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) EscalateWorkOrder(ctx context.Context, workOrderID int64, operatorID int64, reason string) (*models.WorkOrder, error) {
	if reason == "" {
		return nil, ErrReasonRequired
	}

	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.Escalated {
		return nil, ErrAlreadyEscalated
	}

	team, err := e.db.GetTeamByID(ctx, wo.CurrentTeamID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	wo.Escalated = true
	wo.EscalatedAt = &now
	oldHandlerID := wo.CurrentHandlerID
	wo.CurrentHandlerID = &team.LeaderID

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "升级",
		OperatorID:         operatorID,
		OldProcessStatus:   wo.ProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
		AssignedFrom:       oldHandlerID,
		AssignedTo:         &team.LeaderID,
		Reason:             reason,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) AddCommunication(ctx context.Context, workOrderID int64, senderID int64, content string) error {
	_, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return err
	}

	if content == "" {
		return errors.New("content cannot be empty")
	}

	record := &models.CommunicationRecord{
		WorkOrderID: workOrderID,
		SenderID:    senderID,
		Content:     content,
	}

	return e.db.AddCommunicationRecord(ctx, record)
}

func (e *FlowEngine) SubmitForReview(ctx context.Context, workOrderID int64, operatorID int64) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.ProcessStatus != models.ProcessStatusSubmitted && wo.ProcessStatus != models.ProcessStatusRejected {
		return nil, fmt.Errorf("%w: can only submit from 待提交 or 退回修改, current status is %s", ErrInvalidStateTransition, wo.ProcessStatus)
	}

	oldProcessStatus := wo.ProcessStatus
	wo.ProcessStatus = models.ProcessStatusReviewing

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "提交审核",
		OperatorID:         operatorID,
		OldProcessStatus:   oldProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) ApproveWorkOrder(ctx context.Context, workOrderID int64, operatorID int64, comment string) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.ProcessStatus != models.ProcessStatusReviewing {
		return nil, fmt.Errorf("%w: can only approve from 审核中, current status is %s", ErrInvalidStateTransition, wo.ProcessStatus)
	}

	oldProcessStatus := wo.ProcessStatus
	wo.ProcessStatus = models.ProcessStatusApproved

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "审核通过",
		OperatorID:         operatorID,
		OldProcessStatus:   oldProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
		Comment:            comment,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) RejectWorkOrder(ctx context.Context, workOrderID int64, operatorID int64, reason string) (*models.WorkOrder, error) {
	if reason == "" {
		return nil, ErrReasonRequired
	}

	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.ProcessStatus != models.ProcessStatusReviewing {
		return nil, fmt.Errorf("%w: can only reject from 审核中, current status is %s", ErrInvalidStateTransition, wo.ProcessStatus)
	}

	oldProcessStatus := wo.ProcessStatus
	wo.ProcessStatus = models.ProcessStatusRejected

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "审核拒绝",
		OperatorID:         operatorID,
		OldProcessStatus:   oldProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
		Reason:             reason,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) StartExecution(ctx context.Context, workOrderID int64, operatorID int64) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.ProcessStatus != models.ProcessStatusApproved {
		return nil, fmt.Errorf("%w: can only start execution from 已通过, current status is %s", ErrInvalidStateTransition, wo.ProcessStatus)
	}

	if wo.CurrentHandlerID == nil || *wo.CurrentHandlerID != operatorID {
		return nil, ErrNotCurrentHandler
	}

	if wo.Escalated && wo.CurrentHandlerID != nil && *wo.CurrentHandlerID == operatorID {
		return nil, ErrEscalatedCannotOperate
	}

	oldProcessStatus := wo.ProcessStatus
	wo.ProcessStatus = models.ProcessStatusExecuting

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "开始执行",
		OperatorID:         operatorID,
		OldProcessStatus:   oldProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) CompleteExecution(ctx context.Context, workOrderID int64, operatorID int64, comment string) (*models.WorkOrder, error) {
	wo, err := e.db.GetWorkOrderByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}

	if wo.WorkOrderStatus == models.WorkOrderStatusClosed {
		return nil, ErrWorkOrderClosed
	}

	if wo.ProcessStatus != models.ProcessStatusExecuting {
		return nil, fmt.Errorf("%w: can only complete from 执行中, current status is %s", ErrInvalidStateTransition, wo.ProcessStatus)
	}

	if wo.CurrentHandlerID == nil || *wo.CurrentHandlerID != operatorID {
		return nil, ErrNotCurrentHandler
	}

	if wo.Escalated && wo.CurrentHandlerID != nil && *wo.CurrentHandlerID == operatorID {
		return nil, ErrEscalatedCannotOperate
	}

	oldProcessStatus := wo.ProcessStatus
	wo.ProcessStatus = models.ProcessStatusCompleted

	if err := e.db.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	history := &models.HistoryRecord{
		WorkOrderID:        wo.ID,
		OperationType:      "执行完成",
		OperatorID:         operatorID,
		OldProcessStatus:   oldProcessStatus,
		NewProcessStatus:   wo.ProcessStatus,
		OldWorkOrderStatus: wo.WorkOrderStatus,
		NewWorkOrderStatus: wo.WorkOrderStatus,
		Comment:            comment,
	}

	if err := e.db.AddHistoryRecord(ctx, history); err != nil {
		return nil, err
	}

	return wo, nil
}

func (e *FlowEngine) ProcessOverdueWorkOrders(ctx context.Context) error {
	now := time.Now()
	overdue, err := e.db.GetOverdueWorkOrders(ctx, now)
	if err != nil {
		return err
	}

	for _, wo := range overdue {
		if wo.Escalated {
			continue
		}

		team, err := e.db.GetTeamByID(ctx, wo.CurrentTeamID)
		if err != nil {
			continue
		}

		oldHandlerID := wo.CurrentHandlerID
		wo.Escalated = true
		wo.EscalatedAt = &now
		wo.CurrentHandlerID = &team.LeaderID

		if err := e.db.UpdateWorkOrder(ctx, &wo); err != nil {
			continue
		}

		history := &models.HistoryRecord{
			WorkOrderID:        wo.ID,
			OperationType:      "超时升级",
			OperatorID:         0,
			OldProcessStatus:   wo.ProcessStatus,
			NewProcessStatus:   wo.ProcessStatus,
			OldWorkOrderStatus: wo.WorkOrderStatus,
			NewWorkOrderStatus: wo.WorkOrderStatus,
			AssignedFrom:       oldHandlerID,
			AssignedTo:         &team.LeaderID,
			Reason:             "工单超时未处理，自动升级",
		}

		if err := e.db.AddHistoryRecord(ctx, history); err != nil {
			continue
		}
	}

	return nil
}
