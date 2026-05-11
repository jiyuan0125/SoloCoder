package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go-enum-generator/pkg/api"
	"go-enum-generator/pkg/enum"
)

func getPort() string {
	port := flag.String("port", "", "监听端口 (例如: 8080)")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("ENUM_SERVER_PORT"); envPort != "" {
		return envPort
	}

	return "8080"
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.GenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		return
	}

	if req.SourceCode == "" {
		http.Error(w, "source_code不能为空", http.StatusBadRequest)
		return
	}

	filename := req.Filename
	if filename == "" {
		filename = "source.go"
	}

	enums, err := enum.Parse(strings.NewReader(req.SourceCode), filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析源文件失败: %v", err), http.StatusBadRequest)
		return
	}

	if len(enums) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.GenerateResponse{
			Message: "没有找到需要生成的枚举",
		})
		return
	}

	pkgName := "main"
	if idx := strings.LastIndex(filename, "/"); idx >= 0 {
		filename = filename[idx+1:]
	}

	generated, err := enum.Generate(pkgName, enums)
	if err != nil {
		http.Error(w, fmt.Sprintf("生成代码失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.GenerateResponse{
		GeneratedCode: string(generated),
	})
}

func main() {
	port := getPort()

	http.HandleFunc("/api/generate", handleGenerate)

	addr := ":" + port
	fmt.Printf("枚举生成服务启动，监听端口: %s\n", port)
	fmt.Printf("POST /api/generate - 生成枚举代码\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
