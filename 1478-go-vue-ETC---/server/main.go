package main

import (
	"flag"
	"os"
	"strconv"
)

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("ETC_PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "服务端监听端口 (默认 8080)")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	if _, err := strconv.Atoi(port); err != nil {
		port = "8080"
	}

	return ":" + port
}

func main() {
	port := getPort()
	server := NewServer()
	server.Start(port)
}
