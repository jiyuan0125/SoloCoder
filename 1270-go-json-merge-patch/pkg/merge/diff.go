package merge

import (
	"encoding/json"
	"reflect"
)

func Diff(source, target json.RawMessage) (json.RawMessage, error) {
	var sourceVal, targetVal interface{}
	if err := json.Unmarshal(source, &sourceVal); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(target, &targetVal); err != nil {
		return nil, err
	}

	sourceObj, sourceIsObj := sourceVal.(map[string]interface{})
	targetObj, targetIsObj := targetVal.(map[string]interface{})

	if !sourceIsObj || !targetIsObj {
		if reflect.DeepEqual(sourceVal, targetVal) {
			return json.Marshal(map[string]interface{}{})
		}
		return target, nil
	}

	patch := make(map[string]interface{})

	for key, targetFieldVal := range targetObj {
		sourceFieldVal, exists := sourceObj[key]
		if !exists {
			patch[key] = deepCopyValue(targetFieldVal)
			continue
		}

		sourceFieldObj, sourceIsObjField := sourceFieldVal.(map[string]interface{})
		targetFieldObj, targetIsObjField := targetFieldVal.(map[string]interface{})

		if sourceIsObjField && targetIsObjField {
			sourceFieldJSON, _ := json.Marshal(sourceFieldObj)
			targetFieldJSON, _ := json.Marshal(targetFieldObj)
			fieldPatch, err := Diff(sourceFieldJSON, targetFieldJSON)
			if err != nil {
				return nil, err
			}
			var fieldPatchObj map[string]interface{}
			json.Unmarshal(fieldPatch, &fieldPatchObj)
			if len(fieldPatchObj) > 0 {
				patch[key] = fieldPatchObj
			}
			continue
		}

		if !reflect.DeepEqual(sourceFieldVal, targetFieldVal) {
			patch[key] = deepCopyValue(targetFieldVal)
		}
	}

	for key := range sourceObj {
		if _, exists := targetObj[key]; !exists {
			patch[key] = nil
		}
	}

	return json.Marshal(patch)
}
