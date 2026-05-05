package protocol

import (
	"encoding/json"
	"io"
)

type MessageType string

const (
	TypeScanRequest  MessageType = "scan_request"
	TypeScanResponse MessageType = "scan_response"
	TypeError        MessageType = "error"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}

type ScanRequest struct {
	Directory    string   `json:"directory"`
	ConfigFile   string   `json:"config_file,omitempty"`
	ExcludePaths []string `json:"exclude_paths,omitempty"`
	Severity     string   `json:"severity,omitempty"`
	OutputJSON   bool     `json:"output_json,omitempty"`
}

type ScanResult struct {
	FilePath     string        `json:"file_path"`
	Matches      []MatchResult `json:"matches"`
}

type MatchResult struct {
	LineNumber    int    `json:"line_number"`
	RuleName      string `json:"rule_name"`
	Severity      string `json:"severity"`
	MaskedContent string `json:"masked_content"`
}

type ScanSummary struct {
	TotalFilesScanned  int `json:"total_files_scanned"`
	TotalFilesMatched  int `json:"total_files_matched"`
	TotalMatchesFound  int `json:"total_matches_found"`
}

type ScanResponse struct {
	Results []ScanResult `json:"results"`
	Summary ScanSummary  `json:"summary"`
	Error   string       `json:"error,omitempty"`
}

func EncodeMessage(w io.Writer, msg *Message) error {
	return json.NewEncoder(w).Encode(msg)
}

func DecodeMessage(r io.Reader) (*Message, error) {
	var msg Message
	if err := json.NewDecoder(r).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func EncodeScanRequest(req *ScanRequest) (string, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DecodeScanRequest(payload string) (*ScanRequest, error) {
	var req ScanRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func EncodeScanResponse(resp *ScanResponse) (string, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DecodeScanResponse(payload string) (*ScanResponse, error) {
	var resp ScanResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
