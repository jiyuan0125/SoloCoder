package main

import (
	"database/sql"
	"fmt"
	"strings"
	
	_ "modernc.org/sqlite"
)

type Server struct {
	db *sql.DB
}

func NewServer(dbPath string) (*Server, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return &Server{db: db}, nil
}

func (s *Server) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Server) executeSQL(sql string) error {
	_, err := s.db.Exec(sql)
	return err
}

func (s *Server) executeInsert(sql string, values ...interface{}) error {
	_, err := s.db.Exec(sql, values...)
	return err
}

func (s *Server) generateCreateTableSQL(tableName string, headers []string, types []ColumnType) string {
	var columns []string
	for i, header := range headers {
		colName := sanitizeColumnName(header)
		if colName == "" {
			colName = fmt.Sprintf("col_%d", i+1)
		}
		
		var colType string
		switch types[i] {
		case TypeDate:
			colType = "DATE"
		case TypeNumber:
			colType = "TEXT"
		default:
			colType = "TEXT"
		}
		
		columns = append(columns, fmt.Sprintf("%s %s", colName, colType))
	}
	
	columns = append([]string{"id INTEGER PRIMARY KEY AUTOINCREMENT"}, columns...)
	
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columns, ", "))
}

func (s *Server) generateInsertSQL(tableName string, headers []string) string {
	var columns []string
	var placeholders []string
	
	for i, header := range headers {
		colName := sanitizeColumnName(header)
		if colName == "" {
			colName = fmt.Sprintf("col_%d", i+1)
		}
		columns = append(columns, colName)
		placeholders = append(placeholders, "?")
	}
	
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", 
		tableName, 
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))
}

func sanitizeColumnName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	
	resultStr := result.String()
	resultStr = strings.Trim(resultStr, "_")
	
	if len(resultStr) > 0 && (resultStr[0] >= '0' && resultStr[0] <= '9') {
		resultStr = "col_" + resultStr
	}
	
	return resultStr
}
