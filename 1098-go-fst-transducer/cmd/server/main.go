package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"go-fst-transducer/common"
	"go-fst-transducer/fst"
)

type DictManager struct {
	mu    sync.RWMutex
	dicts map[string]*fst.FST
}

func NewDictManager() *DictManager {
	return &DictManager{
		dicts: make(map[string]*fst.FST),
	}
}

func (dm *DictManager) Build(dictName string, items []common.KeyValue) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	fstItems := make([]fst.KeyValue, len(items))
	for i, item := range items {
		fstItems[i] = fst.KeyValue{
			Key:   item.Key,
			Value: item.Value,
		}
	}

	newFst := fst.NewFST()
	if err := newFst.Build(fstItems); err != nil {
		return err
	}

	dm.dicts[dictName] = newFst
	return nil
}

func (dm *DictManager) ExactSearch(dictName string, key string) (uint64, bool, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dict, exists := dm.dicts[dictName]
	if !exists {
		return 0, false, fmt.Errorf("dictionary not found: %s", dictName)
	}

	value, found := dict.ExactSearch(key)
	return value, found, nil
}

func (dm *DictManager) PrefixSearch(dictName string, prefix string) ([]common.KeyValue, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dict, exists := dm.dicts[dictName]
	if !exists {
		return nil, fmt.Errorf("dictionary not found: %s", dictName)
	}

	results := dict.PrefixSearch(prefix)
	items := make([]common.KeyValue, len(results))
	for i, r := range results {
		items[i] = common.KeyValue{
			Key:   r.Key,
			Value: r.Value,
		}
	}
	return items, nil
}

func (dm *DictManager) Dump(dictName string) ([]common.KeyValue, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dict, exists := dm.dicts[dictName]
	if !exists {
		return nil, fmt.Errorf("dictionary not found: %s", dictName)
	}

	items := make([]common.KeyValue, 0)
	err := dict.Walk(func(key string, value uint64) error {
		items = append(items, common.KeyValue{
			Key:   key,
			Value: value,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (dm *DictManager) Save(dictName string, filePath string) error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dict, exists := dm.dicts[dictName]
	if !exists {
		return fmt.Errorf("dictionary not found: %s", dictName)
	}

	data, err := dict.Serialize()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, data, 0644)
}

func (dm *DictManager) Load(dictName string, filePath string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	dict, err := fst.Deserialize(data)
	if err != nil {
		return err
	}

	dm.dicts[dictName] = dict
	return nil
}

func (dm *DictManager) List() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	names := make([]string, 0, len(dm.dicts))
	for name := range dm.dicts {
		names = append(names, name)
	}
	return names
}

var dictManager *DictManager

func main() {
	dictManager = NewDictManager()

	http.HandleFunc("/api/dict/build", handleBuild)
	http.HandleFunc("/api/dict/search", handleSearch)
	http.HandleFunc("/api/dict/prefix", handlePrefix)
	http.HandleFunc("/api/dict/dump", handleDump)
	http.HandleFunc("/api/dict/save", handleSave)
	http.HandleFunc("/api/dict/load", handleLoad)
	http.HandleFunc("/api/dict/list", handleList)

	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	log.Printf("FST Dictionary Server listening on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, common.ErrorResponse{Error: message})
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.DictName == "" {
		sendError(w, http.StatusBadRequest, "dict_name is required")
		return
	}

	if err := dictManager.Build(req.DictName, req.Items); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("build failed: %v", err))
		return
	}

	sendJSON(w, http.StatusOK, common.BuildResponse{
		Success: true,
		Message: "dictionary built successfully",
		Count:   len(req.Items),
	})
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dictName := r.URL.Query().Get("dict_name")
	key := r.URL.Query().Get("key")

	if dictName == "" || key == "" {
		sendError(w, http.StatusBadRequest, "dict_name and key are required")
		return
	}

	value, found, err := dictManager.ExactSearch(dictName, key)
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	resp := common.SearchResponse{
		Found: found,
	}
	if found {
		resp.Value = value
	}
	sendJSON(w, http.StatusOK, resp)
}

func handlePrefix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dictName := r.URL.Query().Get("dict_name")
	prefix := r.URL.Query().Get("prefix")

	if dictName == "" || prefix == "" {
		sendError(w, http.StatusBadRequest, "dict_name and prefix are required")
		return
	}

	items, err := dictManager.PrefixSearch(dictName, prefix)
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, common.PrefixResponse{
		Items: items,
	})
}

func handleDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dictName := r.URL.Query().Get("dict_name")
	if dictName == "" {
		sendError(w, http.StatusBadRequest, "dict_name is required")
		return
	}

	items, err := dictManager.Dump(dictName)
	if err != nil {
		sendError(w, http.StatusNotFound, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, common.DumpResponse{
		Items: items,
	})
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.DictName == "" || req.FilePath == "" {
		sendError(w, http.StatusBadRequest, "dict_name and file_path are required")
		return
	}

	if err := dictManager.Save(req.DictName, req.FilePath); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("save failed: %v", err))
		return
	}

	sendJSON(w, http.StatusOK, common.SaveLoadResponse{
		Success: true,
		Message: "dictionary saved successfully",
	})
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.LoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.DictName == "" || req.FilePath == "" {
		sendError(w, http.StatusBadRequest, "dict_name and file_path are required")
		return
	}

	if err := dictManager.Load(req.DictName, req.FilePath); err != nil {
		sendError(w, http.StatusInternalServerError, fmt.Sprintf("load failed: %v", err))
		return
	}

	sendJSON(w, http.StatusOK, common.SaveLoadResponse{
		Success: true,
		Message: "dictionary loaded successfully",
	})
}

func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dicts := dictManager.List()
	sendJSON(w, http.StatusOK, common.ListResponse{
		Dicts: dicts,
	})
}
