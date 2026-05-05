package service

import (
	"datacleanser/internal/common"
	"datacleanser/internal/server/storage"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CleanerService struct {
	storage *storage.Storage
}

func NewCleanerService(s *storage.Storage) *CleanerService {
	return &CleanerService{storage: s}
}

func (cs *CleanerService) Clean(records []map[string]interface{}, rule common.CleanRule) (*common.CleanResult, error) {
	if len(records) > common.MaxBatchSize {
		return nil, fmt.Errorf("batch size exceeds limit: %d (max %d)", len(records), common.MaxBatchSize)
	}

	result := &common.CleanResult{
		CleanedRecords:   make([]map[string]interface{}, 0),
		DuplicateRecords: make([]map[string]interface{}, 0),
		InvalidRecords:   make([]map[string]interface{}, 0),
		ErrorRecords:     make([]map[string]interface{}, 0),
		Report: common.CleanReport{
			TotalRecords:  len(records),
			RuleHits:      make(map[string]int),
			StartTime:     time.Now(),
		},
	}

	seenKeys := make(map[string]bool)
	prevRecord := make(map[string]interface{})
	batchID := uuid.New().String()

	for _, record := range records {
		recordCopy := deepCopyRecord(record)

		if cs.isInvalidRecord(recordCopy, rule.KeyFields) {
			result.InvalidRecords = append(result.InvalidRecords, recordCopy)
			result.Report.InvalidCount++
			result.Report.RuleHits["invalid"]++
			continue
		}

		if cs.isDuplicate(recordCopy, rule.DedupFields, seenKeys) {
			result.DuplicateRecords = append(result.DuplicateRecords, recordCopy)
			result.Report.DuplicateCount++
			result.Report.RuleHits["duplicate"]++
			continue
		}

		recordModified := false

		if err := cs.fillNulls(recordCopy, rule.DefaultValues, prevRecord); err != nil {
			errorSample := common.ErrorSample{
				Record:    recordCopy,
				Error:     "fill nulls error: " + err.Error(),
				Timestamp: time.Now(),
				BatchID:   batchID,
			}
			cs.storage.AddErrorSample(errorSample)
			result.ErrorRecords = append(result.ErrorRecords, recordCopy)
			result.Report.RuleHits["error"]++
			continue
		}

		if cs.normalizePhone(recordCopy, rule.PhoneFields) {
			recordModified = true
			result.Report.RuleHits["phone_normalize"]++
		}

		if cs.normalizeDate(recordCopy, rule.DateFields) {
			recordModified = true
			result.Report.RuleHits["date_normalize"]++
		}

		if cs.normalizeAmount(recordCopy, rule.AmountFields) {
			recordModified = true
			result.Report.RuleHits["amount_normalize"]++
		}

		if recordModified {
			result.Report.CorrectedCount++
		}

		result.CleanedRecords = append(result.CleanedRecords, recordCopy)
		prevRecord = recordCopy
	}

	result.Report.EndTime = time.Now()
	result.Report.QualityScore = cs.calculateQualityScore(result.Report, len(records))

	return result, nil
}

func (cs *CleanerService) isInvalidRecord(record map[string]interface{}, keyFields []string) bool {
	if len(keyFields) == 0 {
		return false
	}

	allEmpty := true
	for _, field := range keyFields {
		if val, exists := record[field]; exists && !isEmptyValue(val) {
			allEmpty = false
			break
		}
	}

	return allEmpty
}

func isEmptyValue(val interface{}) bool {
	if val == nil {
		return true
	}

	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case int, int64, float64, bool:
		return false
	default:
		return true
	}
}

func (cs *CleanerService) isDuplicate(record map[string]interface{}, dedupFields []string, seenKeys map[string]bool) bool {
	if len(dedupFields) == 0 {
		return false
	}

	keyParts := make([]string, 0, len(dedupFields))
	for _, field := range dedupFields {
		if val, exists := record[field]; exists {
			keyParts = append(keyParts, fmt.Sprintf("%v", val))
		} else {
			keyParts = append(keyParts, "")
		}
	}

	key := strings.Join(keyParts, "|")
	if seenKeys[key] {
		return true
	}

	seenKeys[key] = true
	return false
}

func (cs *CleanerService) fillNulls(record map[string]interface{}, defaults map[string]string, prevRecord map[string]interface{}) error {
	for field, defaultVal := range defaults {
		if val, exists := record[field]; !exists || isEmptyValue(val) {
			if prevVal, prevExists := prevRecord[field]; prevExists && !isEmptyValue(prevVal) {
				record[field] = prevVal
			} else {
				record[field] = defaultVal
			}
		}
	}
	return nil
}

var (
	phoneRegex = regexp.MustCompile(`[\s\-]+`)
	dateRegex  = regexp.MustCompile(`^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$`)
)

func (cs *CleanerService) normalizePhone(record map[string]interface{}, phoneFields []string) bool {
	modified := false
	for _, field := range phoneFields {
		if val, exists := record[field]; exists {
			strVal := fmt.Sprintf("%v", val)
			cleaned := phoneRegex.ReplaceAllString(strVal, "")
			if len(cleaned) == 11 {
				if cleaned != strVal {
					record[field] = cleaned
					modified = true
				}
			}
		}
	}
	return modified
}

func (cs *CleanerService) normalizeDate(record map[string]interface{}, dateFields []string) bool {
	modified := false
	for _, field := range dateFields {
		if val, exists := record[field]; exists {
			strVal := fmt.Sprintf("%v", val)
			matches := dateRegex.FindStringSubmatch(strVal)
			if len(matches) == 4 {
				year := matches[1]
				month := fmt.Sprintf("%02s", matches[2])
				day := fmt.Sprintf("%02s", matches[3])
				normalized := fmt.Sprintf("%s-%s-%s", year, month, day)
				if normalized != strVal {
					record[field] = normalized
					modified = true
				}
			}
		}
	}
	return modified
}

func (cs *CleanerService) normalizeAmount(record map[string]interface{}, amountFields []string) bool {
	modified := false
	for _, field := range amountFields {
		if val, exists := record[field]; exists {
			var floatVal float64
			var err error

			switch v := val.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			case int64:
				floatVal = float64(v)
			case string:
				floatVal, err = strconv.ParseFloat(v, 64)
				if err != nil {
					continue
				}
			default:
				continue
			}

			rounded := math.Round(floatVal*100) / 100
			if rounded != floatVal {
				record[field] = rounded
				modified = true
			}
		}
	}
	return modified
}

func (cs *CleanerService) calculateQualityScore(report common.CleanReport, total int) float64 {
	if total == 0 {
		return 100.0
	}

	validRecords := total - report.DuplicateCount - report.InvalidCount
	if validRecords <= 0 {
		return 0.0
	}

	correctionRatio := float64(report.CorrectedCount) / float64(validRecords)
	duplicateRatio := float64(report.DuplicateCount) / float64(total)
	invalidRatio := float64(report.InvalidCount) / float64(total)

	score := 100.0
	score -= correctionRatio * 10
	score -= duplicateRatio * 30
	score -= invalidRatio * 50

	if score < 0 {
		score = 0
	}

	return math.Round(score*100) / 100
}

func deepCopyRecord(record map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range record {
		switch val := v.(type) {
		case map[string]interface{}:
			copy[k] = deepCopyRecord(val)
		case []interface{}:
			copy[k] = deepCopySlice(val)
		default:
			copy[k] = v
		}
	}
	return copy
}

func deepCopySlice(slice []interface{}) []interface{} {
	copy := make([]interface{}, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case map[string]interface{}:
			copy[i] = deepCopyRecord(val)
		case []interface{}:
			copy[i] = deepCopySlice(val)
		default:
			copy[i] = v
		}
	}
	return copy
}

func (cs *CleanerService) CleanAsync(request common.CleanRequest) (string, error) {
	if len(request.Records) > common.MaxBatchSize {
		return "", fmt.Errorf("batch size exceeds limit: %d (max %d)", len(request.Records), common.MaxBatchSize)
	}

	var rule common.CleanRule
	if request.TemplateName != "" {
		template, exists := cs.storage.GetTemplate(request.TemplateName)
		if !exists {
			return "", fmt.Errorf("template not found: %s", request.TemplateName)
		}
		rule = template.Rule
	} else if request.Rule != nil {
		rule = *request.Rule
	} else {
		return "", fmt.Errorf("either rule or template_name must be provided")
	}

	task := cs.storage.CreateTask(request)

	go cs.runTask(task.ID, request.Records, rule)

	return task.ID, nil
}

func (cs *CleanerService) runTask(taskID string, records []map[string]interface{}, rule common.CleanRule) {
	cs.storage.UpdateTaskStatus(taskID, common.TaskRunning)

	result, err := cs.Clean(records, rule)
	if err != nil {
		cs.storage.UpdateTaskError(taskID, err.Error())
		return
	}

	cs.storage.UpdateTaskResult(taskID, result)
}

func (cs *CleanerService) GetTaskStatus(taskID string) (*common.AsyncTask, bool) {
	return cs.storage.GetTask(taskID)
}

func (cs *CleanerService) CreateTemplate(name string, rule common.CleanRule) (*common.Template, error) {
	template := &common.Template{
		Name: name,
		Rule: rule,
	}

	err := cs.storage.CreateTemplate(template)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (cs *CleanerService) GetTemplate(name string) (*common.Template, bool) {
	return cs.storage.GetTemplate(name)
}

func (cs *CleanerService) ListTemplates() []common.Template {
	return cs.storage.ListTemplates()
}

func (cs *CleanerService) UpdateTemplate(name string, rule common.CleanRule) (*common.Template, error) {
	return cs.storage.UpdateTemplate(name, rule)
}

func (cs *CleanerService) RollbackTemplate(name string, version int) (*common.Template, error) {
	return cs.storage.RollbackTemplate(name, version)
}

func (cs *CleanerService) GetErrorSamples() []common.ErrorSample {
	return cs.storage.GetErrorSamples()
}
