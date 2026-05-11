package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/drivingschool/core"
)

var store = core.NewStore()

func main() {
	var port string
	flag.StringVar(&port, "port", "8908", "服务监听端口")
	flag.Parse()

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		port = envPort
	}

	mux := http.NewServeMux()

	setupStudentRoutes(mux)
	setupCoachRoutes(mux)
	setupExamRoutes(mux)

	fmt.Printf("驾校管理系统服务端启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
