package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gofmt-tool/api"
)

var (
	serverAddr   string
	writeFlag    bool
	lineWidth    int
	removeEmpty  bool
	moduleName   string
	dryRun       bool
)

func main() {
	flag.StringVar(&serverAddr, "server", "http://localhost:8080", "Format server address")
	flag.BoolVar(&writeFlag, "w", false, "Write formatted code back to files")
	flag.IntVar(&lineWidth, "line-width", 120, "Maximum line width")
	flag.BoolVar(&removeEmpty, "remove-empty", true, "Remove excessive empty lines")
	flag.StringVar(&moduleName, "module", "", "Module name for internal package detection")
	flag.BoolVar(&dryRun, "dry-run", true, "Only show diff without modifying files (default)")
	flag.Parse()

	if writeFlag {
		dryRun = false
	}

	args := flag.Args()
	if len(args) == 0 {
		formatStdin()
		return
	}

	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		if info.IsDir() {
			formatDirectory(path)
		} else {
			formatFile(path)
		}
	}
}

func formatStdin() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	source := string(data)
	result, _, err := formatSource(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func formatDirectory(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			formatFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
	}
}

func formatFile(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
		return
	}

	original := string(data)
	formatted, stats, err := formatSource(original)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", path, err)
		return
	}

	if formatted == original {
		fmt.Printf("Unchanged: %s\n", path)
		return
	}

	diff := generateUnifiedDiff(original, formatted, path)
	if diff != "" {
		fmt.Print(diff)
		fmt.Printf("\nStats for %s: %d lines modified, %d imports moved, %d lines split\n",
			path, stats.LinesModified, stats.ImportsMoved, stats.LinesSplit)
	}

	if !dryRun && writeFlag {
		if err := ioutil.WriteFile(path, []byte(formatted), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", path, err)
		} else {
			fmt.Printf("Wrote: %s\n", path)
		}
	}
}

func formatSource(source string) (string, *api.Stats, error) {
	req := api.FormatRequest{
		Source: source,
		Config: api.Config{
			LineWidth:        lineWidth,
			RemoveEmptyLines: removeEmpty,
			ModuleName:       moduleName,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", nil, err
	}

	resp, err := http.Post(serverAddr+"/format", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("server error: %s", string(errBody))
	}

	var result api.FormatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, err
	}

	return result.Source, &result.Stats, nil
}

func generateUnifiedDiff(original, formatted, filename string) string {
	origLines := strings.Split(original, "\n")
	formLines := strings.Split(formatted, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- a/%s\n", filename))
	diff.WriteString(fmt.Sprintf("+++ b/%s\n", filename))

	origIdx := 0
	formIdx := 0
	hunkStarted := false
	hunkStart := 0
	hunkOrigCount := 0
	hunkFormCount := 0

	for origIdx < len(origLines) || formIdx < len(formLines) {
		origLine := ""
		if origIdx < len(origLines) {
			origLine = origLines[origIdx]
		}
		formLine := ""
		if formIdx < len(formLines) {
			formLine = formLines[formIdx]
		}

		if origLine == formLine {
			if hunkStarted {
				diff.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
					hunkStart+1, hunkOrigCount, hunkStart+1, hunkFormCount))
				hunkStarted = false
			}
			origIdx++
			formIdx++
			continue
		}

		if !hunkStarted {
			hunkStart = origIdx
			hunkOrigCount = 0
			hunkFormCount = 0
			hunkStarted = true
		}

		origSame := false
		for i := origIdx + 1; i < len(origLines) && i < origIdx+20; i++ {
			if origLines[i] == formLine {
				origSame = true
				break
			}
		}

		formSame := false
		for j := formIdx + 1; j < len(formLines) && j < formIdx+20; j++ {
			if formLines[j] == origLine {
				formSame = true
				break
			}
		}

		if origIdx < len(origLines) && (formIdx >= len(formLines) || !formSame) {
			diff.WriteString("-")
			diff.WriteString(origLine)
			diff.WriteString("\n")
			origIdx++
			hunkOrigCount++
		}

		if formIdx < len(formLines) && (origIdx >= len(origLines) || !origSame) {
			diff.WriteString("+")
			diff.WriteString(formLine)
			diff.WriteString("\n")
			formIdx++
			hunkFormCount++
		}
	}

	if hunkStarted {
		diff.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			hunkStart+1, hunkOrigCount, hunkStart+1, hunkFormCount))
	}

	return diff.String()
}
