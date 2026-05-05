package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"go-db-dumper/proto"
)

func main() {
	port := flag.Int("port", proto.DefaultPort, "Server port")
	host := flag.String("host", proto.DefaultHost, "Server host")
	tableName := flag.String("table", "", "Table name (defaults to input filename without extension)")
	inputFile := flag.String("input", "", "Input CSV file (required)")
	outputFile := flag.String("output", "", "Output SQL file (required)")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --input and --output are required")
		fmt.Fprintf(os.Stderr, "Usage: %s --input <csv_file> --output <sql_file> [--table <table_name>] [--host <host>] [--port <port>]\n", os.Args[0])
		os.Exit(1)
	}

	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", *inputFile)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to connect to server at %s: %v\n", addr, err)
		fmt.Fprintln(os.Stderr, "Make sure the server is running. Start it with: ./server")
		os.Exit(1)
	}
	defer conn.Close()

	request := &proto.Message{
		Type: proto.MessageTypeRequest,
		Request: &proto.ConvertRequest{
			InputFile:  *inputFile,
			OutputFile: *outputFile,
			TableName:  *tableName,
		},
	}

	fmt.Printf("Connecting to server at %s...\n", addr)
	fmt.Printf("Sending conversion request:\n")
	fmt.Printf("  Input:  %s\n", *inputFile)
	fmt.Printf("  Output: %s\n", *outputFile)
	if *tableName != "" {
		fmt.Printf("  Table:  %s\n", *tableName)
	}

	err = proto.WriteMessage(conn, request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to send request: %v\n", err)
		os.Exit(1)
	}

	response, err := proto.ReadMessage(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if response.Type != proto.MessageTypeResponse || response.Response == nil {
		fmt.Fprintln(os.Stderr, "Error: Invalid response from server")
		os.Exit(1)
	}

	resp := response.Response
	if resp.Status == proto.ResponseStatusSuccess {
		fmt.Println("\nConversion completed successfully!")
		fmt.Printf("  Records read: %d\n", resp.RecordsRead)
		fmt.Printf("  SQL statements generated: %d\n", resp.SQLCount)
		fmt.Printf("  Output written to: %s\n", *outputFile)
	} else {
		fmt.Fprintf(os.Stderr, "\nError: %s\n", resp.Message)
		os.Exit(1)
	}
}
