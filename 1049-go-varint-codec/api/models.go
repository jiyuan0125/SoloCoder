package api

import "varint-codec/codec"

type EncodeRequest struct {
	Numbers  []int64          `json:"numbers"`
	Mode     codec.EncodeMode `json:"mode"`
	NumberType codec.NumberType `json:"number_type"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Error   string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Data       string           `json:"data"`
	Mode       codec.EncodeMode `json:"mode"`
	NumberType codec.NumberType `json:"number_type"`
}

type DecodeResponse struct {
	Success bool    `json:"success"`
	Numbers []int64 `json:"numbers"`
	Error   string  `json:"error,omitempty"`
}

type ValidateRequest struct {
	Data       string           `json:"data"`
	Mode       codec.EncodeMode `json:"mode"`
	NumberType codec.NumberType `json:"number_type"`
}

type ValidateResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type BenchmarkRequest struct {
	Size int `json:"size"`
}

type BenchmarkResponse struct {
	Success       bool    `json:"success"`
	Size          int     `json:"size"`
	EncodeTimeNs  int64   `json:"encode_time_ns"`
	DecodeTimeNs  int64   `json:"decode_time_ns"`
	EncodePerSec  float64 `json:"encode_per_sec"`
	DecodePerSec  float64 `json:"decode_per_sec"`
	Error         string  `json:"error,omitempty"`
}
