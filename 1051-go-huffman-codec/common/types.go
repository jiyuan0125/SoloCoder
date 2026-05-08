package common

import (
	"encoding/base64"
	"huffman-codec/huffman"
)

type EncodeRequest struct {
	Text string `json:"text"`
}

type EncodeResponse struct {
	Data        string           `json:"data"`
	PaddingBits int              `json:"padding_bits"`
	TreeData    string           `json:"tree_data"`
	OriginalSize int             `json:"original_size"`
	EncodedSize  int             `json:"encoded_size"`
	Frequencies map[string]int   `json:"frequencies"`
}

type DecodeRequest struct {
	Data        string `json:"data"`
	PaddingBits int    `json:"padding_bits"`
	TreeData    string `json:"tree_data"`
}

type DecodeResponse struct {
	Text         string `json:"text"`
	OriginalSize int    `json:"original_size"`
	EncodedSize  int    `json:"encoded_size"`
}

type FrequencyResponse struct {
	Frequencies map[string]int `json:"frequencies"`
}

func FrequenciesToJSON(freq huffman.FreqMap) map[string]int {
	result := make(map[string]int, len(freq))
	for ch, count := range freq {
		result[string(ch)] = count
	}
	return result
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64Decode(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
