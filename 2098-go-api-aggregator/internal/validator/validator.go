package validator

import (
	"encoding/json"
	"fmt"
	"net/url"

	"api-aggregator/internal/aggregator"
	"api-aggregator/internal/config"
	"api-aggregator/internal/db"
	"api-aggregator/internal/formatter"
)

type ValidationResult struct {
	OriginalTotal  float64
	VerifiedTotal  float64
	Difference     float64
	Items          []map[string]interface{}
	NeedsUpdate    bool
}

func ValidateAndProportionallyUpdate(configID int64, query url.Values, confirm bool) (*ValidationResult, error) {
	cfg, err := db.GetAggregationConfig(configID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("config not found")
	}

	queryStr := config.GenerateQueryString(query)
	queryHash := hashString(queryStr)

	existing, err := db.GetRecordByQuery(configID, queryHash)
	if err != nil {
		return nil, err
	}

	result, err := aggregator.AggregateWithStoredConfig(configID, query)
	if err != nil {
		return nil, err
	}

	if !result.AllSuccess {
		return nil, fmt.Errorf("validation failed: some endpoints failed")
	}

	verifiedItems, verifiedTotal, err := extractTotalAndItems(result.Results)
	if err != nil {
		return nil, err
	}

	resultObj := &ValidationResult{
		VerifiedTotal: verifiedTotal,
		Items:         verifiedItems,
	}

	if existing != nil {
		resultObj.OriginalTotal = existing.Total
		resultObj.Difference = verifiedTotal - existing.Total
		resultObj.NeedsUpdate = verifiedTotal != existing.Total

		if confirm && resultObj.NeedsUpdate {
			updatedItems, err := proportionallyUpdateItems(existing, verifiedTotal)
			if err != nil {
				return nil, err
			}

			itemsJSON, _ := json.Marshal(updatedItems)
			existing.Total = verifiedTotal
			existing.ItemsJSON = itemsJSON

			if err := db.UpdateRecord(existing); err != nil {
				return nil, err
			}
			resultObj.Items = updatedItems
		}
	} else if confirm {
		itemsJSON, _ := json.Marshal(verifiedItems)
		rec := &db.Record{
			ConfigID:  configID,
			QueryHash: queryHash,
			Total:     verifiedTotal,
			ItemsJSON: itemsJSON,
		}
		if err := db.CreateRecord(rec); err != nil {
			return nil, err
		}
	}

	return resultObj, nil
}

func extractTotalAndItems(results []*formatter.EndpointResult) ([]map[string]interface{}, float64, error) {
	var items []map[string]interface{}
	var total float64

	for _, r := range results {
		if !r.Success {
			continue
		}

		var data interface{}
		if err := json.Unmarshal(r.Body, &data); err != nil {
			continue
		}

		switch v := data.(type) {
		case map[string]interface{}:
			if t, ok := extractFloat(v, "total"); ok {
				total += t
			}
			if itemsData, ok := v["items"].([]interface{}); ok {
				for _, item := range itemsData {
					if itemMap, ok := item.(map[string]interface{}); ok {
						items = append(items, itemMap)
					}
				}
			}
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					items = append(items, itemMap)
					if val, ok := extractFloat(itemMap, "value"); ok {
						total += val
					} else if val, ok := extractFloat(itemMap, "amount"); ok {
						total += val
					}
				}
			}
		}
	}

	return items, total, nil
}

func extractFloat(m map[string]interface{}, key string) (float64, bool) {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val, true
		case int:
			return float64(val), true
		case int64:
			return float64(val), true
		}
	}
	return 0, false
}

func proportionallyUpdateItems(existing *db.Record, newTotal float64) ([]map[string]interface{}, error) {
	var items []map[string]interface{}
	if err := json.Unmarshal(existing.ItemsJSON, &items); err != nil {
		return nil, err
	}

	if existing.Total == 0 {
		return items, nil
	}

	ratio := newTotal / existing.Total

	for _, item := range items {
		updateNumericFieldByRatio(item, "value", ratio)
		updateNumericFieldByRatio(item, "amount", ratio)
		updateNumericFieldByRatio(item, "price", ratio)
		updateNumericFieldByRatio(item, "quantity", ratio)
	}

	return items, nil
}

func updateNumericFieldByRatio(item map[string]interface{}, field string, ratio float64) {
	if v, ok := item[field]; ok {
		switch val := v.(type) {
		case float64:
			item[field] = val * ratio
		case int:
			item[field] = int(float64(val) * ratio)
		case int64:
			item[field] = int64(float64(val) * ratio)
		}
	}
}

func hashString(s string) string {
	h := fnv32a(s)
	return fmt.Sprintf("%x", h)
}

func fnv32a(s string) uint32 {
	const offset32 uint32 = 2166136261
	const prime32 uint32 = 16777619
	hash := offset32
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime32
	}
	return hash
}
