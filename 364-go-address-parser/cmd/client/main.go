package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	
	"address-parser/common"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <address>")
		os.Exit(1)
	}
	
	address := strings.Join(os.Args[1:], " ")
	
	reqBody := common.ParseRequest{
		Address: address,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}
	
	resp, err := http.Post("http://localhost:8080/parse", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	var respBody common.ParseResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}
	
	if !respBody.Success {
		fmt.Printf("Error: %s\n", respBody.Message)
		os.Exit(1)
	}
	
	fmt.Println("Address Parse Result:")
	fmt.Println("======================")
	if respBody.Province != "" {
		fmt.Printf("Province: %s\n", respBody.Province)
	}
	if respBody.City != "" {
		fmt.Printf("City: %s\n", respBody.City)
	}
	if respBody.District != "" {
		fmt.Printf("District: %s\n", respBody.District)
	}
	if respBody.Detail != "" {
		fmt.Printf("Detail: %s\n", respBody.Detail)
	}
}
