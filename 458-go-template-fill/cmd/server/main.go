package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"template-engine/api"
	"template-engine/pkg/template"
)

const (
	DefaultPort = 8080
	Version     = "1.0.0"
)

func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil && port > 0 {
			return port
		}
	}
	return DefaultPort
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	response := api.HealthResponse{
		Status:  "ok",
		Version: Version,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	var req api.RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := api.RenderResponse{
			Success: false,
			Error:   fmt.Sprintf("无效的请求体: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	if req.Template == "" {
		response := api.RenderResponse{
			Success: false,
			Error:   "template 字段不能为空",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	result, err := template.RenderString(req.Template, req.Data)
	if err != nil {
		response := api.RenderResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := api.RenderResponse{
		Success:  true,
		Content:  result.Content,
		Warnings: result.Warnings,
	}
	
	json.NewEncoder(w).Encode(response)
}

func renderFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	var req api.RenderFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := api.RenderFileResponse{
			Success: false,
			Error:   fmt.Sprintf("无效的请求体: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	if req.TemplatePath == "" {
		response := api.RenderFileResponse{
			Success: false,
			Error:   "template_path 字段不能为空",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	result, err := template.RenderFile(req.TemplatePath, req.Data)
	if err != nil {
		response := api.RenderFileResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := api.RenderFileResponse{
		Success:      true,
		Content:      result.Content,
		Warnings:     result.Warnings,
		TemplatePath: req.TemplatePath,
	}
	
	json.NewEncoder(w).Encode(response)
}

func main() {
	port := getPort()
	
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/render", renderHandler)
	http.HandleFunc("/render-file", renderFileHandler)
	
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("模板引擎服务端启动在端口 %d\n", port)
	fmt.Printf("可用端点:\n")
	fmt.Printf("  GET  /health       - 健康检查\n")
	fmt.Printf("  POST /render       - 渲染字符串模板\n")
	fmt.Printf("  POST /render-file  - 从文件渲染模板\n")
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
