package main

import (
	"github.com/bencode/json-converter/pkg/bencode"
)

func decodeBencode(data []byte) (bencode.Value, error) {
	return bencode.Decode(data)
}

func encodeBencode(value bencode.Value) ([]byte, error) {
	return bencode.Encode(value)
}

func toJSON(value bencode.Value) ([]byte, error) {
	return bencode.ToJSON(value)
}

func fromJSON(data []byte) (bencode.Value, error) {
	return bencode.FromJSON(data)
}
