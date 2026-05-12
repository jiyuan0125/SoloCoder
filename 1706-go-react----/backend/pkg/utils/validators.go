package utils

import (
	"errors"
	"regexp"
	"time"

	"hospital-pharmacy/pkg/models"
)

var drugCodeRegex = regexp.MustCompile(`^国药准字[H-Z][0-9]{8}$`)

func ValidateDrugCode(code string) error {
	if !drugCodeRegex.MatchString(code) {
		return errors.New("药品编码格式不正确，应为国药准字H/Z等+8位数字")
	}
	return nil
}

func ValidatePrices(retailPrice, purchasePrice float64) error {
	if retailPrice < 0 {
		return errors.New("零售价不能为负数")
	}
	if purchasePrice < 0 {
		return errors.New("进货价不能为负数")
	}
	return nil
}

func ValidateStockQuantity(quantity int) error {
	if quantity < 0 {
		return errors.New("库存数量不能为负数")
	}
	return nil
}

func GetExpiryStatus(expiryDate time.Time) string {
	now := time.Now()
	daysLeft := int(expiryDate.Sub(now).Hours() / 24)

	if expiryDate.Before(now) {
		return "expired"
	} else if daysLeft <= 90 {
		return "critical"
	} else if daysLeft <= 180 {
		return "near"
	}
	return "normal"
}

func GenerateBatchNumber(supplierCode string, seq int64) string {
	return supplierCode + time.Now().Format("060102") + formatSeq(seq)
}

func formatSeq(seq int64) string {
	return string(rune('0' + seq%10)) + string(rune('0' + (seq/10)%10)) + string(rune('0' + (seq/100)%10)) + string(rune('0' + (seq/1000)%10)) + string(rune('0' + (seq/10000)%10)) + string(rune('0' + (seq/100000)%10))
}

func GeneratePrescriptionNo(seq int64) string {
	return "CF" + time.Now().Format("20060102") + formatSeq(seq)
}

func IsInteger(val float64) bool {
	return val == float64(int64(val))
}

func CheckPrescriptionItem(drug *models.Drug, quantity float64) error {
	if !IsInteger(quantity) {
		if !drug.SupportSplit {
			return errors.New("药品" + drug.GenericName + "不支持拆零销售，数量必须为整数")
		}
	}
	return nil
}
