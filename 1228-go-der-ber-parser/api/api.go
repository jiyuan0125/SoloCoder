package api

import "der-ber-parser/asn1"

type EncodeRequest struct {
	Mode string        `json:"mode"`
	Data asn1.JSONTLV  `json:"data"`
}

type EncodeResponse struct {
	Success bool   `json:"success"`
	Hex     string `json:"hex,omitempty"`
	Error   string `json:"error,omitempty"`
}

type DecodeRequest struct {
	Hex string `json:"hex"`
}

type DecodeResponse struct {
	Success bool           `json:"success"`
	Data    []asn1.JSONTLV `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}
