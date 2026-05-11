package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateTableRequest struct {
	Schema TableSchema `json:"schema"`
}

type CreateTableResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type TableSchema struct {
	Name    string         `json:"name"`
	Columns []ColumnSchema `json:"columns"`
}

type ColumnSchema struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Encoding string `json:"encoding,omitempty"`
}

type InsertRowsRequest struct {
	Table string       `json:"table"`
	Rows  []RowRequest `json:"rows"`
}

type RowRequest struct {
	Values []interface{} `json:"values"`
}

type InsertRowsResponse struct {
	Success   bool   `json:"success"`
	RowCount  int    `json:"row_count"`
	Message   string `json:"message"`
}

type QueryRequest struct {
	Table      string         `json:"table"`
	Columns    []string       `json:"columns,omitempty"`
	Predicates []PredicateRequest `json:"predicates,omitempty"`
}

type PredicateRequest struct {
	Column string      `json:"column"`
	Op     string      `json:"op"`
	Value  interface{} `json:"value,omitempty"`
}

type QueryResponse struct {
	Success bool           `json:"success"`
	Columns []ColumnSchema `json:"columns"`
	Rows    []RowResult    `json:"rows"`
}

type RowResult struct {
	Values []interface{} `json:"values"`
}

type AddColumnRequest struct {
	Table  string       `json:"table"`
	Column ColumnSchema `json:"column"`
}

type AddColumnResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SetEncodingRequest struct {
	Table    string `json:"table"`
	Column   string `json:"column"`
	Encoding string `json:"encoding"`
}

type SetEncodingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type StatsRequest struct {
	Table string `json:"table"`
}

type StatsResponse struct {
	Success bool          `json:"success"`
	Stats   []ColumnStats `json:"stats"`
}

type ColumnStats struct {
	Name             string  `json:"name"`
	Type             string  `json:"type"`
	Encoding         string  `json:"encoding"`
	CompressionRatio float64 `json:"compression_ratio"`
	RowCount         int     `json:"row_count"`
}

type ListTablesResponse struct {
	Success bool     `json:"success"`
	Tables  []string `json:"tables"`
}

type ExportRequest struct {
	Table string `json:"table"`
}

type ExportResponse struct {
	Success bool           `json:"success"`
	Columns []ColumnSchema `json:"columns"`
	Rows    []RowResult    `json:"rows"`
}
