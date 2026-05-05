package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"csv-merger/common"
)

type Config struct {
	ServerHost   string
	ServerPort   string
	OutputFile   string
	InputFiles   []string
	DedupColumns []string
	SortColumn   string
	SortOrder    common.SortOrder
	CustomHeader []string
}

func Parse() (*Config, error) {
	config := &Config{
		ServerHost: common.DefaultServerHost,
		ServerPort: common.DefaultServerPort,
		SortOrder:  common.SortAsc,
	}

	flag.StringVar(&config.ServerHost, "host", common.DefaultServerHost, "Server host")
	flag.StringVar(&config.ServerPort, "port", common.DefaultServerPort, "Server port")
	flag.StringVar(&config.OutputFile, "o", "", "Output file (required)")
	flag.StringVar(&config.OutputFile, "output", "", "Output file (required)")

	var dedupColumns string
	flag.StringVar(&dedupColumns, "dedup", "", "Comma-separated columns for deduplication")

	var sortSpec string
	flag.StringVar(&sortSpec, "sort", "", "Sort specification: column[:asc|:desc]")

	var customHeader string
	flag.StringVar(&customHeader, "header", "", "Comma-separated custom header")

	flag.Usage = usage
	flag.Parse()

	config.InputFiles = flag.Args()

	if config.OutputFile == "" {
		return nil, fmt.Errorf("output file is required (use -o or --output)")
	}

	if len(config.InputFiles) == 0 {
		return nil, fmt.Errorf("at least one input file is required")
	}

	if dedupColumns != "" {
		config.DedupColumns = splitAndTrim(dedupColumns)
	}

	if sortSpec != "" {
		parts := strings.SplitN(sortSpec, ":", 2)
		config.SortColumn = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			order := strings.ToLower(strings.TrimSpace(parts[1]))
			if order == "desc" {
				config.SortOrder = common.SortDesc
			} else {
				config.SortOrder = common.SortAsc
			}
		}
	}

	if customHeader != "" {
		config.CustomHeader = splitAndTrim(customHeader)
	}

	return config, nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func usage() {
	fmt.Fprintf(os.Stderr, `CSV Merger Client

Merge multiple CSV files into one.

Usage:
  csv-merger -o output.csv [options] input1.csv input2.csv ...

Options:
  -o, --output <file>    Output file (required)
  --host <host>          Server host (default: localhost)
  --port <port>          Server port (default: 9999)
  --dedup <cols>         Comma-separated columns for deduplication
                         (later records overwrite earlier ones)
  --sort <spec>          Sort specification: column[:asc|:desc]
                         Numeric columns are sorted numerically
  --header <cols>        Comma-separated custom header to replace original

Examples:
  csv-merger -o yearly.csv jan.csv feb.csv mar.csv
  csv-merger -o yearly.csv --dedup "Order ID" jan.csv feb.csv
  csv-merger -o yearly.csv --sort "Amount:desc" *.csv
  csv-merger -o yearly.csv --header "ID,Name,Amount" *.csv
`)
}
