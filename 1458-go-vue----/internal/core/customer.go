package core

import (
	"errors"
	"fmt"
	"merchant-mgmt-system/pkg/common"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

func formatPhone(phone string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(phone, "")
}

func (s *Service) CreateCustomer(req *common.CreateCustomerRequest) (*common.Customer, error) {
	if req.CustomerName == "" {
		return nil, errors.New("客户名称不能为空")
	}
	if req.ContactPerson == "" {
		return nil, errors.New("联系人不能为空")
	}
	if req.ContactPhone == "" {
		return nil, errors.New("联系电话不能为空")
	}

	formattedPhone := formatPhone(req.ContactPhone)
	if formattedPhone == "" {
		return nil, errors.New("联系电话格式无效")
	}

	s.createMu.Lock()
	defer s.createMu.Unlock()

	if existing, exists := s.storage.FindCustomerByNamePhone(req.CustomerName, formattedPhone); exists {
		return nil, fmt.Errorf("客户已存在 (ID: %s)", existing.ID)
	}

	now := time.Now()
	customer := &common.Customer{
		ID:             uuid.New().String(),
		CustomerName:   req.CustomerName,
		ContactPerson:  req.ContactPerson,
		ContactPhone:   formattedPhone,
		IntentArea:     req.IntentArea,
		IntentBusiness: req.IntentBusiness,
		ExpectedMoveIn: req.ExpectedMoveIn,
		Status:         common.StatusNew,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.storage.SaveCustomer(customer)
	return customer, nil
}

func (s *Service) ListCustomers(status common.CustomerStatus) ([]*common.Customer, error) {
	all := s.storage.GetAllCustomers()
	if status == "" {
		return all, nil
	}

	filtered := []*common.Customer{}
	for _, c := range all {
		if c.Status == status {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

func (s *Service) GetCustomer(id string) (*common.Customer, error) {
	c, ok := s.storage.GetCustomer(id)
	if !ok {
		return nil, errors.New("客户不存在")
	}
	return c, nil
}

func (s *Service) updateCustomerStatus(customerID string, status common.CustomerStatus) error {
	c, ok := s.storage.GetCustomer(customerID)
	if !ok {
		return errors.New("客户不存在")
	}
	c.Status = status
	c.UpdatedAt = time.Now()
	s.storage.SaveCustomer(c)
	return nil
}

func (s *Service) UpdateCustomerStatus(customerID string, status common.CustomerStatus) error {
	return s.updateCustomerStatus(customerID, status)
}

func (s *Service) CheckAndUpdateWarningStatus() {
	now := time.Now()
	warningThreshold := 48 * time.Hour

	customers := s.storage.GetAllCustomers()
	for _, c := range customers {
		if c.Status != common.StatusFollowing && c.Status != common.StatusWarning {
			continue
		}

		follows := s.storage.GetFollowsByCustomer(c.ID)
		if len(follows) == 0 {
			continue
		}

		latestPlan := findLatestOpenPlanTime(follows)
		if latestPlan == nil {
			continue
		}

		latestFollowTime := findLatestFollowTime(follows)

		if now.Sub(*latestPlan) > warningThreshold {
			if latestFollowTime == nil || latestFollowTime.Before(*latestPlan) {
				if c.Status != common.StatusWarning {
					s.updateCustomerStatus(c.ID, common.StatusWarning)
				}
			}
		}
	}
}

func findLatestOpenPlanTime(follows []*common.FollowRecord) *time.Time {
	var latest *time.Time
	for _, f := range follows {
		if f.Status == common.FollowStatusOpen && f.NextPlanTime != nil {
			if latest == nil || f.NextPlanTime.After(*latest) {
				latest = f.NextPlanTime
			}
		}
	}
	return latest
}

func findLatestFollowTime(follows []*common.FollowRecord) *time.Time {
	var latest *time.Time
	for _, f := range follows {
		if latest == nil || f.CreatedAt.After(*latest) {
			t := f.CreatedAt
			latest = &t
		}
	}
	return latest
}

func ValidateBusinessType(bt string) (common.BusinessType, error) {
	bt = strings.ToUpper(bt)
	switch common.BusinessType(bt) {
	case common.BusinessTypeRestaurant,
		common.BusinessTypeRetail,
		common.BusinessTypeOffice,
		common.BusinessTypeWarehouse:
		return common.BusinessType(bt), nil
	default:
		return "", fmt.Errorf("无效的业态类型: %s", bt)
	}
}

func ValidateStatus(st string) (common.CustomerStatus, error) {
	if st == "" {
		return "", nil
	}
	st = strings.ToUpper(st)
	switch common.CustomerStatus(st) {
	case common.StatusNew,
		common.StatusFollowing,
		common.StatusWarning,
		common.StatusSigned:
		return common.CustomerStatus(st), nil
	default:
		return "", fmt.Errorf("无效的状态: %s", st)
	}
}
