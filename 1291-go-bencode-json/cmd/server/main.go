package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bencode/json-converter/pkg/bencode"
	"github.com/bencode/json-converter/pkg/common"
)

const defaultPort = "8100"

func main() {
	portFlag := flag.String("port", "", "服务端监听端口")
	flag.Parse()

	port := getPort(*portFlag)
	fmt.Printf("Bencode 转换服务正在监听 :%s\n", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/decode", handleDecode)
	mux.HandleFunc("/encode", handleEncode)
	mux.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "启动服务端失败: %v\n", err)
		os.Exit(1)
	}
}

func getPort(flagPort string) string {
	if flagPort != "" {
		return flagPort
	}
	if envPort := os.Getenv("BENCODE_PORT"); envPort != "" {
		return envPort
	}
	return defaultPort
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var input []byte
	var filename string

	if strings.Contains(contentType, "multipart/form-data") {
		file, fh, err := r.FormFile("file")
		if err != nil {
			sendError(w, "读取上传文件失败: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename = fh.Filename
		input, err = io.ReadAll(file)
		if err != nil {
			sendError(w, "读取上传文件内容失败: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		var err error
		input, err = io.ReadAll(r.Body)
		if err != nil {
			sendError(w, "读取请求体失败: "+err.Error(), http.StatusBadRequest)
			return
		}
		filename = "input"
	}

	if len(input) == 0 {
		sendError(w, "输入为空", http.StatusBadRequest)
		return
	}

	value, err := bencode.Decode(input)
	if err != nil {
		sendError(w, "解码Bencode失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	jsonData, err := bencode.ToJSON(value)
	if err != nil {
		sendError(w, "转换为JSON失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", strings.TrimSuffix(filename, filepath.Ext(filename))))
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(w, "读取请求体失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		sendError(w, "输入为空", http.StatusBadRequest)
		return
	}

	value, err := bencode.FromJSON(body)
	if err != nil {
		sendError(w, "解析JSON失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	bencodeData, err := bencode.Encode(value)
	if err != nil {
		sendError(w, "编码为Bencode失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-bittorrent")
	w.Header().Set("Content-Disposition", "attachment; filename=output.bencode")
	w.WriteHeader(http.StatusOK)
	w.Write(bencodeData)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := common.DecodeResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(resp)
}
