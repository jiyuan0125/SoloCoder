package main

import (
	"fmt"
	"os"

	"contract-manager/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("错误:", err)
		os.Exit(1)
	}
}
