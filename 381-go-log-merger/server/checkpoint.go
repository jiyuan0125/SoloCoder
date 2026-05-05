package main

import (
	"encoding/json"
	"os"
)

type FileCheckpoint struct {
	Filename string `json:"filename"`
	Position int64  `json:"position"`
}

type Checkpoint struct {
	TimeFormat string          `json:"time_format"`
	Files      []FileCheckpoint `json:"files"`
}

func SaveCheckpoint(filename string, ckpt *Checkpoint) error {
	data, err := json.MarshalIndent(ckpt, "", "  ")
	if err != nil {
		return err
	}

	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempFile, filename)
}

func LoadCheckpoint(filename string) (*Checkpoint, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var ckpt Checkpoint
	if err := json.Unmarshal(data, &ckpt); err != nil {
		return nil, err
	}

	return &ckpt, nil
}
