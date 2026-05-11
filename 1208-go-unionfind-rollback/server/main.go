package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/example/unionfind-rollback/common"
	"github.com/example/unionfind-rollback/unionfind"
)

var uf *unionfind.UnionFind

func handleActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	results := make([]common.ActionResult, 0)
	success := true

	for _, action := range req.Actions {
		var result common.ActionResult
		result.Type = action.Type

		switch action.Type {
		case "add_node":
			err := uf.AddNode(action.Node)
			if err != nil {
				result.Success = false
				result.Message = err.Error()
			} else {
				result.Success = true
				result.Message = fmt.Sprintf("Node %s added", action.Node)
			}
			result.Step = uf.Step()
		case "union":
			merged, err := uf.Union(action.A, action.B)
			if err != nil {
				result.Success = false
				result.Message = err.Error()
			} else {
				if merged {
					result.Success = true
					result.Message = fmt.Sprintf("Union %s and %s: merged", action.A, action.B)
				} else {
					result.Success = true
					result.Message = fmt.Sprintf("Union %s and %s: already connected", action.A, action.B)
				}
			}
			result.Step = uf.Step()
		case "find":
			connected, err := uf.Connected(action.A, action.B)
			if err != nil {
				result.Success = false
				result.Message = err.Error()
			} else {
				result.Success = true
				result.Message = fmt.Sprintf("Find %s and %s", action.A, action.B)
				result.Data = map[string]interface{}{"connected": connected}
			}
			result.Step = uf.Step()
		case "undo":
			undone, err := uf.Undo()
			if err != nil {
				result.Success = false
				result.Message = err.Error()
				success = false
			} else {
				result.Success = true
				if undone {
					result.Message = "Last union operation undone"
				} else {
					result.Message = "No operation to undo"
				}
			}
			result.Step = uf.Step()
		case "undo_to":
			undone, err := uf.UndoTo(action.Step)
			if err != nil {
				result.Success = false
				result.Message = err.Error()
				success = false
			} else {
				result.Success = true
				if undone {
					result.Message = fmt.Sprintf("Undo to step %d", action.Step)
				} else {
					result.Message = fmt.Sprintf("Already at or before step %d", action.Step)
				}
			}
			result.Step = uf.Step()
		default:
			result.Success = false
			result.Message = fmt.Sprintf("Unknown action type: %s", action.Type)
			success = false
		}

		results = append(results, result)
	}

	resp := common.Response{
		Results: results,
		Success: success,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	uf = unionfind.New()

	portFlag := flag.String("port", "", "Server port")
	flag.Parse()

	port := *portFlag
	if port == "" {
		port = os.Getenv("UF_PORT")
	}
	if port == "" {
		port = "8080"
	}

	if _, err := strconv.Atoi(port); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port: %s\n", port)
		os.Exit(1)
	}

	http.HandleFunc("/actions", handleActions)

	addr := ":" + port
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
