package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var baseURL = "http://localhost:8300"

func init() {
	if url := os.Getenv("API_URL"); url != "" {
		baseURL = url
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "submit-paper":
		handleSubmitPaper(args)
	case "assign-reviewer":
		handleAssignReviewer(args)
	case "schedule-paper":
		handleSchedulePaper(args)
	case "export-papers":
		handleExportPapers(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Conference Management System CLI

Usage:
  confman <command> [options]

Commands:
  submit-paper        Submit a paper
  assign-reviewer     Assign a reviewer to a paper
  schedule-paper      Schedule a paper for presentation
  export-papers       Export all papers for a meeting

Examples:
  confman submit-paper --meeting 1 --title "Paper Title" --abstract "..." --authors "John Doe,john@example.com,true,true"
  confman assign-reviewer --paper 1 --reviewer 2 --meeting 1
  confman schedule-paper --paper 1 --meeting 1 --day 1 --start "2026-06-01T09:00:00" --end "2026-06-01T09:20:00"
  confman export-papers --meeting 1 --output papers.json
`)
}

func httpRequest(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func handleSubmitPaper(args []string) {
	fs := flag.NewFlagSet("submit-paper", flag.ExitOnError)
	meetingID := fs.Uint("meeting", 0, "Meeting ID")
	title := fs.String("title", "", "Paper title")
	abstract := fs.String("abstract", "", "Abstract (200-500 words)")
	keywords := fs.String("keywords", "", "Keywords (comma-separated)")
	topic := fs.String("topic", "", "Topic area")
	content := fs.String("content", "", "Paper content")
	authors := fs.String("authors", "", "Authors: name,email,isFirst,isCorresponding;...")
	fs.Parse(args)

	if *meetingID == 0 || *title == "" || *abstract == "" || *authors == "" {
		fmt.Println("Error: --meeting, --title, --abstract, and --authors are required")
		fs.Usage()
		os.Exit(1)
	}

	authorList := parseAuthors(*authors)
	if len(authorList) == 0 {
		fmt.Println("Error: Failed to parse authors")
		os.Exit(1)
	}

	paper := map[string]interface{}{
		"meeting_id": *meetingID,
		"title":      *title,
		"abstract":   *abstract,
		"keywords":   *keywords,
		"topic_area": *topic,
		"content":    *content,
		"authors":    authorList,
	}

	resp, err := httpRequest("POST", "/api/papers", paper)
	if err != nil {
		fmt.Printf("Error creating paper: %v\n", err)
		os.Exit(1)
	}

	var created map[string]interface{}
	if err := json.Unmarshal(resp, &created); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	paperID := uint(created["id"].(float64))

	resp, err = httpRequest("POST", fmt.Sprintf("/api/papers/%d/submit", paperID), nil)
	if err != nil {
		fmt.Printf("Error submitting paper: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Paper submitted successfully!\n")
	fmt.Printf("Paper Number: %s\n", created["paper_number"])
	fmt.Printf("Paper ID: %d\n", paperID)
}

func parseAuthors(authorsStr string) []map[string]interface{} {
	parts := strings.Split(authorsStr, ";")
	result := make([]map[string]interface{}, 0)

	for i, part := range parts {
		fields := strings.Split(part, ",")
		if len(fields) < 2 {
			continue
		}

		author := map[string]interface{}{
			"name":          strings.TrimSpace(fields[0]),
			"email":         strings.TrimSpace(fields[1]),
			"is_first_author": false,
			"is_corresponding": false,
			"order":         i,
		}

		if len(fields) > 2 {
			if b, err := strconv.ParseBool(strings.TrimSpace(fields[2])); err == nil {
				author["is_first_author"] = b
			}
		}
		if len(fields) > 3 {
			if b, err := strconv.ParseBool(strings.TrimSpace(fields[3])); err == nil {
				author["is_corresponding"] = b
			}
		}

		result = append(result, author)
	}

	return result
}

func handleAssignReviewer(args []string) {
	fs := flag.NewFlagSet("assign-reviewer", flag.ExitOnError)
	paperID := fs.Uint("paper", 0, "Paper ID")
	reviewerID := fs.Uint("reviewer", 0, "Reviewer ID")
	meetingID := fs.Uint("meeting", 0, "Meeting ID")
	fs.Parse(args)

	if *paperID == 0 || *reviewerID == 0 || *meetingID == 0 {
		fmt.Println("Error: --paper, --reviewer, and --meeting are required")
		fs.Usage()
		os.Exit(1)
	}

	req := map[string]interface{}{
		"paper_id":    *paperID,
		"reviewer_id": *reviewerID,
		"meeting_id":  *meetingID,
	}

	_, err := httpRequest("POST", "/api/reviews/assign", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Reviewer assigned successfully!")
}

func handleSchedulePaper(args []string) {
	fs := flag.NewFlagSet("schedule-paper", flag.ExitOnError)
	meetingID := fs.Uint("meeting", 0, "Meeting ID")
	paperID := fs.Uint("paper", 0, "Paper ID")
	day := fs.Int("day", 0, "Day (1-3)")
	session := fs.String("session", "", "Session name")
	start := fs.String("start", "", "Start time (ISO format)")
	end := fs.String("end", "", "End time (ISO format)")
	fs.Parse(args)

	if *meetingID == 0 || *paperID == 0 || *day == 0 || *start == "" || *end == "" {
		fmt.Println("Error: --meeting, --paper, --day, --start, and --end are required")
		fs.Usage()
		os.Exit(1)
	}

	req := map[string]interface{}{
		"meeting_id": *meetingID,
		"paper_id":   *paperID,
		"day":        *day,
		"session":    *session,
		"start_time": *start,
		"end_time":   *end,
	}

	resp, err := httpRequest("POST", "/api/schedules", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Paper scheduled successfully!")
	fmt.Println(string(resp))
}

func handleExportPapers(args []string) {
	fs := flag.NewFlagSet("export-papers", flag.ExitOnError)
	meetingID := fs.Uint("meeting", 0, "Meeting ID")
	output := fs.String("output", "", "Output file path")
	fs.Parse(args)

	if *meetingID == 0 {
		fmt.Println("Error: --meeting is required")
		fs.Usage()
		os.Exit(1)
	}

	resp, err := httpRequest("GET", fmt.Sprintf("/api/meetings/%d/papers/export", *meetingID), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, resp, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported to %s\n", *output)
	} else {
		fmt.Println(string(resp))
	}
}
