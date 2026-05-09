package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sqliteparser/pkg/shared"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "open":
		cmdOpen()
	case "tables":
		cmdTables()
	case "schema":
		cmdSchema()
	case "export":
		cmdExport()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`SQLite Parser Client

Usage:
  client open <file>      Upload SQLite file and enter interactive mode
  client tables           List all tables in the current database
  client schema <table>   Show CREATE TABLE statement for a table
  client export <table> <output.csv> Export table to CSV
  client help             Show this help`)
}

func uploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	writer.Close()

	resp, err := http.Post(serverURL+"/api/upload", writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp shared.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("%s", errResp.Error)
	}

	var uploadResp shared.UploadResponse
	json.NewDecoder(resp.Body).Decode(&uploadResp)
	fmt.Println(uploadResp.Message)

	return nil
}

func getTables() ([]shared.TableSchema, error) {
	resp, err := http.Get(serverURL + "/api/tables")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp shared.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var tablesResp shared.TablesResponse
	err = json.NewDecoder(resp.Body).Decode(&tablesResp)
	if err != nil {
		return nil, err
	}

	return tablesResp.Tables, nil
}

func queryTable(tableName string, limit, offset int) (*shared.QueryResponse, error) {
	url := fmt.Sprintf("%s/api/query?table=%s", serverURL, tableName)
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if offset > 0 {
		url += fmt.Sprintf("&offset=%d", offset)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp shared.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	var queryResp shared.QueryResponse
	err = json.NewDecoder(resp.Body).Decode(&queryResp)
	if err != nil {
		return nil, err
	}

	return &queryResp, nil
}

func cmdOpen() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: client open <file>")
		os.Exit(1)
	}

	filePath := os.Args[2]
	err := uploadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	interactiveMode()
}

func interactiveMode() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nInteractive mode. Enter table name to view data.")
	fmt.Println("Commands: tables, schema <table>, exit")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "exit", "quit":
			return
		case "tables":
			tables, err := getTables()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			for _, t := range tables {
				fmt.Printf("  %s\n", t.Name)
			}
		case "schema":
			if len(parts) < 2 {
				fmt.Println("usage: schema <table>")
				continue
			}
			tables, err := getTables()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			found := false
			for _, t := range tables {
				if t.Name == parts[1] {
					fmt.Println(t.CreateSQL)
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("table not found: %s\n", parts[1])
			}
		default:
			limit := 50
			offset := 0
			if len(parts) >= 2 {
				if l, err := strconv.Atoi(parts[1]); err == nil {
					limit = l
				}
			}
			if len(parts) >= 3 {
				if o, err := strconv.Atoi(parts[2]); err == nil {
					offset = o
				}
			}

			result, err := queryTable(cmd, limit, offset)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}

			displayTable(result)
		}
	}
}

func displayTable(result *shared.QueryResponse) {
	if len(result.Rows) == 0 {
		fmt.Println("(empty)")
		return
	}

	columns := make([]string, 0)
	if len(result.Rows) > 0 {
		for k := range result.Rows[0] {
			columns = append(columns, k)
		}
	}

	fmt.Println("\nColumns:", strings.Join(columns, ", "))
	fmt.Printf("\nShowing %d of %d rows\n\n", len(result.Rows), result.Count)

	for i, row := range result.Rows {
		fmt.Printf("Row %d:\n", i+1)
		for _, col := range columns {
			v := row[col]
			valStr := formatValue(v)
			fmt.Printf("  %s: %s\n", col, valStr)
		}
		fmt.Println()
	}
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case map[string]interface{}:
		if val["$type"] == "blob" {
			return fmt.Sprintf("<BLOB %d bytes>", len(val["data"].([]interface{})))
		}
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func cmdTables() {
	tables, err := getTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return
	}

	for _, t := range tables {
		fmt.Printf("%s\n", t.Name)
	}
}

func cmdSchema() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: client schema <table>")
		os.Exit(1)
	}

	tableName := os.Args[2]
	tables, err := getTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, t := range tables {
		if t.Name == tableName {
			fmt.Println(t.CreateSQL)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "table not found: %s\n", tableName)
	os.Exit(1)
}

func cmdExport() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: client export <table> <output.csv>")
		os.Exit(1)
	}

	tableName := os.Args[2]
	outputFile := os.Args[3]

	result, err := queryTable(tableName, 0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	columns := make([]string, 0)
	if len(result.Rows) > 0 {
		for k := range result.Rows[0] {
			columns = append(columns, k)
		}
	}

	writer.Write(columns)

	for _, row := range result.Rows {
		record := make([]string, len(columns))
		for i, col := range columns {
			record[i] = formatValueForCSV(row[col])
		}
		writer.Write(record)
	}

	fmt.Printf("Exported %d rows to %s\n", len(result.Rows), outputFile)
}

func formatValueForCSV(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case map[string]interface{}:
		if val["$type"] == "blob" {
			return "<BLOB>"
		}
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
