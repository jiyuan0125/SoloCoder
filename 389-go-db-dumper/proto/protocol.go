package proto

import (
	"encoding/json"
	"io"
)

type MessageType string

const (
	MessageTypeRequest  MessageType = "request"
	MessageTypeResponse MessageType = "response"
)

type RequestType string

const (
	RequestTypeConvert RequestType = "convert"
	RequestTypeStatus  RequestType = "status"
	RequestTypeStop    RequestType = "stop"
)

type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusError   ResponseStatus = "error"
)

type ConvertRequest struct {
	InputFile  string `json:"input_file"`
	OutputFile string `json:"output_file"`
	TableName  string `json:"table_name,omitempty"`
}

type ConvertResponse struct {
	Status      ResponseStatus `json:"status"`
	Message     string         `json:"message,omitempty"`
	RecordsRead int            `json:"records_read,omitempty"`
	SQLCount    int            `json:"sql_count,omitempty"`
}

type Message struct {
	Type    MessageType     `json:"type"`
	Request *ConvertRequest `json:"request,omitempty"`
	Response *ConvertResponse `json:"response,omitempty"`
}

func ReadMessage(reader io.Reader) (*Message, error) {
	var msg Message
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func WriteMessage(writer io.Writer, msg *Message) error {
	encoder := json.NewEncoder(writer)
	return encoder.Encode(msg)
}
