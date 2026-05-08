package mask

import (
	"regexp"
	"strings"
)

type patterns struct {
	phone    *regexp.Regexp
	idCard   *regexp.Regexp
	bankCard *regexp.Regexp
	email    *regexp.Regexp
	name     *regexp.Regexp
}

var (
	defaultPatterns *patterns
)

func init() {
	defaultPatterns = &patterns{
		phone:    regexp.MustCompile(`\b1[3-9]\d{9}\b`),
		idCard:   regexp.MustCompile(`\b\d{15}\b|\b\d{17}[\dXx]\b`),
		bankCard: regexp.MustCompile(`\b\d{12,19}\b`),
		email:    regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`),
		name:     regexp.MustCompile(`[\u4e00-\u9fa5]{2,4}`),
	}
}

type BatchMasker struct {
	masker *Masker
}

func NewBatchMasker(masker *Masker) *BatchMasker {
	if masker == nil {
		masker = NewMasker(nil)
	}
	return &BatchMasker{masker: masker}
}

func (bm *BatchMasker) MaskText(text string, overrideRules *DefaultRules) string {
	var tempMasker *Masker
	if overrideRules != nil {
		tempMasker = NewMasker(overrideRules)
	} else {
		tempMasker = bm.masker
	}

	result := text

	result = defaultPatterns.phone.ReplaceAllStringFunc(result, func(match string) string {
		rule := tempMasker.rules.Phone
		if overrideRules != nil && overrideRules.Phone != nil {
			rule = overrideRules.Phone
		}
		return tempMasker.MaskPhone(match, rule)
	})

	result = defaultPatterns.idCard.ReplaceAllStringFunc(result, func(match string) string {
		rule := tempMasker.rules.IDCard
		if overrideRules != nil && overrideRules.IDCard != nil {
			rule = overrideRules.IDCard
		}
		return tempMasker.MaskIDCard(match, rule)
	})

	result = defaultPatterns.bankCard.ReplaceAllStringFunc(result, func(match string) string {
		rule := tempMasker.rules.BankCard
		if overrideRules != nil && overrideRules.BankCard != nil {
			rule = overrideRules.BankCard
		}
		if isNotPhone(match) && isNotIDCard(match) {
			return tempMasker.MaskBankCard(match, rule)
		}
		return match
	})

	result = defaultPatterns.email.ReplaceAllStringFunc(result, func(match string) string {
		rule := tempMasker.rules.Email
		if overrideRules != nil && overrideRules.Email != nil {
			rule = overrideRules.Email
		}
		return tempMasker.MaskEmail(match, rule)
	})

	result = defaultPatterns.name.ReplaceAllStringFunc(result, func(match string) string {
		if strings.Contains(result, "的") || strings.Contains(result, "是") {
			rule := tempMasker.rules.Name
			if overrideRules != nil && overrideRules.Name != nil {
				rule = overrideRules.Name
			}
			return tempMasker.MaskName(match, rule)
		}
		return match
	})

	return result
}

func isNotPhone(s string) bool {
	if len(s) != 11 {
		return true
	}
	if s[0] != '1' {
		return true
	}
	if s[1] < '3' || s[1] > '9' {
		return true
	}
	return false
}

func isNotIDCard(s string) bool {
	if len(s) == 15 {
		return false
	}
	if len(s) == 18 {
		last := s[17]
		if (last >= '0' && last <= '9') || last == 'X' || last == 'x' {
			return false
		}
	}
	return true
}
