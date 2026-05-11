package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"isbn-lookup/api"
)

var serverAddr string

func main() {
	flag.StringVar(&serverAddr, "server", "http://localhost:8602", "server address")

	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)
	loadFile := loadCmd.String("file", "", "path to data file")

	queryCmd := flag.NewFlagSet("query", flag.ExitOnError)
	queryISBN := queryCmd.String("isbn", "", "ISBN for exact match")
	queryTitle := queryCmd.String("title", "", "title keyword for fuzzy match")
	queryAuthors := queryCmd.String("authors", "", "authors (semicolon-separated for multiple)")
	queryPublisher := queryCmd.String("publisher", "", "publisher")
	queryPage := queryCmd.Int("page", 1, "page number")
	queryPageSize := queryCmd.Int("page-size", 20, "page size")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "load":
		loadCmd.Parse(os.Args[2:])
		if *loadFile == "" {
			fmt.Println("Error: -file is required")
			loadCmd.PrintDefaults()
			os.Exit(1)
		}
		handleLoad(*loadFile)
	case "query":
		queryCmd.Parse(os.Args[2:])
		handleQuery(*queryISBN, *queryTitle, *queryAuthors, *queryPublisher, *queryPage, *queryPageSize)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client load -file <path>          Load data from file")
	fmt.Println("  client query [options]            Query books")
	fmt.Println("")
	fmt.Println("Query options:")
	fmt.Println("  -isbn string       ISBN for exact match")
	fmt.Println("  -title string      Title keyword for fuzzy match")
	fmt.Println("  -authors string    Authors (semicolon-separated)")
	fmt.Println("  -publisher string  Publisher")
	fmt.Println("  -page int          Page number (default 1)")
	fmt.Println("  -page-size int     Page size (default 20)")
}

func handleLoad(filePath string) {
	req := api.LoadRequest{
		FilePath: filePath,
	}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverAddr+"/api/load", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.LoadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("Load complete:\n")
	fmt.Printf("  Total records: %d\n", result.TotalRecords)
	fmt.Printf("  Loaded: %d\n", result.LoadedRecords)
	fmt.Printf("  Invalid: %d\n", len(result.InvalidRecords))

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if len(result.InvalidRecords) > 0 {
		fmt.Println("\nInvalid records:")
		for _, ir := range result.InvalidRecords {
			fmt.Printf("  Line %d: %s\n", ir.LineNumber, ir.Error)
			fmt.Printf("    %s\n", ir.Line)
		}
	}
}

func handleQuery(isbn, title, authors, publisher string, page, pageSize int) {
	req := api.QueryRequest{
		ISBN:      isbn,
		Title:     title,
		Authors:   authors,
		Publisher: publisher,
		Page:      page,
		PageSize:  pageSize,
	}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverAddr+"/api/query", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result api.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error: Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	if result.Total == 0 {
		fmt.Println("No matching records found.")
		return
	}

	fmt.Printf("Found %d record(s). Page %d/%d (page size: %d)\n",
		result.Total, result.CurrentPage, result.TotalPages, result.PageSize)

	printBooksTable(result.Books)
}

func printBooksTable(books []api.BookDTO) {
	if len(books) == 0 {
		return
	}

	isbnWidth := 4
	titleWidth := 5
	authorWidth := 6
	publisherWidth := 9

	for _, b := range books {
		if len(b.ISBN) > isbnWidth {
			isbnWidth = len(b.ISBN)
		}
		if len(b.Title) > titleWidth {
			titleWidth = len(b.Title)
		}
		authors := strings.Join(b.Authors, ", ")
		if len(authors) > authorWidth {
			authorWidth = len(authors)
		}
		if len(b.Publisher) > publisherWidth {
			publisherWidth = len(b.Publisher)
		}
	}

	format := fmt.Sprintf("%%-%ds | %%-%ds | %%-%ds | %%-%ds\n",
		isbnWidth, titleWidth, authorWidth, publisherWidth)

	separator := strings.Repeat("-", isbnWidth+titleWidth+authorWidth+publisherWidth+9)

	fmt.Println(separator)
	fmt.Printf(format, "ISBN", "Title", "Authors", "Publisher")
	fmt.Println(separator)

	for _, b := range books {
		authors := strings.Join(b.Authors, ", ")
		fmt.Printf(format, b.ISBN, b.Title, authors, b.Publisher)
	}

	fmt.Println(separator)
}
