package main

import (
	"fmt"
	"strings"
	"time"

	"tag-manager/common"
)

func formatTag(t common.Tag) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:          %s\n", t.ID))
	sb.WriteString(fmt.Sprintf("Name:        %s\n", t.Name))
	sb.WriteString(fmt.Sprintf("Type:        %s\n", t.Type))
	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", t.Description))
	}
	if t.GroupID != "" {
		sb.WriteString(fmt.Sprintf("Group ID:    %s\n", t.GroupID))
	}
	sb.WriteString(fmt.Sprintf("Created:     %s\n", formatTime(t.CreatedAt)))
	sb.WriteString(fmt.Sprintf("Updated:     %s\n", formatTime(t.UpdatedAt)))

	switch t.Type {
	case common.TagTypeEnum:
		if t.EnumConfig != nil {
			sb.WriteString(fmt.Sprintf("Enum Values: %v\n", t.EnumConfig.Values))
		}
	case common.TagTypeNumber:
		if t.NumberConfig != nil {
			if t.NumberConfig.Min != nil {
				sb.WriteString(fmt.Sprintf("Min Value:   %v\n", *t.NumberConfig.Min))
			}
			if t.NumberConfig.Max != nil {
				sb.WriteString(fmt.Sprintf("Max Value:   %v\n", *t.NumberConfig.Max))
			}
		}
	case common.TagTypeText:
		if t.TextConfig != nil {
			sb.WriteString(fmt.Sprintf("Max Length:  %d\n", t.TextConfig.MaxLength))
		}
	}

	return sb.String()
}

func formatTagList(tags []common.Tag) string {
	if len(tags) == 0 {
		return "No tags found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tag(s):\n\n", len(tags)))

	for i, t := range tags {
		if i > 0 {
			sb.WriteString("---\n")
		}
		sb.WriteString(formatTag(t))
	}

	return sb.String()
}

func formatTagGroup(g common.TagGroup) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:          %s\n", g.ID))
	sb.WriteString(fmt.Sprintf("Name:        %s\n", g.Name))
	if g.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", g.Description))
	}
	sb.WriteString(fmt.Sprintf("Created:     %s\n", formatTime(g.CreatedAt)))
	return sb.String()
}

func formatTagGroupList(groups []common.TagGroup) string {
	if len(groups) == 0 {
		return "No groups found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d group(s):\n\n", len(groups)))

	for i, g := range groups {
		if i > 0 {
			sb.WriteString("---\n")
		}
		sb.WriteString(formatTagGroup(g))
	}

	return sb.String()
}

func formatUserTagList(tags []common.UserTagWithInfo) string {
	if len(tags) == 0 {
		return "No tags for this user"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tag(s) for this user:\n\n", len(tags)))

	for i, ut := range tags {
		if i > 0 {
			sb.WriteString("---\n")
		}
		sb.WriteString(fmt.Sprintf("Tag ID:   %s\n", ut.TagID))
		sb.WriteString(fmt.Sprintf("Tag Name: %s\n", ut.TagName))
		sb.WriteString(fmt.Sprintf("Type:     %s\n", ut.TagType))
		sb.WriteString(fmt.Sprintf("Value:    %v\n", ut.TagValue))
		sb.WriteString(fmt.Sprintf("Applied:  %s\n", formatTime(ut.CreatedAt)))
	}

	return sb.String()
}

func formatStats(stats []common.TagUsageStats) string {
	if len(stats) == 0 {
		return "No usage stats available"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tag Usage Statistics (%d tags):\n\n", len(stats)))

	maxNameLen := 0
	for _, s := range stats {
		if len(s.TagName) > maxNameLen {
			maxNameLen = len(s.TagName)
		}
	}

	formatStr := fmt.Sprintf("%%-%ds  %%6d users\n", maxNameLen)
	for _, s := range stats {
		sb.WriteString(fmt.Sprintf(formatStr, s.TagName, s.UserCount))
	}

	return sb.String()
}

func formatTrend(trends []common.TagTrend) string {
	if len(trends) == 0 {
		return "No trend data available"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tag Trend (Last 30 Days, %d tags):\n\n", len(trends)))

	maxNameLen := 0
	for _, t := range trends {
		if len(t.TagName) > maxNameLen {
			maxNameLen = len(t.TagName)
		}
	}

	formatStr := fmt.Sprintf("%%-%ds  %%6d new users\n", maxNameLen)
	for _, t := range trends {
		sb.WriteString(fmt.Sprintf(formatStr, t.TagName, t.NewUserCount))
	}

	return sb.String()
}

func formatAuditLogs(logs []common.AuditLog) string {
	if len(logs) == 0 {
		return "No audit logs"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Audit Logs (%d entries):\n\n", len(logs)))

	for i, log := range logs {
		if i > 0 {
			sb.WriteString("---\n")
		}
		sb.WriteString(fmt.Sprintf("Time:     %s\n", formatTime(log.Timestamp)))
		sb.WriteString(fmt.Sprintf("Operator: %s\n", log.Operator))
		sb.WriteString(fmt.Sprintf("Action:   %s\n", log.Action))
		sb.WriteString(fmt.Sprintf("Target:   %s\n", log.Target))
		if log.Details != "" && log.Details != "{}" {
			sb.WriteString(fmt.Sprintf("Details:  %s\n", log.Details))
		}
	}

	return sb.String()
}

func formatFilterResult(userIDs []string) string {
	if len(userIDs) == 0 {
		return "No users matched the filter"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d user(s):\n", len(userIDs)))

	for i, id := range userIDs {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, id))
	}

	return sb.String()
}

func formatTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
