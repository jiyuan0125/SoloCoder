package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"simhash/pkg/common"
)

const serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "fingerprint":
		handleFingerprint()
	case "hamming":
		handleHamming()
	case "similar":
		handleSimilar()
	case "dedup":
		handleDedup()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client fingerprint <text> [-ngram N]")
	fmt.Println("  client hamming <fp1> <fp2>")
	fmt.Println("  client similar <fp1> <fp2> [-threshold N]")
	fmt.Println("  client dedup <file> [-threshold N] [-ngram N]")
}

func handleFingerprint() {
	args := os.Args[2:]
	if len(args) < 1 {
		fmt.Println("usage: client fingerprint <text> [-ngram N]")
		return
	}
	text := args[0]
	var ngram int
	if len(args) >= 3 && args[1] == "-ngram" {
		fmt.Sscanf(args[2], "%d", &ngram)
	}
	req := common.FingerprintRequest{Text: text, NGram: ngram}
	var resp common.FingerprintResponse
	if err := postJSON(serverURL+"/fingerprint", req, &resp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Fingerprint)
}

func handleHamming() {
	if len(os.Args) < 4 {
		fmt.Println("usage: client hamming <fp1> <fp2>")
		return
	}
	fp1 := os.Args[2]
	fp2 := os.Args[3]
	req := common.HammingDistanceRequest{FP1: fp1, FP2: fp2}
	var resp common.HammingDistanceResponse
	if err := postJSON(serverURL+"/hamming", req, &resp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Distance)
}

func handleSimilar() {
	args := os.Args[2:]
	if len(args) < 2 {
		fmt.Println("usage: client similar <fp1> <fp2> [-threshold N]")
		return
	}
	fp1 := args[0]
	fp2 := args[1]
	var threshold int
	if len(args) >= 4 && args[2] == "-threshold" {
		fmt.Sscanf(args[3], "%d", &threshold)
	}
	req := common.IsSimilarRequest{FP1: fp1, FP2: fp2, Threshold: threshold}
	var resp common.IsSimilarResponse
	if err := postJSON(serverURL+"/similar", req, &resp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.Similar)
}

func handleDedup() {
	args := os.Args[2:]
	if len(args) < 1 {
		fmt.Println("usage: client dedup <file> [-threshold N] [-ngram N]")
		return
	}
	filename := args[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	lines := strings.Split(string(content), "\n")
	var docs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			docs = append(docs, line)
		}
	}
	var threshold, ngram int
	for i := 1; i < len(args); i++ {
		if args[i] == "-threshold" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &threshold)
			i++
		}
		if args[i] == "-ngram" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &ngram)
			i++
		}
	}
	req := common.DedupRequest{Docs: docs, Threshold: threshold, NGram: ngram}
	var resp common.DedupResponse
	if err := postJSON(serverURL+"/dedup", req, &resp); err != nil {
		fmt.Println(err)
		return
	}
	for _, doc := range resp.Docs {
		fmt.Println(doc)
	}
}

func postJSON(url string, req, resp interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	r, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	if r.StatusCode >= 400 {
		var errResp common.ErrorResponse
		json.Unmarshal(body, &errResp)
		return fmt.Errorf("%s", errResp.Error)
	}
	return json.Unmarshal(body, resp)
}
