package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"gomod-mvs/pkg/api"
	"gomod-mvs/pkg/gomod"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8440, "Server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		var envPortNum int
		if _, err := fmt.Sscanf(envPort, "%d", &envPortNum); err == nil {
			port = envPortNum
		}
	}

	http.HandleFunc("/analyze", handleAnalyze)
	http.HandleFunc("/graph", handleGraph)
	http.HandleFunc("/verify", handleVerify)
	http.HandleFunc("/why", handleWhy)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := analyzeDependencies(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.GraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := getGraph(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := verifySum(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleWhy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.WhyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := findWhy(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func emptyProvider(path, version string) ([]gomod.Require, error) {
	return []gomod.Require{}, nil
}

func analyzeDependencies(req api.AnalyzeRequest) api.AnalyzeResponse {
	resp := api.AnalyzeResponse{}

	mod, err := gomod.ParseGoMod(req.GoModContent)
	if err != nil {
		resp.Error = fmt.Sprintf("Failed to parse go.mod: %v", err)
		return resp
	}

	resp.Module = mod.Module
	resp.GoVersion = mod.GoVersion

	resp.Dependencies = make([]api.DependencyInfo, len(mod.Requires))
	for i, r := range mod.Requires {
		resp.Dependencies[i] = api.DependencyInfo{
			Path:     r.Path,
			Version:  r.Version,
			Indirect: r.Indirect,
		}
	}

	mvsResult, err := gomod.RunMVS(mod, emptyProvider)
	if err != nil {
		resp.Error = fmt.Sprintf("MVS failed: %v", err)
		return resp
	}

	resp.MVSResult = api.MVSResultInfo{
		Selected:    mvsResult.Selected,
		AllVersions: mvsResult.AllVersions,
	}

	resp.MVSResult.Conflicts = make([]api.ConflictInfo, len(mvsResult.Conflicts))
	for i, c := range mvsResult.Conflicts {
		resp.MVSResult.Conflicts[i] = api.ConflictInfo{
			Path:       c.Path,
			Selected:   c.Selected,
			RequiredBy: c.RequiredBy,
		}
	}

	resp.MVSResult.Retracted = make([]api.RetractedInfo, len(mvsResult.Retracted))
	for i, r := range mvsResult.Retracted {
		resp.MVSResult.Retracted[i] = api.RetractedInfo{
			Path:    r.Path,
			Version: r.Version,
			Reason:  r.Reason,
		}
	}

	graph, err := gomod.BuildDependencyGraph(mod, mvsResult.Selected, emptyProvider)
	if err != nil {
		resp.Error = fmt.Sprintf("Graph build failed: %v", err)
		return resp
	}

	resp.Graph = api.GraphInfo{
		Root:  graph.Root,
		Nodes: make([]api.NodeInfo, len(graph.Nodes)),
		Edges: make([]api.EdgeInfo, len(graph.Edges)),
	}

	for i, n := range graph.Nodes {
		resp.Graph.Nodes[i] = api.NodeInfo{
			ID:       n.ID,
			Path:     n.Path,
			Version:  n.Version,
			IsDirect: n.IsDirect,
		}
	}

	for i, e := range graph.Edges {
		resp.Graph.Edges[i] = api.EdgeInfo{
			From: e.From,
			To:   e.To,
		}
	}

	sumEntries, err := gomod.ParseGoSum(req.GoSumContent)
	if err == nil {
		verifyResult, err := mod.VerifyGoSum(sumEntries)
		if err == nil {
			resp.VerifyResult = api.VerifyResultInfo{
				AllChecked: verifyResult.AllChecked,
			}

			resp.VerifyResult.Orphans = make([]api.OrphanInfo, len(verifyResult.Orphans))
			for i, o := range verifyResult.Orphans {
				resp.VerifyResult.Orphans[i] = api.OrphanInfo{
					Path:    o.Path,
					Version: o.Version,
					Hash:    o.Hash,
				}
			}

			resp.VerifyResult.Mismatches = make([]api.MismatchInfo, len(verifyResult.Mismatches))
			for i, m := range verifyResult.Mismatches {
				resp.VerifyResult.Mismatches[i] = api.MismatchInfo{
					Path:    m.Path,
					Version: m.Version,
					Error:   m.Error,
				}
			}
		}
	}

	return resp
}

func getGraph(req api.GraphRequest) api.GraphResponse {
	resp := api.GraphResponse{}

	mod, err := gomod.ParseGoMod(req.GoModContent)
	if err != nil {
		resp.Error = fmt.Sprintf("Failed to parse go.mod: %v", err)
		return resp
	}

	mvsResult, err := gomod.RunMVS(mod, emptyProvider)
	if err != nil {
		resp.Error = fmt.Sprintf("MVS failed: %v", err)
		return resp
	}

	graph, err := gomod.BuildDependencyGraph(mod, mvsResult.Selected, emptyProvider)
	if err != nil {
		resp.Error = fmt.Sprintf("Graph build failed: %v", err)
		return resp
	}

	resp.Graph = api.GraphInfo{
		Root:  graph.Root,
		Nodes: make([]api.NodeInfo, len(graph.Nodes)),
		Edges: make([]api.EdgeInfo, len(graph.Edges)),
	}

	for i, n := range graph.Nodes {
		resp.Graph.Nodes[i] = api.NodeInfo{
			ID:       n.ID,
			Path:     n.Path,
			Version:  n.Version,
			IsDirect: n.IsDirect,
		}
	}

	for i, e := range graph.Edges {
		resp.Graph.Edges[i] = api.EdgeInfo{
			From: e.From,
			To:   e.To,
		}
	}

	return resp
}

func verifySum(req api.VerifyRequest) api.VerifyResponse {
	resp := api.VerifyResponse{}

	mod, err := gomod.ParseGoMod(req.GoModContent)
	if err != nil {
		resp.Error = fmt.Sprintf("Failed to parse go.mod: %v", err)
		return resp
	}

	sumEntries, err := gomod.ParseGoSum(req.GoSumContent)
	if err != nil {
		resp.Error = fmt.Sprintf("Failed to parse go.sum: %v", err)
		return resp
	}

	verifyResult, err := mod.VerifyGoSum(sumEntries)
	if err != nil {
		resp.Error = fmt.Sprintf("Verify failed: %v", err)
		return resp
	}

	resp.Result = api.VerifyResultInfo{
		AllChecked: verifyResult.AllChecked,
	}

	resp.Result.Orphans = make([]api.OrphanInfo, len(verifyResult.Orphans))
	for i, o := range verifyResult.Orphans {
		resp.Result.Orphans[i] = api.OrphanInfo{
			Path:    o.Path,
			Version: o.Version,
			Hash:    o.Hash,
		}
	}

	resp.Result.Mismatches = make([]api.MismatchInfo, len(verifyResult.Mismatches))
	for i, m := range verifyResult.Mismatches {
		resp.Result.Mismatches[i] = api.MismatchInfo{
			Path:    m.Path,
			Version: m.Version,
			Error:   m.Error,
		}
	}

	return resp
}

func findWhy(req api.WhyRequest) api.WhyResponse {
	resp := api.WhyResponse{
		Target: req.TargetPath,
	}

	mod, err := gomod.ParseGoMod(req.GoModContent)
	if err != nil {
		resp.Error = fmt.Sprintf("Failed to parse go.mod: %v", err)
		return resp
	}

	mvsResult, err := gomod.RunMVS(mod, emptyProvider)
	if err != nil {
		resp.Error = fmt.Sprintf("MVS failed: %v", err)
		return resp
	}

	whyResult, err := gomod.FindDependencyPath(mod, req.TargetPath, mvsResult.Selected, emptyProvider)
	if err != nil {
		resp.Error = fmt.Sprintf("Why query failed: %v", err)
		return resp
	}

	resp.Found = whyResult.Found
	resp.Chains = whyResult.Chains

	return resp
}
