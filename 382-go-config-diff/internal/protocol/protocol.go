package protocol

import "encoding/json"

type MessageType uint8

const (
    TypeRequestCompare MessageType = iota
    TypeResponseCompare
    TypeError
)

type DiffType uint8

const (
    DiffAdded DiffType = iota
    DiffRemoved
    DiffModified
)

type CompareRequest struct {
    File1Path   string   `json:"file1_path"`
    File2Path   string   `json:"file2_path"`
    IgnoreKeys  []string `json:"ignore_keys,omitempty"`
}

type DiffResult struct {
    Type     DiffType `json:"type"`
    Key      string   `json:"key"`
    OldValue string   `json:"old_value,omitempty"`
    NewValue string   `json:"new_value,omitempty"`
}

type CompareResponse struct {
    Success    bool         `json:"success"`
    Identical  bool         `json:"identical,omitempty"`
    Diffs      []DiffResult `json:"diffs,omitempty"`
    Error      string       `json:"error,omitempty"`
}

type Message struct {
    Type    MessageType `json:"type"`
    Payload string      `json:"payload"`
}

func EncodeRequest(req CompareRequest) (Message, error) {
    payload, err := json.Marshal(req)
    if err != nil {
        return Message{}, err
    }
    return Message{
        Type:    TypeRequestCompare,
        Payload: string(payload),
    }, nil
}

func DecodeRequest(msg Message) (CompareRequest, error) {
    var req CompareRequest
    err := json.Unmarshal([]byte(msg.Payload), &req)
    return req, err
}

func EncodeResponse(resp CompareResponse) (Message, error) {
    payload, err := json.Marshal(resp)
    if err != nil {
        return Message{}, err
    }
    return Message{
        Type:    TypeResponseCompare,
        Payload: string(payload),
    }, nil
}

func DecodeResponse(msg Message) (CompareResponse, error) {
    var resp CompareResponse
    err := json.Unmarshal([]byte(msg.Payload), &resp)
    return resp, err
}

func EncodeError(errMsg string) (Message, error) {
    resp := CompareResponse{
        Success: false,
        Error:   errMsg,
    }
    return EncodeResponse(resp)
}
