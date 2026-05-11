package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"splaytree/common"
	"splaytree/splaytree"
)

type Handler struct {
	tree *splaytree.SplayTree
}

func NewHandler() *Handler {
	return &Handler{
		tree: splaytree.NewSplayTree(),
	}
}

func (h *Handler) PutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.tree.Put(req.Key, req.Value)

	resp := common.PutResponse{
		Success: true,
		Message: "put success",
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := h.tree.Get(req.Key)
	if err != nil {
		if errors.Is(err, splaytree.ErrKeyNotFound) {
			resp := common.GetResponse{
				Success: false,
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := common.GetResponse{
		Success: true,
		Value:   value,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.tree.Delete(req.Key)
	if err != nil {
		if errors.Is(err, splaytree.ErrKeyNotFound) {
			resp := common.DeleteResponse{
				Success: false,
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := common.DeleteResponse{
		Success: true,
		Message: "delete success",
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RangeQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RangeQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sum, err := h.tree.RangeQueryByKey(req.Left, req.Right)
	keys := h.tree.KeysInRange(req.Left, req.Right)
	values := h.tree.ValuesInRange(req.Left, req.Right)

	if err != nil && len(keys) == 0 {
		resp := common.RangeQueryResponse{
			Success: false,
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.RangeQueryResponse{
		Success: true,
		Sum:     sum,
		Keys:    keys,
		Values:  values,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RangeAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RangeAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.tree.RangeAddByKey(req.Left, req.Right, req.Delta)
	if err != nil {
		if errors.Is(err, splaytree.ErrInvalidRange) {
			resp := common.RangeAddResponse{
				Success: false,
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := common.RangeAddResponse{
		Success: true,
		Message: "range add success",
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RangeSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.RangeSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.tree.RangeSetByKey(req.Left, req.Right, req.Value)
	if err != nil {
		if errors.Is(err, splaytree.ErrInvalidRange) {
			resp := common.RangeSetResponse{
				Success: false,
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := common.RangeSetResponse{
		Success: true,
		Message: "range set success",
	}
	json.NewEncoder(w).Encode(resp)
}
