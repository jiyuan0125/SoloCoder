package main

import (
	"os"

	"github.com/asn1-parser/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
