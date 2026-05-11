package core

import (
	"errors"
	"fmt"
	"merchant-mgmt-system/pkg/common"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxOpenFollows = 10

func (s *Service) CreateFollow(req *common.CreateFollowRequest) (*common.FollowRecord, error) {
	if req.CustomerID == "" {
		return nil, errors.New("客户ID不能为空")
	}
	if req.Content == "" {
		return nil, errors.New("跟进内容不能为空")
	}

	if _, ok := s.storage.GetCustomer(req.CustomerID); !ok {
		return nil, errors.New("客户不存在")
	}

	openCount := s.storage.GetOpenFollowsByCustomer(req.CustomerID)
	if openCount >= MaxOpenFollows {
		return nil, fmt.Errorf("该客户已有 %d 条未关闭的跟进记录，最多允许 %d 条", openCount, MaxOpenFollows)
	}

	now := time.Now()
	follow := &common.FollowRecord{
		ID:           uuid.New().String(),
		CustomerID:   req.CustomerID,
		Method:       req.Method,
		Content:      req.Content,
		NextPlanTime: req.NextPlanTime,
		Status:       common.FollowStatusOpen,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.storage.SaveFollow(follow)

	c, ok := s.storage.GetCustomer(req.CustomerID)
	if ok && c.Status != common.StatusSigned {
		if c.Status == common.StatusNew || c.Status == common.StatusWarning {
			s.updateCustomerStatus(req.CustomerID, common.StatusFollowing)
		}
	}

	return follow, nil
}

func (s *Service) ListFollows(customerID string) ([]*common.FollowRecord, error) {
	if customerID == "" {
		return s.storage.GetAllFollows(), nil
	}
	return s.storage.GetFollowsByCustomer(customerID), nil
}

func (s *Service) CloseFollow(id string) error {
	f, ok := s.storage.GetFollow(id)
	if !ok {
		return errors.New("跟进记录不存在")
	}
	f.Status = common.FollowStatusClose
	f.UpdatedAt = time.Now()
	s.storage.SaveFollow(f)
	return nil
}

func ValidateFollowMethod(fm string) (common.FollowMethod, error) {
	fm = strings.ToUpper(fm)
	switch common.FollowMethod(fm) {
	case common.FollowMethodPhone,
		common.FollowMethodWeChat,
		common.FollowMethodMeeting,
		common.FollowMethodEmail:
		return common.FollowMethod(fm), nil
	default:
		return "", fmt.Errorf("无效的跟进方式: %s", fm)
	}
}
