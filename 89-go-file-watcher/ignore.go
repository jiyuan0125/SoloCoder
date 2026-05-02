package filewatcher

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type IgnoreConfig struct {
	IncludeHidden bool
}

type IgnorePattern struct {
	Pattern    string
	IsNegation bool
	IsDirOnly  bool
	BaseDir    string
}

type IgnoreMatcher struct {
	config   IgnoreConfig
	patterns []IgnorePattern
}

func NewIgnoreMatcher(config IgnoreConfig) *IgnoreMatcher {
	return &IgnoreMatcher{
		config:   config,
		patterns: []IgnorePattern{},
	}
}

func (m *IgnoreMatcher) LoadGitignore(dir string) error {
	currentDir := filepath.Clean(dir)
	for {
		gitignorePath := filepath.Join(currentDir, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			if err := m.loadGitignoreFile(gitignorePath, currentDir); err != nil {
				return err
			}
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}

	return nil
}

func (m *IgnoreMatcher) loadGitignoreFile(path string, baseDir string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		isNegation := false
		if strings.HasPrefix(line, "!") {
			isNegation = true
			line = strings.TrimPrefix(line, "!")
		}

		isDirOnly := false
		if strings.HasSuffix(line, "/") {
			isDirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		pattern := IgnorePattern{
			Pattern:    line,
			IsNegation: isNegation,
			IsDirOnly:  isDirOnly,
			BaseDir:    baseDir,
		}

		m.patterns = append(m.patterns, pattern)
	}

	return scanner.Err()
}

func (m *IgnoreMatcher) ShouldIgnore(path string, isDir bool) bool {
	if !m.config.IncludeHidden {
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && len(base) > 1 {
			return true
		}
	}

	matched := false
	for _, pattern := range m.patterns {
		if m.matchPattern(pattern, path, isDir) {
			if pattern.IsNegation {
				matched = false
			} else {
				matched = true
			}
		}
	}

	return matched
}

func (m *IgnoreMatcher) matchPattern(pattern IgnorePattern, path string, isDir bool) bool {
	if pattern.IsDirOnly && !isDir {
		return false
	}

	relPath, err := filepath.Rel(pattern.BaseDir, path)
	if err != nil {
		return false
	}
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	patternParts := strings.Split(pattern.Pattern, string(filepath.Separator))
	pathParts := strings.Split(relPath, string(filepath.Separator))

	if len(patternParts) == 1 {
		for _, part := range pathParts {
			if matchGlob(pattern.Pattern, part) {
				return true
			}
		}
		return false
	}

	return matchPatternParts(patternParts, pathParts)
}

func matchPatternParts(patternParts, pathParts []string) bool {
	if len(patternParts) > len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if patternPart == "**" {
			if i == len(patternParts)-1 {
				return true
			}
			remainingPattern := patternParts[i+1:]
			for j := i; j < len(pathParts); j++ {
				if matchPatternParts(remainingPattern, pathParts[j:]) {
					return true
				}
			}
			return false
		}

		if !matchGlob(patternPart, pathParts[i]) {
			return false
		}
	}

	return true
}

func matchGlob(pattern, name string) bool {
	patternParts := splitPattern(pattern)
	nameParts := []rune(name)

	return matchParts(patternParts, nameParts)
}

func splitPattern(pattern string) []rune {
	return []rune(pattern)
}

func matchParts(pattern []rune, name []rune) bool {
	patternIdx := 0
	nameIdx := 0

	for patternIdx < len(pattern) || nameIdx < len(name) {
		if patternIdx < len(pattern) {
			switch pattern[patternIdx] {
			case '*':
				patternIdx++
				for i := nameIdx; i <= len(name); i++ {
					if matchParts(pattern[patternIdx:], name[i:]) {
						return true
					}
				}
				return false
			case '?':
				if nameIdx < len(name) {
					patternIdx++
					nameIdx++
				} else {
					return false
				}
			default:
				if nameIdx < len(name) && pattern[patternIdx] == name[nameIdx] {
					patternIdx++
					nameIdx++
				} else {
					return false
				}
			}
		} else {
			return false
		}
	}

	return patternIdx == len(pattern) && nameIdx == len(name)
}
