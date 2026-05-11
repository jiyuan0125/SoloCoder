package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"phonetic-search/api"
	"phonetic-search/phonetic"
)

var store = phonetic.NewInMemoryStore()

func getPort() string {
	port := "8304"
	if envPort := os.Getenv("PHONETIC_PORT"); envPort != "" {
		port = envPort
	}
	flagPort := flag.String("port", "", "服务监听端口")
	flag.Parse()
	if *flagPort != "" {
		port = *flagPort
	}
	return port
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	name := strings.TrimSpace(req.Name)
	code := phonetic.Encode(name)

	resp := api.EncodeResponse{
		Name:      name,
		Soundex:   code.Soundex,
		Metaphone: code.Metaphone,
	}

	if name == "" {
		resp.Message = "输入为空，返回空编码"
	} else if code.Soundex == "" {
		resp.Message = "输入不包含有效英文字母"
	}

	writeJSON(w, http.StatusOK, resp)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req api.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	name := strings.TrimSpace(req.Name)
	code := phonetic.Encode(name)

	var matches []string
	if code.Soundex != "" {
		matches = store.SearchBySoundex(code.Soundex)
	}

	resp := api.SearchResponse{
		QueryName: name,
		Soundex:   code.Soundex,
		Metaphone: code.Metaphone,
		Matches:   matches,
	}

	if name == "" {
		resp.Message = "输入为空"
	} else if code.Soundex == "" {
		resp.Message = "输入不包含有效英文字母"
	}

	writeJSON(w, http.StatusOK, resp)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	name := strings.TrimSpace(req.Name)
	added := store.Add(name)

	resp := api.AddResponse{
		Name:  name,
		Added: added,
	}

	if added {
		resp.Message = "人名添加成功"
	} else if name == "" {
		resp.Message = "输入为空"
	} else {
		cleanName := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				return r
			}
			return -1
		}, name)
		if cleanName == "" {
			resp.Message = "输入不包含有效英文字母"
		} else {
			resp.Message = "人名已存在"
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req api.RemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	name := strings.TrimSpace(req.Name)
	removed := store.Remove(name)

	resp := api.RemoveResponse{
		Name:    name,
		Removed: removed,
	}

	if removed {
		resp.Message = "人名删除成功"
	} else {
		resp.Message = "人名不存在或输入无效"
	}

	writeJSON(w, http.StatusOK, resp)
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req api.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的请求格式"})
		return
	}

	added := 0
	existing := 0
	invalid := make([]string, 0)

	for _, name := range req.Names {
		cleanName := strings.TrimSpace(name)
		if cleanName == "" {
			continue
		}

		cleaned := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				return r
			}
			return -1
		}, cleanName)
		if cleaned == "" {
			invalid = append(invalid, name)
			continue
		}

		if store.Exists(cleanName) {
			existing++
			continue
		}

		if store.Add(cleanName) {
			added++
		} else {
			existing++
		}
	}

	resp := api.ImportResponse{
		Total:    len(req.Names),
		Added:    added,
		Existing: existing,
		Invalid:  invalid,
		Message:  "导入完成",
	}

	writeJSON(w, http.StatusOK, resp)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	topSoundex := store.TopSoundexCodes(5)
	topMetaphone := store.TopMetaphoneCodes(5)

	codeStatsSoundex := make([]api.CodeStats, len(topSoundex))
	for i, cc := range topSoundex {
		codeStatsSoundex[i] = api.CodeStats{
			Code:  cc.Code,
			Count: cc.Count,
			Names: cc.Names,
		}
	}

	codeStatsMetaphone := make([]api.CodeStats, len(topMetaphone))
	for i, cc := range topMetaphone {
		codeStatsMetaphone[i] = api.CodeStats{
			Code:  cc.Code,
			Count: cc.Count,
			Names: cc.Names,
		}
	}

	resp := api.StatsResponse{
		TotalNames:        store.Total(),
		UniqueSoundex:     store.UniqueSoundexCodes(),
		UniqueMetaphone:   store.UniqueMetaphoneCodes(),
		TopSoundexCodes:   codeStatsSoundex,
		TopMetaphoneCodes: codeStatsMetaphone,
	}

	writeJSON(w, http.StatusOK, resp)
}

func main() {
	port := getPort()

	http.HandleFunc("/encode", encodeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/remove", removeHandler)
	http.HandleFunc("/import", importHandler)
	http.HandleFunc("/stats", statsHandler)

	fmt.Printf("语音编码服务启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
