package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type RaftServer struct {
	node *RaftNode
}

func NewRaftServer(node *RaftNode) *RaftServer {
	return &RaftServer{node: node}
}

func (rs *RaftServer) Start() {
	http.HandleFunc("/requestVote", rs.handleRequestVote)
	http.HandleFunc("/appendEntries", rs.handleAppendEntries)
	http.HandleFunc("/set", rs.handleSet)
	http.HandleFunc("/get", rs.handleGet)
	http.HandleFunc("/status", rs.handleStatus)

	addr := fmt.Sprintf(":%d", rs.node.port)
	log.Printf("Node %d starting HTTP server on %s", rs.node.id, addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Node %d HTTP server error: %v", rs.node.id, err)
		}
	}()
}

func (rs *RaftServer) handleRequestVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var args RequestVoteArgs
	if err := json.Unmarshal(body, &args); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	reply := rs.node.RequestVote(&args)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}

func (rs *RaftServer) handleAppendEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var args AppendEntriesArgs
	if err := json.Unmarshal(body, &args); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	reply := rs.node.AppendEntries(&args)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}

func (rs *RaftServer) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	err := rs.node.Set(key, value)
	if err != nil {
		if _, ok := err.(*NotLeaderError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTemporaryRedirect)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "not leader",
				"leader":  err.(*NotLeaderError).LeaderID,
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"key":     key,
		"value":   value,
	})
}

func (rs *RaftServer) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	value, exists := rs.node.Get(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists": exists,
		"key":    key,
		"value":  value,
	})
}

func (rs *RaftServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rs.node.mu.Lock()
	defer rs.node.mu.Unlock()

	status := map[string]interface{}{
		"nodeId":       rs.node.id,
		"port":         rs.node.port,
		"state":        rs.node.state,
		"currentTerm":  rs.node.currentTerm,
		"commitIndex":  rs.node.commitIndex,
		"lastApplied":  rs.node.lastApplied,
		"logLength":    len(rs.node.log),
		"votedFor":     rs.node.votedFor,
		"stateMachine": rs.node.stateMachine.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func SendRequestVote(peerAddr string, args *RequestVoteArgs) (*RequestVoteReply, error) {
	url := fmt.Sprintf("http://%s/requestVote", peerAddr)
	
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reply RequestVoteReply
	if err := json.Unmarshal(respBody, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}

func SendAppendEntries(peerAddr string, args *AppendEntriesArgs) (*AppendEntriesReply, error) {
	url := fmt.Sprintf("http://%s/appendEntries", peerAddr)
	
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reply AppendEntriesReply
	if err := json.Unmarshal(respBody, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}

func SendSetRequest(peerAddr string, key string, value string) error {
	url := fmt.Sprintf("http://%s/set?key=%s&value=%s", peerAddr, key, value)
	
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func SendGetRequest(peerAddr string, key string) (string, bool, error) {
	url := fmt.Sprintf("http://%s/get?key=%s", peerAddr, key)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", false, err
	}

	exists, _ := result["exists"].(bool)
	value, _ := result["value"].(string)

	return value, exists, nil
}

func SendStatusRequest(peerAddr string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s/status", peerAddr)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
