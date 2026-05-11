package gomod

import (
	"bufio"
	"strings"
)

func ParseGoSum(content string) ([]GoSumEntry, error) {
	entries := make([]GoSumEntry, 0)
	
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		parts := strings.Fields(line)
		
		if len(parts) >= 2 {
			path := parts[0]
			version := parts[1]
			hash := ""
			
			if len(parts) >= 3 {
				hash = parts[2]
			}
			
			version = strings.TrimSuffix(version, "/go.mod")
			
			entries = append(entries, GoSumEntry{
				Path:    path,
				Version: version,
				Hash:    hash,
			})
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return entries, nil
}

func BuildSumMap(entries []GoSumEntry) map[string]map[string]GoSumEntry {
	result := make(map[string]map[string]GoSumEntry)
	
	for _, entry := range entries {
		if result[entry.Path] == nil {
			result[entry.Path] = make(map[string]GoSumEntry)
		}
		result[entry.Path][entry.Version] = entry
	}
	
	return result
}

func (m *GoModFile) VerifyGoSum(sumEntries []GoSumEntry) (*VerifyResult, error) {
	result := &VerifyResult{
		AllChecked: true,
		Orphans:    make([]GoSumEntry, 0),
		Mismatches: make([]Mismatch, 0),
	}
	
	required := make(map[string]map[string]bool)
	
	for _, r := range m.Requires {
		path, version := m.ApplyReplace(r.Path, r.Version)
		if m.IsExcluded(path, version) {
			continue
		}
		
		if required[path] == nil {
			required[path] = make(map[string]bool)
		}
		required[path][version] = true
	}
	
	sumMap := BuildSumMap(sumEntries)
	
	for path, versions := range sumMap {
		for version := range versions {
			if required[path] == nil || !required[path][version] {
				result.Orphans = append(result.Orphans, GoSumEntry{
					Path:    path,
					Version: version,
					Hash:    sumMap[path][version].Hash,
				})
				result.AllChecked = false
			}
		}
	}
	
	for path, versions := range required {
		for version := range versions {
			if sumMap[path] == nil || sumMap[path][version].Hash == "" {
				result.Mismatches = append(result.Mismatches, Mismatch{
					Path:    path,
					Version: version,
					Error:   "missing checksum in go.sum",
				})
				result.AllChecked = false
			}
		}
	}
	
	return result, nil
}

func FindOrphanSumEntries(sumEntries []GoSumEntry, selected map[string]string) []GoSumEntry {
	orphans := make([]GoSumEntry, 0)
	
	for _, entry := range sumEntries {
		selectedVersion, exists := selected[entry.Path]
		if !exists || selectedVersion != entry.Version {
			orphans = append(orphans, entry)
		}
	}
	
	return orphans
}
