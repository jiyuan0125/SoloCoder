package etccore

import (
	"crypto/rand"
	"encoding/hex"
	"etc-system/common"
	"errors"
	"sync"
	"time"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type Storage struct {
	accounts map[string]*common.Account
	records  []common.PassRecord
	mu       sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		accounts: make(map[string]*common.Account),
		records:  make([]common.PassRecord, 0),
	}
}

func (s *Storage) CreateAccount(req common.CreateAccountRequest) (*common.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.LicensePlate == "" {
		return nil, errors.New("车牌号不能为空")
	}
	if req.BankCard == "" {
		return nil, errors.New("银行卡信息不能为空")
	}
	if req.VehicleType != common.Passenger && req.VehicleType != common.Truck {
		return nil, errors.New("无效的车辆类型")
	}

	if _, exists := s.accounts[req.LicensePlate]; exists {
		return nil, errors.New("该车牌号已绑定账户")
	}

	account := &common.Account{
		ID:           generateID(),
		LicensePlate: req.LicensePlate,
		BankCard:     req.BankCard,
		Balance:      req.Balance,
		Status:       common.StatusNormal,
		OweAmount:    0,
		OweDays:      0,
		CreatedAt:    time.Now(),
		VehicleType:  req.VehicleType,
		Seats:        req.Seats,
		LoadWeight:   req.LoadWeight,
	}

	s.accounts[req.LicensePlate] = account
	return account, nil
}

func (s *Storage) GetAccount(licensePlate string) (*common.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.accounts[licensePlate]
	if !exists {
		return nil, errors.New("账户不存在")
	}

	return account, nil
}

func (s *Storage) Recharge(licensePlate string, amount float64) (*common.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if amount <= 0 {
		return nil, errors.New("充值金额必须大于0")
	}

	account, exists := s.accounts[licensePlate]
	if !exists {
		return nil, errors.New("账户不存在")
	}

	account.Balance += amount

	if account.Status == common.StatusOwe || account.Status == common.StatusBlacklist {
		if account.Balance >= account.OweAmount {
			account.Balance -= account.OweAmount
			account.OweAmount = 0
			account.OweDays = 0
			account.Status = common.StatusNormal
		}
	}

	return account, nil
}

func (s *Storage) ProcessPass(req common.PassRequest) (*common.PassRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[req.LicensePlate]
	if !exists {
		return nil, errors.New("账户不存在")
	}

	if account.Status == common.StatusBlacklist {
		return nil, errors.New("账户已进入黑名单，禁止通行")
	}

	var fee float64
	var paymentStatus common.PaymentStatus

	if req.IsFree {
		fee = 0
		paymentStatus = common.PaymentFree
	} else {
		fee = CalculateFee(account, req.Mileage)
		if fee <= 0 {
			fee = 0
			paymentStatus = common.PaymentFree
		} else {
			if account.Balance >= fee {
				account.Balance -= fee
				paymentStatus = common.PaymentSuccess
			} else if account.Balance >= fee*0.8 {
				remaining := fee - account.Balance
				account.Balance = 0
				account.OweAmount += remaining
				account.OweDays = 0
				account.Status = common.StatusOwe
				paymentStatus = common.PaymentOwe

				if account.OweAmount > 500 {
					account.Status = common.StatusBlacklist
				}
			} else {
				return nil, errors.New("账户余额不足，无法通行")
			}
		}
	}

	record := common.PassRecord{
		ID:            generateID(),
		LicensePlate:  req.LicensePlate,
		EntryStation:  req.EntryStation,
		ExitStation:   req.ExitStation,
		PassDate:      req.PassDate,
		Mileage:       req.Mileage,
		Fee:           fee,
		PaymentStatus: paymentStatus,
		CreatedAt:     time.Now(),
	}

	s.records = append(s.records, record)
	s.accounts[req.LicensePlate] = account

	return &record, nil
}

func (s *Storage) GetRecordsInRange(start, end time.Time) []common.PassRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.PassRecord, 0)
	for _, record := range s.records {
		if record.PassDate.After(start) && record.PassDate.Before(end) {
			result = append(result, record)
		}
		if record.PassDate.Equal(start) || record.PassDate.Equal(end) {
			result = append(result, record)
		}
	}

	return result
}
