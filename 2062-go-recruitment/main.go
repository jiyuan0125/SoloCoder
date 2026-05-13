package main

import (
	"fmt"
	"os"
	"path/filepath"

	"recruitment/cmd"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exe)

	dataDir := filepath.Join(exeDir, ".recruitment")
	calendarDir := filepath.Join(dataDir, "calendars")

	rootCmd := cmd.NewRootCmd(dataDir, calendarDir)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
