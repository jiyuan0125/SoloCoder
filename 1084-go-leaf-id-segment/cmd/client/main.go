package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"leaf-segment/pkg/common"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Leaf segment server URL")
	action := flag.String("action", "allocate", "Action: allocate|batch|status")
	count := flag.Int("count", 10, "Batch count (for batch action)")
	timeout := flag.Int("timeout", 10, "Request timeout in seconds")

	flag.Usage = func() {
		fmt.Println("Leaf Segment ID Generator Client")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  client -action allocate                   # Allocate one ID")
		fmt.Println("  client -action batch -count 10            # Allocate 10 IDs")
		fmt.Println("  client -action status                     # Get status")
		fmt.Println()
		flag.PrintDefaults()
	}

	flag.Parse()

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	baseURL := strings.TrimRight(*serverURL, "/")

	switch strings.ToLower(*action) {
	case "allocate":
		doAllocate(client, baseURL)
	case "batch":
		if *count <= 0 {
			fmt.Println("Error: count must be positive")
			return
		}
		doBatch(client, baseURL, *count)
	case "status":
		doStatus(client, baseURL)
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		flag.Usage()
	}
}

func doAllocate(client *http.Client, baseURL string) {
	url := baseURL + "/id/allocate"
	resp, err := client.Post(url, "application/json", strings.NewReader("{}"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		_ = json.Unmarshal(body, &errResp)
		if errResp.Error != "" {
			fmt.Println("Error:", errResp.Error)
		} else {
			fmt.Println("Error status:", resp.StatusCode)
		}
		return
	}
	var result common.AllocateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Parse error:", err)
		return
	}
	fmt.Println("Allocated ID:", result.ID)
}

func doBatch(client *http.Client, baseURL string, count int) {
	url := baseURL + "/id/batch"
	reqBody, _ := json.Marshal(common.AllocateBatchRequest{Count: count})
	resp, err := client.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		_ = json.Unmarshal(body, &errResp)
		if errResp.Error != "" {
			fmt.Println("Error:", errResp.Error)
		} else {
			fmt.Println("Error status:", resp.StatusCode)
		}
		return
	}
	var result common.AllocateBatchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Parse error:", err)
		return
	}
	fmt.Printf("Allocated %d IDs:\n", len(result.IDs))
	for i, id := range result.IDs {
		fmt.Printf("  [%d] %s\n", i+1, id)
	}
}

func doStatus(client *http.Client, baseURL string) {
	url := baseURL + "/id/status"
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error status:", resp.StatusCode)
		return
	}
	var status common.StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		fmt.Println("Parse error:", err)
		return
	}
	fmt.Println("ID Allocator Status:")
	fmt.Println("====================")
	if status.CurrentSegment != nil {
		cs := status.CurrentSegment
		fmt.Printf("Current Segment:  [%d, %d]\n", cs.Start, cs.End)
		fmt.Printf("  Current Value:  %d\n", cs.Current)
		fmt.Printf("  Allocated:      %d\n", cs.Allocated)
		fmt.Printf("  Remaining:      %d\n", cs.Remaining)
	} else {
		fmt.Println("Current Segment:  <none>")
	}
	fmt.Printf("Has Preloaded:    %v\n", status.HasPreloadedSegment)
	fmt.Printf("Is Preloading:    %v\n", status.IsPreloading)
	if status.PreloadedSegment != nil {
		ps := status.PreloadedSegment
		fmt.Printf("Preloaded Segment: [%d, %d]\n", ps.Start, ps.End)
	}
}
