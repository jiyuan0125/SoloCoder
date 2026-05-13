package notification

import (
	"context"
	"fmt"
	"log"

	"workorder-flow/internal/database"
	"workorder-flow/internal/models"
)

type NotificationService struct {
	db *database.DB
}

func NewNotificationService(db *database.DB) *NotificationService {
	return &NotificationService{db: db}
}

func (s *NotificationService) NotifyAssigned(ctx context.Context, workOrderID int64, handlerID int64) {
	user, err := s.db.GetUserByID(ctx, handlerID)
	if err != nil {
		log.Printf("failed to get user %d: %v", handlerID, err)
		return
	}

	log.Printf("[通知] 工单 #%d 已分配给 %s (%s)", workOrderID, user.Username, user.Email)
}

func (s *NotificationService) NotifyEscalated(ctx context.Context, workOrderID int64, oldHandlerID int64, newHandlerID int64, reason string) {
	oldUser, err := s.db.GetUserByID(ctx, oldHandlerID)
	if err != nil {
		log.Printf("failed to get old handler %d: %v", oldHandlerID, err)
	}

	newUser, err := s.db.GetUserByID(ctx, newHandlerID)
	if err != nil {
		log.Printf("failed to get new handler %d: %v", newHandlerID, err)
		return
	}

	if oldUser != nil {
		log.Printf("[通知] 工单 #%d 已从 %s 升级到 %s，原因: %s", workOrderID, oldUser.Username, newUser.Username, reason)
	} else {
		log.Printf("[通知] 工单 #%d 已升级到 %s，原因: %s", workOrderID, newUser.Username, reason)
	}
}

func (s *NotificationService) NotifyTeamLeader(ctx context.Context, workOrderID int64, teamID int64, msg string) {
	team, err := s.db.GetTeamByID(ctx, teamID)
	if err != nil {
		log.Printf("failed to get team %d: %v", teamID, err)
		return
	}

	leader, err := s.db.GetUserByID(ctx, team.LeaderID)
	if err != nil {
		log.Printf("failed to get leader %d: %v", team.LeaderID, err)
		return
	}

	log.Printf("[通知][团队负责人] %s: 工单 #%d - %s", leader.Username, workOrderID, msg)
}

func (s *NotificationService) NotifySubmitter(ctx context.Context, workOrderID int64, submitterID int64, msg string) {
	user, err := s.db.GetUserByID(ctx, submitterID)
	if err != nil {
		log.Printf("failed to get submitter %d: %v", submitterID, err)
		return
	}

	log.Printf("[通知][提交人] %s: 工单 #%d - %s", user.Username, workOrderID, msg)
}

func (s *NotificationService) NotifyStateChanged(ctx context.Context, wo *models.WorkOrder, oldStatus, newStatus models.WorkOrderStatus, operatorID int64) {
	operator, err := s.db.GetUserByID(ctx, operatorID)
	if err != nil {
		log.Printf("failed to get operator %d: %v", operatorID, err)
	}

	operatorName := fmt.Sprintf("用户 %d", operatorID)
	if operator != nil {
		operatorName = operator.Username
	}

	log.Printf("[通知][状态变更] 工单 #%d: %s -> %s (由 %s 操作)", wo.ID, oldStatus, newStatus, operatorName)
}
