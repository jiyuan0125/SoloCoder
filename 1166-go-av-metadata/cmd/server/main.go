package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/example/avmetadata/pkg/api"
	"github.com/example/avmetadata/pkg/metadata"
)

func main() {
	var port string
	var host string

	flag.StringVar(&port, "port", "", "服务端口")
	flag.StringVar(&host, "host", "127.0.0.1", "监听地址")
	flag.Parse()

	if port == "" {
		port = os.Getenv("AVMETA_PORT")
		if port == "" {
			port = "8103"
		}
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	http.HandleFunc("/api/metadata", handleMetadata)
	http.HandleFunc("/api/health", handleHealth)

	fmt.Printf("音视频元数据服务启动在 %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "启动服务失败: %v\n", err)
		os.Exit(1)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "av-metadata-server",
	})
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req api.MetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, false, "请求格式错误: "+err.Error(), nil)
		return
	}

	if req.FilePath == "" {
		sendResponse(w, false, "文件路径不能为空", nil)
		return
	}

	meta, err := metadata.ParseFile(req.FilePath)
	if err != nil {
		sendResponse(w, false, err.Error(), nil)
		return
	}

	apiMeta := &api.Metadata{
		Format:     meta.Format,
		Duration:   meta.Duration,
		Bitrate:    meta.Bitrate,
		FileSize:   meta.FileSize,
		FormatTags: meta.FormatTags,
		AudioTags:  meta.AudioTags,
	}

	sendResponse(w, true, "", apiMeta)
}

func sendResponse(w http.ResponseWriter, success bool, message string, meta *api.Metadata) {
	w.Header().Set("Content-Type", "application/json")

	resp := api.MetadataResponse{
		Success:  success,
		Message:  message,
		Metadata: meta,
	}

	if !success {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}
