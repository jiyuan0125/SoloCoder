package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/example/mimedetect/api"
	"github.com/example/mimedetect/mimedetector"
)

func main() {
	port := getPort()

	http.HandleFunc("/detect", handleDetect)

	log.Printf("MIME检测服务启动，监听端口 %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func getPort() string {
	port := flag.String("port", "", "服务监听端口")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("MIMEDETECT_PORT"); envPort != "" {
		return envPort
	}

	return "8080"
}

func handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.DetectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求格式错误", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.FilePath == "" {
		sendError(w, "缺少file_path参数", http.StatusBadRequest)
		return
	}

	result, err := mimedetector.DetectFile(req.FilePath)
	if err != nil {
		sendError(w, fmt.Sprintf("文件检测失败: %v", err), http.StatusInternalServerError)
		return
	}

	resp := api.DetectResponse{
		Success:       true,
		FilePath:      req.FilePath,
		ExtensionMIME: result.ExtensionMIME,
		MagicMIME:     result.MagicMIME,
		FinalMIME:     result.FinalMIME,
		Warning:       result.Warning,
		Confidence:    result.Confidence,
		IsConsistent:  result.IsConsistent(),
		HasWarning:    result.HasWarning(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.DetectResponse{
		Success: false,
		Error:   message,
	})
}
