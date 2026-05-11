package adapter

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

type CSVAdapter struct {
	converter *TypeConverter
}

func NewCSVAdapter() *CSVAdapter {
	return &CSVAdapter{
		converter: NewTypeConverter(),
	}
}

func (a *CSVAdapter) Name() string {
	return "csv"
}

func (a *CSVAdapter) CanHandle(format string) bool {
	return strings.EqualFold(format, "csv")
}

func (a *CSVAdapter) Decode(data []byte, v interface{}) error {
	cleanData := a.cleanCSVData(data)
	reader := csv.NewReader(strings.NewReader(string(cleanData)))

	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}
	val = val.Elem()

	switch val.Kind() {
	case reflect.Slice:
		elemType := val.Type().Elem()
		isPtr := elemType.Kind() == reflect.Ptr
		if isPtr {
			elemType = elemType.Elem()
		}

		for _, record := range records {
			if len(record) == 0 {
				continue
			}

			elem := reflect.New(elemType).Elem()
			if err := a.fillStructFromCSV(headers, record, elem); err != nil {
				return err
			}

			if isPtr {
				val.Set(reflect.Append(val, elem.Addr()))
			} else {
				val.Set(reflect.Append(val, elem))
			}
		}
	case reflect.Struct:
		if len(records) == 0 {
			return nil
		}
		return a.fillStructFromCSV(headers, records[0], val)
	default:
		if len(records) == 0 {
			return nil
		}
		if len(records[0]) > 0 {
			converted, err := a.converter.Convert(records[0][0], val.Type())
			if err != nil {
				return err
			}
			val.Set(converted)
		}
	}

	return nil
}

func (a *CSVAdapter) cleanCSVData(data []byte) []byte {
	str := string(data)
	str = strings.TrimSpace(str)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
		str = strings.ReplaceAll(str, "\\n", "\n")
		str = strings.ReplaceAll(str, "\\t", "\t")
		str = strings.ReplaceAll(str, "\\r", "\r")
		str = strings.ReplaceAll(str, "\\\"", "\"")
	}
	return []byte(str)
}

func (a *CSVAdapter) Encode(v interface{}) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	switch val.Kind() {
	case reflect.Slice:
		if val.Len() == 0 {
			return nil, nil
		}

		firstElem := val.Index(0)
		if firstElem.Kind() == reflect.Ptr {
			firstElem = firstElem.Elem()
		}

		if firstElem.Kind() == reflect.Struct {
			headers := a.getStructHeaders(firstElem.Type())
			if err := writer.Write(headers); err != nil {
				return nil, err
			}

			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}

				record := a.structToCSVRecord(elem)
				if err := writer.Write(record); err != nil {
					return nil, err
				}
			}
		} else if firstElem.Kind() == reflect.Map {
			headers := a.getMapKeys(val)
			if len(headers) == 0 {
				return nil, nil
			}
			if err := writer.Write(headers); err != nil {
				return nil, err
			}

			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				record := a.mapToCSVRecord(elem, headers)
				if err := writer.Write(record); err != nil {
					return nil, err
				}
			}
		} else {
			return nil, fmt.Errorf("slice elements must be structs or maps")
		}
	case reflect.Struct:
		headers := a.getStructHeaders(val.Type())
		if err := writer.Write(headers); err != nil {
			return nil, err
		}
		record := a.structToCSVRecord(val)
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	case reflect.Map:
		headers := a.getMapKeysFromSingle(val)
		if len(headers) == 0 {
			return nil, nil
		}
		if err := writer.Write(headers); err != nil {
			return nil, err
		}
		record := a.mapToCSVRecord(val, headers)
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	default:
		record := []string{a.valueToString(val)}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return []byte(builder.String()), nil
}

func (a *CSVAdapter) Priority() int {
	return 100
}

func (a *CSVAdapter) fillStructFromCSV(headers, record []string, val reflect.Value) error {
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", val.Kind())
	}

	for i, header := range headers {
		if i >= len(record) {
			continue
		}
		value := record[i]
		if value == "" {
			continue
		}

		if err := a.setFieldByPath(val, header, value); err != nil {
			return fmt.Errorf("error setting field '%s': %w", header, err)
		}
	}

	return nil
}

func (a *CSVAdapter) setFieldByPath(val reflect.Value, path, value string) error {
	parts := strings.Split(path, ".")
	current := val

	for i, part := range parts {
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				current.Set(reflect.New(current.Type().Elem()))
			}
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return fmt.Errorf("path '%s' points to non-struct field", path)
		}

		fieldVal, fieldType, found := a.findFieldByCSVTag(current, part)
		if !found {
			return nil
		}

		if i == len(parts)-1 {
			if fieldType.Kind() == reflect.Slice {
				values := strings.Split(value, "|")
				sliceType := fieldType.Elem()
				slice := reflect.MakeSlice(fieldType, 0, len(values))

				for _, v := range values {
					v = strings.TrimSpace(v)
					if v == "" {
						continue
					}
					elem := reflect.New(sliceType).Elem()
					if err := a.setFieldValue(elem, v); err != nil {
						return err
					}
					slice = reflect.Append(slice, elem)
				}

				fieldVal.Set(slice)
			} else {
				if err := a.setFieldValue(fieldVal, value); err != nil {
					return err
				}
			}
		} else {
			current = fieldVal
		}
	}

	return nil
}

func (a *CSVAdapter) findFieldByCSVTag(val reflect.Value, tagValue string) (reflect.Value, reflect.Type, bool) {
	valType := val.Type()
	lowerTag := strings.ToLower(tagValue)

	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == tagValue || strings.ToLower(csvTag) == lowerTag {
			return fieldVal, field.Type, true
		}

		if strings.ToLower(field.Name) == lowerTag {
			return fieldVal, field.Type, true
		}
	}

	return reflect.Value{}, nil, false
}

func (a *CSVAdapter) setFieldValue(fieldVal reflect.Value, value string) error {
	if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		}
		if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		}
	}

	converted, err := a.converter.Convert(value, fieldVal.Type())
	if err != nil {
		return err
	}
	fieldVal.Set(converted)
	return nil
}

func (a *CSVAdapter) getStructHeaders(typ reflect.Type) []string {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var headers []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !field.IsExported() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if csvTag != "" {
			headers = append(headers, csvTag)
		} else {
			headers = append(headers, strings.ToLower(field.Name))
		}
	}

	return headers
}

func (a *CSVAdapter) structToCSVRecord(val reflect.Value) []string {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var record []string
	valType := val.Type()

	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		csvTag := field.Tag.Get("csv")
		if csvTag == "-" {
			continue
		}

		record = append(record, a.valueToString(fieldVal))
	}

	return record
}

func (a *CSVAdapter) valueToString(val reflect.Value) string {
	if !val.IsValid() {
		return ""
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	if val.Type() == reflect.TypeOf(time.Time{}) {
		t := val.Interface().(time.Time)
		return t.Format(time.RFC3339)
	}

	if val.Kind() == reflect.Slice {
		var values []string
		for i := 0; i < val.Len(); i++ {
			values = append(values, a.valueToString(val.Index(i)))
		}
		return strings.Join(values, "|")
	}

	if val.Kind() == reflect.Interface && !val.IsNil() {
		return a.valueToString(val.Elem())
	}

	return fmt.Sprintf("%v", val.Interface())
}

func (a *CSVAdapter) getMapKeys(slice reflect.Value) []string {
	keySet := make(map[string]bool)
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Map {
			for _, key := range elem.MapKeys() {
				keyStr := fmt.Sprintf("%v", key.Interface())
				keySet[keyStr] = true
			}
		}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	return keys
}

func (a *CSVAdapter) getMapKeysFromSingle(m reflect.Value) []string {
	if m.Kind() == reflect.Interface {
		m = m.Elem()
	}
	keys := make([]string, 0, m.Len())
	for _, key := range m.MapKeys() {
		keys = append(keys, fmt.Sprintf("%v", key.Interface()))
	}
	return keys
}

func (a *CSVAdapter) mapToCSVRecord(m reflect.Value, headers []string) []string {
	if m.Kind() == reflect.Interface {
		m = m.Elem()
	}
	record := make([]string, len(headers))
	for i, header := range headers {
		for _, key := range m.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			if keyStr == header {
				record[i] = a.valueToString(m.MapIndex(key))
				break
			}
		}
	}
	return record
}
