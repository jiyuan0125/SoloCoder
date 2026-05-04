package main

import (
	"flag"
	"fmt"
	"os"

	"realestate/pkg/client"
)

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "服务端地址")
	flag.Parse()

	c := client.NewClient(*serverAddr)
	ui := client.NewUI(c)

	if err := ui.Start(); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
