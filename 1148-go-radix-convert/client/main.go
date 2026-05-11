package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"baseconv/common"
)

func main() {
	var precision int
	flag.IntVar(&precision, "precision", 32, "decimal precision for fractional part")
	flag.Parse()

	args := flag.Args()
	if len(args) != 3 {
		fmt.Println("Usage: baseconv [--precision <n>] <from_base> <to_base> <number>")
		fmt.Println("Example: baseconv 10 16 \"255\"")
		fmt.Println("Example: baseconv --precision 16 10 2 \"0.1\"")
		os.Exit(1)
	}

	fromBase, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Error: invalid from_base: %v\n", err)
		os.Exit(1)
	}

	toBase, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("Error: invalid to_base: %v\n", err)
		os.Exit(1)
	}

	input := args[2]

	req := common.ConvertRequest{
		Input:     input,
		FromBase:  fromBase,
		ToBase:    toBase,
		Precision: precision,
	}

	resp, err := callServer(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		os.Exit(1)
	}

	fmt.Println(resp.Result)
}
