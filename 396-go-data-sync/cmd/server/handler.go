package main

import (
	"fmt"
	"log"
	"time"

	"go-data-sync/pkg/csvutil"
	"go-data-sync/pkg/protocol"
)

func handleSync(req *protocol.Request) (*protocol.Response, error) {
	if req.SourceFile == "" {
		return &protocol.Response{
			Success: false,
			Message: "Source file is required",
		}, nil
	}
	
	if req.TargetFile == "" {
		return &protocol.Response{
			Success: false,
			Message: "Target file is required",
		}, nil
	}
	
	log.Printf("Syncing: %s -> %s (key: %s, apply: %v)", 
		req.SourceFile, req.TargetFile, req.KeyColumn, req.Apply)
	
	result, err := csvutil.SyncCSV(req.SourceFile, req.TargetFile, req.KeyColumn, req.Apply)
	if err != nil {
		return nil, err
	}
	
	for _, w := range result.Warnings {
		log.Printf("Warning: %s", w)
	}
	
	changeSummary := &protocol.ChangeSummary{
		Added:    result.CountByType(csvutil.ChangeAdd),
		Modified: result.CountByType(csvutil.ChangeModify),
		Deleted:  result.CountByType(csvutil.ChangeDelete),
		Total:    result.TotalChanges(),
		ChangeFile: result.ChangeFile,
	}
	
	history := protocol.SyncHistory{
		ID:         generateID(),
		Timestamp:  time.Now().Format(time.RFC3339),
		SourceFile: req.SourceFile,
		TargetFile: req.TargetFile,
		Changes:    *changeSummary,
		Applied:    result.Applied,
	}
	
	addHistory(history)
	
	message := fmt.Sprintf("Sync completed. Changes: %d added, %d modified, %d deleted",
		changeSummary.Added, changeSummary.Modified, changeSummary.Deleted)
	
	if result.Applied {
		message += " (applied)"
	} else {
		message += " (not applied)"
	}
	
	if result.ChangeFile != "" {
		message += fmt.Sprintf("\nChange file: %s", result.ChangeFile)
	}
	
	return &protocol.Response{
		Success: true,
		Message: message,
		Changes: changeSummary,
	}, nil
}

func handleStatus(req *protocol.Request) (*protocol.Response, error) {
	return &protocol.Response{
		Success: true,
		Message: "Server is running",
	}, nil
}

func handleHistory(req *protocol.Request) (*protocol.Response, error) {
	history := getHistory()
	return &protocol.Response{
		Success: true,
		Message: fmt.Sprintf("Found %d history records", len(history)),
		History: history,
	}, nil
}

func generateID() string {
	return time.Now().Format("20060102150405")
}
