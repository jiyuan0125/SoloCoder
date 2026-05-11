package analyzer

import (
	"go-dep-analyzer/pkg/api"
	"sort"
)

func ExportGraph(graph *DependencyGraph) *api.DependencyGraphResponse {
	resp := &api.DependencyGraphResponse{
		Nodes: []api.GraphNode{},
		Edges: []api.GraphEdge{},
	}

	for path, node := range graph.Nodes {
		resp.Nodes = append(resp.Nodes, api.GraphNode{
			Module:     path,
			Version:    node.Version,
			IsRoot:     node.IsRoot,
			IsIndirect: node.IsIndirect,
		})
	}

	for _, edge := range graph.Edges {
		fromVersion := ""
		toVersion := ""
		if fromNode, ok := graph.Nodes[edge.From]; ok {
			fromVersion = fromNode.Version
		}
		if toNode, ok := graph.Nodes[edge.To]; ok {
			toVersion = toNode.Version
		}

		resp.Edges = append(resp.Edges, api.GraphEdge{
			From:       edge.From,
			To:         edge.To,
			IsDirect:   edge.IsDirect,
			FromModule: edge.From + "@" + fromVersion,
			ToModule:   edge.To + "@" + toVersion,
		})
	}

	return resp
}

func ExportCycles(cycles [][]string) *api.CyclesResponse {
	resp := &api.CyclesResponse{
		Cycles: []api.Cycle{},
	}

	for _, cycle := range cycles {
		resp.Cycles = append(resp.Cycles, api.Cycle{
			Modules: cycle,
		})
	}

	return resp
}

func ExportTopology(topology []string) *api.TopologyResponse {
	return &api.TopologyResponse{
		Modules: topology,
	}
}

func ExportDepths(depths map[string]int) *api.DepthAnalysisResponse {
	resp := &api.DepthAnalysisResponse{
		ModuleDepths: []api.ModuleDepth{},
	}

	for module, depth := range depths {
		resp.ModuleDepths = append(resp.ModuleDepths, api.ModuleDepth{
			Module: module,
			Depth:  depth,
		})
	}

	sort.Slice(resp.ModuleDepths, func(i, j int) bool {
		return resp.ModuleDepths[i].Depth > resp.ModuleDepths[j].Depth
	})

	return resp
}

func ExportConflicts(conflicts []VersionConflict) *api.VersionConflictResponse {
	resp := &api.VersionConflictResponse{
		Conflicts: []api.VersionConflict{},
	}

	for _, c := range conflicts {
		resp.Conflicts = append(resp.Conflicts, api.VersionConflict{
			Module:          c.Module,
			RequiredVersion: c.RequiredVersion,
			ActualVersion:   c.ActualVersion,
			Reason:          c.Reason,
		})
	}

	return resp
}

func ExportDependencyChains(graph *DependencyGraph, chains [][]string) *api.DependencyChainResponse {
	resp := &api.DependencyChainResponse{
		Chains: []api.DependencyChain{},
	}

	if len(chains) == 0 {
		return resp
	}

	resp.TargetModule = chains[0][len(chains[0])-1]

	for _, chain := range chains {
		nodes := []api.ChainNode{}
		for _, module := range chain {
			version := ""
			if node, ok := graph.Nodes[module]; ok {
				version = node.Version
			}
			nodes = append(nodes, api.ChainNode{
				Module:  module,
				Version: version,
			})
		}
		resp.Chains = append(resp.Chains, api.DependencyChain{
			Path: nodes,
		})
	}

	return resp
}

func ExportFullReport(analysis *FullAnalysis) *api.FullReportResponse {
	totalDeps := 0
	directDeps := 0
	indirectDeps := 0

	for _, node := range analysis.Graph.Nodes {
		if node.IsRoot {
			continue
		}
		totalDeps++
		if node.IsIndirect {
			indirectDeps++
		} else {
			directDeps++
		}
	}

	deepestModules := []api.ModuleDepth{}
	for module, depth := range analysis.Depths {
		deepestModules = append(deepestModules, api.ModuleDepth{
			Module: module,
			Depth:  depth,
		})
	}
	sort.Slice(deepestModules, func(i, j int) bool {
		return deepestModules[i].Depth > deepestModules[j].Depth
	})
	if len(deepestModules) > 5 {
		deepestModules = deepestModules[:5]
	}

	cycles := []api.Cycle{}
	for _, c := range analysis.Cycles {
		cycles = append(cycles, api.Cycle{Modules: c})
	}

	conflicts := []api.VersionConflict{}
	for _, c := range analysis.Conflicts {
		conflicts = append(conflicts, api.VersionConflict{
			Module:          c.Module,
			RequiredVersion: c.RequiredVersion,
			ActualVersion:   c.ActualVersion,
			Reason:          c.Reason,
		})
	}

	suggestions := []api.SlimmingSuggestion{}
	for _, s := range analysis.SlimmingSuggestions {
		suggestions = append(suggestions, api.SlimmingSuggestion{
			Module:     s.Module,
			Suggestion: s.Suggestion,
		})
	}

	return &api.FullReportResponse{
		TotalDeps:          totalDeps,
		DirectDeps:         directDeps,
		IndirectDeps:       indirectDeps,
		Cycles:             cycles,
		Conflicts:          conflicts,
		DeepestModules:     deepestModules,
		SlimmingSuggestions: suggestions,
	}
}
