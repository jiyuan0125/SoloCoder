package main

import (
	"encoding/json"
	"net/http"

	"bplus-tree/bptree"
	"bplus-tree/common"
)

type Handler struct {
	tree *bptree.BPTree
}

func NewHandler() *Handler {
	opts := &bptree.Options{
		Order:           4,
		AllowDuplicates: false,
	}
	return &Handler{
		tree: bptree.New(opts),
	}
}

func (h *Handler) Insert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "仅支持POST方法")
		return
	}

	var req common.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "键不能为空")
		return
	}

	h.tree.Insert(req.Key, req.Value)
	writeJSON(w, http.StatusOK, &common.InsertResponse{Success: true})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "仅支持POST方法")
		return
	}

	var req common.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "键不能为空")
		return
	}

	value, found := h.tree.Search(req.Key)
	resp := &common.SearchResponse{
		Found: found,
	}
	if found {
		resp.Key = req.Key
		resp.Value = value
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "仅支持POST方法")
		return
	}

	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "键不能为空")
		return
	}

	success := h.tree.Delete(req.Key)
	writeJSON(w, http.StatusOK, &common.DeleteResponse{Success: success})
}

func (h *Handler) Range(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "仅支持POST方法")
		return
	}

	var req common.RangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if req.Start > req.End {
		writeError(w, http.StatusBadRequest, "起始键不能大于结束键")
		return
	}

	items := h.tree.Range(req.Start, req.End)
	commonItems := make([]*common.KeyValue, len(items))
	for i, item := range items {
		commonItems[i] = &common.KeyValue{
			Key:   item.Key,
			Value: item.Value,
		}
	}

	writeJSON(w, http.StatusOK, &common.RangeResponse{
		Count: len(commonItems),
		Items: commonItems,
	})
}

func (h *Handler) Scan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "仅支持POST方法")
		return
	}

	var req common.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体解析失败")
		return
	}

	if req.BatchSize <= 0 {
		req.BatchSize = 10
	}

	var direction int
	switch req.Direction {
	case "backward":
		direction = bptree.Backward
	case "forward", "":
		direction = bptree.Forward
	default:
		writeError(w, http.StatusBadRequest, "无效的遍历方向")
		return
	}

	items := h.tree.Scan(direction, req.BatchSize+1, req.StartOffset)
	
	hasMore := len(items) > req.BatchSize
	if hasMore {
		items = items[:req.BatchSize]
	}

	commonItems := make([]*common.KeyValue, len(items))
	for i, item := range items {
		commonItems[i] = &common.KeyValue{
			Key:   item.Key,
			Value: item.Value,
		}
	}

	writeJSON(w, http.StatusOK, &common.ScanResponse{
		Count:   len(commonItems),
		Items:   commonItems,
		HasMore: hasMore,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, &common.ErrorResponse{Error: message})
}
