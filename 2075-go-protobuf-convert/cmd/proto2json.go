package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"protoconv/internal/proto"
	"protoconv/internal/resource"

	"github.com/spf13/cobra"
)

var proto2jsonCmd = &cobra.Command{
	Use:   "proto2json",
	Short: "Convert protobuf binary to JSON",
	Long:  `Convert protobuf binary data to JSON format using a .proto file definition.`,
	RunE:  runProto2JSON,
}

func init() {
	rootCmd.AddCommand(proto2jsonCmd)

	proto2jsonCmd.Flags().StringP("proto", "p", "", "Path to .proto file (required)")
	proto2jsonCmd.Flags().StringP("message", "m", "", "Message type name (e.g., package.Message)")
	proto2jsonCmd.Flags().StringP("input", "i", "", "Path to input protobuf binary file (or stdin if not specified)")
	proto2jsonCmd.Flags().StringP("output", "o", "", "Path to output JSON file (or stdout if not specified)")
	proto2jsonCmd.Flags().Bool("emit-defaults", false, "Emit default values in JSON output")
	proto2jsonCmd.Flags().BoolP("pretty", "P", true, "Pretty print JSON output")
	proto2jsonCmd.Flags().String("resource-id", "", "Associate operation with a resource ID")
	proto2jsonCmd.Flags().Bool("roundtrip", false, "Validate round-trip conversion")

	proto2jsonCmd.MarkFlagRequired("proto")
	proto2jsonCmd.MarkFlagRequired("message")
}

func runProto2JSON(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	protoPath, _ := cmd.Flags().GetString("proto")
	messageType, _ := cmd.Flags().GetString("message")
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	emitDefaults, _ := cmd.Flags().GetBool("emit-defaults")
	pretty, _ := cmd.Flags().GetBool("pretty")
	resourceID, _ := cmd.Flags().GetString("resource-id")
	validateRoundtrip, _ := cmd.Flags().GetBool("roundtrip")

	storagePath, _ := cmd.Flags().GetString("storage")

	protoInfo, err := proto.ParseFile(protoPath)
	if err != nil {
		if pe, ok := err.(*proto.ParseError); ok {
			if pe.Line > 0 {
				fmt.Fprintf(os.Stderr, "Proto syntax error at line %d:%d: %s\n",
					pe.Line, pe.Column, pe.Message)
			} else {
				fmt.Fprintf(os.Stderr, "Proto syntax error: %s\n", pe.Message)
			}
			recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
		os.Exit(1)
	}

	msg, err := protoInfo.NewMessage(messageType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
		os.Exit(1)
	}

	var binaryData []byte
	if inputPath != "" {
		binaryData, err = os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
			os.Exit(1)
		}
	} else {
		binaryData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
			os.Exit(1)
		}
	}

	parseResult, parseErr := proto.UnmarshalProtoBinary(binaryData, msg)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Parse error at offset %d: %s\n",
			parseResult.ErrorOffset, parseResult.ErrorMessage)
		if len(parseResult.ParsedFields) > 0 {
			partialJSON, _ := json.MarshalIndent(parseResult.ParsedFields, "", "  ")
			fmt.Fprintf(os.Stderr, "Partially parsed fields:\n%s\n", string(partialJSON))
		}
		recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", parseErr.Error())
		os.Exit(1)
	}

	if validateRoundtrip {
		if err := proto.ValidateRoundTrip(binaryData, msg); err != nil {
			fmt.Fprintf(os.Stderr, "Round-trip validation failed: %v\n", err)
			recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
			os.Exit(1)
		}
	}

	result, err := proto.ProtoToJSON(msg, proto.ConvertOptions{
		EmitDefaultValues: emitDefaults,
		PrettyPrint:       pretty,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
		os.Exit(1)
	}

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s - %s\n", w.Field, w.Message)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, result.Data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			recordOperation(resourceID, storagePath, "proto2json", startTime, "failed", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println(string(result.Data))
	}

	recordOperation(resourceID, storagePath, "proto2json", startTime, "success", "")
	return nil
}

func recordOperation(resourceID, storagePath, opType string, startTime time.Time, status, errorMsg string) {
	if resourceID == "" {
		return
	}

	rm, err := resource.NewManager(storagePath)
	if err != nil {
		return
	}

	rm.RecordOperation(resource.Operation{
		ResourceID:   resourceID,
		Type:         opType,
		StartTime:    startTime,
		EndTime:      time.Now(),
		Status:       status,
		ErrorMessage: errorMsg,
	})
}
