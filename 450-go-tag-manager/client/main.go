package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"tag-manager/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	client := NewClient("http://localhost:8080")

	switch command {
	case "help", "-h", "--help":
		printUsage()

	case "tag-create":
		cmdTagCreate(client, args)

	case "tag-list":
		cmdTagList(client, args)

	case "tag-get":
		cmdTagGet(client, args)

	case "tag-update":
		cmdTagUpdate(client, args)

	case "tag-delete":
		cmdTagDelete(client, args)

	case "tag-impact":
		cmdTagImpact(client, args)

	case "group-create":
		cmdGroupCreate(client, args)

	case "group-list":
		cmdGroupList(client, args)

	case "group-get":
		cmdGroupGet(client, args)

	case "tag-apply":
		cmdTagApply(client, args)

	case "tag-batch-apply":
		cmdTagBatchApply(client, args)

	case "tag-remove":
		cmdTagRemove(client, args)

	case "user-tags":
		cmdUserTags(client, args)

	case "users-filter":
		cmdUsersFilter(client, args)

	case "stats":
		cmdStats(client, args)

	case "trend":
		cmdTrend(client, args)

	case "audit":
		cmdAudit(client, args)

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Tag Manager Client")
	fmt.Println()
	fmt.Println("Usage: tag-client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Tag Management:")
	fmt.Println("  tag-create         Create a new tag")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -name=TAG        Tag name (unique)")
	fmt.Println("    -type=TYPE       Type: enum, number, text")
	fmt.Println("    -desc=TEXT       Description")
	fmt.Println("    -group=ID        Group ID")
	fmt.Println("    -values=LIST     Enum values (comma-separated)")
	fmt.Println("    -min=NUMBER      Number min value")
	fmt.Println("    -max=NUMBER      Number max value")
	fmt.Println("    -maxlen=NUMBER   Text max length (default 200)")
	fmt.Println()
	fmt.Println("  tag-list           List all tags")
	fmt.Println("    -group=ID        Filter by group ID")
	fmt.Println()
	fmt.Println("  tag-get            Get tag details")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println()
	fmt.Println("  tag-update         Update a tag")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println("    -name=TAG        New tag name")
	fmt.Println("    -desc=TEXT       New description")
	fmt.Println("    -group=ID        New group ID")
	fmt.Println("    -values=LIST     New enum values")
	fmt.Println()
	fmt.Println("  tag-delete         Delete a tag (cascade)")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println()
	fmt.Println("  tag-impact         Check tag deletion impact")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println()
	fmt.Println("Group Management:")
	fmt.Println("  group-create       Create a tag group")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -name=NAME       Group name")
	fmt.Println("    -desc=TEXT       Description")
	fmt.Println()
	fmt.Println("  group-list         List all groups")
	fmt.Println()
	fmt.Println("  group-get          Get group details")
	fmt.Println("    -id=GROUPID      Group ID")
	fmt.Println()
	fmt.Println("User Tagging:")
	fmt.Println("  tag-apply          Apply tag to a user")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println("    -user=USERID     User ID")
	fmt.Println("    -value=VALUE     Tag value")
	fmt.Println()
	fmt.Println("  tag-batch-apply    Batch apply tag to users")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println("    -users=LIST      User IDs (comma-separated, max 100)")
	fmt.Println("    -value=VALUE     Tag value")
	fmt.Println()
	fmt.Println("  tag-remove         Remove tag from a user")
	fmt.Println("    -operator=NAME   Operator name")
	fmt.Println("    -id=TAGID        Tag ID")
	fmt.Println("    -user=USERID     User ID")
	fmt.Println()
	fmt.Println("  user-tags          Get user's tags")
	fmt.Println("    -user=USERID     User ID")
	fmt.Println()
	fmt.Println("Filtering:")
	fmt.Println("  users-filter       Filter users by tags (JSON filter)")
	fmt.Println("    -json=JSON       Filter JSON (see below)")
	fmt.Println()
	fmt.Println("Statistics:")
	fmt.Println("  stats              Show tag usage statistics")
	fmt.Println()
	fmt.Println("  trend              Show tag trend (last 30 days)")
	fmt.Println()
	fmt.Println("  audit              Show audit logs")
	fmt.Println()
	fmt.Println("Filter JSON Format:")
	fmt.Println(`  {
    "logic": "AND",
    "groups": [
      {
        "logic": "OR",
        "conditions": [
          {"tag_id": "tag-1", "operator": "=", "value": "vip"}
        ]
      },
      {
        "logic": "AND",
        "conditions": [
          {"tag_id": "tag-2", "operator": ">", "value": 100},
          {"tag_id": "tag-3", "operator": "contains", "value": "test"}
        ]
      }
    ]
  }`)
	fmt.Println()
	fmt.Println("Operators by type:")
	fmt.Println("  enum:   =, ==, eq, !=, ne")
	fmt.Println("  number: =, ==, eq, !=, ne, >, gt, >=, gte, <, lt, <=, lte")
	fmt.Println("  text:   =, ==, eq, !=, ne, contains, starts_with, ends_with")
	fmt.Println()
}

func cmdTagCreate(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-create", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	name := fs.String("name", "", "Tag name")
	tagType := fs.String("type", "", "Type: enum, number, text")
	desc := fs.String("desc", "", "Description")
	group := fs.String("group", "", "Group ID")
	values := fs.String("values", "", "Enum values (comma-separated)")
	minVal := fs.String("min", "", "Number min value")
	maxVal := fs.String("max", "", "Number max value")
	maxLen := fs.Int("maxlen", 200, "Text max length")

	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *name == "" {
		fatalError("-name is required")
	}
	if *tagType == "" {
		fatalError("-type is required")
	}

	req := &common.CreateTagRequest{
		Operator:    *operator,
		Name:        *name,
		Type:        common.TagType(*tagType),
		Description: *desc,
		GroupID:     *group,
	}

	switch req.Type {
	case common.TagTypeEnum:
		if *values == "" {
			fatalError("-values is required for enum type")
		}
		enumValues := strings.Split(*values, ",")
		for i, v := range enumValues {
			enumValues[i] = strings.TrimSpace(v)
		}
		req.EnumConfig = &common.EnumConfig{Values: enumValues}

	case common.TagTypeNumber:
		config := &common.NumberConfig{}
		if *minVal != "" {
			min, err := strconv.ParseFloat(*minVal, 64)
			if err != nil {
				fatalError("invalid -min value: %v", err)
			}
			config.Min = &min
		}
		if *maxVal != "" {
			max, err := strconv.ParseFloat(*maxVal, 64)
			if err != nil {
				fatalError("invalid -max value: %v", err)
			}
			config.Max = &max
		}
		req.NumberConfig = config

	case common.TagTypeText:
		req.TextConfig = &common.TextConfig{MaxLength: *maxLen}
	}

	resp, err := client.CreateTag(req)
	if err != nil {
		fatalError("Failed to create tag: %v", err)
	}

	fmt.Println("Tag created successfully!")
	fmt.Printf("Tag ID: %s\n", resp.TagID)
}

func cmdTagList(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-list", flag.ExitOnError)
	group := fs.String("group", "", "Filter by group ID")
	fs.Parse(args)

	resp, err := client.ListTags(*group)
	if err != nil {
		fatalError("Failed to list tags: %v", err)
	}

	fmt.Println(formatTagList(resp.Tags))
}

func cmdTagGet(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-get", flag.ExitOnError)
	tagID := fs.String("id", "", "Tag ID")
	fs.Parse(args)

	if *tagID == "" {
		fatalError("-id is required")
	}

	tag, err := client.GetTag(*tagID)
	if err != nil {
		fatalError("Failed to get tag: %v", err)
	}

	fmt.Println(formatTag(*tag))
}

func cmdTagUpdate(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-update", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	tagID := fs.String("id", "", "Tag ID")
	name := fs.String("name", "", "New tag name")
	desc := fs.String("desc", "", "New description")
	group := fs.String("group", "", "New group ID")
	values := fs.String("values", "", "New enum values")
	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *tagID == "" {
		fatalError("-id is required")
	}

	req := &common.UpdateTagRequest{
		Operator:    *operator,
		Name:        *name,
		Description: *desc,
		GroupID:     *group,
	}

	if *values != "" {
		enumValues := strings.Split(*values, ",")
		for i, v := range enumValues {
			enumValues[i] = strings.TrimSpace(v)
		}
		req.EnumConfig = &common.EnumConfig{Values: enumValues}
	}

	err := client.UpdateTag(*tagID, req)
	if err != nil {
		fatalError("Failed to update tag: %v", err)
	}

	fmt.Println("Tag updated successfully!")
}

func cmdTagDelete(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-delete", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	tagID := fs.String("id", "", "Tag ID")
	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *tagID == "" {
		fatalError("-id is required")
	}

	impact, err := client.GetTagImpact(*tagID)
	if err != nil {
		fatalError("Failed to check impact: %v", err)
	}

	fmt.Printf("Deleting tag will affect %d user(s).\n", impact.ImpactedUserCount)
	fmt.Printf("Proceed? (y/N): ")

	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("Delete cancelled.")
		return
	}

	count, err := client.DeleteTag(*tagID, *operator)
	if err != nil {
		fatalError("Failed to delete tag: %v", err)
	}

	fmt.Println("Tag deleted successfully!")
	fmt.Printf("Removed from %d user(s).\n", count)
}

func cmdTagImpact(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-impact", flag.ExitOnError)
	tagID := fs.String("id", "", "Tag ID")
	fs.Parse(args)

	if *tagID == "" {
		fatalError("-id is required")
	}

	impact, err := client.GetTagImpact(*tagID)
	if err != nil {
		fatalError("Failed to get impact: %v", err)
	}

	fmt.Printf("Tag deletion impact:\n")
	fmt.Printf("  Affected users: %d\n", impact.ImpactedUserCount)
}

func cmdGroupCreate(client *Client, args []string) {
	fs := flag.NewFlagSet("group-create", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	name := fs.String("name", "", "Group name")
	desc := fs.String("desc", "", "Description")
	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *name == "" {
		fatalError("-name is required")
	}

	req := &common.CreateTagGroupRequest{
		Operator:    *operator,
		Name:        *name,
		Description: *desc,
	}

	resp, err := client.CreateTagGroup(req)
	if err != nil {
		fatalError("Failed to create group: %v", err)
	}

	fmt.Println("Group created successfully!")
	fmt.Printf("Group ID: %s\n", resp.GroupID)
}

func cmdGroupList(client *Client, args []string) {
	fs := flag.NewFlagSet("group-list", flag.ExitOnError)
	fs.Parse(args)

	resp, err := client.ListTagGroups()
	if err != nil {
		fatalError("Failed to list groups: %v", err)
	}

	fmt.Println(formatTagGroupList(resp.Groups))
}

func cmdGroupGet(client *Client, args []string) {
	fs := flag.NewFlagSet("group-get", flag.ExitOnError)
	groupID := fs.String("id", "", "Group ID")
	fs.Parse(args)

	if *groupID == "" {
		fatalError("-id is required")
	}

	group, err := client.GetTagGroup(*groupID)
	if err != nil {
		fatalError("Failed to get group: %v", err)
	}

	fmt.Println(formatTagGroup(*group))
}

func cmdTagApply(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-apply", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	tagID := fs.String("id", "", "Tag ID")
	userID := fs.String("user", "", "User ID")
	value := fs.String("value", "", "Tag value")
	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *tagID == "" {
		fatalError("-id is required")
	}
	if *userID == "" {
		fatalError("-user is required")
	}

	var tagValue interface{}
	if *value != "" {
		if num, err := strconv.ParseFloat(*value, 64); err == nil {
			tagValue = num
		} else {
			tagValue = *value
		}
	}

	req := &common.ApplyTagToUserRequest{
		Operator: *operator,
		UserID:   *userID,
		TagValue: tagValue,
	}

	err := client.ApplyTagToUser(*tagID, req)
	if err != nil {
		fatalError("Failed to apply tag: %v", err)
	}

	fmt.Println("Tag applied successfully!")
}

func cmdTagBatchApply(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-batch-apply", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	tagID := fs.String("id", "", "Tag ID")
	users := fs.String("users", "", "User IDs (comma-separated)")
	value := fs.String("value", "", "Tag value")
	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *tagID == "" {
		fatalError("-id is required")
	}
	if *users == "" {
		fatalError("-users is required")
	}

	userIDs := strings.Split(*users, ",")
	for i, id := range userIDs {
		userIDs[i] = strings.TrimSpace(id)
	}

	var tagValue interface{}
	if *value != "" {
		if num, err := strconv.ParseFloat(*value, 64); err == nil {
			tagValue = num
		} else {
			tagValue = *value
		}
	}

	req := &common.BatchApplyTagRequest{
		Operator: *operator,
		UserIDs:  userIDs,
		TagValue: tagValue,
	}

	count, err := client.BatchApplyTag(*tagID, req)
	if err != nil {
		fatalError("Failed to batch apply tag: %v", err)
	}

	fmt.Printf("Batch applied tag to %d user(s) successfully!\n", count)
}

func cmdTagRemove(client *Client, args []string) {
	fs := flag.NewFlagSet("tag-remove", flag.ExitOnError)
	operator := fs.String("operator", "", "Operator name")
	tagID := fs.String("id", "", "Tag ID")
	userID := fs.String("user", "", "User ID")
	fs.Parse(args)

	if *operator == "" {
		fatalError("-operator is required")
	}
	if *tagID == "" {
		fatalError("-id is required")
	}
	if *userID == "" {
		fatalError("-user is required")
	}

	err := client.RemoveUserTag(*tagID, *userID, *operator)
	if err != nil {
		fatalError("Failed to remove tag: %v", err)
	}

	fmt.Println("Tag removed successfully!")
}

func cmdUserTags(client *Client, args []string) {
	fs := flag.NewFlagSet("user-tags", flag.ExitOnError)
	userID := fs.String("user", "", "User ID")
	fs.Parse(args)

	if *userID == "" {
		fatalError("-user is required")
	}

	resp, err := client.GetUserTags(*userID)
	if err != nil {
		fatalError("Failed to get user tags: %v", err)
	}

	fmt.Println(formatUserTagList(resp.Tags))
}

func cmdUsersFilter(client *Client, args []string) {
	fs := flag.NewFlagSet("users-filter", flag.ExitOnError)
	jsonStr := fs.String("json", "", "Filter JSON")
	fs.Parse(args)

	if *jsonStr == "" {
		fatalError("-json is required")
	}

	var req common.FilterUsersRequest
	if err := json.Unmarshal([]byte(*jsonStr), &req); err != nil {
		fatalError("Invalid JSON: %v", err)
	}

	resp, err := client.FilterUsers(&req)
	if err != nil {
		fatalError("Failed to filter users: %v", err)
	}

	fmt.Println(formatFilterResult(resp.UserIDs))
}

func cmdStats(client *Client, args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	fs.Parse(args)

	resp, err := client.GetTagStats()
	if err != nil {
		fatalError("Failed to get stats: %v", err)
	}

	fmt.Println(formatStats(resp.Stats))
}

func cmdTrend(client *Client, args []string) {
	fs := flag.NewFlagSet("trend", flag.ExitOnError)
	fs.Parse(args)

	resp, err := client.GetTagTrend()
	if err != nil {
		fatalError("Failed to get trend: %v", err)
	}

	fmt.Println(formatTrend(resp.Trends))
}

func cmdAudit(client *Client, args []string) {
	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	fs.Parse(args)

	resp, err := client.ListAuditLogs()
	if err != nil {
		fatalError("Failed to get audit logs: %v", err)
	}

	fmt.Println(formatAuditLogs(resp.Logs))
}

func fatalError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
