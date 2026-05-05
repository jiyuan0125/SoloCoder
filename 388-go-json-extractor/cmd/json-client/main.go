package main

import (
	"fmt"
	"os"
)

func main() {
	cli := NewClientFromArgs()
	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
