package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"dataconverter/config"
	"dataconverter/utils"
)

type ConvertResult struct {
	Data        map[string]interface{} `json:"data"`
	Errors      []FieldError           `json:"errors,omitempty"`
	Success     bool                   `json:"success"`
	RuleName    string                 `json:"rule_name"`
	Partial     bool                   `json:"partial,omitempty"`
}

type FieldError struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Error      string `json:"error"`
}

type RuleEngine struct {
	rules    map[string]config.Rule
	rwMutex  sync.RWMutex
}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules: make(map[string]config.Rule),
	}
}

func (e *RuleEngine) LoadRules(rules []config.Rule) {
	e.rwMutex.Lock()
	defer e.rwMutex.Unlock()
	
	e.rules = make(map[string]config.Rule)
	for _, rule := range rules {
		if rule.Enabled {
			e.rules[rule.Name] = rule
		}
	}
}

func (e *RuleEngine) GetRuleNames() []string {
	e.rwMutex.RLock()
	defer e.rwMutex.RUnlock()
	
	names := make([]string, 0, len(e.rules))
	for name := range e.rules {
		names = append(names, name)
	}
	return names
}

func (e *RuleEngine) HasRule(name string) bool {
	e.rwMutex.RLock()
	defer e.rwMutex.RUnlock()
	
	_, exists := e.rules[name]
	return exists
}

func (e *RuleEngine) Convert(data []byte, ruleName string) (*ConvertResult, error) {
	e.rwMutex.RLock()
	rule, exists := e.rules[ruleName]
	e.rwMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleName)
	}
	
	var sourceData map[string]interface{}
	if err := json.Unmarshal(data, &sourceData); err != nil {
		return nil, fmt.Errorf("failed to parse input JSON: %w", err)
	}
	
	result := &ConvertResult{
		Data:     make(map[string]interface{}),
		Errors:   []FieldError{},
		Success:  true,
		RuleName: ruleName,
	}
	
	errorStrategy := strings.ToLower(rule.ErrorStrategy)
	if errorStrategy == "" {
		errorStrategy = "partial"
	}
	
	failFast := errorStrategy == "fail_fast"
	failAll := errorStrategy == "fail_all"
	
	for _, mapping := range rule.FieldMappings {
		sourceValue, found := utils.GetByPath(sourceData, mapping.SourcePath)
		
		if !found {
			if mapping.SkipIfMissing {
				continue
			}
			
			if mapping.Required {
				err := FieldError{
					SourcePath: mapping.SourcePath,
					TargetPath: mapping.TargetPath,
					Error:      "required field missing",
				}
				result.Errors = append(result.Errors, err)
				result.Success = false
				result.Partial = true
				
				if failFast || failAll {
					if failAll {
						result.Data = map[string]interface{}{}
					}
					return result, nil
				}
				continue
			}
			
			if mapping.DefaultValue != nil {
				utils.SetByPath(result.Data, mapping.TargetPath, mapping.DefaultValue)
			}
			continue
		}
		
		convertedValue, err := utils.ConvertType(sourceValue, mapping.SourceType, mapping.TargetType)
		if err != nil {
			fieldErr := FieldError{
				SourcePath: mapping.SourcePath,
				TargetPath: mapping.TargetPath,
				Error:      fmt.Sprintf("type conversion failed: %v", err),
			}
			result.Errors = append(result.Errors, fieldErr)
			result.Success = false
			result.Partial = true
			
			if failFast || failAll {
				if failAll {
					result.Data = map[string]interface{}{}
				}
				return result, nil
			}
			
			if mapping.DefaultValue != nil {
				utils.SetByPath(result.Data, mapping.TargetPath, mapping.DefaultValue)
			}
			continue
		}
		
		if mapping.Transform != nil {
			transformed, transformErr := applyTransform(convertedValue, mapping.Transform)
			if transformErr != nil {
				fieldErr := FieldError{
					SourcePath: mapping.SourcePath,
					TargetPath: mapping.TargetPath,
					Error:      fmt.Sprintf("transform failed: %v", transformErr),
				}
				result.Errors = append(result.Errors, fieldErr)
				result.Success = false
				result.Partial = true
				
				if failFast || failAll {
					if failAll {
						result.Data = map[string]interface{}{}
					}
					return result, nil
				}
				
				if mapping.DefaultValue != nil {
					utils.SetByPath(result.Data, mapping.TargetPath, mapping.DefaultValue)
				}
				continue
			}
			convertedValue = transformed
		}
		
		utils.SetByPath(result.Data, mapping.TargetPath, convertedValue)
	}
	
	if !result.Success && failAll {
		result.Data = map[string]interface{}{}
	}
	
	return result, nil
}

func applyTransform(value interface{}, transform *config.Transform) (interface{}, error) {
	switch strings.ToLower(transform.Type) {
	case "uppercase":
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("uppercase requires string value")
		}
		return strings.ToUpper(s), nil
	case "lowercase":
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("lowercase requires string value")
		}
		return strings.ToLower(s), nil
	case "trim":
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("trim requires string value")
		}
		return strings.TrimSpace(s), nil
	case "prefix":
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("prefix requires string value")
		}
		prefix := transform.Params["prefix"]
		return prefix + s, nil
	case "suffix":
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("suffix requires string value")
		}
		suffix := transform.Params["suffix"]
		return s + suffix, nil
	case "replace":
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("replace requires string value")
		}
		old := transform.Params["old"]
		newVal := transform.Params["new"]
		return strings.ReplaceAll(s, old, newVal), nil
	default:
		return value, nil
	}
}
