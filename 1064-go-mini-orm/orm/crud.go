package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func (db *DB) collectAllColumns(fields []*FieldInfo) []string {
	var columns []string
	for _, field := range fields {
		if field.IsEmbedded {
			columns = append(columns, db.collectAllColumns(field.EmbeddedFields)...)
		} else {
			columns = append(columns, field.ColumnName)
		}
	}
	return columns
}

func (db *DB) collectAllValues(model reflect.Value, fields []*FieldInfo) ([]interface{}, error) {
	var values []interface{}
	for _, field := range fields {
		fieldValue := model.FieldByName(field.Name)

		if field.IsEmbedded {
			if field.IsPointer {
				if fieldValue.IsNil() {
					for range field.EmbeddedFields {
						values = append(values, nil)
					}
					continue
				}
				fieldValue = fieldValue.Elem()
			}
			embeddedValues, err := db.collectAllValues(fieldValue, field.EmbeddedFields)
			if err != nil {
				return nil, err
			}
			values = append(values, embeddedValues...)
			continue
		}

		dbValue, err := db.toDBValue(field, fieldValue)
		if err != nil {
			return nil, err
		}
		values = append(values, dbValue)
	}
	return values, nil
}

func isDefaultZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func (db *DB) collectInsertColumnsAndValues(
	modelVal reflect.Value,
	fields []*FieldInfo,
	pkColumn string,
) ([]string, []interface{}, error) {
	var columns []string
	var values []interface{}

	for _, field := range fields {
		fieldValue := modelVal.FieldByName(field.Name)

		if field.IsEmbedded {
			var embedVal reflect.Value
			if field.IsPointer {
				if fieldValue.IsNil() {
					for _, ef := range field.EmbeddedFields {
						columns = append(columns, ef.ColumnName)
						values = append(values, nil)
					}
					continue
				}
				embedVal = fieldValue.Elem()
			} else {
				embedVal = fieldValue
			}
			embedColumns, embedValues, err := db.collectInsertColumnsAndValues(
				embedVal, field.EmbeddedFields, pkColumn,
			)
			if err != nil {
				return nil, nil, err
			}
			columns = append(columns, embedColumns...)
			values = append(values, embedValues...)
			continue
		}

		if field.ColumnName == pkColumn && isDefaultZeroValue(fieldValue) {
			continue
		}

		dbValue, err := db.toDBValue(field, fieldValue)
		if err != nil {
			return nil, nil, err
		}
		columns = append(columns, field.ColumnName)
		values = append(values, dbValue)
	}
	return columns, values, nil
}

func (db *DB) Insert(model interface{}) error {
	modelInfo, err := db.ParseModel(model)
	if err != nil {
		return err
	}

	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
	}

	columns, values, err := db.collectInsertColumnsAndValues(
		modelVal, modelInfo.Fields, modelInfo.PrimaryKey,
	)
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return fmt.Errorf("no columns to insert")
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		modelInfo.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = db.sqlDB.Exec(query, values...)
	return err
}

func (db *DB) GetByID(model interface{}, id interface{}) error {
	modelInfo, err := db.ParseModel(model)
	if err != nil {
		return err
	}

	columns := db.collectAllColumns(modelInfo.Fields)
	if len(columns) == 0 {
		return fmt.Errorf("no columns to select")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ?",
		strings.Join(columns, ", "),
		modelInfo.TableName,
		modelInfo.PrimaryKey,
	)

	row := db.sqlDB.QueryRow(query, id)

	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() != reflect.Ptr || modelVal.IsNil() {
		return fmt.Errorf("model must be a non-nil pointer")
	}
	modelVal = modelVal.Elem()

	scanDest, targets, err := db.prepareScanDest(modelInfo.Fields)
	if err != nil {
		return err
	}

	if err := row.Scan(scanDest...); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("record not found")
		}
		return err
	}

	return db.applyScanResults(modelVal, targets)
}

func (db *DB) prepareScanDestCount(fields []*FieldInfo) int {
	count := 0
	for _, field := range fields {
		if field.IsEmbedded {
			count += db.prepareScanDestCount(field.EmbeddedFields)
		} else {
			count++
		}
	}
	return count
}

func (db *DB) prepareScanDest(fields []*FieldInfo) ([]interface{}, []*scanTarget, error) {
	var dest []interface{}
	var targets []*scanTarget

	for _, field := range fields {
		if field.IsEmbedded {
			embedDest, embedTargets, err := db.prepareScanDest(field.EmbeddedFields)
			if err != nil {
				return nil, nil, err
			}
			dest = append(dest, embedDest...)
			for _, t := range embedTargets {
				t.embeddedPath = append([]string{field.Name}, t.embeddedPath...)
				t.embeddedPointers = append([]bool{field.IsPointer}, t.embeddedPointers...)
			}
			targets = append(targets, embedTargets...)
			continue
		}

		var holder interface{}
		if field.GoType == reflect.TypeOf(time.Time{}) {
			holder = new(sql.NullString)
		} else if field.IsPointer {
			holder = new(interface{})
		} else {
			holder = new(interface{})
		}

		dest = append(dest, holder)
		targets = append(targets, &scanTarget{
			field:      field,
			holder:     holder,
		})
	}
	return dest, targets, nil
}

type scanTarget struct {
	field            *FieldInfo
	holder           interface{}
	embeddedPath     []string
	embeddedPointers []bool
}

func (db *DB) applyScanResults(modelVal reflect.Value, targets []*scanTarget) error {
	for _, t := range targets {
		dest := modelVal
		validPath := true

		for i, fieldName := range t.embeddedPath {
			embedField := dest.FieldByName(fieldName)
			if !embedField.IsValid() {
				validPath = false
				break
			}
			if t.embeddedPointers[i] {
				if embedField.IsNil() {
					embedField.Set(reflect.New(embedField.Type().Elem()))
				}
				embedField = embedField.Elem()
			}
			dest = embedField
		}

		if !validPath {
			continue
		}

		destField := dest.FieldByName(t.field.Name)
		if !destField.IsValid() {
			continue
		}

		var dbValue interface{}
		switch h := t.holder.(type) {
		case *sql.NullString:
			if h.Valid {
				dbValue = h.String
			} else {
				dbValue = nil
			}
		case *interface{}:
			dbValue = *h
		default:
			dbValue = reflect.ValueOf(t.holder).Elem().Interface()
		}

		if err := db.fromDBValue(t.field, dbValue, destField); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Update(model interface{}, fieldsToUpdate map[string]interface{}) error {
	modelInfo, err := db.ParseModel(model)
	if err != nil {
		return err
	}

	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
	}

	if len(fieldsToUpdate) == 0 {
		return fmt.Errorf("no fields to update")
	}

	var setClauses []string
	var values []interface{}

	for colName, value := range fieldsToUpdate {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", colName))
		values = append(values, value)
	}

	idField := modelVal.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("model must have ID field")
	}
	values = append(values, idField.Interface())

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = ?",
		modelInfo.TableName,
		strings.Join(setClauses, ", "),
		modelInfo.PrimaryKey,
	)

	_, err = db.sqlDB.Exec(query, values...)
	return err
}

func (db *DB) Delete(model interface{}, id interface{}) error {
	modelInfo, err := db.ParseModel(model)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?",
		modelInfo.TableName,
		modelInfo.PrimaryKey,
	)

	_, err = db.sqlDB.Exec(query, id)
	return err
}

func (db *DB) List(models interface{}, options *ListOptions) error {
	modelsVal := reflect.ValueOf(models)
	if modelsVal.Kind() != reflect.Ptr || modelsVal.IsNil() {
		return fmt.Errorf("models must be a non-nil pointer to slice")
	}
	modelsVal = modelsVal.Elem()

	if modelsVal.Kind() != reflect.Slice {
		return fmt.Errorf("models must be a pointer to slice")
	}

	elemType := modelsVal.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}

	example := reflect.New(elemType).Interface()
	modelInfo, err := db.ParseModel(example)
	if err != nil {
		return err
	}

	columns := db.collectAllColumns(modelInfo.Fields)
	if len(columns) == 0 {
		return fmt.Errorf("no columns to select")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		strings.Join(columns, ", "),
		modelInfo.TableName,
	)

	var whereValues []interface{}
	if options != nil && len(options.Where) > 0 {
		var whereClauses []string
		for col, val := range options.Where {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", col))
			whereValues = append(whereValues, val)
		}
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if options != nil && options.OrderBy != "" {
		query += " ORDER BY " + options.OrderBy
	}

	if options != nil && options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", options.Limit)
	}

	if options != nil && options.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", options.Offset)
	}

	rows, err := db.sqlDB.Query(query, whereValues...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		newElem := reflect.New(elemType)
		elemVal := newElem.Elem()

		scanDest, targets, err := db.prepareScanDest(modelInfo.Fields)
		if err != nil {
			return err
		}

		if err := rows.Scan(scanDest...); err != nil {
			return err
		}

		if err := db.applyScanResults(elemVal, targets); err != nil {
			return err
		}

		if isPtr {
			modelsVal.Set(reflect.Append(modelsVal, newElem))
		} else {
			modelsVal.Set(reflect.Append(modelsVal, elemVal))
		}
	}

	return rows.Err()
}
