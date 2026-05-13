package comparator

import (
	"config-diff/internal/diff"
	"reflect"
	"strconv"
)

func Compare(oldData, newData map[string]interface{}) []diff.Change {
	var changes []diff.Change
	changes = append(changes, compareMap("", oldData, newData)...)
	return changes
}

func compareMap(path string, oldMap, newMap map[string]interface{}) []diff.Change {
	var changes []diff.Change

	for key, oldVal := range oldMap {
		newVal, exists := newMap[key]
		currentPath := joinPath(path, key)

		if !exists {
			changes = append(changes, diff.Change{
				Path:     currentPath,
				Action:   diff.ActionRemove,
				OldValue: oldVal,
			})
		} else {
			changes = append(changes, compareValue(currentPath, oldVal, newVal)...)
		}
	}

	for key, newVal := range newMap {
		if _, exists := oldMap[key]; !exists {
			currentPath := joinPath(path, key)
			changes = append(changes, diff.Change{
				Path:     currentPath,
				Action:   diff.ActionAdd,
				NewValue: newVal,
			})
		}
	}

	return changes
}

func compareValue(path string, oldVal, newVal interface{}) []diff.Change {
	var changes []diff.Change

	if reflect.DeepEqual(oldVal, newVal) {
		return changes
	}

	oldIsMap := isMap(oldVal)
	newIsMap := isMap(newVal)

	if oldIsMap && newIsMap {
		return compareMap(path, toMap(oldVal), toMap(newVal))
	}

	oldIsSlice := isSlice(oldVal)
	newIsSlice := isSlice(newVal)

	if oldIsSlice && newIsSlice {
		return compareSlice(path, toSlice(oldVal), toSlice(newVal))
	}

	changes = append(changes, diff.Change{
		Path:     path,
		Action:   diff.ActionModify,
		OldValue: oldVal,
		NewValue: newVal,
	})

	return changes
}

func compareSlice(path string, oldSlice, newSlice []interface{}) []diff.Change {
	var changes []diff.Change

	maxLen := max(len(oldSlice), len(newSlice))
	for i := 0; i < maxLen; i++ {
		idx := i
		indexPath := joinIndexPath(path, idx)

		if i >= len(oldSlice) {
			changes = append(changes, diff.Change{
				Path:     indexPath,
				Action:   diff.ActionAdd,
				NewValue: newSlice[i],
				Index:    &idx,
			})
			continue
		}

		if i >= len(newSlice) {
			changes = append(changes, diff.Change{
				Path:     indexPath,
				Action:   diff.ActionRemove,
				OldValue: oldSlice[i],
				Index:    &idx,
			})
			continue
		}

		oldVal := oldSlice[i]
		newVal := newSlice[i]

		if !reflect.DeepEqual(oldVal, newVal) {
			oldIsMap := isMap(oldVal)
			newIsMap := isMap(newVal)
			oldIsSlice := isSlice(oldVal)
			newIsSlice := isSlice(newVal)

			if oldIsMap && newIsMap {
				changes = append(changes, compareMap(indexPath, toMap(oldVal), toMap(newVal))...)
			} else if oldIsSlice && newIsSlice {
				changes = append(changes, compareSlice(indexPath, toSlice(oldVal), toSlice(newVal))...)
			} else {
				changes = append(changes, diff.Change{
					Path:     indexPath,
					Action:   diff.ActionModify,
					OldValue: oldVal,
					NewValue: newVal,
					Index:    &idx,
				})
			}
		}
	}

	return changes
}

func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	if key == "" {
		return base
	}
	return base + "." + key
}

func joinIndexPath(base string, index int) string {
	if base == "" {
		return "[" + strconv.Itoa(index) + "]"
	}
	return base + "[" + strconv.Itoa(index) + "]"
}

func isMap(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}

func toMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func isSlice(v interface{}) bool {
	_, ok := v.([]interface{})
	return ok
}

func toSlice(v interface{}) []interface{} {
	if s, ok := v.([]interface{}); ok {
		return s
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
