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

	"wire-di-gen/internal/api"
	"wire-di-gen/internal/wiregen"
)

const (
	defaultServerURL = "http://localhost:8080"
	envServerURL     = "WIREGEN_SERVER_URL"
)

func main() {
	var pkgPath string
	var serverURL string
	var useRemote bool

	flag.StringVar(&pkgPath, "pkg", ".", "package path to scan")
	flag.StringVar(&serverURL, "server", "", "server URL (default: http://localhost:8080 or WIREGEN_SERVER_URL)")
	flag.BoolVar(&useRemote, "remote", false, "use remote server instead of local generation")
	flag.Parse()

	if flag.NArg() > 0 {
		pkgPath = flag.Arg(0)
	}

	if serverURL == "" {
		serverURL = os.Getenv(envServerURL)
		if serverURL == "" {
			serverURL = defaultServerURL
		}
	}

	var err error
	if useRemote {
		err = generateRemote(pkgPath, serverURL)
	} else {
		err = generateLocal(pkgPath)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func generateLocal(pkgPath string) error {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Generating wire_gen.go for package: %s\n", absPath)

	err = wiregen.GenerateAndWrite(absPath)
	if err != nil {
		if wiregen.IsCircularDependencyError(err) {
			path := wiregen.GetCircularDependencyPath(err)
			fmt.Fprintf(os.Stderr, "circular dependency detected:\n  %s\n", strings.Join(path, " -> "))
		}
		return err
	}

	outputPath := filepath.Join(absPath, "wire_gen.go")
	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

func generateRemote(pkgPath string, serverURL string) error {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Generating wire_gen.go using remote server: %s\n", serverURL)
	fmt.Printf("Package: %s\n", absPath)

	files, err := collectGoFiles(absPath)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	req := api.GenerateRequest{
		Files: files,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(respBody))
	}

	var genResp api.GenerateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !genResp.Success {
		if genResp.CyclePath != nil && len(genResp.CyclePath) > 0 {
			fmt.Fprintf(os.Stderr, "circular dependency detected:\n  %s\n", strings.Join(genResp.CyclePath, " -> "))
		}
		return fmt.Errorf("generation failed: %s", genResp.Error)
	}

	outputPath := filepath.Join(absPath, "wire_gen.go")
	if err := os.WriteFile(outputPath, []byte(genResp.Code), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

func collectGoFiles(pkgPath string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_gen.go") {
			return nil
		}

		relPath, err := filepath.Rel(pkgPath, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files[relPath] = string(content)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
