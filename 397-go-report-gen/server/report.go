package main

import (
	"fmt"
	"strings"
	"time"

	"go-report-gen/protocol"
)

func GenerateReport(repoPath, since, author, format string) (*protocol.Report, error) {
	commits, err := GetCommits(repoPath, since, author)
	if err != nil {
		return nil, err
	}

	groups := GroupCommitsByDate(commits)
	stats := CalculateStats(commits, groups)
	reportID := generateReportID()

	report := &protocol.Report{
		ID:           reportID,
		RepoPath:     repoPath,
		Since:        since,
		Author:       author,
		GeneratedAt:  time.Now(),
		Stats:        stats,
		DailyCommits: groups,
	}

	if stats.TotalCommits == 0 {
		report.Content = "该时间段内无提交记录"
		report.ContentMarkdown = "该时间段内无提交记录"
		return report, nil
	}

	report.Content = generateTextReport(report, groups)
	if format == "markdown" {
		report.ContentMarkdown = generateMarkdownReport(report, groups)
	}

	return report, nil
}

func generateReportID() string {
	return fmt.Sprintf("report_%s", time.Now().Format("20060102_150405"))
}

func generateTextReport(report *protocol.Report, groups map[string][]*protocol.Commit) string {
	var sb strings.Builder

	sb.WriteString("========================================\n")
	sb.WriteString("            工作日报/周报\n")
	sb.WriteString("========================================\n")
	sb.WriteString(fmt.Sprintf("仓库路径: %s\n", report.RepoPath))
	sb.WriteString(fmt.Sprintf("时间范围: 最近 %s\n", report.Since))
	if report.Author != "" {
		sb.WriteString(fmt.Sprintf("作者: %s\n", report.Author))
	}
	sb.WriteString(fmt.Sprintf("生成时间: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("========================================\n\n")

	sortedDates := GetSortedDates(groups)

	for _, date := range sortedDates {
		commits := groups[date]
		sb.WriteString(fmt.Sprintf("【%s】 共 %d 个提交\n", date, len(commits)))
		sb.WriteString("----------------------------------------\n")

		for i, commit := range commits {
			sb.WriteString(fmt.Sprintf("\n  提交 %d:\n", i+1))
			sb.WriteString(fmt.Sprintf("    哈希: %s\n", commit.ShortHash))
			sb.WriteString(fmt.Sprintf("    作者: %s\n", commit.Author))
			sb.WriteString(fmt.Sprintf("    时间: %s\n", commit.Date.Format("2006-01-02 15:04:05")))
			sb.WriteString(fmt.Sprintf("    信息: %s\n", commit.ShortMessage))

			if len(commit.Issues) > 0 {
				sb.WriteString(fmt.Sprintf("    关联 Issue: %s\n", strings.Join(commit.Issues, ", ")))
			}

			if len(commit.Files) > 0 {
				sb.WriteString(fmt.Sprintf("    修改文件: %d 个\n", len(commit.Files)))
			}
		}

		sb.WriteString("\n")
	}

	sb.WriteString("========================================\n")
	sb.WriteString("              统计信息\n")
	sb.WriteString("========================================\n")
	sb.WriteString(fmt.Sprintf("总提交数: %d\n", report.Stats.TotalCommits))
	sb.WriteString(fmt.Sprintf("修改文件数: %d\n", report.Stats.TotalFiles))
	sb.WriteString(fmt.Sprintf("有提交的天数: %d\n", report.Stats.DaysWithCommits))
	sb.WriteString("========================================\n")

	return sb.String()
}

func generateMarkdownReport(report *protocol.Report, groups map[string][]*protocol.Commit) string {
	var sb strings.Builder

	sb.WriteString("# 工作日报/周报\n\n")
	sb.WriteString("## 基本信息\n\n")
	sb.WriteString(fmt.Sprintf("- **仓库路径**: %s\n", report.RepoPath))
	sb.WriteString(fmt.Sprintf("- **时间范围**: 最近 %s\n", report.Since))
	if report.Author != "" {
		sb.WriteString(fmt.Sprintf("- **作者**: %s\n", report.Author))
	}
	sb.WriteString(fmt.Sprintf("- **生成时间**: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))

	sortedDates := GetSortedDates(groups)

	for _, date := range sortedDates {
		commits := groups[date]
		sb.WriteString(fmt.Sprintf("## %s (共 %d 个提交)\n\n", date, len(commits)))

		for i, commit := range commits {
			sb.WriteString(fmt.Sprintf("### 提交 %d\n\n", i+1))
			sb.WriteString(fmt.Sprintf("- **哈希**: `%s`\n", commit.ShortHash))
			sb.WriteString(fmt.Sprintf("- **作者**: %s\n", commit.Author))
			sb.WriteString(fmt.Sprintf("- **时间**: %s\n", commit.Date.Format("2006-01-02 15:04:05")))
			sb.WriteString(fmt.Sprintf("- **信息**: %s\n", commit.ShortMessage))

			if len(commit.Issues) > 0 {
				sb.WriteString(fmt.Sprintf("- **关联 Issue**: %s\n", strings.Join(commit.Issues, ", ")))
			}

			if len(commit.Files) > 0 {
				sb.WriteString(fmt.Sprintf("- **修改文件**: %d 个\n", len(commit.Files)))
			}

			sb.WriteString("\n")
		}
	}

	sb.WriteString("## 统计信息\n\n")
	sb.WriteString(fmt.Sprintf("- **总提交数**: %d\n", report.Stats.TotalCommits))
	sb.WriteString(fmt.Sprintf("- **修改文件数**: %d\n", report.Stats.TotalFiles))
	sb.WriteString(fmt.Sprintf("- **有提交的天数**: %d\n", report.Stats.DaysWithCommits))

	return sb.String()
}

func SaveReportToFile(report *protocol.Report, filePath string) error {
	return nil
}
