package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go-report-gen/protocol"
)

const (
	gitDateFormat = "2006-01-02 15:04:05 -0700"
)

var (
	issuePattern = regexp.MustCompile(`#(\d+)`)
)

func IsGitRepo(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func GetCommits(repoPath, since, author string) ([]*protocol.Commit, error) {
	if !IsGitRepo(repoPath) {
		return nil, fmt.Errorf("指定的路径不是一个 Git 仓库: %s", repoPath)
	}

	args := []string{
		"log",
		"--pretty=format:%H|%h|%s|%an|%ai",
		"--name-status",
		"--since=" + since,
	}

	if author != "" {
		args = append(args, "--author="+author)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("执行 git log 失败: %w", err)
	}

	return parseGitLogOutput(out.String())
}

func parseGitLogOutput(output string) ([]*protocol.Commit, error) {
	if output == "" {
		return []*protocol.Commit{}, nil
	}

	var commits []*protocol.Commit
	scanner := bufio.NewScanner(strings.NewReader(output))

	var currentCommit *protocol.Commit
	var commitFiles []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "|") && !strings.HasPrefix(line, "A\t") && !strings.HasPrefix(line, "M\t") && 
		   !strings.HasPrefix(line, "D\t") && !strings.HasPrefix(line, "R\t") && !strings.HasPrefix(line, "C\t") &&
		   !strings.HasPrefix(line, "U\t") && !strings.HasPrefix(line, "T\t") && !strings.HasPrefix(line, "X\t") {
			
			if currentCommit != nil {
				currentCommit.Files = commitFiles
				commits = append(commits, currentCommit)
			}

			parts := strings.SplitN(line, "|", 5)
			if len(parts) != 5 {
				continue
			}

			hash := parts[0]
			shortHash := parts[1]
			message := parts[2]
			commitAuthor := parts[3]
			dateStr := parts[4]

			date, err := time.Parse(gitDateFormat, dateStr)
			if err != nil {
				continue
			}

			isMerge := strings.HasPrefix(message, "Merge ")
			issues := extractIssues(message)
			shortMessage := truncateMessage(message, 100)

			currentCommit = &protocol.Commit{
				Hash:         hash,
				ShortHash:    shortHash,
				Message:      message,
				ShortMessage: shortMessage,
				Author:       commitAuthor,
				Date:         date,
				DateStr:      dateStr,
				Issues:       issues,
				IsMerge:      isMerge,
			}

			commitFiles = []string{}
		} else if currentCommit != nil && line != "" {
			fileParts := strings.SplitN(line, "\t", 2)
			if len(fileParts) >= 2 {
				filePath := fileParts[1]
				if !contains(commitFiles, filePath) {
					commitFiles = append(commitFiles, filePath)
				}
			}
		}
	}

	if currentCommit != nil {
		currentCommit.Files = commitFiles
		commits = append(commits, currentCommit)
	}

	return filterAndSortCommits(commits), nil
}

func filterAndSortCommits(commits []*protocol.Commit) []*protocol.Commit {
	var filtered []*protocol.Commit
	for _, commit := range commits {
		if !commit.IsMerge {
			filtered = append(filtered, commit)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date.After(filtered[j].Date)
	})

	return filtered
}

func extractIssues(message string) []string {
	matches := issuePattern.FindAllStringSubmatch(message, -1)
	if len(matches) == 0 {
		return nil
	}

	var issues []string
	for _, match := range matches {
		if len(match) > 1 {
			issues = append(issues, match[0])
		}
	}

	return issues
}

func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength] + "..."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GroupCommitsByDate(commits []*protocol.Commit) map[string][]*protocol.Commit {
	groups := make(map[string][]*protocol.Commit)
	
	for _, commit := range commits {
		dateKey := commit.Date.Format("2006-01-02")
		groups[dateKey] = append(groups[dateKey], commit)
	}

	return groups
}

func GetSortedDates(groups map[string][]*protocol.Commit) []string {
	dates := make([]string, 0, len(groups))
	for date := range groups {
		dates = append(dates, date)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates
}

func CalculateStats(commits []*protocol.Commit, groups map[string][]*protocol.Commit) protocol.ReportStats {
	uniqueFiles := make(map[string]struct{})
	for _, commit := range commits {
		for _, file := range commit.Files {
			uniqueFiles[file] = struct{}{}
		}
	}

	return protocol.ReportStats{
		TotalCommits:    len(commits),
		TotalFiles:      len(uniqueFiles),
		DaysWithCommits: len(groups),
	}
}
