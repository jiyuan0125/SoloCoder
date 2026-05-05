package protocol

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MessageTypeRequest  MessageType = "request"
	MessageTypeResponse MessageType = "response"
	MessageTypeError    MessageType = "error"
)

type RequestType string

const (
	RequestTypeGenerateReport RequestType = "generate_report"
	RequestTypeListHistory    RequestType = "list_history"
	RequestTypeGetReport      RequestType = "get_report"
)

type Request struct {
	Type       RequestType     `json:"type"`
	Generate   *GenerateRequest `json:"generate,omitempty"`
	ListHistory *ListHistoryRequest `json:"list_history,omitempty"`
	GetReport  *GetReportRequest `json:"get_report,omitempty"`
}

type GenerateRequest struct {
	RepoPath string `json:"repo_path"`
	Since    string `json:"since"`
	Author   string `json:"author,omitempty"`
	Format   string `json:"format,omitempty"`
}

type ListHistoryRequest struct {
	Limit int `json:"limit,omitempty"`
}

type GetReportRequest struct {
	ReportID string `json:"report_id"`
}

type Response struct {
	Type    MessageType `json:"type"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Report  *Report     `json:"report,omitempty"`
	History []HistoryEntry `json:"history,omitempty"`
}

type HistoryEntry struct {
	ID        string    `json:"id"`
	RepoPath  string    `json:"repo_path"`
	GeneratedAt time.Time `json:"generated_at"`
	CommitCount int     `json:"commit_count"`
}

type Report struct {
	ID           string                `json:"id"`
	RepoPath     string                `json:"repo_path"`
	Since        string                `json:"since"`
	Author       string                `json:"author,omitempty"`
	GeneratedAt  time.Time             `json:"generated_at"`
	Content      string                `json:"content"`
	ContentMarkdown string              `json:"content_markdown,omitempty"`
	Stats        ReportStats           `json:"stats"`
	DailyCommits map[string][]*Commit  `json:"daily_commits"`
}

type ReportStats struct {
	TotalCommits    int `json:"total_commits"`
	TotalFiles      int `json:"total_files"`
	DaysWithCommits int `json:"days_with_commits"`
}

type Commit struct {
	Hash        string    `json:"hash"`
	ShortHash   string    `json:"short_hash"`
	Message     string    `json:"message"`
	ShortMessage string   `json:"short_message"`
	Author      string    `json:"author"`
	Date        time.Time `json:"date"`
	DateStr     string    `json:"date_str"`
	Issues      []string  `json:"issues,omitempty"`
	Files       []string  `json:"files,omitempty"`
	IsMerge     bool      `json:"is_merge"`
}

func SerializeRequest(req *Request) ([]byte, error) {
	return json.Marshal(req)
}

func DeserializeRequest(data []byte) (*Request, error) {
	var req Request
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func SerializeResponse(resp *Response) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeResponse(data []byte) (*Response, error) {
	var resp Response
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
