package analyzer

import (
	"bufio"
	"regexp"
	"strings"
)

func ParseGoSum(content string) ([]GoSumEntry, error) {
	var entries []GoSumEntry

	sumRegex := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)$`)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		match := sumRegex.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		modulePath := match[1]
		versionPart := match[2]
		hash := match[3]

		isMod := false
		version := versionPart
		if strings.HasSuffix(versionPart, "/go.mod") {
			isMod = true
			version = strings.TrimSuffix(versionPart, "/go.mod")
		}

		entries = append(entries, GoSumEntry{
			Module:  modulePath,
			Version: version,
			Hash:    hash,
			IsMod:   isMod,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func GetModuleVersionsFromSum(entries []GoSumEntry) map[string]string {
	moduleVersions := make(map[string]string)
	moduleVersionSet := make(map[string]map[string]bool)

	for _, entry := range entries {
		if moduleVersionSet[entry.Module] == nil {
			moduleVersionSet[entry.Module] = make(map[string]bool)
		}
		moduleVersionSet[entry.Module][entry.Version] = true
	}

	for module, versions := range moduleVersionSet {
		var latest string
		for v := range versions {
			if latest == "" || compareVersions(v, latest) > 0 {
				latest = v
			}
		}
		moduleVersions[module] = latest
	}

	return moduleVersions
}

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] < parts2[i] {
			return -1
		}
		if parts1[i] > parts2[i] {
			return 1
		}
	}
	if len(parts1) < len(parts2) {
		return -1
	}
	if len(parts1) > len(parts2) {
		return 1
	}
	return 0
}
