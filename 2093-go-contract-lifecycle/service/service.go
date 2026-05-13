package service

import (
	"fmt"
	"time"

	"contract-lifecycle/models"
	"contract-lifecycle/storage"
)

var statusFlow = map[models.ContractStatus]models.ContractStatus{
	models.StatusDraft:       models.StatusLegalReview,
	models.StatusLegalReview: models.StatusPendingSign,
	models.StatusModifying:   models.StatusLegalReview,
	models.StatusPendingSign: models.StatusSigned,
	models.StatusSigned:      models.StatusPerforming,
	models.StatusPerforming:  models.StatusExpiryRemind,
	models.StatusExpiryRemind: models.StatusExpired,
	models.StatusExpired:     models.StatusTerminated,
}

type ContractService struct {
	storage *storage.Storage
}

func NewContractService(storage *storage.Storage) *ContractService {
	return &ContractService{storage: storage}
}

func (s *ContractService) CreateContract(contract *models.Contract) error {
	if contract.Amount <= 0 {
		return fmt.Errorf("合同金额必须大于0，当前金额: %.2f", contract.Amount)
	}

	if contract.EffectiveDate.Before(contract.SignDate) {
		fmt.Println("警告: 生效日期早于签署日期")
	}

	contract.ID = storage.GenerateID()
	contract.Status = models.StatusDraft
	contract.CreatedAt = time.Now()
	contract.UpdatedAt = time.Now()
	contract.Archived = false

	if err := s.storage.SaveContract(contract); err != nil {
		return err
	}

	return s.logOperation(contract.ID, "create", "", models.StatusDraft, "", "创建合同")
}

func (s *ContractService) UpdateContract(id string, updates *models.Contract) error {
	contract, err := s.storage.GetContract(id)
	if err != nil {
		return err
	}

	if contract.Archived {
		return fmt.Errorf("合同已归档，无法修改。当前状态: %s", contract.Status)
	}

	if contract.Status == models.StatusTerminated {
		return fmt.Errorf("合同已终止，无法修改。当前状态: %s", contract.Status)
	}

	if updates.Title != "" {
		contract.Title = updates.Title
	}
	if len(updates.Parties) > 0 {
		contract.Parties = updates.Parties
	}
	if !updates.SignDate.IsZero() {
		contract.SignDate = updates.SignDate
	}
	if !updates.EffectiveDate.IsZero() {
		contract.EffectiveDate = updates.EffectiveDate
	}
	if !updates.ExpiryDate.IsZero() {
		contract.ExpiryDate = updates.ExpiryDate
	}
	if updates.Amount > 0 {
		contract.Amount = updates.Amount
	}
	if updates.Content != "" {
		contract.Content = updates.Content
	}
	if len(updates.ResourceIDs) > 0 {
		contract.ResourceIDs = updates.ResourceIDs
	}

	if contract.Amount <= 0 {
		return fmt.Errorf("合同金额必须大于0，当前金额: %.2f", contract.Amount)
	}

	if contract.EffectiveDate.Before(contract.SignDate) {
		fmt.Println("警告: 生效日期早于签署日期")
	}

	contract.UpdatedAt = time.Now()

	if err := s.storage.SaveContract(contract); err != nil {
		return err
	}

	return s.logOperation(id, "update", contract.Status, contract.Status, "", "更新合同信息")
}

func (s *ContractService) TransitionStatus(id string, action string, resourceID string) error {
	contract, err := s.storage.GetContract(id)
	if err != nil {
		return err
	}

	if contract.Archived {
		return fmt.Errorf("合同已归档，无法变更状态。当前状态: %s", contract.Status)
	}

	var nextStatus models.ContractStatus
	fromStatus := contract.Status
	var logDetails string

	switch action {
	case "submit":
		if contract.Status == models.StatusDraft {
			nextStatus = models.StatusLegalReview
			logDetails = "提交法务审核"
		} else if contract.Status == models.StatusModifying {
			nextStatus = models.StatusLegalReview
			logDetails = "修改后重新提交审核"
		} else {
			return fmt.Errorf("无法从当前状态 %s 提交", contract.Status)
		}

	case "reject":
		if contract.Status == models.StatusLegalReview {
			nextStatus = models.StatusModifying
			logDetails = "法务审核退回修改"
		} else {
			return fmt.Errorf("无法从当前状态 %s 退回", contract.Status)
		}

	case "approve":
		if contract.Status == models.StatusLegalReview {
			nextStatus = models.StatusPendingSign
			logDetails = "法务审核通过"
		} else {
			return fmt.Errorf("无法从当前状态 %s 批准", contract.Status)
		}

	case "sign":
		if contract.Status == models.StatusPendingSign {
			nextStatus = models.StatusSigned
			logDetails = "合同签署完成"
		} else {
			return fmt.Errorf("无法从当前状态 %s 签署", contract.Status)
		}

	case "perform":
		if contract.Status == models.StatusSigned {
			nextStatus = models.StatusPerforming
			logDetails = "开始履行合同"
		} else {
			return fmt.Errorf("无法从当前状态 %s 开始履行", contract.Status)
		}

	case "expire":
		if contract.Status == models.StatusPerforming || contract.Status == models.StatusExpiryRemind {
			nextStatus = models.StatusExpired
			logDetails = "合同到期"
		} else {
			return fmt.Errorf("无法从当前状态 %s 标记到期", contract.Status)
		}

	default:
		return fmt.Errorf("未知操作: %s。当前状态: %s", action, contract.Status)
	}

	contract.Status = nextStatus
	contract.UpdatedAt = time.Now()

	if err := s.storage.SaveContract(contract); err != nil {
		return err
	}

	return s.logOperation(id, action, fromStatus, nextStatus, resourceID, logDetails)
}

func (s *ContractService) TerminateContract(id string, resourceID string) error {
	contract, err := s.storage.GetContract(id)
	if err != nil {
		return err
	}

	if contract.Status != models.StatusExpired {
		return fmt.Errorf("只有已到期的合同可以终止。当前状态: %s", contract.Status)
	}

	contract.Status = models.StatusTerminated
	contract.Archived = true
	contract.UpdatedAt = time.Now()

	if err := s.storage.SaveContract(contract); err != nil {
		return err
	}

	return s.logOperation(id, "terminate", models.StatusExpired, models.StatusTerminated, resourceID, "合同终止归档")
}

func (s *ContractService) RenewContract(originalID string, newContract *models.Contract, resourceID string) error {
	original, err := s.storage.GetContract(originalID)
	if err != nil {
		return err
	}

	if original.Status != models.StatusExpired {
		return fmt.Errorf("只有已到期的合同可以续约。当前状态: %s", original.Status)
	}

	if newContract.Amount <= 0 {
		return fmt.Errorf("合同金额必须大于0，当前金额: %.2f", newContract.Amount)
	}

	if newContract.EffectiveDate.Before(newContract.SignDate) {
		fmt.Println("警告: 生效日期早于签署日期")
	}

	newContract.ID = storage.GenerateID()
	newContract.Status = models.StatusDraft
	newContract.RelatedContract = originalID
	newContract.CreatedAt = time.Now()
	newContract.UpdatedAt = time.Now()
	newContract.Archived = false

	if err := s.storage.SaveContract(newContract); err != nil {
		return err
	}

	return s.logOperation(newContract.ID, "renew", "", models.StatusDraft, resourceID, "续约合同 (原合同: "+originalID+")")
}

func (s *ContractService) RecordBreach(contractID string, description string, date time.Time, resourceID string) error {
	contract, err := s.storage.GetContract(contractID)
	if err != nil {
		return err
	}

	if contract.Status != models.StatusPerforming {
		return fmt.Errorf("只有履行中的合同可以记录违约。当前状态: %s", contract.Status)
	}

	breach := models.BreachRecord{
		ID:          storage.GenerateID(),
		ContractID:  contractID,
		Description: description,
		Date:        date,
		CreatedAt:   time.Now(),
	}

	contract.BreachRecords = append(contract.BreachRecords, breach)
	contract.UpdatedAt = time.Now()

	if err := s.storage.SaveContract(contract); err != nil {
		return err
	}

	return s.logOperation(contractID, "breach", contract.Status, contract.Status, resourceID, "记录违约: "+description)
}

func (s *ContractService) CheckReminders(resourceID string) ([]string, error) {
	contracts, err := s.storage.LoadContracts()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var reminders []string

	for i := range contracts {
		contract := &contracts[i]
		if contract.Status != models.StatusPerforming {
			continue
		}

		daysToExpiry := int(contract.ExpiryDate.Sub(now).Hours() / 24)
		needsSave := false

		if daysToExpiry <= 30 && daysToExpiry > 14 && !contract.Reminded30 {
			reminders = append(reminders, fmt.Sprintf("[30天提醒] 合同 '%s' (ID: %s) 还剩 %d 天到期", contract.Title, contract.ID, daysToExpiry))
			contract.Reminded30 = true
			needsSave = true
			s.logOperation(contract.ID, "remind", contract.Status, contract.Status, resourceID, "30天到期提醒")
		}

		if daysToExpiry <= 14 && daysToExpiry > 7 && !contract.Reminded14 {
			reminders = append(reminders, fmt.Sprintf("[14天提醒] 合同 '%s' (ID: %s) 还剩 %d 天到期", contract.Title, contract.ID, daysToExpiry))
			contract.Reminded14 = true
			needsSave = true
			s.logOperation(contract.ID, "remind", contract.Status, contract.Status, resourceID, "14天到期提醒")
		}

		if daysToExpiry <= 7 && daysToExpiry > 0 && !contract.Reminded7 {
			reminders = append(reminders, fmt.Sprintf("[7天提醒] 合同 '%s' (ID: %s) 还剩 %d 天到期", contract.Title, contract.ID, daysToExpiry))
			contract.Reminded7 = true
			needsSave = true
			s.logOperation(contract.ID, "remind", contract.Status, contract.Status, resourceID, "7天到期提醒")
		}

		if needsSave {
			s.storage.SaveContract(contract)
		}
	}

	return reminders, nil
}

func (s *ContractService) QueryContracts(status string, party string, startDate, endDate *time.Time) ([]models.Contract, error) {
	contracts, err := s.storage.LoadContracts()
	if err != nil {
		return nil, err
	}

	var result []models.Contract

	for _, contract := range contracts {
		if status != "" && string(contract.Status) != status {
			continue
		}

		if party != "" {
			found := false
			for _, p := range contract.Parties {
				if p == party {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if startDate != nil && contract.ExpiryDate.Before(*startDate) {
			continue
		}
		if endDate != nil && contract.ExpiryDate.After(*endDate) {
			continue
		}

		result = append(result, contract)
	}

	return result, nil
}

func (s *ContractService) GetContract(id string) (*models.Contract, error) {
	return s.storage.GetContract(id)
}

func (s *ContractService) ListContracts() ([]models.Contract, error) {
	return s.storage.LoadContracts()
}

func (s *ContractService) GetLogs(contractID string) ([]models.OperationLog, error) {
	logs, err := s.storage.LoadLogs()
	if err != nil {
		return nil, err
	}

	if contractID == "" {
		return logs, nil
	}

	var result []models.OperationLog
	for _, log := range logs {
		if log.ContractID == contractID {
			result = append(result, log)
		}
	}

	return result, nil
}

func (s *ContractService) CreateResource(resource *models.Resource) error {
	resource.ID = storage.GenerateID()
	return s.storage.SaveResource(resource)
}

func (s *ContractService) GetResource(id string) (*models.Resource, error) {
	return s.storage.GetResource(id)
}

func (s *ContractService) ListResources() ([]models.Resource, error) {
	return s.storage.LoadResources()
}

func (s *ContractService) GetContractsByResource(resourceID string) ([]models.Contract, error) {
	contracts, err := s.storage.LoadContracts()
	if err != nil {
		return nil, err
	}

	var result []models.Contract
	for _, contract := range contracts {
		for _, rid := range contract.ResourceIDs {
			if rid == resourceID {
				result = append(result, contract)
				break
			}
		}
	}

	return result, nil
}

func (s *ContractService) LinkContractToResource(contractID, resourceID string) error {
	contract, err := s.storage.GetContract(contractID)
	if err != nil {
		return err
	}

	if _, err := s.storage.GetResource(resourceID); err != nil {
		return err
	}

	for _, rid := range contract.ResourceIDs {
		if rid == resourceID {
			return fmt.Errorf("合同已关联到此资源")
		}
	}

	contract.ResourceIDs = append(contract.ResourceIDs, resourceID)
	contract.UpdatedAt = time.Now()

	if err := s.storage.SaveContract(contract); err != nil {
		return err
	}

	return s.logOperation(contractID, "link_resource", contract.Status, contract.Status, resourceID, "关联资源")
}

func (s *ContractService) logOperation(contractID, action string, fromStatus, toStatus models.ContractStatus, resourceID, details string) error {
	log := models.OperationLog{
		ID:         storage.GenerateID(),
		ContractID: contractID,
		Action:     action,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ResourceID: resourceID,
		Details:    details,
		Timestamp:  time.Now(),
	}

	return s.storage.AddLog(log)
}
