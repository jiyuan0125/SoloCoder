package query

import (
	"fmt"
	"strings"

	"report-query/internal/db"
	"report-query/internal/parser"
	"report-query/internal/security"
)

type QueryBuilder struct {
	DB         *db.DB
	Table      string
	Fields     []string
	Conditions string
	Where      string
	GroupBy    string
}

func (qb *QueryBuilder) BuildAndValidate() (string, *db.TableInfo, error) {
	if !qb.DB.HasTable(qb.Table) {
		return "", nil, fmt.Errorf("表不存在: %s", qb.Table)
	}

	tableInfo, err := qb.DB.GetTableInfo(qb.Table)
	if err != nil {
		return "", nil, err
	}

	if len(qb.Fields) == 0 || (len(qb.Fields) == 1 && qb.Fields[0] == "*") {
		qb.Fields = make([]string, 0, len(tableInfo.Columns))
		for _, col := range tableInfo.Columns {
			qb.Fields = append(qb.Fields, col.Name)
		}
	}

	invalidFields := qb.DB.ValidateColumns(tableInfo, qb.Fields)
	if len(invalidFields) > 0 {
		return "", tableInfo, fmt.Errorf("字段不存在: %s", strings.Join(invalidFields, ", "))
	}

	if qb.GroupBy != "" {
		groupFields := splitAndTrim(qb.GroupBy)
		invalidGroup := qb.DB.ValidateColumns(tableInfo, groupFields)
		if len(invalidGroup) > 0 {
			return "", tableInfo, fmt.Errorf("GROUP BY 字段不存在: %s", strings.Join(invalidGroup, ", "))
		}
	}

	var whereClause string
	if qb.Where != "" {
		if err := security.ValidateWhereClause(qb.Where); err != nil {
			return "", tableInfo, err
		}
		whereClause = qb.Where
	} else if qb.Conditions != "" {
		parsed, err := parser.ParseConditions(qb.Conditions)
		if err != nil {
			return "", tableInfo, err
		}
		whereClause = parsed
	}

	sql := qb.buildSQL(whereClause)

	if err := security.ValidateWhereClause(sql); err != nil {
		return "", tableInfo, err
	}

	return sql, tableInfo, nil
}

func (qb *QueryBuilder) buildSQL(whereClause string) string {
	fieldsStr := strings.Join(qb.Fields, ", ")
	sql := fmt.Sprintf("SELECT %s FROM `%s`", fieldsStr, qb.Table)

	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	if qb.GroupBy != "" {
		sql += " GROUP BY " + qb.GroupBy
	}

	return sql
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func ParseFields(fields string) []string {
	return splitAndTrim(fields)
}
