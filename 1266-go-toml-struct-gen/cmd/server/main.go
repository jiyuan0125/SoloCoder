package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/example/toml2struct/pkg/common"
	"github.com/example/toml2struct/pkg/generator"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务器端口 (默认8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SERVER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/codegen/generate", generateHandler)

	log.Printf("服务器启动，监听端口 %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req common.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求体解析失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.TomlContent == "" {
		sendError(w, "TOML内容不能为空", http.StatusBadRequest)
		return
	}

	gen := generator.NewGenerator(req.PackageName)
	code, err := gen.Generate(req.TomlContent)
	if err != nil {
		sendError(w, "生成失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := common.GenerateResponse{
		Code:    code,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, status int) {
	resp := common.GenerateResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
