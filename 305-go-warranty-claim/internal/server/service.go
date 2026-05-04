package server

import (
	"fmt"
	"time"
	"warranty-claim/pkg/common"

	"github.com/google/uuid"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) SubmitApplication(req common.SubmitApplicationRequest) common.SubmitApplicationResponse {
	if !common.IsValidSerialNumber(req.SerialNumber) {
		return common.SubmitApplicationResponse{
			Success: false,
			Message: fmt.Sprintf("产品序列号格式错误，必须为%d位字母数字", common.SerialNumberLength),
		}
	}

	if !common.IsValidDescription(req.Description) {
		return common.SubmitApplicationResponse{
			Success: false,
			Message: fmt.Sprintf("故障描述不能为空且不能超过%d字", common.MaxDescriptionLength),
		}
	}

	if !common.IsValidPurchaseDate(req.PurchaseDate) {
		return common.SubmitApplicationResponse{
			Success: false,
			Message: "购买日期格式错误，应为YYYY-MM-DD",
		}
	}

	if !common.IsPurchaseDateNotFuture(req.PurchaseDate) {
		return common.SubmitApplicationResponse{
			Success: false,
			Message: "购买日期不能晚于当前日期",
		}
	}

	if s.store.HasActiveApplication(req.SerialNumber) {
		return common.SubmitApplicationResponse{
			Success: false,
			Message: "该产品序列号已有处理中的保修申请，请等待处理完成",
		}
	}

	withinWarranty := common.IsWithinWarranty(req.PurchaseDate)
	warrantyExpiry := common.CalculateWarrantyExpiry(req.PurchaseDate)

	var status string
	var autoApproved bool
	var rejectReason string

	if withinWarranty {
		status = common.ApplicationStatusAutoApproved
		autoApproved = true
	} else {
		status = common.ApplicationStatusRejected
		rejectReason = "超过一年保修期"
	}

	app := common.WarrantyApplication{
		ID:             uuid.New().String(),
		UserID:         req.UserID,
		SerialNumber:   req.SerialNumber,
		PurchaseDate:   req.PurchaseDate,
		Description:    req.Description,
		Status:         status,
		RejectReason:   rejectReason,
		SubmittedAt:    time.Now().Unix(),
		AutoApproved:   autoApproved,
		WarrantyExpiry: warrantyExpiry,
	}

	s.store.CreateApplication(app)

	return common.SubmitApplicationResponse{
		Success:       true,
		ApplicationID: app.ID,
		Status:        app.Status,
		Message:       s.getMessageForStatus(app.Status, rejectReason),
	}
}

func (s *Service) getMessageForStatus(status, reason string) string {
	switch status {
	case common.ApplicationStatusAutoApproved:
		return "申请已自动通过，进入待处理队列"
	case common.ApplicationStatusRejected:
		return fmt.Sprintf("申请已拒绝: %s", reason)
	default:
		return "申请已提交"
	}
}

func (s *Service) GetApplication(id string) (common.WarrantyApplication, bool) {
	return s.store.GetApplication(id)
}

func (s *Service) GetUserApplications(userID string) []common.WarrantyApplication {
	return s.store.GetApplicationsByUser(userID)
}

func (s *Service) GetAllApplications() []common.WarrantyApplication {
	return s.store.GetAllApplications()
}

func (s *Service) GetPendingApplications() []common.WarrantyApplication {
	return s.store.GetPendingApplications()
}

func (s *Service) ReviewApplication(req common.ReviewApplicationRequest) (bool, string) {
	app, exists := s.store.GetApplication(req.ApplicationID)
	if !exists {
		return false, "申请不存在"
	}

	if app.Status != common.ApplicationStatusPending && app.Status != common.ApplicationStatusAutoApproved {
		return false, "该申请状态不允许审核"
	}

	if req.Approved {
		app.Status = common.ApplicationStatusApproved
		app.ApprovedAt = time.Now().Unix()
		app.RejectReason = ""
	} else {
		app.Status = common.ApplicationStatusRejected
		app.RejectReason = req.Reason
	}

	s.store.UpdateApplication(app)
	return true, "审核成功"
}

func (s *Service) SubmitAppeal(req common.SubmitAppealRequest, userID string) (bool, string) {
	app, exists := s.store.GetApplication(req.ApplicationID)
	if !exists {
		return false, "申请不存在"
	}

	if app.UserID != userID {
		return false, "无权对此申请发起申诉"
	}

	if app.Status != common.ApplicationStatusRejected && app.Status != common.ApplicationStatusApproved {
		return false, "只能对已审核的申请发起申诉"
	}

	if len(req.Reason) == 0 {
		return false, "申诉原因不能为空"
	}

	appeal := common.Appeal{
		ID:            uuid.New().String(),
		ApplicationID: req.ApplicationID,
		Reason:        req.Reason,
		Status:        common.AppealStatusPending,
		SubmittedAt:   time.Now().Unix(),
	}

	app.Status = common.ApplicationStatusAppealed
	s.store.UpdateApplication(app)
	s.store.CreateAppeal(appeal)

	return true, "申诉已提交"
}

func (s *Service) GetAppealsByApplication(applicationID string) []common.Appeal {
	return s.store.GetAppealsByApplication(applicationID)
}

func (s *Service) GetPendingAppeals() []common.AppealWithApplication {
	appeals := s.store.GetPendingAppeals()
	var result []common.AppealWithApplication

	for _, appeal := range appeals {
		if app, exists := s.store.GetApplication(appeal.ApplicationID); exists {
			result = append(result, common.AppealWithApplication{
				Appeal:      appeal,
				Application: app,
			})
		}
	}

	return result
}

func (s *Service) ResolveAppeal(appealID string, resolved bool, adminNote string) (bool, string) {
	appeal, exists := s.store.GetAppeal(appealID)
	if !exists {
		return false, "申诉不存在"
	}

	if appeal.Status != common.AppealStatusPending {
		return false, "该申诉已处理"
	}

	if resolved {
		appeal.Status = common.AppealStatusResolved
	} else {
		appeal.Status = common.AppealStatusDismissed
	}

	appeal.AdminNote = adminNote
	appeal.ResolvedAt = time.Now().Unix()

	s.store.UpdateAppeal(appeal)

	app, exists := s.store.GetApplication(appeal.ApplicationID)
	if exists {
		if resolved {
			app.Status = common.ApplicationStatusApproved
			app.ApprovedAt = time.Now().Unix()
		} else {
			app.Status = common.ApplicationStatusRejected
		}
		s.store.UpdateApplication(app)
	}

	return true, "申诉已处理"
}

func (s *Service) GetStatistics() common.StatisticsResponse {
	return s.store.GetStatistics()
}
