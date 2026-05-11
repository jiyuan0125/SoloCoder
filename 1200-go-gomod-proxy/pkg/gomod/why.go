package gomod

func FindDependencyPath(mod *GoModFile, targetPath string, selected map[string]string, provider DepProvider) (*WhyResult, error) {
	result := &WhyResult{
		Found:  false,
		Target: targetPath,
		Chains: make([][]string, 0),
	}
	
	for _, r := range mod.GetDirectRequires() {
		path, version := mod.ApplyReplace(r.Path, r.Version)
		
		if selectedVersion, exists := selected[path]; exists {
			version = selectedVersion
		}
		
		if mod.IsExcluded(path, version) {
			continue
		}
		
		chain := []string{mod.Module, path + "@" + version}
		
		if path == targetPath {
			result.Found = true
			result.Chains = append(result.Chains, chain)
			continue
		}
		
		visited := make(map[string]bool)
		visitedKey := path + "@" + version
		visited[visitedKey] = true
		
		subChains, err := findPathRecursive(path, version, targetPath, chain, visited, mod, selected, provider)
		if err != nil {
			continue
		}
		
		if len(subChains) > 0 {
			result.Found = true
			result.Chains = append(result.Chains, subChains...)
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
		
		if path == targetPath {
			result.Found = true
			chain := []string{mod.Module, path + "@" + version + " (indirect)"}
			result.Chains = append(result.Chains, chain)
		}
	}
	
	return result, nil
}

func findPathRecursive(path, version, targetPath string, currentChain []string, visited map[string]bool, mod *GoModFile, selected map[string]string, provider DepProvider) ([][]string, error) {
	result := make([][]string, 0)
	
	deps, err := provider(path, version)
	if err != nil {
		return result, err
	}
	
	for _, dep := range deps {
		depPath, depVersion := mod.ApplyReplace(dep.Path, dep.Version)
		
		if selectedVersion, exists := selected[depPath]; exists {
			depVersion = selectedVersion
		}
		
		if mod.IsExcluded(depPath, depVersion) {
			continue
		}
		
		key := depPath + "@" + depVersion
		if visited[key] {
			continue
		}
		
		visited[key] = true
		newChain := append([]string(nil), currentChain...)
		newChain = append(newChain, key)
		
		if depPath == targetPath {
			result = append(result, newChain)
			continue
		}
		
		subChains, err := findPathRecursive(depPath, depVersion, targetPath, newChain, visited, mod, selected, provider)
		if err != nil {
			continue
		}
		
		if len(subChains) > 0 {
			result = append(result, subChains...)
		}
	}
	
	return result, nil
}
