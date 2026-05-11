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

	"github.com/example/errcode-gen/pkg/codegen"
	"github.com/example/errcode-gen/pkg/common"
)

func main() {
	var yamlPath string
	var serverURL string
	var outputDir string
	var packageName string

	flag.StringVar(&yamlPath, "file", "", "path to YAML definition file (required)")
	flag.StringVar(&yamlPath, "f", "", "path to YAML definition file (short)")
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "server URL")
	flag.StringVar(&serverURL, "s", "http://localhost:8080", "server URL (short)")
	flag.StringVar(&outputDir, "output", "", "output directory (defaults to YAML file directory)")
	flag.StringVar(&outputDir, "o", "", "output directory (short)")
	flag.StringVar(&packageName, "package", "", "package name (defaults to inferred from YAML directory)")
	flag.Parse()

	if yamlPath == "" {
		fmt.Println("Error: YAML file path is required (use -file or -f)")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		os.Exit(1)
	}

	if packageName == "" {
		packageName, err = codegen.InferPackageName(yamlPath)
		if err != nil {
			fmt.Printf("Error inferring package name: %v\n", err)
			os.Exit(1)
		}
	}

	req := common.GenerateRequest{
		YAMLContent: string(yamlData),
		PackageName: packageName,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error encoding request: %v\n", err)
		os.Exit(1)
	}

	generateURL := strings.TrimRight(serverURL, "/") + "/generate"
	resp, err := http.Post(generateURL, "application/json", bytes.NewReader(reqData))
	if err != nil {
		fmt.Printf("Error calling server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var apiResp common.GenerateResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(body))
		os.Exit(1)
	}

	if !apiResp.Success {
		fmt.Printf("Error generating code: %s\n", apiResp.Error)
		os.Exit(1)
	}

	if len(apiResp.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range apiResp.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	outDir := outputDir
	if outDir == "" {
		absPath, _ := filepath.Abs(yamlPath)
		outDir = filepath.Dir(absPath)
	}

	outPath := filepath.Join(outDir, "errors_generated.go")
	if err := os.WriteFile(outPath, []byte(apiResp.Code), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated: %s\n", outPath)
}
