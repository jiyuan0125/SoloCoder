package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"mask-engine/pkg/api"
)

func handleMask(args []string) {
	fs := flag.NewFlagSet("mask", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "Server URL")
	phonePrefix := fs.Int("phone-prefix", -1, "Phone prefix keep")
	phoneSuffix := fs.Int("phone-suffix", -1, "Phone suffix keep")
	idcardPrefix := fs.Int("idcard-prefix", -1, "ID card prefix keep")
	idcardSuffix := fs.Int("idcard-suffix", -1, "ID card suffix keep")
	bankcardPrefix := fs.Int("bankcard-prefix", -1, "Bank card prefix keep")
	bankcardSuffix := fs.Int("bankcard-suffix", -1, "Bank card suffix keep")
	emailPrefix := fs.Int("email-prefix", -1, "Email prefix keep")
	namePrefix := fs.Int("name-prefix", -1, "Name prefix keep")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Println("Error: text is required for mask command")
		os.Exit(1)
		return
	}
	text := fs.Arg(0)

	req := api.MaskRequest{Text: text}

	hasOverride := false
	overrides := &api.Rules{}
	if *phonePrefix >= 0 || *phoneSuffix >= 0 {
		hasOverride = true
		overrides.Phone = &api.MaskRule{}
		if *phonePrefix >= 0 {
			overrides.Phone.PrefixKeep = *phonePrefix
		}
		if *phoneSuffix >= 0 {
			overrides.Phone.SuffixKeep = *phoneSuffix
		}
	}
	if *idcardPrefix >= 0 || *idcardSuffix >= 0 {
		hasOverride = true
		overrides.IDCard = &api.MaskRule{}
		if *idcardPrefix >= 0 {
			overrides.IDCard.PrefixKeep = *idcardPrefix
		}
		if *idcardSuffix >= 0 {
			overrides.IDCard.SuffixKeep = *idcardSuffix
		}
	}
	if *bankcardPrefix >= 0 || *bankcardSuffix >= 0 {
		hasOverride = true
		overrides.BankCard = &api.MaskRule{}
		if *bankcardPrefix >= 0 {
			overrides.BankCard.PrefixKeep = *bankcardPrefix
		}
		if *bankcardSuffix >= 0 {
			overrides.BankCard.SuffixKeep = *bankcardSuffix
		}
	}
	if *emailPrefix >= 0 {
		hasOverride = true
		overrides.Email = &api.MaskRule{PrefixKeep: *emailPrefix}
	}
	if *namePrefix >= 0 {
		hasOverride = true
		overrides.Name = &api.MaskRule{PrefixKeep: *namePrefix}
	}

	if hasOverride {
		req.OverrideRules = overrides
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	respBody, err := httpPost(*serverURL+"/mask", body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	var resp api.MaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Original:", resp.Original)
	fmt.Println("Masked:  ", resp.Masked)
}
