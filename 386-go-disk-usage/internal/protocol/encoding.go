package protocol

import (
	"bufio"
	"encoding/json"
	"io"
)

func EncodeMessage(msg *Message, w io.Writer) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	writer := bufio.NewWriter(w)
	writer.Write(data)
	writer.WriteByte('\n')
	return writer.Flush()
}

func DecodeMessage(r io.Reader) (*Message, error) {
	reader := bufio.NewReader(r)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	
	var msg Message
	err = json.Unmarshal(line, &msg)
	if err != nil {
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
	err := json.Unmarshal([]byte(payload), &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func EncodeScanResult(res *ScanResult) (string, error) {
	data, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DecodeScanResult(payload string) (*ScanResult, error) {
	var res ScanResult
	err := json.Unmarshal([]byte(payload), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func EncodeCompareResult(res *CompareResult) (string, error) {
	data, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DecodeCompareResult(payload string) (*CompareResult, error) {
	var res CompareResult
	err := json.Unmarshal([]byte(payload), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
