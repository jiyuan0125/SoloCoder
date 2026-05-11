package merge

import (
	"encoding/json"
)

func Apply(original, patch json.RawMessage) (json.RawMessage, error) {
	var patchVal interface{}
	if err := json.Unmarshal(patch, &patchVal); err != nil {
		return nil, err
	}

	patchObj, isPatchObj := patchVal.(map[string]interface{})
	if !isPatchObj {
		return patch, nil
	}

	var originalVal interface{}
	if err := json.Unmarshal(original, &originalVal); err != nil {
		return nil, err
	}

	originalObj, isOriginalObj := originalVal.(map[string]interface{})
	if !isOriginalObj {
		return patch, nil
	}

	result := deepCopyMap(originalObj)

	for key, patchFieldVal := range patchObj {
		if patchFieldVal == nil {
			delete(result, key)
			continue
		}

		patchFieldObj, patchFieldIsObj := patchFieldVal.(map[string]interface{})
		if !patchFieldIsObj {
			result[key] = patchFieldVal
			continue
		}

		originalFieldVal, exists := result[key]
		if !exists {
			result[key] = deepCopyValue(patchFieldObj)
			continue
		}

		originalFieldObj, originalFieldIsObj := originalFieldVal.(map[string]interface{})
		if !originalFieldIsObj {
			result[key] = deepCopyValue(patchFieldObj)
			continue
		}

		originalFieldJSON, _ := json.Marshal(originalFieldObj)
		patchFieldJSON, _ := json.Marshal(patchFieldObj)
		merged, err := Apply(originalFieldJSON, patchFieldJSON)
		if err != nil {
			return nil, err
		}
		var mergedObj map[string]interface{}
		json.Unmarshal(merged, &mergedObj)
		result[key] = mergedObj
	}

	return json.Marshal(result)
}

func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

func deepCopyValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return deepCopyMap(val)
	case []interface{}:
		copySlice := make([]interface{}, len(val))
		for i, item := range val {
			copySlice[i] = deepCopyValue(item)
		}
		return copySlice
	default:
		return val
	}
}
