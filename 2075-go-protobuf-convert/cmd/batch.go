package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"protoconv/internal/proto"

	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Batch convert protobuf files",
	Long:  `Convert multiple protobuf files in a directory`,
}

var batchProto2JSONCmd = &cobra.Command{
	Use:   "proto2json",
	Short: "Batch convert protobuf binary to JSON",
	RunE:  runBatchProto2JSON,
}

var batchJSON2ProtoCmd = &cobra.Command{
	Use:   "json2proto",
	Short: "Batch convert JSON to protobuf binary",
	RunE:  runBatchJSON2Proto,
}

func init() {
	rootCmd.AddCommand(batchCmd)
	batchCmd.AddCommand(batchProto2JSONCmd)
	batchCmd.AddCommand(batchJSON2ProtoCmd)

	batchProto2JSONCmd.Flags().StringP("proto", "p", "", "Path to .proto file (required)")
	batchProto2JSONCmd.Flags().StringP("message", "m", "", "Message type name (required)")
	batchProto2JSONCmd.Flags().StringP("input-dir", "i", "", "Input directory containing .pb files (required)")
	batchProto2JSONCmd.Flags().StringP("output-dir", "o", "", "Output directory for .json files (required)")
	batchProto2JSONCmd.Flags().Bool("emit-defaults", false, "Emit default values")
	batchProto2JSONCmd.Flags().BoolP("pretty", "P", true, "Pretty print JSON")

	batchProto2JSONCmd.MarkFlagRequired("proto")
	batchProto2JSONCmd.MarkFlagRequired("message")
	batchProto2JSONCmd.MarkFlagRequired("input-dir")
	batchProto2JSONCmd.MarkFlagRequired("output-dir")

	batchJSON2ProtoCmd.Flags().StringP("proto", "p", "", "Path to .proto file (required)")
	batchJSON2ProtoCmd.Flags().StringP("message", "m", "", "Message type name (required)")
	batchJSON2ProtoCmd.Flags().StringP("input-dir", "i", "", "Input directory containing .json files (required)")
	batchJSON2ProtoCmd.Flags().StringP("output-dir", "o", "", "Output directory for .pb files (required)")

	batchJSON2ProtoCmd.MarkFlagRequired("proto")
	batchJSON2ProtoCmd.MarkFlagRequired("message")
	batchJSON2ProtoCmd.MarkFlagRequired("input-dir")
	batchJSON2ProtoCmd.MarkFlagRequired("output-dir")
}

func runBatchProto2JSON(cmd *cobra.Command, args []string) error {
	protoPath, _ := cmd.Flags().GetString("proto")
	messageType, _ := cmd.Flags().GetString("message")
	inputDir, _ := cmd.Flags().GetString("input-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	emitDefaults, _ := cmd.Flags().GetBool("emit-defaults")
	pretty, _ := cmd.Flags().GetBool("pretty")

	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input directory not found: %s\n", inputDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	protoInfo, err := proto.ParseFile(protoPath)
	if err != nil {
		if pe, ok := err.(*proto.ParseError); ok {
			if pe.Line > 0 {
				fmt.Fprintf(os.Stderr, "Proto syntax error at line %d:%d: %s\n",
					pe.Line, pe.Column, pe.Message)
			} else {
				fmt.Fprintf(os.Stderr, "Proto syntax error: %s\n", pe.Message)
			}
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	successCount := 0
	failCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".pb") &&
			!strings.HasSuffix(strings.ToLower(name), ".bin") {
			continue
		}

		inputPath := filepath.Join(inputDir, name)
		baseName := strings.TrimSuffix(name, filepath.Ext(name))
		outputPath := filepath.Join(outputDir, baseName+".json")

		fmt.Printf("Converting: %s -> %s\n", name, baseName+".json")

		binaryData, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error reading %s: %v\n", name, err)
			failCount++
			continue
		}

		msg, err := protoInfo.NewMessage(messageType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error creating message: %v\n", err)
			failCount++
			continue
		}

		_, parseErr := proto.UnmarshalProtoBinary(binaryData, msg)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "  Error parsing %s: %v\n", name, parseErr)
			failCount++
			continue
		}

		result, err := proto.ProtoToJSON(msg, proto.ConvertOptions{
			EmitDefaultValues: emitDefaults,
			PrettyPrint:       pretty,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error converting %s: %v\n", name, err)
			failCount++
			continue
		}

		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  Warning [%s] %s: %s\n", name, w.Field, w.Message)
		}

		if err := os.WriteFile(outputPath, result.Data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  Error writing %s: %v\n", baseName+".json", err)
			failCount++
			continue
		}

		successCount++
	}

	fmt.Printf("\nDone: %d succeeded, %d failed\n", successCount, failCount)

	if failCount > 0 {
		os.Exit(1)
	}

	return nil
}

func runBatchJSON2Proto(cmd *cobra.Command, args []string) error {
	protoPath, _ := cmd.Flags().GetString("proto")
	messageType, _ := cmd.Flags().GetString("message")
	inputDir, _ := cmd.Flags().GetString("input-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input directory not found: %s\n", inputDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	protoInfo, err := proto.ParseFile(protoPath)
	if err != nil {
		if pe, ok := err.(*proto.ParseError); ok {
			if pe.Line > 0 {
				fmt.Fprintf(os.Stderr, "Proto syntax error at line %d:%d: %s\n",
					pe.Line, pe.Column, pe.Message)
			} else {
				fmt.Fprintf(os.Stderr, "Proto syntax error: %s\n", pe.Message)
			}
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	md, err := protoInfo.GetMessageDescriptor(messageType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	successCount := 0
	failCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}

		inputPath := filepath.Join(inputDir, name)
		baseName := strings.TrimSuffix(name, filepath.Ext(name))
		outputPath := filepath.Join(outputDir, baseName+".pb")

		fmt.Printf("Converting: %s -> %s\n", name, baseName+".pb")

		jsonData, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error reading %s: %v\n", name, err)
			failCount++
			continue
		}

		msg, err := proto.JSONToProto(jsonData, md, proto.ConvertOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error converting %s: %v\n", name, err)
			failCount++
			continue
		}

		binaryData, err := proto.MarshalProtoBinary(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error marshaling %s: %v\n", name, err)
			failCount++
			continue
		}

		if err := os.WriteFile(outputPath, binaryData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  Error writing %s: %v\n", baseName+".pb", err)
			failCount++
			continue
		}

		successCount++
	}

	fmt.Printf("\nDone: %d succeeded, %d failed\n", successCount, failCount)

	if failCount > 0 {
		os.Exit(1)
	}

	return nil
}
