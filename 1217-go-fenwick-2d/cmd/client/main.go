package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"fenwick2d/api"
)

var serverURL = "http://localhost:8300"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Usage: create <rows> <cols>")
			os.Exit(1)
		}
		rows, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid rows: %v\n", err)
			os.Exit(1)
		}
		cols, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid cols: %v\n", err)
			os.Exit(1)
		}
		handleCreate(rows, cols)
	case "update":
		if len(args) != 3 {
			fmt.Fprintln(os.Stderr, "Usage: update <x> <y> <delta>")
			os.Exit(1)
		}
		x, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid x: %v\n", err)
			os.Exit(1)
		}
		y, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid y: %v\n", err)
			os.Exit(1)
		}
		delta, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid delta: %v\n", err)
			os.Exit(1)
		}
		handleUpdate(x, y, delta)
	case "sum":
		if len(args) != 4 {
			fmt.Fprintln(os.Stderr, "Usage: sum <lx> <ly> <rx> <ry>")
			os.Exit(1)
		}
		lx, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid lx: %v\n", err)
			os.Exit(1)
		}
		ly, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid ly: %v\n", err)
			os.Exit(1)
		}
		rx, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid rx: %v\n", err)
			os.Exit(1)
		}
		ry, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid ry: %v\n", err)
			os.Exit(1)
		}
		handleSum(lx, ly, rx, ry)
	case "init":
		handleInit()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: client <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  create <rows> <cols>     Create a new matrix")
	fmt.Println("  update <x> <y> <delta>   Update a position")
	fmt.Println("  sum <lx> <ly> <rx> <ry>  Query sum in range")
	fmt.Println("  init                     Initialize with matrix (reads from stdin)")
}

func handleCreate(rows, cols int) {
	req := api.CreateRequest{Rows: rows, Cols: cols}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/create", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var res api.CreateResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		os.Exit(1)
	}
	fmt.Println("Created successfully")
}

func handleUpdate(x, y int, delta int64) {
	req := api.UpdateRequest{X: x, Y: y, Delta: delta}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/update", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var res api.UpdateResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		os.Exit(1)
	}
	fmt.Println("Updated successfully")
}

func handleSum(lx, ly, rx, ry int) {
	req := api.SumRequest{Lx: lx, Ly: ly, Rx: rx, Ry: ry}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/sum", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var res api.SumResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		os.Exit(1)
	}
	fmt.Println(res.Sum)
}

func handleInit() {
	scanner := bufio.NewScanner(os.Stdin)
	var matrix [][]int64

	fmt.Println("Enter matrix dimensions first (rows cols), then the matrix data (space-separated values per line):")
	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "failed to read dimensions\n")
		os.Exit(1)
	}

	dimLine := scanner.Text()
	parts := strings.Fields(dimLine)
	if len(parts) != 2 {
		fmt.Fprintf(os.Stderr, "invalid dimensions format\n")
		os.Exit(1)
	}

	rows, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid rows: %v\n", err)
		os.Exit(1)
	}
	cols, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid cols: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < rows; i++ {
		if !scanner.Scan() {
			fmt.Fprintf(os.Stderr, "unexpected EOF\n")
			os.Exit(1)
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != cols {
			fmt.Fprintf(os.Stderr, "row %d has %d elements, expected %d\n", i+1, len(parts), cols)
			os.Exit(1)
		}
		row := make([]int64, cols)
		for j, s := range parts {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid value at (%d,%d): %v\n", i+1, j+1, err)
				os.Exit(1)
			}
			row[j] = val
		}
		matrix = append(matrix, row)
	}

	req := api.InitRequest{Matrix: matrix}
	body, _ := json.Marshal(req)

	resp, err := http.Post(serverURL+"/init", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var res api.InitResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Success {
		fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		os.Exit(1)
	}
	fmt.Println("Initialized successfully")
}
