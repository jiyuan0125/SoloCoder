package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
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

func (db *DB) Insert(model interface{}) error {
	modelInfo, err := db.ParseModel(model)
	if err != nil {
		return err
	}

	modelVal := reflect.ValueOf(model)
	if modelVal.Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
	}

	columns := db.collectAllColumns(modelInfo.Fields)
	if len(columns) == 0 {
		return fmt.Errorf("no columns to insert")
	}

	values, err := db.collectAllValues(modelVal, modelInfo.Fields)
	if err != nil {
		return err
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

	scanDest, err := db.prepareScanDest(modelVal, modelInfo.Fields)
	if err != nil {
		return err
	}

	if err := row.Scan(scanDest...); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("record not found")
		}
		return err
	}

	return db.afterScan(modelVal, modelInfo.Fields)
}

func (db *DB) prepareScanDest(modelVal reflect.Value, fields []*FieldInfo) ([]interface{}, error) {
	var dest []interface{}
	for _, field := range fields {
		fieldValue := modelVal.FieldByName(field.Name)

		if field.IsEmbedded {
			if field.IsPointer {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}
				fieldValue = fieldValue.Elem()
			}
			embeddedDest, err := db.prepareScanDest(fieldValue, field.EmbeddedFields)
			if err != nil {
				return nil, err
			}
			dest = append(dest, embeddedDest...)
			continue
		}

		dest = append(dest, reflect.New(fieldValue.Type()).Interface())
	}
	return dest, nil
}

func (db *DB) afterScan(modelVal reflect.Value, fields []*FieldInfo) error {
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

		scanDest, err := db.prepareScanDest(elemVal, modelInfo.Fields)
		if err != nil {
			return err
		}

		if err := rows.Scan(scanDest...); err != nil {
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
