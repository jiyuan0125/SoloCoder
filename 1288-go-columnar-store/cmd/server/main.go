package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"columnar-store/pkg/api"
	"columnar-store/pkg/columnstore"
)

var store = columnstore.NewStore()

func main() {
	port := flag.String("port", "", "Server port (default: 8080)")
	flag.Parse()

	listenPort := *port
	if listenPort == "" {
		listenPort = os.Getenv("COLUMNAR_STORE_PORT")
	}
	if listenPort == "" {
		listenPort = "8515"
	}

	http.HandleFunc("/api/tables/create", handleCreateTable)
	http.HandleFunc("/api/tables/list", handleListTables)
	http.HandleFunc("/api/tables/insert", handleInsertRows)
	http.HandleFunc("/api/tables/query", handleQuery)
	http.HandleFunc("/api/tables/addColumn", handleAddColumn)
	http.HandleFunc("/api/tables/setEncoding", handleSetEncoding)
	http.HandleFunc("/api/tables/stats", handleStats)
	http.HandleFunc("/api/tables/export", handleExport)

	fmt.Printf("Columnar Store Server listening on :%s\n", listenPort)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, api.ErrorResponse{Error: err})
}

func handleCreateTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req api.CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Schema.Name == "" {
		writeError(w, http.StatusBadRequest, "table name is required")
		return
	}
	if len(req.Schema.Columns) == 0 {
		writeError(w, http.StatusBadRequest, "at least one column is required")
		return
	}
	cs := make([]columnstore.ColumnSchema, len(req.Schema.Columns))
	for i, c := range req.Schema.Columns {
		colType, err := columnstore.ParseColumnType(c.Type)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		var encType columnstore.EncodingType = columnstore.EncodingUnknown
		if c.Encoding != "" {
			et, err := columnstore.ParseEncodingType(c.Encoding)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			encType = et
		}
		cs[i] = columnstore.ColumnSchema{
			Name:     c.Name,
			Type:     colType,
			Encoding: encType,
		}
	}
	schema := columnstore.TableSchema{
		Name:    req.Schema.Name,
		Columns: cs,
	}
	if err := store.CreateTable(schema); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, api.CreateTableResponse{
		Success: true,
		Message: fmt.Sprintf("table %s created successfully", req.Schema.Name),
	})
}

func handleListTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, api.ListTablesResponse{
		Success: true,
		Tables:  store.ListTables(),
	})
}

func handleInsertRows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req api.InsertRowsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tbl, err := store.GetTable(req.Table)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	schema := tbl.Schema()
	rows := make([]columnstore.Row, len(req.Rows))
	for i, rowReq := range req.Rows {
		if len(rowReq.Values) != len(schema.Columns) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("row %d has %d values, expected %d", i, len(rowReq.Values), len(schema.Columns)))
			return
		}
		values := make([]columnstore.Value, len(rowReq.Values))
		for j, v := range rowReq.Values {
			val, err := parseValue(v, schema.Columns[j].Type)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("row %d, column %d: %s", i, j, err.Error()))
				return
			}
			values[j] = val
		}
		rows[i] = columnstore.Row{Values: values}
	}
	if err := tbl.InsertRows(rows); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, api.InsertRowsResponse{
		Success:  true,
		RowCount: len(rows),
		Message:  fmt.Sprintf("inserted %d rows", len(rows)),
	})
}

func parseValue(v interface{}, colType columnstore.ColumnType) (columnstore.Value, error) {
	if v == nil {
		return columnstore.Value{Null: true}, nil
	}
	switch colType {
	case columnstore.ColumnTypeString:
		if s, ok := v.(string); ok {
			if s == "" {
				return columnstore.Value{Null: true}, nil
			}
			return columnstore.Value{Str: s}, nil
		}
		return columnstore.Value{}, fmt.Errorf("expected string")
	case columnstore.ColumnTypeInt64:
		switch val := v.(type) {
		case float64:
			return columnstore.Value{Int: int64(val)}, nil
		case int:
			return columnstore.Value{Int: int64(val)}, nil
		case int64:
			return columnstore.Value{Int: val}, nil
		case string:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return columnstore.Value{}, fmt.Errorf("invalid int64: %s", val)
			}
			return columnstore.Value{Int: n}, nil
		}
		return columnstore.Value{}, fmt.Errorf("expected int64")
	case columnstore.ColumnTypeFloat64:
		switch val := v.(type) {
		case float64:
			return columnstore.Value{Float: val}, nil
		case int:
			return columnstore.Value{Float: float64(val)}, nil
		case int64:
			return columnstore.Value{Float: float64(val)}, nil
		case string:
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return columnstore.Value{}, fmt.Errorf("invalid float64: %s", val)
			}
			return columnstore.Value{Float: f}, nil
		}
		return columnstore.Value{}, fmt.Errorf("expected float64")
	default:
		return columnstore.Value{}, fmt.Errorf("unknown column type")
	}
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tbl, err := store.GetTable(req.Table)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	schema := tbl.Schema()
	predicates := make([]columnstore.Predicate, len(req.Predicates))
	for i, p := range req.Predicates {
		colIdx := schema.ColumnIndex(p.Column)
		if colIdx == -1 {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("column %s not found", p.Column))
			return
		}
		op, err := parsePredicateOp(p.Op)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		var val columnstore.Value
		if op != columnstore.OpIsNull && op != columnstore.OpIsNotNull {
			val, err = parseValue(p.Value, schema.Columns[colIdx].Type)
			if err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("predicate column %s: %s", p.Column, err.Error()))
				return
			}
		}
		predicates[i] = columnstore.Predicate{
			Column: p.Column,
			Op:     op,
			Value:  val,
		}
	}
	result, err := tbl.Query(req.Columns, predicates)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := api.QueryResponse{
		Success: true,
		Columns: make([]api.ColumnSchema, len(result.Columns)),
		Rows:    make([]api.RowResult, len(result.Rows)),
	}
	for i, c := range result.Columns {
		resp.Columns[i] = api.ColumnSchema{
			Name:     c.Name,
			Type:     c.Type.String(),
			Encoding: c.Encoding.String(),
		}
	}
	for i, row := range result.Rows {
		resp.Rows[i] = api.RowResult{
			Values: make([]interface{}, len(row.Values)),
		}
		for j, val := range row.Values {
			resp.Rows[i].Values[j] = valueToInterface(val)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func parsePredicateOp(op string) (columnstore.PredicateOp, error) {
	switch strings.ToLower(op) {
	case "=", "eq", "==":
		return columnstore.OpEq, nil
	case "!=", "<>", "neq":
		return columnstore.OpNeq, nil
	case "<", "lt":
		return columnstore.OpLt, nil
	case "<=", "lte":
		return columnstore.OpLte, nil
	case ">", "gt":
		return columnstore.OpGt, nil
	case ">=", "gte":
		return columnstore.OpGte, nil
	case "isnull", "is_null", "null":
		return columnstore.OpIsNull, nil
	case "isnotnull", "is_not_null", "notnull":
		return columnstore.OpIsNotNull, nil
	default:
		return 0, fmt.Errorf("unknown predicate operator: %s", op)
	}
}

func valueToInterface(v columnstore.Value) interface{} {
	if v.Null {
		return nil
	}
	if v.Str != "" {
		return v.Str
	}
	if v.Int != 0 {
		return v.Int
	}
	return v.Float
}

func handleAddColumn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req api.AddColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tbl, err := store.GetTable(req.Table)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	colType, err := columnstore.ParseColumnType(req.Column.Type)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var encType columnstore.EncodingType = columnstore.EncodingUnknown
	if req.Column.Encoding != "" {
		et, err := columnstore.ParseEncodingType(req.Column.Encoding)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		encType = et
	}
	colSchema := columnstore.ColumnSchema{
		Name:     req.Column.Name,
		Type:     colType,
		Encoding: encType,
	}
	if err := tbl.AddColumn(colSchema); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, api.AddColumnResponse{
		Success: true,
		Message: fmt.Sprintf("column %s added successfully", req.Column.Name),
	})
}

func handleSetEncoding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req api.SetEncodingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tbl, err := store.GetTable(req.Table)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	colIdx := tbl.ColumnIndex(req.Column)
	if colIdx == -1 {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("column %s not found", req.Column))
		return
	}
	encType, err := columnstore.ParseEncodingType(req.Encoding)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := tbl.SetColumnEncoding(colIdx, encType); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, api.SetEncodingResponse{
		Success: true,
		Message: fmt.Sprintf("column %s encoding set to %s", req.Column, req.Encoding),
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeError(w, http.StatusBadRequest, "table parameter is required")
		return
	}
	tbl, err := store.GetTable(tableName)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	cs := tbl.ColumnStats()
	stats := make([]api.ColumnStats, len(cs))
	for i, s := range cs {
		stats[i] = api.ColumnStats{
			Name:             s.Name,
			Type:             s.Type,
			Encoding:         s.Encoding,
			CompressionRatio: s.CompressionRatio,
			RowCount:         s.RowCount,
		}
	}
	writeJSON(w, http.StatusOK, api.StatsResponse{
		Success: true,
		Stats:   stats,
	})
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		writeError(w, http.StatusBadRequest, "table parameter is required")
		return
	}
	tbl, err := store.GetTable(tableName)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	result, err := tbl.ExportAll()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := api.ExportResponse{
		Success: true,
		Columns: make([]api.ColumnSchema, len(result.Columns)),
		Rows:    make([]api.RowResult, len(result.Rows)),
	}
	for i, c := range result.Columns {
		resp.Columns[i] = api.ColumnSchema{
			Name:     c.Name,
			Type:     c.Type.String(),
			Encoding: c.Encoding.String(),
		}
	}
	for i, row := range result.Rows {
		resp.Rows[i] = api.RowResult{
			Values: make([]interface{}, len(row.Values)),
		}
		for j, val := range row.Values {
			resp.Rows[i].Values[j] = valueToInterface(val)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
