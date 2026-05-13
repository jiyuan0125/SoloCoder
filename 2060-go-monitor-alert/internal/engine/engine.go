package engine

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"monitor-alert/internal/types"
)

type RuleEngine struct {
	rules []types.AlertRule
}

func NewRuleEngine(rules []types.AlertRule) (*RuleEngine, error) {
	for i := range rules {
		parsed, logicalOp, err := parseCondition(rules[i].Condition)
		if err != nil {
			return nil, fmt.Errorf("解析规则 '%s' 的条件失败: %w", rules[i].Name, err)
		}
		rules[i].ParsedConditions = parsed
		rules[i].LogicalOp = logicalOp
	}
	
	if err := validateRules(rules); err != nil {
		return nil, err
	}
	
	return &RuleEngine{rules: rules}, nil
}

func (e *RuleEngine) Rules() []types.AlertRule {
	return e.rules
}

func validateRules(rules []types.AlertRule) error {
	names := make(map[string]bool)
	for _, r := range rules {
		if r.Name == "" {
			return fmt.Errorf("规则名称不能为空")
		}
		if names[r.Name] {
			return fmt.Errorf("规则名称重复: %s", r.Name)
		}
		names[r.Name] = true
		
		if r.Duration < 0 {
			return fmt.Errorf("规则 '%s' 的持续时间不能为负数", r.Name)
		}
		
		switch r.Severity {
		case types.SeverityInfo, types.SeverityWarning, types.SeverityCritical:
		default:
			return fmt.Errorf("规则 '%s' 的严重级别无效: %s (有效值: info, warning, critical)", r.Name, r.Severity)
		}
	}
	return nil
}

var (
	conditionPattern = regexp.MustCompile(`^\s*(\w+)\s*(>|<|>=|<=|==|!=)\s*([\d.]+)\s*$`)
	operatorMap      = map[string]string{">": ">", "<": "<", ">=": ">=", "<=": "<=", "==": "==", "!=": "!="}
)

func parseCondition(condition string) ([]types.Condition, string, error) {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return nil, "", fmt.Errorf("条件为空")
	}
	
	var logicalOp string
	var parts []string
	
	if strings.Contains(condition, " AND ") {
		parts = strings.Split(condition, " AND ")
		logicalOp = "AND"
	} else if strings.Contains(condition, " OR ") {
		parts = strings.Split(condition, " OR ")
		logicalOp = "OR"
	} else {
		parts = []string{condition}
		logicalOp = ""
	}
	
	conditions := make([]types.Condition, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		matches := conditionPattern.FindStringSubmatch(part)
		if matches == nil {
			return nil, "", fmt.Errorf("条件语法错误: '%s'，正确格式: 指标 运算符 值 (如: cpu_usage > 80)", part)
		}
		
		var threshold float64
		if _, err := fmt.Sscanf(matches[3], "%f", &threshold); err != nil {
			return nil, "", fmt.Errorf("阈值解析失败: '%s'", matches[3])
		}
		
		conditions = append(conditions, types.Condition{
			Metric:    matches[1],
			Operator:  matches[2],
			Threshold: threshold,
		})
	}
	
	return conditions, logicalOp, nil
}

func (e *RuleEngine) Evaluate(rule types.AlertRule, metrics []types.Metric, start time.Time) (bool, string, error) {
	if len(rule.ParsedConditions) == 0 {
		return false, "", fmt.Errorf("规则 '%s' 没有解析的条件", rule.Name)
	}
	
	metricMap := make(map[string]float64)
	for _, m := range metrics {
		metricMap[m.Name] = m.Value
	}
	
	results := make([]bool, 0, len(rule.ParsedConditions))
	var messageParts []string
	
	for _, cond := range rule.ParsedConditions {
		value, exists := metricMap[cond.Metric]
		if !exists {
			return false, "", fmt.Errorf("指标 '%s' 不存在", cond.Metric)
		}
		
		result := evaluateSingleCondition(cond, value)
		results = append(results, result)
		
		rel := "不满足"
		if result {
			rel = "满足"
		}
		messageParts = append(messageParts, fmt.Sprintf("%s %.2f %s %.2f (%s)", cond.Metric, value, cond.Operator, cond.Threshold, rel))
	}
	
	var finalResult bool
	if len(results) == 1 {
		finalResult = results[0]
	} else if rule.LogicalOp == "AND" {
		finalResult = true
		for _, r := range results {
			if !r {
				finalResult = false
				break
			}
		}
	} else if rule.LogicalOp == "OR" {
		finalResult = false
		for _, r := range results {
			if r {
				finalResult = true
				break
			}
		}
	}
	
	duration := time.Since(start)
	if duration < rule.Duration {
		finalResult = false
	}
	
	message := strings.Join(messageParts, " "+rule.LogicalOp+" ")
	return finalResult, message, nil
}

func evaluateSingleCondition(cond types.Condition, value float64) bool {
	switch cond.Operator {
	case ">":
		return value > cond.Threshold
	case "<":
		return value < cond.Threshold
	case ">=":
		return value >= cond.Threshold
	case "<=":
		return value <= cond.Threshold
	case "==":
		return value == cond.Threshold
	case "!=":
		return value != cond.Threshold
	default:
		return false
	}
}
