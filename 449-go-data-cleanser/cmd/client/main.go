package main

import (
	"bytes"
	"datacleanser/internal/common"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
)

const defaultServerURL = "http://localhost:8080"

var serverURL string

func main() {
	flag.StringVar(&serverURL, "server", defaultServerURL, "server URL (e.g., http://localhost:8080)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	var err error

	switch cmd {
	case "clean":
		err = handleClean(args[1:])
	case "template":
		err = handleTemplate(args[1:])
	case "task":
		err = handleTask(args[1:])
	case "errors":
		err = handleErrors(args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Data Cleanser CLI

Usage:
  dc [flags] <command> [arguments]

Flags:
  -server string    Server URL (default "http://localhost:8080")

Commands:
  clean      Clean data records
  template   Manage cleaning templates
  task       Check async task status
  errors     View error samples
  help       Show this help message

Use "dc <command> -help" for more information about a command.`)
}

func handleClean(args []string) error {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	var async bool
	var templateName string
	var ruleFile string
	var outputFile string

	fs.BoolVar(&async, "async", false, "run as async task")
	fs.StringVar(&templateName, "template", "", "use existing template by name")
	fs.StringVar(&ruleFile, "rule", "", "rule configuration file (JSON)")
	fs.StringVar(&outputFile, "output", "", "output file for cleaned data")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("input data file is required")
	}
	inputFile := remaining[0]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to parse input JSON: %w", err)
	}

	if len(records) > common.MaxBatchSize {
		return fmt.Errorf("batch size exceeds limit: %d (max %d)", len(records), common.MaxBatchSize)
	}

	var rule *common.CleanRule
	if ruleFile != "" {
		ruleData, err := os.ReadFile(ruleFile)
		if err != nil {
			return fmt.Errorf("failed to read rule file: %w", err)
		}
		rule = &common.CleanRule{}
		if err := json.Unmarshal(ruleData, rule); err != nil {
			return fmt.Errorf("failed to parse rule JSON: %w", err)
		}
	}

	req := common.CleanRequest{
		Records:      records,
		Rule:         rule,
		TemplateName: templateName,
		Async:        async,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/api/v1/clean", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var cleanResp common.CleanResponse
	if err := json.Unmarshal(respBody, &cleanResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !cleanResp.Success {
		return fmt.Errorf("clean failed: %s", cleanResp.Message)
	}

	if async {
		fmt.Printf("Async task created. Task ID: %s\n", cleanResp.TaskID)
		fmt.Printf("Use 'dc task %s' to check status.\n", cleanResp.TaskID)
		return nil
	}

	printCleanResult(cleanResp.Data)

	if outputFile != "" {
		output := map[string]interface{}{
			"cleaned_records":   cleanResp.Data.CleanedRecords,
			"duplicate_records": cleanResp.Data.DuplicateRecords,
			"invalid_records":   cleanResp.Data.InvalidRecords,
			"error_records":     cleanResp.Data.ErrorRecords,
			"report":            cleanResp.Data.Report,
		}
		outputData, _ := json.MarshalIndent(output, "", "  ")
		if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("\nResults written to: %s\n", outputFile)
	}

	return nil
}

func printCleanResult(result *common.CleanResult) {
	fmt.Println("\n=== Clean Result ===")
	fmt.Printf("Cleaned Records:   %d\n", len(result.CleanedRecords))
	fmt.Printf("Duplicate Records: %d\n", len(result.DuplicateRecords))
	fmt.Printf("Invalid Records:   %d\n", len(result.InvalidRecords))
	fmt.Printf("Error Records:     %d\n", len(result.ErrorRecords))

	fmt.Println("\n=== Clean Report ===")
	report := result.Report
	fmt.Printf("Total Records:     %d\n", report.TotalRecords)
	fmt.Printf("Duplicate Count:   %d\n", report.DuplicateCount)
	fmt.Printf("Corrected Count:   %d\n", report.CorrectedCount)
	fmt.Printf("Invalid Count:     %d\n", report.InvalidCount)
	fmt.Printf("Quality Score:     %.2f\n", report.QualityScore)

	fmt.Println("\nRule Hits:")
	for rule, count := range report.RuleHits {
		fmt.Printf("  %s: %d\n", rule, count)
	}
}

func handleTemplate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subcommand required (create, list, get, update, rollback)")
	}

	subCmd := args[0]
	switch subCmd {
	case "create":
		return handleTemplateCreate(args[1:])
	case "list":
		return handleTemplateList(args[1:])
	case "get":
		return handleTemplateGet(args[1:])
	case "update":
		return handleTemplateUpdate(args[1:])
	case "rollback":
		return handleTemplateRollback(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subCmd)
	}
}

func handleTemplateCreate(args []string) error {
	fs := flag.NewFlagSet("template create", flag.ExitOnError)
	var name string
	var ruleFile string

	fs.StringVar(&name, "name", "", "template name (required)")
	fs.StringVar(&ruleFile, "rule", "", "rule configuration file (JSON, required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("template name is required (-name)")
	}
	if ruleFile == "" {
		return fmt.Errorf("rule file is required (-rule)")
	}

	ruleData, err := os.ReadFile(ruleFile)
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	var rule common.CleanRule
	if err := json.Unmarshal(ruleData, &rule); err != nil {
		return fmt.Errorf("failed to parse rule JSON: %w", err)
	}

	req := common.CreateTemplateRequest{
		Name: name,
		Rule: rule,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/api/v1/templates", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var tplResp common.TemplateResponse
	if err := json.Unmarshal(respBody, &tplResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Println("Template created successfully!")
	printTemplate(tplResp.Data)
	return nil
}

func handleTemplateList(args []string) error {
	resp, err := http.Get(serverURL + "/api/v1/templates")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var listResp common.ListTemplatesResponse
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(listResp.Data) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tCREATED AT")
	for _, t := range listResp.Data {
		fmt.Fprintf(w, "%s\t%d\t%s\n", t.Name, t.Version, t.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	w.Flush()

	return nil
}

func handleTemplateGet(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("template name is required")
	}
	name := args[0]

	resp, err := http.Get(serverURL + "/api/v1/templates/" + name)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var tplResp common.TemplateResponse
	if err := json.Unmarshal(respBody, &tplResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	printTemplate(tplResp.Data)
	return nil
}

func printTemplate(t *common.Template) {
	if t == nil {
		return
	}
	fmt.Printf("Name:        %s\n", t.Name)
	fmt.Printf("Version:     %d\n", t.Version)
	fmt.Printf("Created At:  %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At:  %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Rule:\n")
	ruleJSON, _ := json.MarshalIndent(t.Rule, "  ", "  ")
	fmt.Println("  " + string(ruleJSON))

	if len(t.Versions) > 0 {
		fmt.Printf("\nVersion History:\n")
		for _, v := range t.Versions {
			fmt.Printf("  Version %d (created: %s)\n", v.Version, v.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
}

func handleTemplateUpdate(args []string) error {
	fs := flag.NewFlagSet("template update", flag.ExitOnError)
	var name string
	var ruleFile string

	fs.StringVar(&name, "name", "", "template name (required)")
	fs.StringVar(&ruleFile, "rule", "", "rule configuration file (JSON, required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("template name is required (-name)")
	}
	if ruleFile == "" {
		return fmt.Errorf("rule file is required (-rule)")
	}

	ruleData, err := os.ReadFile(ruleFile)
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	var rule common.CleanRule
	if err := json.Unmarshal(ruleData, &rule); err != nil {
		return fmt.Errorf("failed to parse rule JSON: %w", err)
	}

	req := common.UpdateTemplateRequest{
		Name: name,
		Rule: rule,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPut, serverURL+"/api/v1/templates/"+name, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var tplResp common.TemplateResponse
	if err := json.Unmarshal(respBody, &tplResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Println("Template updated successfully!")
	printTemplate(tplResp.Data)
	return nil
}

func handleTemplateRollback(args []string) error {
	fs := flag.NewFlagSet("template rollback", flag.ExitOnError)
	var name string
	var version int

	fs.StringVar(&name, "name", "", "template name (required)")
	fs.IntVar(&version, "version", 0, "target version number (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if name == "" {
		return fmt.Errorf("template name is required (-name)")
	}
	if version == 0 {
		return fmt.Errorf("version is required (-version)")
	}

	req := map[string]int{"version": version}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverURL+"/api/v1/templates/"+name+"/rollback", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var tplResp common.TemplateResponse
	if err := json.Unmarshal(respBody, &tplResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Println("Template rolled back successfully!")
	printTemplate(tplResp.Data)
	return nil
}

func handleTask(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}
	taskID := args[0]

	resp, err := http.Get(serverURL + "/api/v1/tasks/" + taskID)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var taskResp common.TaskResponse
	if err := json.Unmarshal(respBody, &taskResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	task := taskResp.Data
	fmt.Println("\n=== Task Status ===")
	fmt.Printf("Task ID:    %s\n", task.ID)
	fmt.Printf("Status:     %s\n", task.Status)
	fmt.Printf("Created At: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))

	if task.Error != "" {
		fmt.Printf("\nError: %s\n", task.Error)
	}

	if task.Status == common.TaskCompleted && task.Result != nil {
		fmt.Println("\n=== Task Result ===")
		printCleanResult(task.Result)
	}

	return nil
}

func handleErrors(args []string) error {
	resp, err := http.Get(serverURL + "/api/v1/errors")
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool                     `json:"success"`
		Data    []common.ErrorSample `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Data) == 0 {
		fmt.Println("No error samples found.")
		return nil
	}

	fmt.Printf("\n=== Error Samples (%d total) ===\n", len(result.Data))
	for i, sample := range result.Data {
		fmt.Printf("\n--- Sample %d ---\n", i+1)
		fmt.Printf("Batch ID:  %s\n", sample.BatchID)
		fmt.Printf("Timestamp: %s\n", sample.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Error:     %s\n", sample.Error)
		recordJSON, _ := json.MarshalIndent(sample.Record, "  ", "  ")
		fmt.Printf("Record:\n  %s\n", string(recordJSON))
	}

	return nil
}
