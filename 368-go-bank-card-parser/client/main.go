package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: client [options] <card_number>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cardNumber := args[0]

	client := NewClient(*serverURL)

	result, err := client.ParseCard(cardNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 银行卡解析结果 ===")
	fmt.Printf("卡号: %s\n", result.CardNumber)
	fmt.Printf("格式化卡号: %s\n", result.FormattedNumber)
	fmt.Printf("脱敏卡号: %s\n", result.MaskedNumber)
	fmt.Printf("有效性: ")
	if result.IsValid {
		fmt.Println("有效")
	} else {
		fmt.Println("无效")
	}
	if result.Error != "" {
		fmt.Printf("错误信息: %s\n", result.Error)
	}
	if result.IsValid {
		fmt.Printf("发卡银行: %s\n", result.BankName)
		fmt.Printf("卡种: %s\n", result.CardType)
	}
}
