package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

const defaultBaseURL = "http://localhost:8300"

type Client struct {
	BaseURL string
}

func NewClient() *Client {
	baseURL := os.Getenv("API_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{BaseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) AddQuestion(questionNum, content, qType string, kpID uint, difficulty int, correctAnswer, options, explanation string, score float64) error {
	body := map[string]interface{}{
		"question_number":    questionNum,
		"content":          content,
		"type":             qType,
		"knowledge_point_id": kpID,
		"difficulty":       difficulty,
		"correct_answer":   correctAnswer,
		"options":          options,
		"explanation":      explanation,
		"score":            score,
	}

	resp, err := c.doRequest("POST", "/api/questions", body)
	if err != nil {
		return err
	}

	fmt.Println("Question added successfully:")
	fmt.Println(string(resp))
	return nil
}

func (c *Client) StartTest(studentID string, billID *uint) error {
	body := map[string]interface{}{
		"student_id": studentID,
	}
	if billID != nil {
		body["bill_id"] = *billID
	}

	resp, err := c.doRequest("POST", "/api/exams/start", body)
	if err != nil {
		return err
	}

	fmt.Println("Test started successfully:")
	fmt.Println(string(resp))
	return nil
}

func (c *Client) SubmitAnswer(examID uint, answer string, timeSpent int) error {
	body := map[string]interface{}{
		"student_answer": answer,
		"time_spent":    timeSpent,
	}

	resp, err := c.doRequest("POST", fmt.Sprintf("/api/exams/%d/submit", examID), body)
	if err != nil {
		return err
	}

	fmt.Println("Answer submitted successfully:")
	fmt.Println(string(resp))
	return nil
}

func (c *Client) ExportRecords(studentID string) error {
	resp, err := c.doRequest("GET", "/api/reports/student?student_id="+studentID, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	fileName := fmt.Sprintf("student_records_%s.json", studentID)
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(fileName, output, 0644); err != nil {
		return err
	}

	fmt.Printf("Records exported to %s\n", fileName)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient()
	command := os.Args[1]

	switch command {
	case "add-question":
		handleAddQuestion(client, os.Args[2:])
	case "start-test":
		handleStartTest(client, os.Args[2:])
	case "submit-answer":
		handleSubmitAnswer(client, os.Args[2:])
	case "export-records":
		handleExportRecords(client, os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleAddQuestion(client *Client, args []string) {
	fs := flag.NewFlagSet("add-question", flag.ExitOnError)
	questionNum := fs.String("number", "", "Question number (required)")
	content := fs.String("content", "", "Question content (required)")
	qType := fs.String("type", "single_choice", "Question type: single_choice, multiple_choice, true_false, fill_blank, essay")
	kpID := fs.Uint("kp-id", 0, "Knowledge point ID (required)")
	difficulty := fs.Int("difficulty", 3, "Difficulty level (1-5)")
	correctAnswer := fs.String("answer", "", "Correct answer (required)")
	options := fs.String("options", "", "Question options (JSON format for choice questions)")
	explanation := fs.String("explanation", "", "Answer explanation")
	score := fs.Float64("score", 1.0, "Question score")

	fs.Parse(args)

	if *questionNum == "" || *content == "" || *kpID == 0 || *correctAnswer == "" {
		fmt.Println("Error: --number, --content, --kp-id, and --answer are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if *difficulty < 1 || *difficulty > 5 {
		fmt.Println("Error: difficulty must be between 1 and 5")
		os.Exit(1)
	}

	if err := client.AddQuestion(*questionNum, *content, *qType, *kpID, *difficulty, *correctAnswer, *options, *explanation, *score); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStartTest(client *Client, args []string) {
	fs := flag.NewFlagSet("start-test", flag.ExitOnError)
	studentID := fs.String("student-id", "", "Student ID (required)")
	billID := fs.Uint("bill-id", 0, "Bill ID (optional)")

	fs.Parse(args)

	if *studentID == "" {
		fmt.Println("Error: --student-id is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	var billPtr *uint
	if *billID > 0 {
		billPtr = billID
	}

	if err := client.StartTest(*studentID, billPtr); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleSubmitAnswer(client *Client, args []string) {
	fs := flag.NewFlagSet("submit-answer", flag.ExitOnError)
	examID := fs.Uint("exam-id", 0, "Exam ID (required)")
	answer := fs.String("answer", "", "Student answer (required)")
	timeSpent := fs.Int("time-spent", 0, "Time spent on the question (seconds)")

	fs.Parse(args)

	if *examID == 0 || *answer == "" {
		fmt.Println("Error: --exam-id and --answer are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if err := client.SubmitAnswer(*examID, *answer, *timeSpent); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleExportRecords(client *Client, args []string) {
	fs := flag.NewFlagSet("export-records", flag.ExitOnError)
	studentID := fs.String("student-id", "", "Student ID (required)")

	fs.Parse(args)

	if *studentID == "" {
		fmt.Println("Error: --student-id is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if err := client.ExportRecords(*studentID); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Smart Exam CLI Client")
	fmt.Println("")
	fmt.Println("Usage: client <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  add-question    Add a new question to the question bank")
	fmt.Println("  start-test      Start a new adaptive test")
	fmt.Println("  submit-answer   Submit an answer for the current question")
	fmt.Println("  export-records  Export student's exam records")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  API_URL         Base URL of the API server (default: http://localhost:8300)")
}
