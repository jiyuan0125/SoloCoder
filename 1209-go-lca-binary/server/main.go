package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/example/lca-family-tree/common"
	"github.com/example/lca-family-tree/lca"
)

func main() {
	port := flag.String("port", "8080", "Server port (or use PORT environment variable)")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = &envPort
	}

	http.HandleFunc("/query", handleQuery)
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("Family Tree LCA Server starting on port %s...\n", *port)
	fmt.Printf("Health check: http://localhost:%s/health\n", *port)
	fmt.Printf("Query endpoint: http://localhost:%s/query\n", *port)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	req, err := common.ParseRequest(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	tree, err := buildFamilyTree(req.Tree)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid tree: %v", err))
		return
	}

	tree.Preprocess()

	response := &common.Response{
		Results: make([]common.ResultItem, 0, len(req.Queries)),
	}

	for _, query := range req.Queries {
		result := common.ResultItem{Query: query}
		lcaID, err := tree.LCA(query.NodeA, query.NodeB)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.LCA = lcaID
		}
		response.Results = append(response.Results, result)
	}

	w.WriteHeader(http.StatusOK)
	jsonData, _ := response.ToJSON()
	w.Write(jsonData)
}

func buildFamilyTree(treeData common.TreeData) (*lca.FamilyTree, error) {
	if len(treeData.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes provided")
	}

	sortedNodes := make([]common.Node, len(treeData.Nodes))
	copy(sortedNodes, treeData.Nodes)

	sort.SliceStable(sortedNodes, func(i, j int) bool {
		return (sortedNodes[i].ParentID == nil || *sortedNodes[i].ParentID == "") &&
			(sortedNodes[j].ParentID != nil && *sortedNodes[j].ParentID != "")
	})

	processed := make(map[string]bool)
	tree := lca.NewFamilyTree()
	remaining := make([]common.Node, 0, len(sortedNodes))

	for len(sortedNodes) > 0 {
		remaining = remaining[:0]
		progress := false

		for _, node := range sortedNodes {
			if processed[node.ID] {
				continue
			}

			if node.ParentID == nil || *node.ParentID == "" || processed[*node.ParentID] {
				if err := tree.AddNode(node.ID, node.ParentID); err != nil {
					return nil, err
				}
				processed[node.ID] = true
				progress = true
			} else {
				remaining = append(remaining, node)
			}
		}

		if !progress && len(remaining) > 0 {
			var unprocessed []string
			for _, node := range remaining {
				unprocessed = append(unprocessed, node.ID)
			}
			return nil, fmt.Errorf("cannot build tree due to missing parents or cycle: %s", strings.Join(unprocessed, ", "))
		}

		sortedNodes, remaining = remaining, sortedNodes
	}

	if err := tree.ValidateAndBuild(); err != nil {
		return nil, err
	}

	return tree, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	response := &common.Response{
		Error: message,
	}
	jsonData, _ := response.ToJSON()
	w.Write(jsonData)
}
