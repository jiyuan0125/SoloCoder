package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"contract-manager/internal/models"
)

const (
	ContractsFile       = "contracts.json"
	RenewalHistoryFile  = "renewal-history.json"
	CleanupLogFile      = "cleanup-log.json"
	dateFormat          = "2006-01-02"
)

var (
	ErrContractExists    = errors.New("合同编号已存在")
	ErrContractNotFound  = errors.New("合同不存在")
	ErrContractTerminated= errors.New("合同已终止")
	ErrInvalidExpiry     = errors.New("新到期日期不能早于原到期日期")
)

type Store struct {
	contracts       []models.Contract
	renewalHistory  []models.RenewalHistory
	cleanupLog      []models.CleanupRecord
}

func Load() (*Store, error) {
	s := &Store{
		contracts:       []models.Contract{},
		renewalHistory:  []models.RenewalHistory{},
		cleanupLog:      []models.CleanupRecord{},
	}

	if data, err := os.ReadFile(ContractsFile); err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(data, &s.contracts); err != nil {
				return nil, fmt.Errorf("加载合同数据失败: %v", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取合同文件失败: %v", err)
	}

	if data, err := os.ReadFile(RenewalHistoryFile); err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(data, &s.renewalHistory); err != nil {
				return nil, fmt.Errorf("加载续约历史失败: %v", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取续约历史文件失败: %v", err)
	}

	if data, err := os.ReadFile(CleanupLogFile); err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(data, &s.cleanupLog); err != nil {
				return nil, fmt.Errorf("加载清理日志失败: %v", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取清理日志文件失败: %v", err)
	}

	return s, nil
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.contracts, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化合同数据失败: %v", err)
	}
	if err := os.WriteFile(ContractsFile, data, 0644); err != nil {
		return fmt.Errorf("保存合同数据失败: %v", err)
	}

	data, err = json.MarshalIndent(s.renewalHistory, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化续约历史失败: %v", err)
	}
	if err := os.WriteFile(RenewalHistoryFile, data, 0644); err != nil {
		return fmt.Errorf("保存续约历史失败: %v", err)
	}

	data, err = json.MarshalIndent(s.cleanupLog, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化清理日志失败: %v", err)
	}
	if err := os.WriteFile(CleanupLogFile, data, 0644); err != nil {
		return fmt.Errorf("保存清理日志失败: %v", err)
	}

	return nil
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.ParseInLocation(dateFormat, dateStr, time.Local)
}

func FormatDate(t time.Time) string {
	return t.Format(dateFormat)
}

func (s *Store) AddContract(contract *models.Contract) error {
	for _, c := range s.contracts {
		if c.ID == contract.ID {
			return ErrContractExists
		}
	}

	s.contracts = append(s.contracts, *contract)
	if err := s.save(); err != nil {
		return err
	}

	s.logCleanup(contract.ID, "add", "添加新合同，触发清理操作")
	return s.save()
}

func (s *Store) GetContracts(party string, sortByExpiry bool) []models.Contract {
	var result []models.Contract
	for _, c := range s.contracts {
		if party == "" || c.Party == party {
			result = append(result, c)
		}
	}

	if sortByExpiry {
		sort.Slice(result, func(i, j int) bool {
			return result[i].ExpiryDate.Before(result[j].ExpiryDate)
		})
	}

	return result
}

func (s *Store) FindContract(id string) (*models.Contract, int, error) {
	for i, c := range s.contracts {
		if c.ID == id {
			return &s.contracts[i], i, nil
		}
	}
	return nil, -1, ErrContractNotFound
}

func (s *Store) getNextRenewalID(originalID string) string {
	maxNum := 1
	for _, c := range s.contracts {
		if strings.HasPrefix(c.ID, originalID+"-R") {
			parts := strings.Split(c.ID, "-R")
			if len(parts) == 2 {
				if num, err := strconv.Atoi(parts[1]); err == nil && num > maxNum {
					maxNum = num
				}
			}
		}
	}
	return fmt.Sprintf("%s-R%d", originalID, maxNum+1)
}

func (s *Store) RenewContract(contractID string, newExpiryStr string, isAuto bool) (*models.Contract, error) {
	contract, idx, err := s.FindContract(contractID)
	if err != nil {
		return nil, err
	}

	if contract.Terminated {
		return nil, ErrContractTerminated
	}

	var newExpiry time.Time
	if isAuto {
		newExpiry = contract.ExpiryDate.AddDate(1, 0, 0)
	} else {
		newExpiry, err = ParseDate(newExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("无效的日期格式: %v", err)
		}
		if newExpiry.Before(contract.ExpiryDate) {
			return nil, ErrInvalidExpiry
		}
	}

	originalID := getOriginalID(contract.ID)
	newID := s.getNextRenewalID(originalID)

	prevExpiry := contract.ExpiryDate

	contract.Terminated = true
	contract.TerminatedAt = time.Now()
	s.contracts[idx] = *contract

	newContract := &models.Contract{
		ID:         newID,
		Name:       contract.Name,
		Party:      contract.Party,
		SignDate:   contract.SignDate,
		ExpiryDate: newExpiry,
		AutoRenew:  contract.AutoRenew,
		Terminated: false,
	}
	s.contracts = append(s.contracts, *newContract)

	s.renewalHistory = append(s.renewalHistory, models.RenewalHistory{
		ContractID:     newID,
		OriginalID:     originalID,
		PreviousExpiry: prevExpiry,
		NewExpiry:      newExpiry,
		IsAutoRenew:    isAuto,
		RenewedAt:      time.Now(),
	})

	if err := s.save(); err != nil {
		return nil, err
	}

	renewType := "手动"
	if isAuto {
		renewType = "自动"
	}
	s.logCleanup(newID, "renewal", fmt.Sprintf("%s续约完成，新合同号: %s，原到期日: %s，新到期日: %s",
		renewType, newID, FormatDate(prevExpiry), FormatDate(newExpiry)))
	return newContract, s.save()
}

func (s *Store) CheckAndProcess() ([]string, error) {
	alerts := []string{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	var contractsToTerminate []int
	var contractsToRenew []int

	for i, c := range s.contracts {
		if c.Terminated {
			continue
		}

		daysUntil := int(c.ExpiryDate.Sub(today).Hours() / 24)

		switch {
		case daysUntil == 30:
			alerts = append(alerts, fmt.Sprintf("[提醒] 合同 %s (%s) 将在 30 天后到期 (%s)",
				c.ID, c.Name, FormatDate(c.ExpiryDate)))
		case daysUntil == 14:
			alerts = append(alerts, fmt.Sprintf("[提醒] 合同 %s (%s) 将在 14 天后到期 (%s)",
				c.ID, c.Name, FormatDate(c.ExpiryDate)))
		case daysUntil == 3:
			alerts = append(alerts, fmt.Sprintf("[提醒] 合同 %s (%s) 将在 3 天后到期 (%s)",
				c.ID, c.Name, FormatDate(c.ExpiryDate)))
		case daysUntil == 0:
			alerts = append(alerts, fmt.Sprintf("[提醒] 合同 %s (%s) 今日到期 (%s)",
				c.ID, c.Name, FormatDate(c.ExpiryDate)))
		}

		if daysUntil < 0 {
			if c.AutoRenew {
				contractsToRenew = append(contractsToRenew, i)
			} else {
				contractsToTerminate = append(contractsToTerminate, i)
			}
		}
	}

	for _, idx := range contractsToRenew {
		c := s.contracts[idx]
		alerts = append(alerts, fmt.Sprintf("[自动续约] 合同 %s (%s) 已到期，自动续约一年",
			c.ID, c.Name))
		if _, err := s.RenewContract(c.ID, "", true); err != nil {
			return alerts, fmt.Errorf("自动续约失败: %v", err)
		}
	}

	for _, idx := range contractsToTerminate {
		c := s.contracts[idx]
		alerts = append(alerts, fmt.Sprintf("[终止] 合同 %s (%s) 已到期，标记为已终止",
			c.ID, c.Name))
		s.contracts[idx].Terminated = true
		s.contracts[idx].TerminatedAt = time.Now()
		s.logCleanup(c.ID, "terminate", fmt.Sprintf("合同已到期且未续约，标记为已终止"))
	}

	if err := s.save(); err != nil {
		return alerts, err
	}

	return alerts, nil
}

func (s *Store) logCleanup(contractID, action, details string) {
	s.cleanupLog = append(s.cleanupLog, models.CleanupRecord{
		ContractID: contractID,
		Action:     action,
		Details:    details,
		ExecutedAt: time.Now(),
	})
}

func getOriginalID(id string) string {
	re := regexp.MustCompile(`-R\d+$`)
	return re.ReplaceAllString(id, "")
}
