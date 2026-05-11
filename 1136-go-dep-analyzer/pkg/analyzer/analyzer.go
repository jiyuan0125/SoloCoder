package analyzer

func DetectVersionConflicts(mod *GoModFile, sumVersions map[string]string, graph *DependencyGraph) []VersionConflict {
	var conflicts []VersionConflict

	for _, req := range mod.Requires {
		targetPath := req.Path

		replaceKey := req.Path
		if rep, ok := graph.Replaces[replaceKey]; ok {
			targetPath = rep.NewPath
		}
		replaceKeyWithVersion := req.Path + "@" + req.Version
		if rep, ok := graph.Replaces[replaceKeyWithVersion]; ok {
			targetPath = rep.NewPath
		}

		actualVersion, existsInSum := sumVersions[targetPath]
		if existsInSum && actualVersion != req.Version {
			conflicts = append(conflicts, VersionConflict{
				Module:          req.Path,
				RequiredVersion: req.Version,
				ActualVersion:   actualVersion,
				Reason:          "go.mod 中声明的版本与 go.sum 中实际使用的版本不一致，可能由间接依赖升级或 MVS 算法选择导致",
			})
		}
	}

	return conflicts
}

func GenerateSlimmingSuggestions(mod *GoModFile, graph *DependencyGraph) []SlimmingSuggestion {
	var suggestions []SlimmingSuggestion

	incomingEdges := make(map[string]int)
	for _, edge := range graph.Edges {
		incomingEdges[edge.To]++
	}

	for _, req := range mod.Requires {
		if req.IsIndirect {
			continue
		}

		targetPath := req.Path
		replaceKey := req.Path
		if rep, ok := graph.Replaces[replaceKey]; ok {
			targetPath = rep.NewPath
		}
		replaceKeyWithVersion := req.Path + "@" + req.Version
		if rep, ok := graph.Replaces[replaceKeyWithVersion]; ok {
			targetPath = rep.NewPath
		}

		node, exists := graph.Nodes[targetPath]
		if !exists {
			continue
		}

		if node.IsIndirect {
			suggestions = append(suggestions, SlimmingSuggestion{
				Module:     req.Path,
				Suggestion: "该模块被声明为直接依赖，但在依赖图中实际表现为间接依赖。建议添加 // indirect 注释或将其从 go.mod 中移除",
			})
		}
	}

	return suggestions
}

func Analyze(goModContent string, goSumContent string) (*FullAnalysis, error) {
	mod, err := ParseGoMod(goModContent)
	if err != nil {
		return nil, err
	}

	sumEntries, err := ParseGoSum(goSumContent)
	if err != nil {
		return nil, err
	}

	sumVersions := GetModuleVersionsFromSum(sumEntries)
	graph := BuildDependencyGraph(mod, sumVersions)
	cycles := DetectCycles(graph)
	topology := TopologicalSort(graph)
	depths := CalculateDepths(graph)
	conflicts := DetectVersionConflicts(mod, sumVersions, graph)
	slimmingSuggestions := GenerateSlimmingSuggestions(mod, graph)

	return &FullAnalysis{
		Graph:               graph,
		Cycles:              cycles,
		Topology:            topology,
		Depths:              depths,
		Conflicts:           conflicts,
		SlimmingSuggestions: slimmingSuggestions,
	}, nil
}
