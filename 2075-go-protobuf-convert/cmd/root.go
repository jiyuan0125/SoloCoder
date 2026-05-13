package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "protoconv",
	Short: "Protobuf conversion system",
	Long: `A command-line tool for converting between Protobuf binary and JSON formats.
Supports proto3, nested messages, enums, repeated, map, and oneof fields.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("storage", "", "Path to resource storage file")
}
