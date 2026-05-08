package tomlmerge

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type MergeStrategy int

const (
	StrategyReplace MergeStrategy = iota
	StrategyAppend
)

func MergeFiles(basePath, envPath string) (map[string]interface{}, error) {
	var base, env map[string]interface{}
	if _, err := toml.DecodeFile(basePath, &base); err != nil {
		return nil, fmt.Errorf("failed to decode base config: %w", err)
	}
	if _, err := toml.DecodeFile(envPath, &env); err != nil {
		return nil, fmt.Errorf("failed to decode env config: %w", err)
	}
	return Merge(base, env)
}

func MergeStrings(baseStr, envStr string) (map[string]interface{}, error) {
	var base, env map[string]interface{}
	if _, err := toml.Decode(baseStr, &base); err != nil {
		return nil, fmt.Errorf("failed to decode base config: %w", err)
	}
	if _, err := toml.Decode(envStr, &env); err != nil {
		return nil, fmt.Errorf("failed to decode env config: %w", err)
	}
	return Merge(base, env)
}

func Merge(base, env map[string]interface{}) (map[string]interface{}, error) {
	result := deepCopy(base).(map[string]interface{})
	if err := mergeMaps(result, env); err != nil {
		return nil, err
	}
	return result, nil
}

func mergeMaps(dst, src map[string]interface{}) error {
	for key, srcVal := range src {
		if strings.HasPrefix(key, "+") {
			realKey := strings.TrimPrefix(key, "+")
			appendStrategy := StrategyAppend
			if err := handleAppendKey(dst, realKey, srcVal, appendStrategy); err != nil {
				return err
			}
		} else if parts := parseIndexedKey(key); parts != nil {
			realKey := parts[0]
			index, _ := strconv.Atoi(parts[1])
			if err := handleIndexedKey(dst, realKey, index, srcVal); err != nil {
				return err
			}
		} else {
			if err := handleRegularKey(dst, key, srcVal); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseIndexedKey(key string) []string {
	if !strings.Contains(key, ".") {
		return nil
	}
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return nil
	}
	return parts
}

func handleRegularKey(dst map[string]interface{}, key string, srcVal interface{}) error {
	dstVal, exists := dst[key]
	if !exists {
		dst[key] = normalizeValue(srcVal)
		return nil
	}
	srcVal = normalizeValue(srcVal)
	dstVal = normalizeValue(dstVal)
	if !sameValueType(dstVal, srcVal) {
		dst[key] = srcVal
		return nil
	}
	switch typedSrc := srcVal.(type) {
	case map[string]interface{}:
		typedDst, ok := dstVal.(map[string]interface{})
		if !ok {
			dst[key] = typedSrc
			return nil
		}
		return mergeMaps(typedDst, typedSrc)
	case []interface{}:
		dst[key] = typedSrc
		return nil
	default:
		dst[key] = srcVal
		return nil
	}
}

func handleAppendKey(dst map[string]interface{}, key string, srcVal interface{}, strategy MergeStrategy) error {
	dstVal, exists := dst[key]
	if !exists {
		dst[key] = normalizeValue(srcVal)
		return nil
	}
	srcVal = normalizeValue(srcVal)
	dstVal = normalizeValue(dstVal)
	switch typedSrc := srcVal.(type) {
	case map[string]interface{}:
		return handleAppendMap(dst, key, typedSrc, dstVal)
	case []interface{}:
		return handleAppendSlice(dst, key, typedSrc, dstVal)
	default:
		dst[key] = srcVal
		return nil
	}
}

func handleAppendMap(dst map[string]interface{}, key string, srcMap map[string]interface{}, dstVal interface{}) error {
	dstMap, ok := dstVal.(map[string]interface{})
	if !ok {
		dst[key] = srcMap
		return nil
	}
	for k, v := range srcMap {
		dstMap[k] = normalizeValue(v)
	}
	return nil
}

func handleAppendSlice(dst map[string]interface{}, key string, srcSlice []interface{}, dstVal interface{}) error {
	dstSlice, ok := dstVal.([]interface{})
	if !ok {
		dst[key] = srcSlice
		return nil
	}
	if len(srcSlice) == 0 {
		return nil
	}
	firstSrc := normalizeValue(srcSlice[0])
	if _, ok := firstSrc.(map[string]interface{}); ok {
		for _, srcItem := range srcSlice {
			typedSrcItem, ok := normalizeValue(srcItem).(map[string]interface{})
			if !ok {
				dstSlice = append(dstSlice, srcItem)
				continue
			}
			found := false
			for i, dstItem := range dstSlice {
				typedDstItem, ok := normalizeValue(dstItem).(map[string]interface{})
				if !ok {
					continue
				}
				if canMergeTableItems(typedDstItem, typedSrcItem) {
					if err := mergeMaps(typedDstItem, typedSrcItem); err != nil {
						return err
					}
					dstSlice[i] = typedDstItem
					found = true
					break
				}
			}
			if !found {
				dstSlice = append(dstSlice, typedSrcItem)
			}
		}
		dst[key] = dstSlice
		return nil
	}
	dst[key] = append(dstSlice, srcSlice...)
	return nil
}

func canMergeTableItems(dst, src map[string]interface{}) bool {
	commonKeys := 0
	for k := range src {
		if _, exists := dst[k]; exists {
			commonKeys++
		}
	}
	return commonKeys > 0
}

func handleIndexedKey(dst map[string]interface{}, key string, index int, srcVal interface{}) error {
	sliceVal, exists := dst[key]
	if !exists {
		return fmt.Errorf("array table %s does not exist for index modification", key)
	}
	sliceVal = normalizeValue(sliceVal)
	slice, ok := sliceVal.([]interface{})
	if !ok {
		return fmt.Errorf("%s is not an array table", key)
	}
	if index < 0 || index >= len(slice) {
		return fmt.Errorf("index %d out of bounds for array table %s (length %d)", index, key, len(slice))
	}
	srcVal = normalizeValue(srcVal)
	srcMap, ok := srcVal.(map[string]interface{})
	if !ok {
		return fmt.Errorf("indexed modification requires a table value")
	}
	dstItem := normalizeValue(slice[index])
	if dstMap, ok := dstItem.(map[string]interface{}); ok {
		if err := mergeMaps(dstMap, srcMap); err != nil {
			return err
		}
		slice[index] = dstMap
	} else {
		slice[index] = srcMap
	}
	dst[key] = slice
	return nil
}

func normalizeValue(v interface{}) interface{} {
	switch typed := v.(type) {
	case time.Time:
		return typed.UTC()
	case bool:
		return typed
	default:
		return v
	}
}

func sameValueType(a, b interface{}) bool {
	if a == nil || b == nil {
		return false
	}
	_, aMap := a.(map[string]interface{})
	_, bMap := b.(map[string]interface{})
	if aMap && bMap {
		return true
	}
	if aMap || bMap {
		return false
	}
	_, aSlice := a.([]interface{})
	_, bSlice := b.([]interface{})
	if aSlice && bSlice {
		return true
	}
	if aSlice || bSlice {
		return false
	}
	_, aTime := a.(time.Time)
	_, bTime := b.(time.Time)
	if aTime && bTime {
		return true
	}
	if aTime || bTime {
		return false
	}
	_, aBool := a.(bool)
	_, bBool := b.(bool)
	if aBool && bBool {
		return true
	}
	if aBool || bBool {
		return false
	}
	_, aStr := a.(string)
	_, bStr := b.(string)
	if aStr && bStr {
		return true
	}
	if aStr || bStr {
		return false
	}
	_, aInt := a.(int64)
	_, aFloat := a.(float64)
	_, bInt := b.(int64)
	_, bFloat := b.(float64)
	if (aInt || aFloat) && (bInt || bFloat) {
		return true
	}
	return false
}

func deepCopy(v interface{}) interface{} {
	switch typed := v.(type) {
	case map[string]interface{}:
		cp := make(map[string]interface{}, len(typed))
		for k, val := range typed {
			cp[k] = deepCopy(val)
		}
		return cp
	case []interface{}:
		cp := make([]interface{}, len(typed))
		for i, val := range typed {
			cp[i] = deepCopy(val)
		}
		return cp
	default:
		return v
	}
}

func Encode(w io.Writer, data map[string]interface{}) error {
	enc := toml.NewEncoder(w)
	enc.Indent = ""
	return enc.Encode(data)
}

func DecodeString(s string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if _, err := toml.Decode(s, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func DecodeFile(path string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if _, err := toml.DecodeFile(path, &result); err != nil {
		return nil, err
	}
	return result, nil
}
