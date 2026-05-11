package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gomock-generator/pkg/api"
)

func main() {
	var filePath string
	var ifaceName string
	var serverURL string
	var listOnly bool

	flag.StringVar(&filePath, "file", "", "path to Go source file")
	flag.StringVar(&ifaceName, "interface", "", "interface name to generate mock for")
	flag.StringVar(&serverURL, "server", "", "mock server URL (e.g. http://localhost:8501)")
	flag.BoolVar(&listOnly, "list", false, "list all interfaces in file without generating")
	flag.Parse()

	if serverURL == "" {
		serverURL = os.Getenv("GOMOCK_SERVER_URL")
	}
	if serverURL == "" {
		log.Fatal("server URL must be provided via -server flag or GOMOCK_SERVER_URL env variable")
	}

	if filePath == "" {
		log.Fatal("file path must be provided via -file flag")
	}

	source, err := readFile(filePath)
	if err != nil {
		log.Fatalf("read file: %v", err)
	}

	if listOnly {
		ifaces, err := listInterfaces(serverURL, source)
		if err != nil {
			log.Fatalf("list interfaces: %v", err)
		}
		fmt.Println("Found interfaces:")
		for _, name := range ifaces {
			fmt.Println("  -", name)
		}
		return
	}

	if ifaceName == "" {
		log.Fatal("interface name must be provided via -interface flag")
	}

	mockCode, err := generateMock(serverURL, source, ifaceName)
	if err != nil {
		log.Fatalf("generate mock: %v", err)
	}

	outputPath := filepath.Join(".", "mock_"+ifaceName+".go")
	if err := writeFile(outputPath, mockCode); err != nil {
		log.Fatalf("write output: %v", err)
	}

	fmt.Printf("Mock generated successfully: %s\n", outputPath)
}

func readFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeFile(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0644)
}

func listInterfaces(serverURL, source string) ([]string, error) {
	url := strings.TrimRight(serverURL, "/") + "/parse"
	resp, err := http.Post(url, "text/plain", bytes.NewBufferString(source))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.ParseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Err != "" {
		return nil, fmt.Errorf(result.Err)
	}

	return result.Interfaces, nil
}

func generateMock(serverURL, source, ifaceName string) (string, error) {
	req := api.GenerateRequest{
		Source:    source,
		Interface: ifaceName,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(serverURL, "/") + "/generate"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result api.GenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Err != "" {
		return "", fmt.Errorf(result.Err)
	}

	return result.Code, nil
}
