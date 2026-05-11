package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go-struct-flatten/common"
)

func main() {
	var (
		serverURL  string
		sourceFile string
		structName string
		listMode   bool
		jsonOutput bool
	)

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "flatten server URL")
	flag.StringVar(&sourceFile, "file", "", "Go source file to parse (required)")
	flag.StringVar(&structName, "struct", "", "struct name to flatten")
	flag.BoolVar(&listMode, "list", false, "list all structs in the file")
	flag.BoolVar(&jsonOutput, "json", false, "output in JSON format")
	flag.Parse()

	if sourceFile == "" {
		fmt.Println("Error: -file is required")
		flag.Usage()
		os.Exit(1)
	}

	if !listMode && structName == "" {
		fmt.Println("Error: -struct is required (or use -list)")
		flag.Usage()
		os.Exit(1)
	}

	sourceCode, err := readSourceFile(sourceFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	if listMode {
		err = listStructs(serverURL, sourceCode, jsonOutput)
	} else {
		err = flattenStruct(serverURL, sourceCode, structName, jsonOutput)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func readSourceFile(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func listStructs(serverURL, sourceCode string, jsonOutput bool) error {
	reqBody, err := json.Marshal(common.ListStructsRequest{
		SourceCode: sourceCode,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(strings.TrimRight(serverURL, "/")+"/api/list", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result common.ListStructsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("server returned invalid response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	if jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		if len(result.Structs) == 0 {
			fmt.Println("No structs found")
		} else {
			fmt.Println("Structs found:")
			for _, s := range result.Structs {
				fmt.Printf("  - %s\n", s)
			}
		}
	}

	return nil
}

func flattenStruct(serverURL, sourceCode, structName string, jsonOutput bool) error {
	reqBody, err := json.Marshal(common.FlattenRequest{
		SourceCode: sourceCode,
		StructName: structName,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(strings.TrimRight(serverURL, "/")+"/api/flatten", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result common.FlattenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("server returned invalid response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	if jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		printFields(result.Fields)
	}

	return nil
}

func printFields(fields []common.FieldInfo) {
	if len(fields) == 0 {
		fmt.Println("No fields found")
		return
	}

	for _, f := range fields {
		line := fmt.Sprintf("%s -> %s (%s)", f.FlatKey, f.OriginalPath, f.Type)

		var annotations []string
		if f.IsCircular {
			annotations = append(annotations, "[circular ref]")
		}
		if f.HasMultiplePaths {
			annotations = append(annotations, "[multiple paths]")
		}
		if f.HasOmitempty {
			annotations = append(annotations, "[omitempty]")
		}
		if f.IsArrayIndex {
			annotations = append(annotations, "[array]")
		}
		if f.IsMapKey {
			annotations = append(annotations, "[map]")
		}

		if len(annotations) > 0 {
			line += " " + strings.Join(annotations, " ")
		}

		if len(f.Notes) > 0 {
			line += " " + strings.Join(f.Notes, "; ")
		}

		fmt.Println(line)
	}
}
