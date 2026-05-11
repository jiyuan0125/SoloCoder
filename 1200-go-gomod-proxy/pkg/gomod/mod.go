package gomod

import (
	"bufio"
	"strings"
)

func ParseGoMod(content string) (*GoModFile, error) {
	mod := NewGoModFile()
	
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	var inBlock string
	var blockType string
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		if inBlock != "" {
			if line == ")" {
				inBlock = ""
				blockType = ""
				continue
			}
			
			if strings.HasPrefix(line, "//") {
				continue
			}
			
			parseBlockLine(line, blockType, mod)
			continue
		}
		
		if strings.HasPrefix(line, "module") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				mod.Module = parts[1]
			}
		} else if strings.HasPrefix(line, "go") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				mod.GoVersion = parts[1]
			}
		} else if strings.HasPrefix(line, "require (") {
			inBlock = "("
			blockType = "require"
		} else if strings.HasPrefix(line, "replace (") {
			inBlock = "("
			blockType = "replace"
		} else if strings.HasPrefix(line, "exclude (") {
			inBlock = "("
			blockType = "exclude"
		} else if strings.HasPrefix(line, "retract (") {
			inBlock = "("
			blockType = "retract"
		} else if strings.HasPrefix(line, "require") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				indirect := len(parts) >= 4 && parts[3] == "//" && strings.Contains(strings.Join(parts[4:], " "), "indirect")
				mod.Requires = append(mod.Requires, Require{
					Path:     parts[1],
					Version:  parts[2],
					Indirect: indirect,
				})
			}
		} else if strings.HasPrefix(line, "replace") {
			parseReplaceLine(line, mod)
		} else if strings.HasPrefix(line, "exclude") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				mod.Excludes = append(mod.Excludes, Exclude{
					Path:    parts[1],
					Version: parts[2],
				})
			}
		} else if strings.HasPrefix(line, "retract") {
			parseRetractLine(line, mod)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return mod, nil
}

func parseBlockLine(line, blockType string, mod *GoModFile) {
	switch blockType {
	case "require":
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			indirect := len(parts) >= 3 && parts[2] == "//" && strings.Contains(strings.Join(parts[3:], " "), "indirect")
			mod.Requires = append(mod.Requires, Require{
				Path:     parts[0],
				Version:  parts[1],
				Indirect: indirect,
			})
		}
	case "replace":
		parseReplaceLine("replace "+line, mod)
	case "exclude":
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			mod.Excludes = append(mod.Excludes, Exclude{
				Path:    parts[0],
				Version: parts[1],
			})
		}
	case "retract":
		parseRetractLine("retract "+line, mod)
	}
}

func parseReplaceLine(line string, mod *GoModFile) {
	line = strings.TrimSpace(strings.TrimPrefix(line, "replace"))
	
	var oldPath, oldVersion, newPath, newVersion string
	
	if strings.Contains(line, "=>") {
		parts := strings.Split(line, "=>")
		oldPart := strings.TrimSpace(parts[0])
		newPart := strings.TrimSpace(parts[1])
		
		oldFields := strings.Fields(oldPart)
		if len(oldFields) >= 1 {
			oldPath = oldFields[0]
			if len(oldFields) >= 2 {
				oldVersion = oldFields[1]
			}
		}
		
		newFields := strings.Fields(newPart)
		if len(newFields) >= 1 {
			newPath = newFields[0]
			if len(newFields) >= 2 {
				newVersion = newFields[1]
			}
		}
	}
	
	if oldPath != "" {
		mod.Replaces = append(mod.Replaces, Replace{
			OldPath:    oldPath,
			OldVersion: oldVersion,
			NewPath:    newPath,
			NewVersion: newVersion,
		})
	}
}

func parseRetractLine(line string, mod *GoModFile) {
	line = strings.TrimSpace(strings.TrimPrefix(line, "retract"))
	
	line = strings.TrimSpace(line)
	
	if strings.HasPrefix(line, "[") {
		line = strings.Trim(line, "[]")
		parts := strings.Split(line, ",")
		if len(parts) == 1 {
			mod.Retracts = append(mod.Retracts, Retract{
				LowVersion: strings.TrimSpace(parts[0]),
			})
		} else if len(parts) == 2 {
			mod.Retracts = append(mod.Retracts, Retract{
				LowVersion:  strings.TrimSpace(parts[0]),
				HighVersion: strings.TrimSpace(parts[1]),
			})
		}
	} else {
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			mod.Retracts = append(mod.Retracts, Retract{
				LowVersion: parts[0],
			})
		}
	}
}

func (m *GoModFile) ApplyReplace(path, version string) (string, string) {
	for _, r := range m.Replaces {
		if r.OldPath == path {
			if r.OldVersion == "" || r.OldVersion == version {
				if r.NewVersion == "" {
					return r.NewPath, version
				}
				return r.NewPath, r.NewVersion
			}
		}
	}
	return path, version
}

func (m *GoModFile) IsExcluded(path, version string) bool {
	for _, e := range m.Excludes {
		if e.Path == path && (e.Version == "" || e.Version == version) {
			return true
		}
	}
	return false
}

func (m *GoModFile) IsRetracted(path, version string) bool {
	for _, r := range m.Retracts {
		if r.Contains(version) {
			return true
		}
	}
	return false
}

func (m *GoModFile) GetDirectRequires() []Require {
	result := make([]Require, 0)
	for _, r := range m.Requires {
		if !r.Indirect {
			result = append(result, r)
		}
	}
	return result
}

func (m *GoModFile) GetIndirectRequires() []Require {
	result := make([]Require, 0)
	for _, r := range m.Requires {
		if r.Indirect {
			result = append(result, r)
		}
	}
	return result
}
