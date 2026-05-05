package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"xlsx-reader/pkg/common"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "upload":
		uploadCommand()
	case "list":
		listCommand()
	case "read":
		readCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client upload <file.xlsx>          Upload an Excel file")
	fmt.Println("  client list <filename>              List all sheets in a file")
	fmt.Println("  client read <filename> <sheetname> [options]  Read sheet data")
	fmt.Println("")
	fmt.Println("Read options:")
	fmt.Println("  -start=N     Start row (1-based, default: 1)")
	fmt.Println("  -end=N       End row (default: all)")
	fmt.Println("  -no-skip     Don't skip empty rows")
}

func uploadCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client upload <file.xlsx>")
		os.Exit(1)
	}

	filename := os.Args[2]
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", filename)
		os.Exit(1)
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		fmt.Printf("Failed to create form file: %v\n", err)
		os.Exit(1)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Failed to copy file: %v\n", err)
		os.Exit(1)
	}

	writer.Close()

	resp, err := http.Post(serverURL+"/upload", writer.FormDataContentType(), body)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
}

func listCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client list <filename>")
		os.Exit(1)
	}

	filename := os.Args[2]

	resp, err := http.Get(fmt.Sprintf("%s/list-sheets?filename=%s", serverURL, filename))
	if err != nil {
		fmt.Printf("Failed to list sheets: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.ListSheetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}

	fmt.Println("Available sheets:")
	for i, name := range result.SheetNames {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
}

func readCommand() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: client read <filename> <sheetname> [options]")
		os.Exit(1)
	}

	filename := os.Args[2]
	sheetname := os.Args[3]

	flags := flag.NewFlagSet("read", flag.ExitOnError)
	startRow := flags.Int("start", 0, "Start row (1-based)")
	endRow := flags.Int("end", 0, "End row")
	noSkip := flags.Bool("no-skip", false, "Don't skip empty rows")

	remainingArgs := os.Args[4:]
	if err := flags.Parse(remainingArgs); err != nil {
		os.Exit(1)
	}

	req := common.ReadSheetRequest{
		SheetName:     sheetname,
		StartRow:      *startRow,
		EndRow:        *endRow,
		SkipEmptyRows: !*noSkip,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/read-sheet?filename=%s", serverURL, filename)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to read sheet: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.ReadSheetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to parse response: %v\nResponse: %s\n", err, string(respBody))
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Error)
		if len(result.SheetNames) > 0 {
			fmt.Println("\nAvailable sheets:")
			for i, name := range result.SheetNames {
				fmt.Printf("  %d. %s\n", i+1, name)
			}
		}
		os.Exit(1)
	}

	printTable(result.Data)
}

func printTable(data [][]string) {
	if len(data) == 0 {
		fmt.Println("(No data)")
		return
	}

	colWidths := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
			if colWidths[i] > 50 {
				colWidths[i] = 50
			}
		}
	}

	printSeparator(colWidths)

	for rowIdx, row := range data {
		fmt.Print("|")
		for i, cell := range row {
			display := cell
			if len(display) > 50 {
				display = display[:47] + "..."
			}
			fmt.Printf(" %-*s |", colWidths[i], display)
		}
		fmt.Println()

		if rowIdx == 0 {
			printSeparator(colWidths)
		}
	}

	printSeparator(colWidths)
	fmt.Printf("\nTotal: %d rows\n", len(data))
}

func printSeparator(colWidths []int) {
	fmt.Print("+")
	for _, w := range colWidths {
		fmt.Print(strings.Repeat("-", w+2))
		fmt.Print("+")
	}
	fmt.Println()
}
