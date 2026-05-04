package main

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"time"
)

var serialNumberRegex = regexp.MustCompile(`^[a-zA-Z0-9]{12}$`)

func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func ValidateSerialNumber(serialNumber string) error {
	if !serialNumberRegex.MatchString(serialNumber) {
		return &ValidationError{
			Field:   "serial_number",
			Message: "产品序列号必须是12位字母数字",
		}
	}
	return nil
}

func ValidateFailureDesc(desc string) error {
	if len(desc) == 0 {
		return &ValidationError{
			Field:   "failure_desc",
			Message: "故障描述不能为空",
		}
	}
	if len(desc) > 200 {
		return &ValidationError{
			Field:   "failure_desc",
			Message: "故障描述不能超过200字",
		}
	}
	return nil
}

type ApplicationService struct {
	storage *Storage
}

func NewApplicationService(storage *Storage) *ApplicationService {
	return &ApplicationService{storage: storage}
}

func (s *ApplicationService) CreateClaim(req CreateClaimRequest) (*WarrantyClaim, error) {
	if err := ValidateSerialNumber(req.SerialNumber); err != nil {
		return nil, err
	}

	if err := ValidateFailureDesc(req.FailureDesc); err != nil {
		return nil, err
	}

	purchaseDate, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return nil, &ValidationError{
			Field:   "purchase_date",
			Message: "购买日期格式错误，应为YYYY-MM-DD",
		}
	}

	currentDate := time.Now()
	if err := ValidatePurchaseDate(purchaseDate, currentDate); err != nil {
		return nil, err
	}

	if s.storage.HasActiveClaim(req.SerialNumber) {
		return nil, &ValidationError{
			Field:   "serial_number",
			Message: "该产品序列号已有处理中的保修申请",
		}
	}

	claim := &WarrantyClaim{
		ID:           GenerateID(),
		SerialNumber: req.SerialNumber,
		PurchaseDate: purchaseDate,
		FailureDesc:  req.FailureDesc,
		CreatedAt:    currentDate,
		UserID:       req.UserID,
		Appeals:      []Appeal{},
	}

	if IsUnderWarranty(purchaseDate, currentDate) {
		claim.Status = StatusPending
	} else {
		claim.Status = StatusRejected
		claim.RejectReason = "已超过一年保修期"
	}

	if err := s.storage.SaveClaim(claim); err != nil {
		return nil, err
	}

	return claim, nil
}
