package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"smtp-client/common"
)

var (
	serverURL string
)

func main() {
	flag.StringVar(&serverURL, "server", getEnvOrDefault("SERVER_URL", "http://localhost:8080"), "Server URL")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== SMTP Client CLI ===")
	fmt.Println()

	req := common.SendEmailRequest{}

	fmt.Print("From address: ")
	req.From = readLine(reader)

	fmt.Print("To addresses (comma separated): ")
	toStr := readLine(reader)
	req.To = splitCommaSeparated(toStr)

	fmt.Print("Cc addresses (comma separated, optional): ")
	ccStr := readLine(reader)
	req.Cc = splitCommaSeparated(ccStr)

	fmt.Print("Subject: ")
	req.Subject = readLine(reader)

	fmt.Print("Use HTML body? (y/n): ")
	useHTML := strings.ToLower(readLine(reader)) == "y"

	if useHTML {
		fmt.Println("Enter HTML body (end with a single line containing '.'):")
		req.HTMLBody = readMultiline(reader)
	}

	fmt.Print("Use plain text body? (y/n): ")
	useText := strings.ToLower(readLine(reader)) == "y"

	if useText {
		fmt.Println("Enter plain text body (end with a single line containing '.'):")
		req.TextBody = readMultiline(reader)
	}

	fmt.Print("Add attachments? (y/n): ")
	addAttachments := strings.ToLower(readLine(reader)) == "y"

	if addAttachments {
		for {
			fmt.Print("Attachment file path (or press Enter to finish): ")
			path := strings.TrimSpace(readLine(reader))
			if path == "" {
				break
			}

			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				continue
			}

			att := common.Attachment{
				Filename: filepath.Base(path),
				Data:     data,
			}
			req.Attachments = append(req.Attachments, att)
			fmt.Printf("Added attachment: %s\n", att.Filename)
		}
	}

	fmt.Println()
	fmt.Println("Sending email...")

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/emails", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result common.SendEmailResponse
	json.Unmarshal(body, &result)

	if result.Success {
		fmt.Printf("Success! Email queued with ID: %s\n", result.ID)
	} else {
		fmt.Printf("Failed: %s\n", result.Message)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func readMultiline(reader *bufio.Reader) string {
	var lines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line == "." {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
