package orm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

func toSnakeCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; i++ {
		current := runes[i]
		currentIsUpper := unicode.IsUpper(current)

		if i > 0 && currentIsUpper {
			prev := runes[i-1]
			prevIsUpper := unicode.IsUpper(prev)

			needUnderscore := false
			if !prevIsUpper {
				needUnderscore = true
			} else if i+1 < n && unicode.IsLower(runes[i+1]) {
				needUnderscore = true
			}

			if needUnderscore {
				result.WriteByte('_')
			}
		}

		result.WriteRune(unicode.ToLower(current))
	}

	return result.String()
}

func toPlural(s string) string {
	if strings.HasSuffix(s, "y") {
		return strings.TrimSuffix(s, "y") + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

func isSupportedType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int64, reflect.Float64, reflect.String, reflect.Bool:
		return true
	}
	if t == reflect.TypeOf(time.Time{}) {
		return true
	}
	return false
}

func (db *DB) ParseModel(model interface{}) (*ModelInfo, error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct, got %s", modelType.Kind())
	}

	modelInfo := &ModelInfo{
		ModelType: model,
		ColumnMap: make(map[string]*FieldInfo),
		Fields:    []*FieldInfo{},
	}

	tableName := toSnakeCase(modelType.Name())
	tableName = toPlural(tableName)

	if modelType.NumField() > 0 {
		firstField := modelType.Field(0)
		tagValue := firstField.Tag.Get("db")
		if strings.HasPrefix(tagValue, "table=") {
			tableName = strings.TrimPrefix(tagValue, "table=")
		}
	}

	modelInfo.TableName = tableName

	fields, err := db.parseFields(modelType, "")
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		modelInfo.Fields = append(modelInfo.Fields, field)
		modelInfo.ColumnMap[field.ColumnName] = field

		if strings.ToLower(field.Name) == "id" {
			modelInfo.PrimaryKey = field.ColumnName
		}
	}

	if modelInfo.PrimaryKey == "" {
		modelInfo.PrimaryKey = "id"
	}

	return modelInfo, nil
}

func (db *DB) parseFields(t reflect.Type, prefix string) ([]*FieldInfo, error) {
	var fields []*FieldInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		tagValue := field.Tag.Get("db")
		if tagValue == "-" {
			continue
		}

		var converterName string
		var columnName string

		if strings.HasPrefix(tagValue, "table=") {
			continue
		}

		tagParts := strings.Split(tagValue, ",")
		for _, part := range tagParts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "converter=") {
				converterName = strings.TrimPrefix(part, "converter=")
			} else if part != "" && !strings.HasPrefix(part, "table=") {
				columnName = part
			}
		}

		fieldType := field.Type
		isPointer := fieldType.Kind() == reflect.Ptr
		if isPointer {
			fieldType = fieldType.Elem()
		}

		isSlice := fieldType.Kind() == reflect.Slice

		var converter TypeConverter
		if converterName != "" {
			converter = db.converters[converterName]
		}

		if columnName == "" {
			columnName = toSnakeCase(field.Name)
		}
		if prefix != "" {
			columnName = prefix + "_" + columnName
		}

		if fieldType.Kind() == reflect.Struct && fieldType != reflect.TypeOf(time.Time{}) {
			embeddedFields, err := db.parseFields(fieldType, toSnakeCase(field.Name))
			if err != nil {
				return nil, err
			}

			fieldInfo := &FieldInfo{
				Name:           field.Name,
				ColumnName:     columnName,
				IsPointer:      isPointer,
				IsEmbedded:     true,
				EmbeddedFields: embeddedFields,
				ConverterName:  converterName,
				Converter:      converter,
			}
			fields = append(fields, fieldInfo)

			for _, ef := range embeddedFields {
				ef.GoType = fieldType
			}
			continue
		}

		if !isSupportedType(fieldType) && !isSlice && converter == nil {
			return nil, fmt.Errorf("unsupported field type: %s for field %s", fieldType.Name(), field.Name)
		}

		fieldInfo := &FieldInfo{
			Name:          field.Name,
			ColumnName:    columnName,
			GoType:        fieldType,
			IsPointer:     isPointer,
			IsSlice:       isSlice,
			IsEmbedded:    false,
			ConverterName: converterName,
			Converter:     converter,
		}

		if converter == nil {
			if globalConv, ok := db.typeConverters[fieldType]; ok {
				fieldInfo.Converter = globalConv
			}
		}

		fields = append(fields, fieldInfo)
	}

	return fields, nil
}

func (db *DB) toDBValue(field *FieldInfo, value reflect.Value) (interface{}, error) {
	if field.IsPointer {
		if value.IsNil() {
			return nil, nil
		}
		value = value.Elem()
	}

	if field.IsSlice {
		jsonBytes, err := json.Marshal(value.Interface())
		if err != nil {
			return nil, err
		}
		return string(jsonBytes), nil
	}

	if field.Converter != nil {
		return field.Converter.ToDB(value.Interface())
	}

	return value.Interface(), nil
}

func (db *DB) fromDBValue(field *FieldInfo, dbValue interface{}, dest reflect.Value) error {
	if field.IsPointer {
		if dbValue == nil {
			dest.Set(reflect.Zero(dest.Type()))
			return nil
		}

		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		dest = dest.Elem()
	}

	if field.IsSlice {
		var jsonStr string
		switch v := dbValue.(type) {
		case []byte:
			jsonStr = string(v)
		case string:
			jsonStr = v
		case nil:
			return nil
		default:
			return fmt.Errorf("unexpected type for slice field: %T", dbValue)
		}

		sliceValue := reflect.New(field.GoType).Elem()
		if err := json.Unmarshal([]byte(jsonStr), sliceValue.Addr().Interface()); err != nil {
			return err
		}
		dest.Set(sliceValue)
		return nil
	}

	if field.Converter != nil {
		converted, err := field.Converter.FromDB(dbValue)
		if err != nil {
			return err
		}
		dest.Set(reflect.ValueOf(converted))
		return nil
	}

	if dbValue == nil {
		dest.Set(reflect.Zero(dest.Type()))
		return nil
	}

	if dest.Type() == reflect.TypeOf(time.Time{}) {
		var timeStr string
		switch v := dbValue.(type) {
		case time.Time:
			dest.Set(reflect.ValueOf(v))
			return nil
		case []byte:
			timeStr = string(v)
		case string:
			timeStr = v
		default:
			return fmt.Errorf("cannot convert %T to time.Time", dbValue)
		}
		for _, layout := range []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05-07:00",
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02",
		} {
			if t, err := time.Parse(layout, timeStr); err == nil {
				dest.Set(reflect.ValueOf(t))
				return nil
			}
		}
		return fmt.Errorf("cannot parse time from string: %s", timeStr)
	}

	if dest.Kind() == reflect.Bool {
		switch v := dbValue.(type) {
		case bool:
			dest.SetBool(v)
			return nil
		case int64:
			dest.SetBool(v != 0)
			return nil
		case int:
			dest.SetBool(v != 0)
			return nil
		}
	}

	dbReflect := reflect.ValueOf(dbValue)
	if dbReflect.Type().ConvertibleTo(dest.Type()) {
		dest.Set(dbReflect.Convert(dest.Type()))
		return nil
	}

	return fmt.Errorf("cannot convert %T to %s", dbValue, dest.Type().Name())
}
