package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn   *sql.DB
	dbPath string
}

type TableInfo struct {
	Name       string
	Columns    []ColumnInfo
	ColumnMap  map[string]bool
}

type ColumnInfo struct {
	Name string
	Type string
}

func NewDB(dbPath string) (*DB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, errors.New("数据库文件不存在: " + dbPath)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法连接数据库: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	return &DB{conn: conn, dbPath: dbPath}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) ListTables() ([]string, error) {
	rows, err := db.conn.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (db *DB) GetTableInfo(tableName string) (*TableInfo, error) {
	tables, err := db.ListTables()
	if err != nil {
		return nil, err
	}

	found := false
	for _, t := range tables {
		if strings.EqualFold(t, tableName) {
			tableName = t
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("表不存在: %s", tableName)
	}

	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", quoteIdentifier(tableName)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{
		Name:      tableName,
		ColumnMap: make(map[string]bool),
	}

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt_value interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return nil, err
		}
		info.Columns = append(info.Columns, ColumnInfo{Name: name, Type: ctype})
		info.ColumnMap[strings.ToLower(name)] = true
	}

	if len(info.Columns) == 0 {
		return nil, fmt.Errorf("无法获取表 %s 的列信息", tableName)
	}

	return info, rows.Err()
}

func (db *DB) HasTable(tableName string) bool {
	tables, err := db.ListTables()
	if err != nil {
		return false
	}
	for _, t := range tables {
		if strings.EqualFold(t, tableName) {
			return true
		}
	}
	return false
}

func (db *DB) ValidateColumns(tableInfo *TableInfo, columns []string) []string {
	var invalid []string
	for _, col := range columns {
		colLower := strings.ToLower(col)
		isAggregate := false
		aggFuncs := []string{"count(", "sum(", "avg(", "max(", "min(", "count(*"}
		for _, agg := range aggFuncs {
			if strings.HasPrefix(colLower, agg) {
				isAggregate = true
				break
			}
		}
		if !isAggregate && !tableInfo.ColumnMap[colLower] {
			invalid = append(invalid, col)
		}
	}
	return invalid
}

func (db *DB) ExecuteQuery(query string) (*sql.Rows, error) {
	return db.conn.Query(query)
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
