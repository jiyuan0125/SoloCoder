package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var format string
	var fields string
	var fieldMapping string
	var serverURL string
	var outputFile string

	flag.StringVar(&format, "format", "csv", "Export format: csv, json, jsonl")
	flag.StringVar(&fields, "fields", "", "Comma-separated list of fields to export")
	flag.StringVar(&fieldMapping, "mapping", "", "Field mapping in format: key1:value1,key2:value2")
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.StringVar(&outputFile, "output", "", "Output file (default: stdout)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  cat data.json | %s -format csv -fields id,name -mapping id:编号,name:姓名 > output.csv\n", os.Args[0])
	}

	flag.Parse()

	if err := run(format, fields, fieldMapping, serverURL, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
