package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"segment-tree/model"
	"segment-tree/segmenttree"
)

const (
	defaultPort = "8080"
)

func getPort() string {
	port := flag.String("port", "", "服务端口")
	flag.Parse()

	if *port != "" {
		return *port
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		if _, err := strconv.Atoi(envPort); err == nil {
			return envPort
		}
	}

	return defaultPort
}

func execute(st *segmenttree.SegmentTree, req *model.Request) []int {
	var results []int
	for _, action := range req.Actions {
		switch action.Type {
		case model.ActionRangeAdd:
			st.RangeAdd(action.Left, action.Right, action.Value)
		case model.ActionRangeSum:
			results = append(results, st.RangeSum(action.Left, action.Right))
		case model.ActionRangeMax:
			results = append(results, st.RangeMax(action.Left, action.Right))
		}
	}
	return results
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "只接受 POST 请求")
		return
	}

	var req model.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON 解析失败: "+err.Error())
		return
	}
	defer r.Body.Close()

	st := segmenttree.New(req.InitValues)
	results := execute(st, &req)

	resp := model.Response{Results: results}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}

func main() {
	port := getPort()

	http.HandleFunc("/process", handleProcess)

	addr := ":" + port
	fmt.Printf("线段树统计分析服务启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
