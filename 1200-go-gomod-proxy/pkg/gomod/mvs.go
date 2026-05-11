package gomod

import (
	"fmt"
)

type DepProvider func(path, version string) ([]Require, error)

func RunMVS(mod *GoModFile, provider DepProvider) (*MVSResult, error) {
	result := &MVSResult{
		Selected:    make(map[string]string),
		Conflicts:   make([]Conflict, 0),
		Retracted:   make([]RetractedVersion, 0),
		AllVersions: make(map[string][]string),
	}
	
	requiredBy := make(map[string]map[string]string)
	
	visited := make(map[string]bool)
	queue := make([]Require, 0)
	
	for _, r := range mod.Requires {
		path, version := mod.ApplyReplace(r.Path, r.Version)
		
		if mod.IsExcluded(path, version) {
			continue
		}
		
		if requiredBy[path] == nil {
			requiredBy[path] = make(map[string]string)
		}
		requiredBy[path][version] = mod.Module
		
		key := fmt.Sprintf("%s@%s", path, version)
		if !visited[key] {
			visited[key] = true
			queue = append(queue, Require{Path: path, Version: version, Indirect: r.Indirect})
		}
	}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if result.AllVersions[current.Path] == nil {
			result.AllVersions[current.Path] = make([]string, 0)
		}
		result.AllVersions[current.Path] = append(result.AllVersions[current.Path], current.Version)
		
		deps, err := provider(current.Path, current.Version)
		if err != nil {
			continue
		}
		
		for _, dep := range deps {
			depPath, depVersion := mod.ApplyReplace(dep.Path, dep.Version)
			
			if mod.IsExcluded(depPath, depVersion) {
				continue
			}
			
			if requiredBy[depPath] == nil {
				requiredBy[depPath] = make(map[string]string)
			}
			requiredBy[depPath][depVersion] = current.Path + "@" + current.Version
			
			key := fmt.Sprintf("%s@%s", depPath, depVersion)
			if !visited[key] {
				visited[key] = true
				queue = append(queue, Require{Path: depPath, Version: depVersion, Indirect: true})
			}
		}
	}
	
	for path, versionsMap := range requiredBy {
		versions := make([]string, 0, len(versionsMap))
		for v := range versionsMap {
			versions = append(versions, v)
		}
		
		if len(versions) == 0 {
			continue
		}
		
		maxVer, err := MaxVersion(versions)
		if err != nil {
			return nil, err
		}
		
		result.Selected[path] = maxVer
		
		if len(versions) > 1 {
			conflict := Conflict{
				Path:       path,
				Selected:   maxVer,
				RequiredBy: make(map[string]string),
			}
			for v, requirer := range versionsMap {
				conflict.RequiredBy[v] = requirer
			}
			result.Conflicts = append(result.Conflicts, conflict)
		}
		
		if mod.IsRetracted(path, maxVer) {
			result.Retracted = append(result.Retracted, RetractedVersion{
				Path:    path,
				Version: maxVer,
				Reason:  "version has been retracted by module author",
			})
		}
	}
	
	return result, nil
}

func BuildDependencyGraph(mod *GoModFile, selected map[string]string, provider DepProvider) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Root:  mod.Module,
		Nodes: make([]GraphNode, 0),
		Edges: make([]GraphEdge, 0),
	}
	
	nodeIDs := make(map[string]string)
	nodeCounter := 0
	
	getNodeID := func(path, version string) string {
		key := fmt.Sprintf("%s@%s", path, version)
		if id, exists := nodeIDs[key]; exists {
			return id
		}
		nodeCounter++
		id := fmt.Sprintf("n%d", nodeCounter)
		nodeIDs[key] = id
		return id
	}
	
	rootID := "root"
	graph.Nodes = append(graph.Nodes, GraphNode{
		ID:      rootID,
		Path:    mod.Module,
		Version: "",
		IsDirect: false,
	})
	
	for _, r := range mod.GetDirectRequires() {
		path, version := mod.ApplyReplace(r.Path, r.Version)
		
		if selectedVersion, exists := selected[path]; exists {
			version = selectedVersion
		}
		
		if mod.IsExcluded(path, version) {
			continue
		}
		
		nodeID := getNodeID(path, version)
		graph.Nodes = append(graph.Nodes, GraphNode{
			ID:       nodeID,
			Path:     path,
			Version:  version,
			IsDirect: true,
		})
		
		graph.Edges = append(graph.Edges, GraphEdge{
			From: rootID,
			To:   nodeID,
		})
		
		err := buildSubGraph(path, version, nodeID, graph, selected, provider, mod, nodeIDs, getNodeID)
		if err != nil {
			return nil, err
		}
	}
	
	for _, r := range mod.GetIndirectRequires() {
		path, version := mod.ApplyReplace(r.Path, r.Version)
		
		if selectedVersion, exists := selected[path]; exists {
			version = selectedVersion
		}
		
		if mod.IsExcluded(path, version) {
			continue
		}
		
		key := fmt.Sprintf("%s@%s", path, version)
		if _, exists := nodeIDs[key]; exists {
			continue
		}
		
		nodeID := getNodeID(path, version)
		graph.Nodes = append(graph.Nodes, GraphNode{
			ID:       nodeID,
			Path:     path,
			Version:  version,
			IsDirect: false,
		})
		
		graph.Edges = append(graph.Edges, GraphEdge{
			From: rootID,
			To:   nodeID,
		})
	}
	
	return graph, nil
}

func buildSubGraph(path, version, parentID string, graph *DependencyGraph, selected map[string]string, provider DepProvider, mod *GoModFile, nodeIDs map[string]string, getNodeID func(string, string) string) error {
	deps, err := provider(path, version)
	if err != nil {
		return err
	}
	
	for _, dep := range deps {
		depPath, depVersion := mod.ApplyReplace(dep.Path, dep.Version)
		
		if selectedVersion, exists := selected[depPath]; exists {
			depVersion = selectedVersion
		}
		
		if mod.IsExcluded(depPath, depVersion) {
			continue
		}
		
		key := fmt.Sprintf("%s@%s", depPath, depVersion)
		if _, exists := nodeIDs[key]; exists {
			childID := nodeIDs[key]
			graph.Edges = append(graph.Edges, GraphEdge{
				From: parentID,
				To:   childID,
			})
			continue
		}
		
		childID := getNodeID(depPath, depVersion)
		graph.Nodes = append(graph.Nodes, GraphNode{
			ID:       childID,
			Path:     depPath,
			Version:  depVersion,
			IsDirect: false,
		})
		
		graph.Edges = append(graph.Edges, GraphEdge{
			From: parentID,
			To:   childID,
		})
		
		err := buildSubGraph(depPath, depVersion, childID, graph, selected, provider, mod, nodeIDs, getNodeID)
		if err != nil {
			return err
		}
	}
	
	return nil
}
