package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"mask-engine/pkg/api"
)

func handleRules(args []string) {
	fs := flag.NewFlagSet("rules", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8102", "Server URL")
	update := fs.Bool("update", false, "Update rules")
	phonePrefix := fs.Int("phone-prefix", -1, "Phone prefix keep")
	phoneSuffix := fs.Int("phone-suffix", -1, "Phone suffix keep")
	idcardPrefix := fs.Int("idcard-prefix", -1, "ID card prefix keep")
	idcardSuffix := fs.Int("idcard-suffix", -1, "ID card suffix keep")
	bankcardPrefix := fs.Int("bankcard-prefix", -1, "Bank card prefix keep")
	bankcardSuffix := fs.Int("bankcard-suffix", -1, "Bank card suffix keep")
	emailPrefix := fs.Int("email-prefix", -1, "Email prefix keep")
	namePrefix := fs.Int("name-prefix", -1, "Name prefix keep")
	fs.Parse(args)

	if !*update {
		respBody, err := httpGet(*serverURL + "/mask/rules")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
			return
		}
		var resp api.RulesResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			os.Exit(1)
			return
		}
		data, _ := json.MarshalIndent(resp.Rules, "", "  ")
		fmt.Println(string(data))
		return
	}

	req := api.UpdateRulesRequest{}
	hasUpdate := false
	if *phonePrefix >= 0 || *phoneSuffix >= 0 {
		hasUpdate = true
		req.Phone = &api.MaskRule{}
		if *phonePrefix >= 0 {
			req.Phone.PrefixKeep = *phonePrefix
		}
		if *phoneSuffix >= 0 {
			req.Phone.SuffixKeep = *phoneSuffix
		}
	}
	if *idcardPrefix >= 0 || *idcardSuffix >= 0 {
		hasUpdate = true
		req.IDCard = &api.MaskRule{}
		if *idcardPrefix >= 0 {
			req.IDCard.PrefixKeep = *idcardPrefix
		}
		if *idcardSuffix >= 0 {
			req.IDCard.SuffixKeep = *idcardSuffix
		}
	}
	if *bankcardPrefix >= 0 || *bankcardSuffix >= 0 {
		hasUpdate = true
		req.BankCard = &api.MaskRule{}
		if *bankcardPrefix >= 0 {
			req.BankCard.PrefixKeep = *bankcardPrefix
		}
		if *bankcardSuffix >= 0 {
			req.BankCard.SuffixKeep = *bankcardSuffix
		}
	}
	if *emailPrefix >= 0 {
		hasUpdate = true
		req.Email = &api.MaskRule{PrefixKeep: *emailPrefix}
	}
	if *namePrefix >= 0 {
		hasUpdate = true
		req.Name = &api.MaskRule{PrefixKeep: *namePrefix}
	}

	if !hasUpdate {
		fmt.Println("Error: at least one rule parameter is required for update")
		os.Exit(1)
		return
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	respBody, err := httpPut(*serverURL+"/mask/rules", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	var resp api.UpdateRulesResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
		return
	}

	if resp.Success {
		fmt.Println("Rules updated successfully")
	} else {
		fmt.Println("Failed to update rules")
		os.Exit(1)
	}
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, body []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPut(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
