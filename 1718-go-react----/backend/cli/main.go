package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const BaseURL = "http://localhost:8080"

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func makeRequest(method, url string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, BaseURL+url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func cmdCreatePath(args []string) {
	fs := flag.NewFlagSet("create-path", flag.ExitOnError)
	code := fs.String("code", "", "Path code")
	name := fs.String("name", "", "Path name")
	icd := fs.String("icd", "", "ICD codes (comma separated)")
	minDays := fs.Int("min-days", 0, "Minimum days")
	maxDays := fs.Int("max-days", 0, "Maximum days")
	minCost := fs.Float64("min-cost", 0, "Minimum cost")
	maxCost := fs.Float64("max-cost", 0, "Maximum cost")
	fs.Parse(args)

	if *code == "" || *name == "" || *icd == "" || *minDays == 0 || *maxDays == 0 {
		fmt.Println("Usage: create-path --code CODE --name NAME --icd ICD_CODES --min-days MIN --max-days MAX [--min-cost MIN --max-cost MAX]")
		os.Exit(1)
	}

	path := map[string]interface{}{
		"code":     *code,
		"name":     *name,
		"icd_codes": *icd,
		"min_days": *minDays,
		"max_days": *maxDays,
		"min_cost": *minCost,
		"max_cost": *maxCost,
	}

	respData, err := makeRequest("POST", "/api/r", path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp APIResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		fmt.Println(string(respData))
		os.Exit(1)
	}

	if resp.Success {
		data, _ := json.MarshalIndent(resp.Data, "", "  ")
		fmt.Println("Path created successfully:")
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}
}

func cmdEnrollPatient(args []string) {
	fs := flag.NewFlagSet("enroll-patient", flag.ExitOnError)
	patientID := fs.Int64("patient-id", 0, "Patient ID")
	pathID := fs.Int64("path-id", 0, "Path ID")
	fs.Parse(args)

	if *patientID == 0 || *pathID == 0 {
		fmt.Println("Usage: enroll-patient --patient-id ID --path-id ID")
		os.Exit(1)
	}

	req := map[string]interface{}{
		"patient_id": *patientID,
		"path_id":    *pathID,
	}

	respData, err := makeRequest("POST", "/api/enrollments", req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp APIResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		fmt.Println(string(respData))
		os.Exit(1)
	}

	if resp.Success {
		data, _ := json.MarshalIndent(resp.Data, "", "  ")
		fmt.Println("Patient enrolled successfully:")
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}
}

func cmdRecordVariant(args []string) {
	fs := flag.NewFlagSet("record-variant", flag.ExitOnError)
	enrollmentID := fs.Int64("enrollment-id", 0, "Enrollment ID")
	date := fs.String("date", time.Now().Format("2006-01-02"), "Date (YYYY-MM-DD)")
	content := fs.String("content", "", "Variation content")
	reason := fs.String("reason", "", "Variation reason")
	variationType := fs.String("type", "可控变异", "Variation type (可控变异 or 不可控变异)")
	action := fs.String("action", "", "Action taken")
	fs.Parse(args)

	if *enrollmentID == 0 || *content == "" || *reason == "" || *action == "" {
		fmt.Println("Usage: record-variant --enrollment-id ID --content CONTENT --reason REASON --action ACTION [--date DATE --type TYPE]")
		os.Exit(1)
	}

	req := map[string]interface{}{
		"date":    *date,
		"content": *content,
		"reason":  *reason,
		"type":    *variationType,
		"action":  *action,
	}

	respData, err := makeRequest("POST", fmt.Sprintf("/api/enrollments/%d/variations", *enrollmentID), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp APIResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		fmt.Println(string(respData))
		os.Exit(1)
	}

	if resp.Success {
		data, _ := json.MarshalIndent(resp.Data, "", "  ")
		fmt.Println("Variation recorded successfully:")
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}
}

func cmdShowQuality(args []string) {
	fs := flag.NewFlagSet("show-quality", flag.ExitOnError)
	year := fs.Int("year", time.Now().Year(), "Year")
	month := fs.Int("month", int(time.Now().Month()), "Month (1-12)")
	fs.Parse(args)

	url := fmt.Sprintf("/api/quality?year=%d&month=%d", *year, *month)
	respData, err := makeRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var resp APIResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		fmt.Println(string(respData))
		os.Exit(1)
	}

	if resp.Success {
		data, _ := json.MarshalIndent(resp.Data, "", "  ")
		fmt.Printf("Quality metrics for %04d-%02d:\n", *year, *month)
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Clinical Path Management System CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  cpath <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-path     Create a new clinical path")
	fmt.Println("  enroll-patient  Enroll a patient into a path")
	fmt.Println("  record-variant  Record a variation for an enrollment")
	fmt.Println("  show-quality    Show quality metrics")
	fmt.Println()
	fmt.Println("Use 'cpath <command> -help' for command-specific options")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch cmd {
	case "create-path":
		cmdCreatePath(args)
	case "enroll-patient":
		cmdEnrollPatient(args)
	case "record-variant":
		cmdRecordVariant(args)
	case "show-quality":
		cmdShowQuality(args)
	case "help", "-help", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
