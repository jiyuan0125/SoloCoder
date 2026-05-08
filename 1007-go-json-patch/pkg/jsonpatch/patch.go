package jsonpatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrTestFailed         = errors.New("test operation failed")
	ErrInvalidOperation   = errors.New("invalid patch operation")
	ErrMissingField       = errors.New("missing required field in patch operation")
	ErrMoveToDescendant   = errors.New("cannot move to a descendant path")
	ErrSamePath           = errors.New("source and target paths are the same")
	ErrPathMustExist      = errors.New("target path must exist for this operation")
	ErrIndexOutOfBounds   = errors.New("array index out of bounds")
)

type Operation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	From  string          `json:"from,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

type Patch []Operation

func Apply(doc interface{}, patch Patch) (interface{}, error) {
	docCopy, err := deepCopy(doc)
	if err != nil {
		return nil, err
	}
	
	for i, op := range patch {
		newDoc, err := applyOperation(docCopy, op)
		if err != nil {
			return nil, fmt.Errorf("operation %d failed: %w", i, err)
		}
		docCopy = newDoc
	}
	
	return docCopy, nil
}

func Validate(patch Patch) error {
	for i, op := range patch {
		if err := validateOperation(op); err != nil {
			return fmt.Errorf("operation %d invalid: %w", i, err)
		}
	}
	return nil
}

func validateOperation(op Operation) error {
	switch op.Op {
	case "add":
		if op.Path == "" {
			return ErrMissingField
		}
		if len(op.Value) == 0 {
			return ErrMissingField
		}
		if _, err := ParsePointer(op.Path); err != nil {
			return err
		}
	case "remove":
		if op.Path == "" {
			return ErrMissingField
		}
		if _, err := ParsePointer(op.Path); err != nil {
			return err
		}
	case "replace":
		if op.Path == "" {
			return ErrMissingField
		}
		if len(op.Value) == 0 {
			return ErrMissingField
		}
		if _, err := ParsePointer(op.Path); err != nil {
			return err
		}
	case "move":
		if op.Path == "" || op.From == "" {
			return ErrMissingField
		}
		fromPtr, err := ParsePointer(op.From)
		if err != nil {
			return err
		}
		toPtr, err := ParsePointer(op.Path)
		if err != nil {
			return err
		}
		if op.From == op.Path {
			return ErrSamePath
		}
		if fromPtr.IsParentOf(toPtr) {
			return ErrMoveToDescendant
		}
	case "copy":
		if op.Path == "" || op.From == "" {
			return ErrMissingField
		}
		if _, err := ParsePointer(op.From); err != nil {
			return err
		}
		if _, err := ParsePointer(op.Path); err != nil {
			return err
		}
	case "test":
		if op.Path == "" {
			return ErrMissingField
		}
		if len(op.Value) == 0 {
			return ErrMissingField
		}
		if _, err := ParsePointer(op.Path); err != nil {
			return err
		}
	default:
		return ErrInvalidOperation
	}
	return nil
}

func applyOperation(doc interface{}, op Operation) (interface{}, error) {
	switch op.Op {
	case "add":
		return applyAdd(doc, op)
	case "remove":
		return applyRemove(doc, op)
	case "replace":
		return applyReplace(doc, op)
	case "move":
		return applyMove(doc, op)
	case "copy":
		return applyCopy(doc, op)
	case "test":
		return applyTest(doc, op)
	default:
		return nil, ErrInvalidOperation
	}
}

func applyAdd(doc interface{}, op Operation) (interface{}, error) {
	ptr, err := ParsePointer(op.Path)
	if err != nil {
		return nil, err
	}
	
	var value interface{}
	if len(op.Value) > 0 {
		if err := json.Unmarshal(op.Value, &value); err != nil {
			return nil, err
		}
	}
	
	if len(ptr) == 0 {
		return value, nil
	}
	
	parentPtr := ptr.parent()
	parent, err := getValue(doc, parentPtr)
	if err != nil {
		return nil, err
	}
	
	key := ptr.lastKey()
	
	switch parentType := parent.(type) {
	case map[string]interface{}:
		newParent := make(map[string]interface{})
		for k, v := range parentType {
			newParent[k] = v
		}
		newParent[key] = value
		return setValue(doc, parentPtr, newParent)
	case []interface{}:
		idx, err := parseArrayIndex(key, len(parentType))
		if err != nil {
			return nil, err
		}
		if idx > len(parentType) {
			return nil, ErrIndexOutOfBounds
		}
		newArr := make([]interface{}, 0, len(parentType)+1)
		newArr = append(newArr, parentType[:idx]...)
		newArr = append(newArr, value)
		newArr = append(newArr, parentType[idx:]...)
		return setValue(doc, parentPtr, newArr)
	default:
		return nil, ErrPointerNotFound
	}
}

func applyRemove(doc interface{}, op Operation) (interface{}, error) {
	ptr, err := ParsePointer(op.Path)
	if err != nil {
		return nil, err
	}
	
	if len(ptr) == 0 {
		return nil, errors.New("cannot remove entire document")
	}
	
	if _, err := getValue(doc, ptr); err != nil {
		return nil, err
	}
	
	parentPtr := ptr.parent()
	parent, err := getValue(doc, parentPtr)
	if err != nil {
		return nil, err
	}
	
	key := ptr.lastKey()
	
	switch parentType := parent.(type) {
	case map[string]interface{}:
		if _, ok := parentType[key]; !ok {
			return nil, ErrPointerNotFound
		}
		newParent := make(map[string]interface{})
		for k, v := range parentType {
			if k != key {
				newParent[k] = v
			}
		}
		return setValue(doc, parentPtr, newParent)
	case []interface{}:
		idx, err := parseArrayIndex(key, len(parentType)-1)
		if err != nil {
			return nil, err
		}
		if idx >= len(parentType) {
			return nil, ErrIndexOutOfBounds
		}
		newArr := make([]interface{}, 0, len(parentType)-1)
		newArr = append(newArr, parentType[:idx]...)
		newArr = append(newArr, parentType[idx+1:]...)
		return setValue(doc, parentPtr, newArr)
	default:
		return nil, ErrPointerNotFound
	}
}

func applyReplace(doc interface{}, op Operation) (interface{}, error) {
	ptr, err := ParsePointer(op.Path)
	if err != nil {
		return nil, err
	}
	
	if _, err := getValue(doc, ptr); err != nil {
		return nil, ErrPathMustExist
	}
	
	var value interface{}
	if len(op.Value) > 0 {
		if err := json.Unmarshal(op.Value, &value); err != nil {
			return nil, err
		}
	}
	
	if len(ptr) == 0 {
		return value, nil
	}
	
	parentPtr := ptr.parent()
	parent, err := getValue(doc, parentPtr)
	if err != nil {
		return nil, err
	}
	
	key := ptr.lastKey()
	
	switch parentType := parent.(type) {
	case map[string]interface{}:
		newParent := make(map[string]interface{})
		for k, v := range parentType {
			newParent[k] = v
		}
		newParent[key] = value
		return setValue(doc, parentPtr, newParent)
	case []interface{}:
		idx, err := parseArrayIndex(key, len(parentType)-1)
		if err != nil {
			return nil, err
		}
		if idx >= len(parentType) {
			return nil, ErrIndexOutOfBounds
		}
		newArr := make([]interface{}, len(parentType))
		copy(newArr, parentType)
		newArr[idx] = value
		return setValue(doc, parentPtr, newArr)
	default:
		return nil, ErrPointerNotFound
	}
}

func applyMove(doc interface{}, op Operation) (interface{}, error) {
	fromPtr, err := ParsePointer(op.From)
	if err != nil {
		return nil, err
	}
	toPtr, err := ParsePointer(op.Path)
	if err != nil {
		return nil, err
	}
	
	if op.From == op.Path {
		return nil, ErrSamePath
	}
	if fromPtr.IsParentOf(toPtr) {
		return nil, ErrMoveToDescendant
	}
	
	value, err := getValue(doc, fromPtr)
	if err != nil {
		return nil, err
	}
	
	docWithoutFrom, err := applyRemove(doc, Operation{Op: "remove", Path: op.From})
	if err != nil {
		return nil, err
	}
	
	var valueJSON json.RawMessage
	if value != nil {
		valueJSON, err = json.Marshal(value)
		if err != nil {
			return nil, err
		}
	}
	
	return applyAdd(docWithoutFrom, Operation{Op: "add", Path: op.Path, Value: valueJSON})
}

func applyCopy(doc interface{}, op Operation) (interface{}, error) {
	fromPtr, err := ParsePointer(op.From)
	if err != nil {
		return nil, err
	}
	
	value, err := getValue(doc, fromPtr)
	if err != nil {
		return nil, err
	}
	
	valueCopy, err := deepCopy(value)
	if err != nil {
		return nil, err
	}
	
	var valueJSON json.RawMessage
	if valueCopy != nil {
		valueJSON, err = json.Marshal(valueCopy)
		if err != nil {
			return nil, err
		}
	}
	
	return applyAdd(doc, Operation{Op: "add", Path: op.Path, Value: valueJSON})
}

func applyTest(doc interface{}, op Operation) (interface{}, error) {
	ptr, err := ParsePointer(op.Path)
	if err != nil {
		return nil, err
	}
	
	actual, err := getValue(doc, ptr)
	if err != nil {
		return nil, ErrTestFailed
	}
	
	var expected interface{}
	if len(op.Value) > 0 {
		if err := json.Unmarshal(op.Value, &expected); err != nil {
			return nil, err
		}
	}
	
	if !deepEqual(actual, expected) {
		return nil, ErrTestFailed
	}
	
	return doc, nil
}

func setValue(doc interface{}, ptr Pointer, value interface{}) (interface{}, error) {
	if len(ptr) == 0 {
		return value, nil
	}
	
	root, err := deepCopy(doc)
	if err != nil {
		return nil, err
	}
	
	current := root
	for i, key := range ptr[:len(ptr)-1] {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		case []interface{}:
			idx, err := parseArrayIndex(key, len(v)-1)
			if err != nil {
				return nil, err
			}
			current = v[idx]
		default:
			return nil, ErrPointerNotFound
		}
		_ = i
	}
	
	key := ptr[len(ptr)-1]
	switch parent := current.(type) {
	case map[string]interface{}:
		parent[key] = value
	case []interface{}:
		idx, err := parseArrayIndex(key, len(parent)-1)
		if err != nil {
			return nil, err
		}
		parent[idx] = value
	default:
		return nil, ErrPointerNotFound
	}
	
	return root, nil
}

func deepCopy(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return reflect.DeepEqual(normalizeForCompare(a), normalizeForCompare(b))
}

func normalizeForCompare(v interface{}) interface{} {
	switch val := v.(type) {
	case float64:
		if float64(int64(val)) == val {
			return int64(val)
		}
		return val
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = normalizeForCompare(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = normalizeForCompare(v)
		}
		return result
	default:
		return v
	}
}
