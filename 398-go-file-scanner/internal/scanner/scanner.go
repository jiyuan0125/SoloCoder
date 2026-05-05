package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"filescanner/internal/config"
	"filescanner/internal/protocol"
)

type Scanner struct {
	config *config.Config
}

type fileMatch struct {
	path    string
	matches []lineMatch
}

type lineMatch struct {
	lineNum  int
	content  string
	ruleName string
	severity config.SeverityLevel
}

func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config: cfg,
	}
}

func (s *Scanner) Scan(directory string, excludePaths []string, severityFilter string) (*protocol.ScanResponse, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if s.config.IsExcludedPath(path, excludePaths) {
			return nil
		}
		if config.IsBinaryFile(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	var wg sync.WaitGroup
	resultsChan := make(chan fileMatch, len(files))
	errorsChan := make(chan error, len(files))

	rules := s.config.GetRulesBySeverity(config.SeverityLevel(severityFilter))

	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			matches, err := s.scanFile(filePath, rules)
			if err != nil {
				errorsChan <- fmt.Errorf("error scanning %s: %w", filePath, err)
				return
			}
			if len(matches) > 0 {
				resultsChan <- fileMatch{
					path:    filePath,
					matches: matches,
				}
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	var scanResults []protocol.ScanResult
	totalMatches := 0

	for result := range resultsChan {
		sort.Slice(result.matches, func(i, j int) bool {
			return result.matches[i].lineNum < result.matches[j].lineNum
		})

		uniqueMatches := deduplicateMatches(result.matches)

		protocolMatches := make([]protocol.MatchResult, len(uniqueMatches))
		for i, m := range uniqueMatches {
			protocolMatches[i] = protocol.MatchResult{
				LineNumber:    m.lineNum,
				RuleName:      m.ruleName,
				Severity:      string(m.severity),
				MaskedContent: maskSensitiveContent(m.content),
			}
		}

		scanResults = append(scanResults, protocol.ScanResult{
			FilePath: result.path,
			Matches:  protocolMatches,
		})
		totalMatches += len(protocolMatches)
	}

	sort.Slice(scanResults, func(i, j int) bool {
		return scanResults[i].FilePath < scanResults[j].FilePath
	})

	var scanErrors []string
	for err := range errorsChan {
		scanErrors = append(scanErrors, err.Error())
	}

	summary := protocol.ScanSummary{
		TotalFilesScanned: len(files),
		TotalFilesMatched: len(scanResults),
		TotalMatchesFound: totalMatches,
	}

	response := &protocol.ScanResponse{
		Results: scanResults,
		Summary: summary,
	}

	if len(scanErrors) > 0 {
		response.Error = strings.Join(scanErrors, "; ")
	}

	return response, nil
}

func (s *Scanner) scanFile(filePath string, rules []config.Rule) ([]lineMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []lineMatch
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, rule := range rules {
			if rule.Compiled == nil {
				continue
			}

			if rule.Compiled.MatchString(line) {
				matches = append(matches, lineMatch{
					lineNum:  lineNum,
					content:  line,
					ruleName: rule.Name,
					severity: rule.Severity,
				})
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

func deduplicateMatches(matches []lineMatch) []lineMatch {
	if len(matches) == 0 {
		return matches
	}

	var unique []lineMatch
	seen := make(map[int]bool)

	for _, m := range matches {
		if !seen[m.lineNum] {
			seen[m.lineNum] = true
			unique = append(unique, m)
		}
	}

	return unique
}

func maskSensitiveContent(content string) string {
	maskers := []func(string) string{
		maskPasswordPattern,
		maskAPIKeyPattern,
		maskSecretPattern,
		maskDBConnectionPattern,
		maskMySQLConnectionPattern,
		maskAWSKeyPattern,
		maskGitHubTokenPattern,
		maskAuthHeaderPattern,
		maskCommentCredentialPattern,
	}

	result := content
	for _, masker := range maskers {
		result = masker(result)
	}

	return result
}

func maskPasswordPattern(content string) string {
	re := regexp.MustCompile(`(?i)(password|passwd|pwd)(\s*[=:]\s*['"]?)([^\s'";,]+)(['"]?)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		return parts[1] + parts[2] + maskValue(parts[3]) + parts[4]
	})
}

func maskAPIKeyPattern(content string) string {
	re := regexp.MustCompile(`(?i)(api[_-]?key|apikey)(\s*[=:]\s*['"]?)([^\s'";,]+)(['"]?)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		return parts[1] + parts[2] + maskValue(parts[3]) + parts[4]
	})
}

func maskSecretPattern(content string) string {
	re := regexp.MustCompile(`(?i)(secret|token)(\s*[=:]\s*['"]?)([^\s'";,]+)(['"]?)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		return parts[1] + parts[2] + maskValue(parts[3]) + parts[4]
	})
}

func maskDBConnectionPattern(content string) string {
	re := regexp.MustCompile(`(?i)(mysql|postgres|postgresql|mongodb|mssql|oracle)://([^:@]+):([^@]+)@`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		return parts[1] + "://" + maskValue(parts[2]) + ":" + maskValue(parts[3]) + "@"
	})
}

func maskMySQLConnectionPattern(content string) string {
	re := regexp.MustCompile(`(?i)([^:\s]+):([^@\s]+)@tcp\([^)]+\)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		return maskValue(parts[1]) + ":" + maskValue(parts[2]) + "@" + match[strings.Index(match, "@tcp"):]
	})
}

func maskAWSKeyPattern(content string) string {
	re := regexp.MustCompile(`(?i)(aws_access_key_id|aws_secret_access_key)(\s*[=:]\s*['"]?)([^\s'";,]+)(['"]?)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		return parts[1] + parts[2] + maskValue(parts[3]) + parts[4]
	})
}

func maskGitHubTokenPattern(content string) string {
	re := regexp.MustCompile(`(?i)gh[ps]_[a-zA-Z0-9]{36,}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		return maskValue(match)
	})
}

func maskAuthHeaderPattern(content string) string {
	re := regexp.MustCompile(`(?i)(authorization|auth)(\s*:\s*(bearer|basic|token)\s+)([^\s'";,]+)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 5 {
			return match
		}
		return parts[1] + parts[2] + maskValue(parts[4])
	})
}

func maskCommentCredentialPattern(content string) string {
	re := regexp.MustCompile(`(?i)(//\s*(password|api[_-]?key|secret|token)\s*[:=]\s*)([^\s'";,]+)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		return parts[1] + maskValue(parts[3])
	})
}

func maskValue(value string) string {
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
