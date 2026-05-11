package tagparser

import (
	"strconv"

	"go-reflect-tags/pkg/common"
)

func ParseValidateTag(tag string, fieldType string) *common.ValidateTagInfo {
	if tag == "" {
		return nil
	}

	info := &common.ValidateTagInfo{}
	options := splitTagOptions(tag)

	for _, opt := range options {
		key, value := parseKeyValue(opt)
		rule := common.ValidateRule{Name: key}

		switch key {
		case "required":
			rule.Value = ""

		case "email":
			rule.Value = ""

		case "min", "max":
			if value != "" {
				rule.Value = value
				if isNumericType(fieldType) || isStringType(fieldType) {
					if num, err := strconv.ParseInt(value, 10, 64); err == nil {
						rule.Number = num
					}
				}
			}

		case "regex":
			if value != "" {
				rule.Value = value
			}

		default:
			if value != "" {
				rule.Value = value
				if num, err := strconv.ParseInt(value, 10, 64); err == nil {
					rule.Number = num
				}
			}
		}

		info.Rules = append(info.Rules, rule)
	}

	return info
}
