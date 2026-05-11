package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"mask-engine/pkg/api"
)

func handleMaskType(args []string) {
	fs := flag.NewFlagSet("mask-type", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8102", "Server URL")
	dataType := fs.String("type", "", "Data type: phone, idcard, bankcard, email, name")
	prefixKeep := fs.Int("prefix", -1, "Prefix keep (optional override)")
	suffixKeep := fs.Int("suffix", -1, "Suffix keep (optional override)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: value is required for mask-type command")
		os.Exit(1)
		return
	}
	value := fs.Arg(0)

	if *dataType == "" {
		fmt.Println("Error: type is required. Valid types: phone, idcard, bankcard, email, name")
		os.Exit(1)
		return
	}

	req := api.MaskByTypeRequest{
		Value: value,
		Type:  *dataType,
	}

	if *prefixKeep >= 0 || *suffixKeep >= 0 {
		req.OverrideRule = &api.MaskRule{}
		if *prefixKeep >= 0 {
			req.OverrideRule.PrefixKeep = *prefixKeep
		}
		if *suffixKeep >= 0 {
			req.OverrideRule.SuffixKeep = *suffixKeep
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	respBody, err := httpPost(*serverURL+"/mask/type", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	var resp api.MaskByTypeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Type:    ", resp.Type)
	fmt.Println("Original:", resp.Original)
	fmt.Println("Masked:  ", resp.Masked)
}
