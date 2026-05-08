package mask

import (
	"strings"
	"unicode/utf8"
)

const maskChar = '*'

type Masker struct {
	rules *DefaultRules
}

func NewMasker(rules *DefaultRules) *Masker {
	if rules == nil {
		rules = NewDefaultRules()
	}
	return &Masker{rules: rules}
}

func (m *Masker) MaskPhone(phone string, rule *MaskRule) string {
	if rule == nil {
		rule = m.rules.Phone
	}
	return maskString(phone, rule.PrefixKeep, rule.SuffixKeep)
}

func (m *Masker) MaskIDCard(idCard string, rule *MaskRule) string {
	if rule == nil {
		rule = m.rules.IDCard
	}
	normalized := strings.ToUpper(strings.TrimSpace(idCard))
	n := utf8.RuneCountInString(normalized)
	if n == 15 {
		return maskString(normalized, rule.PrefixKeep, rule.SuffixKeep)
	}
	if n == 18 {
		prefixKeep := rule.PrefixKeep
		suffixKeep := rule.SuffixKeep
		if prefixKeep > 17 {
			prefixKeep = 17
		}
		if suffixKeep > 1 {
			suffixKeep = 1
		}
		return maskString(normalized, prefixKeep, suffixKeep)
	}
	return idCard
}

func (m *Masker) MaskBankCard(bankCard string, rule *MaskRule) string {
	if rule == nil {
		rule = m.rules.BankCard
	}
	n := len(strings.TrimSpace(bankCard))
	if n < 12 || n > 19 {
		return bankCard
	}
	return maskString(bankCard, rule.PrefixKeep, rule.SuffixKeep)
}

func (m *Masker) MaskEmail(email string, rule *MaskRule) string {
	if rule == nil {
		rule = m.rules.Email
	}
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return email
	}
	username := email[:atIndex]
	domain := email[atIndex:]
	usernameLen := utf8.RuneCountInString(username)
	if usernameLen == 0 {
		return email
	}
	prefixKeep := rule.PrefixKeep
	if prefixKeep > usernameLen {
		prefixKeep = usernameLen
	}
	maskedUsername := maskString(username, prefixKeep, 0)
	return maskedUsername + domain
}

func (m *Masker) MaskName(name string, rule *MaskRule) string {
	if rule == nil {
		rule = m.rules.Name
	}
	return maskString(name, rule.PrefixKeep, rule.SuffixKeep)
}

func maskString(s string, prefixKeep, suffixKeep int) string {
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return s
	}
	if prefixKeep < 0 {
		prefixKeep = 0
	}
	if suffixKeep < 0 {
		suffixKeep = 0
	}
	if prefixKeep >= n || suffixKeep >= n {
		return s
	}
	if prefixKeep+suffixKeep >= n {
		suffixKeep = n - prefixKeep - 1
		if suffixKeep < 0 {
			suffixKeep = 0
		}
	}
	result := make([]rune, 0, n)
	for i := 0; i < prefixKeep; i++ {
		result = append(result, runes[i])
	}
	maskCount := n - prefixKeep - suffixKeep
	for i := 0; i < maskCount; i++ {
		result = append(result, maskChar)
	}
	for i := n - suffixKeep; i < n; i++ {
		result = append(result, runes[i])
	}
	return string(result)
}
