package main

import (
	"fmt"
	"os"

	"qc-process/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
