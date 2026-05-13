package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"protoconv/internal/proto"

	"github.com/spf13/cobra"
)

var json2protoCmd = &cobra.Command{
	Use:   "json2proto",
	Short: "Convert JSON to protobuf binary",
	Long:  `Convert JSON data to protobuf binary format using a .proto file definition.`,
	RunE:  runJSON2Proto,
}

func init() {
	rootCmd.AddCommand(json2protoCmd)

	json2protoCmd.Flags().StringP("proto", "p", "", "Path to .proto file (required)")
	json2protoCmd.Flags().StringP("message", "m", "", "Message type name (e.g., package.Message)")
	json2protoCmd.Flags().StringP("input", "i", "", "Path to input JSON file (or stdin if not specified)")
	json2protoCmd.Flags().StringP("output", "o", "", "Path to output protobuf binary file (or stdout if not specified)")
	json2protoCmd.Flags().String("resource-id", "", "Associate operation with a resource ID")

	json2protoCmd.MarkFlagRequired("proto")
	json2protoCmd.MarkFlagRequired("message")
}

func runJSON2Proto(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	protoPath, _ := cmd.Flags().GetString("proto")
	messageType, _ := cmd.Flags().GetString("message")
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	resourceID, _ := cmd.Flags().GetString("resource-id")
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
			recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
		os.Exit(1)
	}

	md, err := protoInfo.GetMessageDescriptor(messageType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
		os.Exit(1)
	}

	var jsonData []byte
	if inputPath != "" {
		jsonData, err = os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
			os.Exit(1)
		}
	} else {
		jsonData, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
			os.Exit(1)
		}
	}

	msg, err := proto.JSONToProto(jsonData, md, proto.ConvertOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
		os.Exit(1)
	}

	binaryData, err := proto.MarshalProtoBinary(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to binary: %v\n", err)
		recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
		os.Exit(1)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, binaryData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			recordOperation(resourceID, storagePath, "json2proto", startTime, "failed", err.Error())
			os.Exit(1)
		}
	} else {
		os.Stdout.Write(binaryData)
	}

	recordOperation(resourceID, storagePath, "json2proto", startTime, "success", "")
	return nil
}
