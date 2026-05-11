package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"kahn-toposort/pkg/common"
	"kahn-toposort/pkg/toposort"
	"log"
	"net/http"
	"os"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flag.StringVar(&port, "port", "8080", "服务端监听端口")
		flag.Parse()
	}
	return ":" + port
}

func sortHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	var req common.SortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的JSON请求体", http.StatusBadRequest)
		return
	}

	sorter := toposort.NewTopoSorter()
	if err := sorter.Import(req.Tasks, req.Dependencies); err != nil {
		response := common.SortResponse{
			Levels:   nil,
			HasCycle: false,
			Error:    err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	levels, hasCycle := sorter.Sort()

	response := common.SortResponse{
		Levels:   levels,
		HasCycle: hasCycle,
	}

	if hasCycle {
		cyclePath := sorter.FindCycle()
		if cyclePath != nil {
			response.Error = fmt.Sprintf("检测到循环依赖，参与循环的任务: %v", cyclePath)
		} else {
			response.Error = "检测到循环依赖"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	port := getPort()

	http.HandleFunc("/sort", sortHandler)

	log.Printf("拓扑排序服务已启动，监听端口 %s", port)
	log.Printf("API端点: POST http://localhost%s/sort", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
