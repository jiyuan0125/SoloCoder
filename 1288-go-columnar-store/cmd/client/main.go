package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"columnar-store/pkg/api"
)

var serverURL = "http://localhost:8515"

func main() {
	urlFlag := flag.String("server", "", "Server URL (default: http://localhost:8515)")
	flag.Parse()

	if *urlFlag != "" {
		serverURL = *urlFlag
	} else if envURL := os.Getenv("COLUMNAR_STORE_URL"); envURL != "" {
		serverURL = envURL
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "create":
		cmdCreate(args[1:])
	case "list":
		cmdList()
	case "import":
		cmdImport(args[1:])
	case "insert":
		cmdInsert(args[1:])
	case "query":
		cmdQuery(args[1:])
	case "addColumn":
		cmdAddColumn(args[1:])
	case "setEncoding":
		cmdSetEncoding(args[1:])
	case "stats":
		cmdStats(args[1:])
	case "export":
		cmdExport(args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Columnar Store Client

Usage:
  client [flags] command [arguments]

Flags:
  -server string    Server URL (default: http://localhost:8515)

Commands:
  create <table> <col1:type[:encoding]> <col2:type[:encoding]> ...
      Create a new table. Types: string, int64, float64. Encodings: dictionary, runlength.

  list
      List all tables.

  import <table> <csv-file> <col1:type[:encoding]> <col2:type[:encoding]> ...
      Create table and import data from CSV file.

  insert <table> <csv-file>
      Insert data from CSV file into existing table.

  query <table> [column1,column2,...] [where col op value ...]
      Query table with optional column selection and predicates.
      Operators: =, !=, <, <=, >, >=, isnull, notnull.

  addColumn <table> <name:type[:encoding]>
      Add a new column to existing table.

  setEncoding <table> <column> <encoding>
      Set encoding type for a column.

  stats <table>
      Show column statistics including compression ratios.

  export <table> <output-csv>
      Export entire table to CSV file.`)
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func parseColumnSpec(spec string) (api.ColumnSchema, error) {
	parts := strings.Split(spec, ":")
	if len(parts) < 2 {
		return api.ColumnSchema{}, fmt.Errorf("invalid column spec: %s (expected name:type[:encoding])", spec)
	}
	col := api.ColumnSchema{
		Name: parts[0],
		Type: parts[1],
	}
	if len(parts) >= 3 {
		col.Encoding = parts[2]
	}
	return col, nil
}

func cmdCreate(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: create <table> <col1:type[:encoding]> ...")
		os.Exit(1)
	}
	tableName := args[0]
	cols := make([]api.ColumnSchema, len(args)-1)
	for i := 1; i < len(args); i++ {
		col, err := parseColumnSpec(args[i])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cols[i-1] = col
	}
	req := api.CreateTableRequest{
		Schema: api.TableSchema{
			Name:    tableName,
			Columns: cols,
		},
	}
	respData, err := httpPost(serverURL+"/api/tables/create", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.CreateTableResponse
	json.Unmarshal(respData, &resp)
	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func cmdList() {
	respData, err := httpGet(serverURL + "/api/tables/list")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.ListTablesResponse
	json.Unmarshal(respData, &resp)
	if resp.Success {
		if len(resp.Tables) == 0 {
			fmt.Println("No tables found.")
		} else {
			fmt.Println("Tables:")
			for _, t := range resp.Tables {
				fmt.Printf("  - %s\n", t)
			}
		}
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func cmdImport(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: import <table> <csv-file> <col1:type[:encoding]> ...")
		os.Exit(1)
	}
	tableName := args[0]
	csvFile := args[1]
	cols := make([]api.ColumnSchema, len(args)-2)
	for i := 2; i < len(args); i++ {
		col, err := parseColumnSpec(args[i])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cols[i-2] = col
	}
	rows, err := readCSVFile(csvFile, len(cols))
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}
	createReq := api.CreateTableRequest{
		Schema: api.TableSchema{
			Name:    tableName,
			Columns: cols,
		},
	}
	respData, err := httpPost(serverURL+"/api/tables/create", createReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var createResp api.CreateTableResponse
	json.Unmarshal(respData, &createResp)
	if !createResp.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error creating table: %s\n", errResp.Error)
		os.Exit(1)
	}
	fmt.Println(createResp.Message)
	if len(rows) > 0 {
		insertReq := api.InsertRowsRequest{
			Table: tableName,
			Rows:  rows,
		}
		respData, err = httpPost(serverURL+"/api/tables/insert", insertReq)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		var insertResp api.InsertRowsResponse
		json.Unmarshal(respData, &insertResp)
		if insertResp.Success {
			fmt.Println(insertResp.Message)
		} else {
			var errResp api.ErrorResponse
			json.Unmarshal(respData, &errResp)
			fmt.Printf("Error inserting rows: %s\n", errResp.Error)
			os.Exit(1)
		}
	}
}

func cmdInsert(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: insert <table> <csv-file>")
		os.Exit(1)
	}
	tableName := args[0]
	csvFile := args[1]
	rows, err := readCSVFileAuto(csvFile)
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}
	if len(rows) == 0 {
		fmt.Println("No data to insert.")
		return
	}
	insertReq := api.InsertRowsRequest{
		Table: tableName,
		Rows:  rows,
	}
	respData, err := httpPost(serverURL+"/api/tables/insert", insertReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var insertResp api.InsertRowsResponse
	json.Unmarshal(respData, &insertResp)
	if insertResp.Success {
		fmt.Println(insertResp.Message)
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func readCSVFile(filename string, expectedCols int) ([]api.RowRequest, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	rows := make([]api.RowRequest, 0)
	for i, rec := range records {
		if len(rec) != expectedCols {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", i, len(rec), expectedCols)
		}
		values := make([]interface{}, len(rec))
		for j, v := range rec {
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				values[j] = nil
			} else {
				values[j] = trimmed
			}
		}
		rows = append(rows, api.RowRequest{Values: values})
	}
	return rows, nil
}

func readCSVFileAuto(filename string) ([]api.RowRequest, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	rows := make([]api.RowRequest, 0)
	for _, rec := range records {
		values := make([]interface{}, len(rec))
		for j, v := range rec {
			trimmed := strings.TrimSpace(v)
			if trimmed == "" {
				values[j] = nil
			} else {
				values[j] = trimmed
			}
		}
		rows = append(rows, api.RowRequest{Values: values})
	}
	return rows, nil
}

func cmdQuery(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: query <table> [columns] [where col op value ...]")
		os.Exit(1)
	}
	tableName := args[0]
	req := api.QueryRequest{
		Table:      tableName,
		Columns:    nil,
		Predicates: nil,
	}
	remaining := args[1:]
	if len(remaining) > 0 && remaining[0] != "where" {
		colStr := remaining[0]
		remaining = remaining[1:]
		if colStr != "" {
			req.Columns = strings.Split(colStr, ",")
			for i := range req.Columns {
				req.Columns[i] = strings.TrimSpace(req.Columns[i])
			}
		}
	}
	if len(remaining) > 0 && remaining[0] == "where" {
		remaining = remaining[1:]
		for len(remaining) >= 2 {
			col := remaining[0]
			op := remaining[1]
			remaining = remaining[2:]
			pred := api.PredicateRequest{
				Column: col,
				Op:     op,
			}
			opLower := strings.ToLower(op)
			if opLower != "isnull" && opLower != "is_null" && opLower != "null" &&
				opLower != "isnotnull" && opLower != "is_not_null" && opLower != "notnull" {
				if len(remaining) < 1 {
					fmt.Println("Missing value for predicate")
					os.Exit(1)
				}
				pred.Value = remaining[0]
				remaining = remaining[1:]
			}
			req.Predicates = append(req.Predicates, pred)
		}
	}
	respData, err := httpPost(serverURL+"/api/tables/query", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.QueryResponse
	json.Unmarshal(respData, &resp)
	if resp.Success {
		printQueryResult(resp)
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func printQueryResult(resp api.QueryResponse) {
	if len(resp.Columns) == 0 {
		fmt.Println("Empty result.")
		return
	}
	header := make([]string, len(resp.Columns))
	for i, c := range resp.Columns {
		header[i] = c.Name
	}
	fmt.Println(strings.Join(header, "\t"))
	for _, row := range resp.Rows {
		vals := make([]string, len(row.Values))
		for i, v := range row.Values {
			if v == nil {
				vals[i] = "NULL"
			} else {
				vals[i] = fmt.Sprintf("%v", v)
			}
		}
		fmt.Println(strings.Join(vals, "\t"))
	}
	fmt.Printf("\n%d rows returned.\n", len(resp.Rows))
}

func cmdAddColumn(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: addColumn <table> <name:type[:encoding]>")
		os.Exit(1)
	}
	col, err := parseColumnSpec(args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req := api.AddColumnRequest{
		Table:  args[0],
		Column: col,
	}
	respData, err := httpPost(serverURL+"/api/tables/addColumn", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.AddColumnResponse
	json.Unmarshal(respData, &resp)
	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func cmdSetEncoding(args []string) {
	if len(args) != 3 {
		fmt.Println("Usage: setEncoding <table> <column> <encoding>")
		os.Exit(1)
	}
	req := api.SetEncodingRequest{
		Table:    args[0],
		Column:   args[1],
		Encoding: args[2],
	}
	respData, err := httpPost(serverURL+"/api/tables/setEncoding", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.SetEncodingResponse
	json.Unmarshal(respData, &resp)
	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func cmdStats(args []string) {
	if len(args) != 1 {
		fmt.Println("Usage: stats <table>")
		os.Exit(1)
	}
	respData, err := httpGet(serverURL + "/api/tables/stats?table=" + args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.StatsResponse
	json.Unmarshal(respData, &resp)
	if resp.Success {
		fmt.Printf("Table: %s, Rows: %d\n", args[0], resp.Stats[0].RowCount)
		fmt.Println("Columns:")
		for _, s := range resp.Stats {
			fmt.Printf("  %-20s type=%-10s encoding=%-12s compression=%.2f\n",
				s.Name, s.Type, s.Encoding, s.CompressionRatio)
		}
	} else {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
}

func cmdExport(args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: export <table> <output-csv>")
		os.Exit(1)
	}
	tableName := args[0]
	outputFile := args[1]
	respData, err := httpGet(serverURL + "/api/tables/export?table=" + tableName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	var resp api.ExportResponse
	json.Unmarshal(respData, &resp)
	if !resp.Success {
		var errResp api.ErrorResponse
		json.Unmarshal(respData, &errResp)
		fmt.Printf("Error: %s\n", errResp.Error)
		os.Exit(1)
	}
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	header := make([]string, len(resp.Columns))
	for i, c := range resp.Columns {
		header[i] = c.Name
	}
	writer.Write(header)
	for _, row := range resp.Rows {
		vals := make([]string, len(row.Values))
		for i, v := range row.Values {
			if v == nil {
				vals[i] = ""
			} else {
				vals[i] = fmt.Sprintf("%v", v)
			}
		}
		writer.Write(vals)
	}
	writer.Flush()
	fmt.Printf("Exported %d rows to %s\n", len(resp.Rows), outputFile)
}
