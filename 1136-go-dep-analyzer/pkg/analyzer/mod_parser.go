package analyzer

import (
	"bufio"
	"regexp"
	"strings"
)

func ParseGoMod(content string) (*GoModFile, error) {
	mod := &GoModFile{
		Requires: []Require{},
		Excludes: []Exclude{},
		Replaces: []Replace{},
		Retracts: []Retract{},
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentBlock string

	moduleRegex := regexp.MustCompile(`^module\s+(\S+)`)
	goVersionRegex := regexp.MustCompile(`^go\s+(\S+)`)
	requireSingleRegex := regexp.MustCompile(`^require\s+(\S+)\s+(\S+)(?:\s+//\s*(indirect))?$`)
	excludeSingleRegex := regexp.MustCompile(`^exclude\s+(\S+)\s+(\S+)`)
	replaceSingleRegex := regexp.MustCompile(`^replace\s+(\S+)(?:\s+(\S+))?\s*=>\s*(\S+)(?:\s+(\S+))?$`)
	retractSingleRegex := regexp.MustCompile(`^retract\s+(\S+)`)

	blockEntryRegex := regexp.MustCompile(`^\s*(\S+)\s+(\S+)(?:\s+//\s*(indirect))?$`)
	blockReplaceEntryRegex := regexp.MustCompile(`^\s*(\S+)(?:\s+(\S+))?\s*=>\s*(\S+)(?:\s+(\S+))?$`)
	blockExcludeEntryRegex := regexp.MustCompile(`^\s*(\S+)\s+(\S+)$`)
	blockRetractEntryRegex := regexp.MustCompile(`^\s*(\S+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if line == "(" {
			continue
		}
		if line == ")" {
			currentBlock = ""
			continue
		}

		if match := moduleRegex.FindStringSubmatch(line); match != nil {
			mod.Module = match[1]
			continue
		}

		if match := goVersionRegex.FindStringSubmatch(line); match != nil {
			mod.GoVersion = match[1]
			continue
		}

		if strings.HasPrefix(line, "require (") {
			currentBlock = "require"
			continue
		}
		if strings.HasPrefix(line, "exclude (") {
			currentBlock = "exclude"
			continue
		}
		if strings.HasPrefix(line, "replace (") {
			currentBlock = "replace"
			continue
		}
		if strings.HasPrefix(line, "retract (") {
			currentBlock = "retract"
			continue
		}

		if currentBlock != "" {
			switch currentBlock {
			case "require":
				if match := blockEntryRegex.FindStringSubmatch(line); match != nil {
					mod.Requires = append(mod.Requires, Require{
						Path:       match[1],
						Version:    match[2],
						IsIndirect: match[3] == "indirect",
					})
				}
			case "exclude":
				if match := blockExcludeEntryRegex.FindStringSubmatch(line); match != nil {
					mod.Excludes = append(mod.Excludes, Exclude{
						Path:    match[1],
						Version: match[2],
					})
				}
			case "replace":
				if match := blockReplaceEntryRegex.FindStringSubmatch(line); match != nil {
					isLocal := !strings.HasPrefix(match[3], "./") && !strings.HasPrefix(match[3], "../") && !strings.Contains(match[3], "/") == false
					isLocal = strings.HasPrefix(match[3], "./") || strings.HasPrefix(match[3], "../")
					mod.Replaces = append(mod.Replaces, Replace{
						OldPath:    match[1],
						OldVersion: match[2],
						NewPath:    match[3],
						NewVersion: match[4],
						IsLocal:    isLocal,
					})
				}
			case "retract":
				if match := blockRetractEntryRegex.FindStringSubmatch(line); match != nil {
					mod.Retracts = append(mod.Retracts, Retract{
						Version: match[1],
					})
				}
			}
			continue
		}

		if match := requireSingleRegex.FindStringSubmatch(line); match != nil {
			mod.Requires = append(mod.Requires, Require{
				Path:       match[1],
				Version:    match[2],
				IsIndirect: match[3] == "indirect",
			})
			continue
		}

		if match := excludeSingleRegex.FindStringSubmatch(line); match != nil {
			mod.Excludes = append(mod.Excludes, Exclude{
				Path:    match[1],
				Version: match[2],
			})
			continue
		}

		if match := replaceSingleRegex.FindStringSubmatch(line); match != nil {
			isLocal := strings.HasPrefix(match[3], "./") || strings.HasPrefix(match[3], "../")
			mod.Replaces = append(mod.Replaces, Replace{
				OldPath:    match[1],
				OldVersion: match[2],
				NewPath:    match[3],
				NewVersion: match[4],
				IsLocal:    isLocal,
			})
			continue
		}

		if match := retractSingleRegex.FindStringSubmatch(line); match != nil {
			mod.Retracts = append(mod.Retracts, Retract{
				Version: match[1],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return mod, nil
}
