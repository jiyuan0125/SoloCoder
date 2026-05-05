package security

import (
	"regexp"
	"strings"
)

var (
	phoneRegex  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	idCardRegex = regexp.MustCompile(`^(\d{6})(\d{8})(\d{3}[\dXx])$`)
)

type Masker struct{}

func NewMasker() *Masker {
	return &Masker{}
}

func (m *Masker) MaskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) == 11 && phoneRegex.MatchString(phone) {
		return phone[:3] + "****" + phone[7:]
	}
	if len(phone) > 7 {
		return phone[:3] + "****" + phone[len(phone)-4:]
	}
	return m.maskGeneral(phone)
}

func (m *Masker) MaskIDCard(idCard string) string {
	idCard = strings.TrimSpace(idCard)
	if len(idCard) == 18 {
		if idCardRegex.MatchString(idCard) {
			return idCard[:3] + "**************" + idCard[15:]
		}
		return idCard[:3] + "**************" + idCard[15:]
	}
	if len(idCard) == 15 {
		return idCard[:3] + "*********" + idCard[12:]
	}
	if len(idCard) > 6 {
		return idCard[:3] + strings.Repeat("*", len(idCard)-6) + idCard[len(idCard)-3:]
	}
	return m.maskGeneral(idCard)
}

func (m *Masker) maskGeneral(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 2 {
		return "*"
	}
	if len(value) <= 4 {
		return value[:1] + strings.Repeat("*", len(value)-1)
	}
	showLen := len(value) / 4
	return value[:showLen] + strings.Repeat("*", len(value)-showLen*2) + value[len(value)-showLen:]
}

func (m *Masker) Mask(value string, fieldType string) string {
	switch fieldType {
	case "phone":
		return m.MaskPhone(value)
	case "idcard":
		return m.MaskIDCard(value)
	default:
		return value
	}
}

func (m *Masker) MaskIfSensitive(value string, sensitive bool, fieldType string) string {
	if !sensitive {
		return value
	}
	return m.Mask(value, fieldType)
}
