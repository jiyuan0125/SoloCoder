package datasource

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"reportgen/internal/model"
)

type DatabaseReader struct {
	dsType   model.DataSourceType
	host     string
	port     int
	database string
	username string
	password string
	table    string
	query    string
	columns  []model.Column
}

func NewDatabaseReader(ds model.DataSource) *DatabaseReader {
	return &DatabaseReader{
		dsType:   ds.Type,
		host:     ds.Host,
		port:     ds.Port,
		database: ds.Database,
		username: ds.Username,
		password: ds.Password,
		table:    ds.Table,
		query:    ds.Query,
		columns:  ds.Columns,
	}
}

func (r *DatabaseReader) Read() ([]model.DataRow, error) {
	db, err := r.connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := r.getQuery()
	if query == "" {
		return nil, fmt.Errorf("neither table name nor query specified")
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	if len(r.columns) == 0 {
		for _, c := range cols {
			r.columns = append(r.columns, model.Column{
				Name: c,
				Type: model.FieldTypeString,
			})
		}
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	typeMap := make(map[string]string)
	for i, ct := range columnTypes {
		if i < len(cols) {
			typeMap[cols[i]] = ct.DatabaseTypeName()
		}
	}

	var results []model.DataRow
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := model.DataRow{
			Values: make(map[string]interface{}),
		}

		for i, colName := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row.Values[colName] = r.convertValue(val, colName, typeMap[colName])
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (r *DatabaseReader) connect() (*sql.DB, error) {
	var dsn string

	switch r.dsType {
	case model.DataSourceMySQL:
		port := r.port
		if port == 0 {
			port = 3306
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			r.username, r.password, r.host, port, r.database)
	case model.DataSourcePostgres:
		port := r.port
		if port == 0 {
			port = 5432
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			r.host, port, r.username, r.password, r.database)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", r.dsType)
	}

	db, err := sql.Open(string(r.dsType), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func (r *DatabaseReader) getQuery() string {
	if r.query != "" {
		return r.query
	}
	if r.table != "" {
		return fmt.Sprintf("SELECT * FROM %s", r.table)
	}
	return ""
}

func (r *DatabaseReader) convertValue(val interface{}, colName string, dbType string) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case time.Time:
		return v
	case []byte:
		return string(v)
	case float64, float32, int, int64, int32, bool:
		return v
	case string:
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t
		}
		return v
	default:
		return val
	}
}
