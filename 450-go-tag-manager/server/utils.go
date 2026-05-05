package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tag-manager/common"
)

type DuplicateNameError struct {
	Name string
}

func (e *DuplicateNameError) Error() string {
	return fmt.Sprintf("tag name '%s' already exists", e.Name)
}

type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%s' not found", e.Resource, e.ID)
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

const (
	MaxBatchUsers = 100
	TextMaxLength = 200
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func nowUnix() int64 {
	return time.Now().Unix()
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, common.Response{
		Success: false,
		Error:   err.Error(),
	})
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func getIDFromPath(r *http.Request, prefix string) (string, error) {
	path := r.URL.Path
	if len(path) <= len(prefix)+1 {
		return "", errors.New("missing id in path")
	}

	remaining := path[len(prefix):]
	if len(remaining) > 0 && remaining[0] == '/' {
		remaining = remaining[1:]
	}

	slashIdx := strings.Index(remaining, "/")
	if slashIdx == -1 {
		return remaining, nil
	}

	return remaining[:slashIdx], nil
}

func validateTagValue(tag *common.Tag, value interface{}) error {
	switch tag.Type {
	case common.TagTypeEnum:
		strVal, ok := value.(string)
		if !ok {
			return &ValidationError{Message: "enum tag value must be a string"}
		}
		if tag.EnumConfig == nil {
			return &ValidationError{Message: "enum tag has no values configured"}
		}
		found := false
		for _, v := range tag.EnumConfig.Values {
			if v == strVal {
				found = true
				break
			}
		}
		if !found {
			return &ValidationError{Message: fmt.Sprintf("value '%s' is not a valid enum value", strVal)}
		}

	case common.TagTypeNumber:
		var numVal float64
		var ok bool

		switch v := value.(type) {
		case float64:
			numVal = v
			ok = true
		case int:
			numVal = float64(v)
			ok = true
		case int64:
			numVal = float64(v)
			ok = true
		case string:
			parsed, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return &ValidationError{Message: "number tag value must be a number"}
			}
			numVal = parsed
			ok = true
		}

		if !ok {
			return &ValidationError{Message: "number tag value must be a number"}
		}

		if tag.NumberConfig != nil {
			if tag.NumberConfig.Min != nil && numVal < *tag.NumberConfig.Min {
				return &ValidationError{Message: fmt.Sprintf("value %v is below minimum %v", numVal, *tag.NumberConfig.Min)}
			}
			if tag.NumberConfig.Max != nil && numVal > *tag.NumberConfig.Max {
				return &ValidationError{Message: fmt.Sprintf("value %v is above maximum %v", numVal, *tag.NumberConfig.Max)}
			}
		}

	case common.TagTypeText:
		strVal, ok := value.(string)
		if !ok {
			return &ValidationError{Message: "text tag value must be a string"}
		}
		maxLen := TextMaxLength
		if tag.TextConfig != nil && tag.TextConfig.MaxLength > 0 {
			maxLen = tag.TextConfig.MaxLength
		}
		if len(strVal) > maxLen {
			return &ValidationError{Message: fmt.Sprintf("text length %d exceeds maximum %d", len(strVal), maxLen)}
		}
	}

	return nil
}

func validateCreateTagRequest(req *common.CreateTagRequest) error {
	if req.Operator == "" {
		return &ValidationError{Message: "operator is required"}
	}
	if req.Name == "" {
		return &ValidationError{Message: "tag name is required"}
	}

	switch req.Type {
	case common.TagTypeEnum:
		if req.EnumConfig == nil || len(req.EnumConfig.Values) == 0 {
			return &ValidationError{Message: "enum type requires values"}
		}
	case common.TagTypeNumber:
		if req.NumberConfig != nil {
			if req.NumberConfig.Min != nil && req.NumberConfig.Max != nil {
				if *req.NumberConfig.Min > *req.NumberConfig.Max {
					return &ValidationError{Message: "min cannot be greater than max"}
				}
			}
		}
	case common.TagTypeText:
		if req.TextConfig != nil && req.TextConfig.MaxLength <= 0 {
			return &ValidationError{Message: "text max_length must be positive"}
		}
	default:
		return &ValidationError{Message: fmt.Sprintf("invalid tag type: %s", req.Type)}
	}

	return nil
}

func validateFilterGroup(group *common.FilterGroup) error {
	if group.Logic != "AND" && group.Logic != "OR" {
		return &ValidationError{Message: "filter logic must be 'AND' or 'OR'"}
	}

	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return &ValidationError{Message: "filter group must have at least one condition or subgroup"}
	}

	if len(group.Conditions) > 0 && len(group.Groups) > 0 {
		return &ValidationError{Message: "filter group cannot have both conditions and subgroups at the same level"}
	}

	for _, cond := range group.Conditions {
		if cond.TagID == "" {
			return &ValidationError{Message: "condition tag_id is required"}
		}
		if cond.Operator == "" {
			return &ValidationError{Message: "condition operator is required"}
		}
	}

	for _, g := range group.Groups {
		if err := validateFilterGroup(&g); err != nil {
			return err
		}
	}

	return nil
}

func evaluateFilter(group *common.FilterGroup, userID string, userTags map[string]*common.UserTag, tags map[string]*common.Tag) (bool, error) {
	if len(group.Groups) > 0 {
		results := make([]bool, 0, len(group.Groups))
		for _, g := range group.Groups {
			result, err := evaluateFilter(&g, userID, userTags, tags)
			if err != nil {
				return false, err
			}
			results = append(results, result)
		}

		switch group.Logic {
		case "AND":
			for _, r := range results {
				if !r {
					return false, nil
				}
			}
			return true, nil
		case "OR":
			for _, r := range results {
				if r {
					return true, nil
				}
			}
			return false, nil
		}
	}

	for _, cond := range group.Conditions {
		ut := userTags[cond.TagID]
		tag := tags[cond.TagID]

		if ut == nil || tag == nil {
			switch group.Logic {
			case "AND":
				return false, nil
			case "OR":
				continue
			}
		}

		match, err := evaluateCondition(cond, ut, tag)
		if err != nil {
			return false, err
		}

		switch group.Logic {
		case "AND":
			if !match {
				return false, nil
			}
		case "OR":
			if match {
				return true, nil
			}
		}
	}

	switch group.Logic {
	case "AND":
		return true, nil
	case "OR":
		return false, nil
	}

	return false, nil
}

func evaluateCondition(cond common.FilterCondition, ut *common.UserTag, tag *common.Tag) (bool, error) {
	switch tag.Type {
	case common.TagTypeEnum:
		strVal, ok := ut.TagValue.(string)
		if !ok {
			return false, nil
		}
		condVal, ok := cond.Value.(string)
		if !ok {
			return false, nil
		}
		switch cond.Operator {
		case "=", "==", "eq":
			return strVal == condVal, nil
		case "!=", "ne":
			return strVal != condVal, nil
		default:
			return false, &ValidationError{Message: fmt.Sprintf("unsupported operator '%s' for enum type", cond.Operator)}
		}

	case common.TagTypeNumber:
		utNum, ok := toFloat64(ut.TagValue)
		if !ok {
			return false, nil
		}
		condNum, ok := toFloat64(cond.Value)
		if !ok {
			return false, nil
		}
		switch cond.Operator {
		case "=", "==", "eq":
			return utNum == condNum, nil
		case "!=", "ne":
			return utNum != condNum, nil
		case ">", "gt":
			return utNum > condNum, nil
		case ">=", "gte":
			return utNum >= condNum, nil
		case "<", "lt":
			return utNum < condNum, nil
		case "<=", "lte":
			return utNum <= condNum, nil
		default:
			return false, &ValidationError{Message: fmt.Sprintf("unsupported operator '%s' for number type", cond.Operator)}
		}

	case common.TagTypeText:
		strVal, ok := ut.TagValue.(string)
		if !ok {
			return false, nil
		}
		condVal, ok := cond.Value.(string)
		if !ok {
			return false, nil
		}
		switch cond.Operator {
		case "=", "==", "eq":
			return strVal == condVal, nil
		case "!=", "ne":
			return strVal != condVal, nil
		case "contains":
			return contains(strVal, condVal), nil
		case "starts_with":
			return startsWith(strVal, condVal), nil
		case "ends_with":
			return endsWith(strVal, condVal), nil
		default:
			return false, &ValidationError{Message: fmt.Sprintf("unsupported operator '%s' for text type", cond.Operator)}
		}
	}

	return false, nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		parsed, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	}
	return 0, false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
