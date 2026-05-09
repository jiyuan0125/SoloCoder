package sqlite

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
)

func OpenFile(path string) (*Database, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return OpenBytes(data)
}

func OpenBytes(data []byte) (*Database, error) {
	header, err := parseHeader(data)
	if err != nil {
		return nil, err
	}

	pageSize := uint32(header.PageSize)
	if pageSize == 1 {
		pageSize = 65536
	}
	pages := make(map[uint32][]byte)

	totalPages := uint32(len(data)) / pageSize
	if header.PageCount > 0 && header.PageCount < totalPages {
		totalPages = header.PageCount
	}

	for i := uint32(0); i < totalPages; i++ {
		start := i * pageSize
		end := start + pageSize
		if end > uint32(len(data)) {
			break
		}
		pages[i+1] = data[start:end]
	}

	db := &Database{
		pages:    pages,
		pageSize: pageSize,
		header:   header,
		tables:   make(map[string]*TableInfo),
	}

	err = db.parseSchema()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *Database) Header() *Header {
	return db.header
}

func (db *Database) GetTables() []*TableInfo {
	result := make([]*TableInfo, 0, len(db.tables))
	for _, t := range db.tables {
		result = append(result, t)
	}
	return result
}

func (db *Database) GetTable(name string) (*TableInfo, bool) {
	t, ok := db.tables[name]
	return t, ok
}

func (db *Database) parseSchema() error {
	var rows []*Row
	err := db.scanTablePage(1, &rows)
	if err != nil {
		return err
	}

	for _, row := range rows {
		values := row.Values()
		if len(values) < 5 {
			continue
		}

		typeVal, _ := values[0].(string)
		nameVal, _ := values[2].(string)
		sqlVal, _ := values[4].(string)
		rootPageVal := values[3]

		if strings.ToLower(typeVal) != "table" {
			continue
		}
		if nameVal == "sqlite_sequence" {
			continue
		}

		var rootPage uint32
		switch v := rootPageVal.(type) {
		case int:
			rootPage = uint32(v)
		case int32:
			rootPage = uint32(v)
		case int64:
			rootPage = uint32(v)
		case uint32:
			rootPage = v
		case uint64:
			rootPage = uint32(v)
		case float64:
			rootPage = uint32(v)
		}

		db.tables[nameVal] = &TableInfo{
			Name:     nameVal,
			SQL:      sqlVal,
			RootPage: rootPage,
		}
	}

	return nil
}

func (db *Database) QueryTable(tableName string) ([]map[string]interface{}, error) {
	table, ok := db.tables[tableName]
	if !ok {
		return nil, nil
	}

	columns := extractColumnsFromSQL(table.SQL)

	var rows []*Row
	err := db.scanTablePage(table.RootPage, &rows)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		values := row.Values()
		rowMap := make(map[string]interface{})

		for i, col := range columns {
			if i < len(values) {
				rowMap[col] = values[i]
			} else {
				rowMap[col] = nil
			}
		}

		result = append(result, rowMap)
	}

	return result, nil
}

func (db *Database) QueryTableWithLimit(tableName string, limit, offset int) ([]map[string]interface{}, error) {
	allRows, err := db.QueryTable(tableName)
	if err != nil {
		return nil, err
	}

	if offset < 0 {
		offset = 0
	}
	if offset >= len(allRows) {
		return []map[string]interface{}{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(allRows) {
		end = len(allRows)
	}

	return allRows[offset:end], nil
}

func extractColumnsFromSQL(createSQL string) []string {
	createSQL = strings.TrimSpace(createSQL)
	lower := strings.ToLower(createSQL)

	createIdx := strings.Index(lower, "create")
	if createIdx == -1 {
		return nil
	}

	tableIdx := strings.Index(lower[createIdx:], "table")
	if tableIdx == -1 {
		return nil
	}

	openParen := strings.IndexByte(createSQL, '(')
	closeParen := strings.LastIndexByte(createSQL, ')')
	if openParen == -1 || closeParen == -1 || openParen >= closeParen {
		return nil
	}

	columnDefs := createSQL[openParen+1 : closeParen]

	var columns []string
	var inQuotes bool
	var inParens int
	var currentDef strings.Builder

	for i := 0; i < len(columnDefs); i++ {
		c := columnDefs[i]

		switch c {
		case '\'':
			inQuotes = !inQuotes
			currentDef.WriteByte(c)
		case '(':
			if !inQuotes {
				inParens++
			}
			currentDef.WriteByte(c)
		case ')':
			if !inQuotes && inParens > 0 {
				inParens--
			}
			currentDef.WriteByte(c)
		case ',':
			if !inQuotes && inParens == 0 {
				columns = append(columns, extractColumnName(currentDef.String()))
				currentDef.Reset()
			} else {
				currentDef.WriteByte(c)
			}
		default:
			currentDef.WriteByte(c)
		}
	}

	if currentDef.Len() > 0 {
		columns = append(columns, extractColumnName(currentDef.String()))
	}

	return columns
}

func extractColumnName(def string) string {
	def = strings.TrimSpace(def)

	lower := strings.ToLower(def)
	if strings.HasPrefix(lower, "constraint") ||
		strings.HasPrefix(lower, "primary key") ||
		strings.HasPrefix(lower, "unique") ||
		strings.HasPrefix(lower, "foreign key") ||
		strings.HasPrefix(lower, "check") {
		return ""
	}

	if len(def) == 0 {
		return ""
	}

	var name strings.Builder
	inQuote := false

	for i := 0; i < len(def); i++ {
		c := def[i]

		if c == '"' || c == '`' || c == '[' {
			inQuote = true
			continue
		}
		if inQuote {
			if c == '"' || c == '`' || c == ']' {
				break
			}
			name.WriteByte(c)
			continue
		}

		if c == ' ' || c == '\t' || c == '\n' {
			if name.Len() > 0 {
				break
			}
			continue
		}

		name.WriteByte(c)
	}

	return name.String()
}

func (db *Database) readUint16(data []byte, offset int) uint16 {
	return binary.BigEndian.Uint16(data[offset : offset+2])
}

func (db *Database) readUint32(data []byte, offset int) uint32 {
	return binary.BigEndian.Uint32(data[offset : offset+4])
}
