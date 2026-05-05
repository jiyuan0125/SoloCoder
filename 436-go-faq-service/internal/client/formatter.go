package client

import (
	"encoding/json"
	"fmt"
	"go-faq-service/pkg/protocol"
	"strings"
)

func FormatResponse(resp *protocol.Response) string {
	var builder strings.Builder

	if resp.Success {
		builder.WriteString("✓ Success\n")
	} else {
		builder.WriteString("✗ Failed\n")
	}

	if resp.Message != "" {
		builder.WriteString(fmt.Sprintf("Message: %s\n", resp.Message))
	}

	if resp.Data != nil {
		builder.WriteString("\nData:\n")
		dataJSON, _ := json.MarshalIndent(resp.Data, "  ", "  ")
		builder.WriteString(string(dataJSON))
		builder.WriteString("\n")
	}

	return builder.String()
}

func FormatSearchResult(resp *protocol.Response) string {
	var builder strings.Builder

	if !resp.Success {
		return FormatResponse(resp)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return FormatResponse(resp)
	}

	total, _ := dataMap["total"].(float64)
	page, _ := dataMap["page"].(float64)
	pageSize, _ := dataMap["page_size"].(float64)
	totalPages, _ := dataMap["total_pages"].(float64)

	builder.WriteString(fmt.Sprintf("✓ Found %d results (Page %d/%d, %d per page)\n\n",
		int(total), int(page), int(totalPages), int(pageSize)))

	items, ok := dataMap["items"].([]interface{})
	if !ok || len(items) == 0 {
		builder.WriteString("  No results found.\n")
		return builder.String()
	}

	for i, item := range items {
		itemMap, _ := item.(map[string]interface{})
		builder.WriteString(fmt.Sprintf("--- Result %d ---\n", i+1))

		if id, ok := itemMap["id"].(string); ok {
			builder.WriteString(fmt.Sprintf("  ID: %s\n", id))
		}
		if catName, ok := itemMap["category_name"].(string); ok && catName != "" {
			builder.WriteString(fmt.Sprintf("  Category: %s\n", catName))
		}

		content, ok := itemMap["content"].(map[string]interface{})
		if ok {
			for lang, langContent := range content {
				langMap, _ := langContent.(map[string]interface{})
				question, _ := langMap["question"].(string)
				answer, _ := langMap["answer"].(string)

				builder.WriteString(fmt.Sprintf("\n  [%s]\n", strings.ToUpper(lang)))
				builder.WriteString(fmt.Sprintf("  Q: %s\n", question))
				if len(answer) > 100 {
					builder.WriteString(fmt.Sprintf("  A: %s...\n", answer[:100]))
				} else {
					builder.WriteString(fmt.Sprintf("  A: %s\n", answer))
				}
			}
		}

		builder.WriteString("\n")
		if isPinned, ok := itemMap["is_pinned"].(bool); ok && isPinned {
			builder.WriteString("  [PINNED] ")
		}
		if isHot, ok := itemMap["is_hot"].(bool); ok && isHot {
			builder.WriteString("🔥 [HOT] ")
		}
		if needOpt, ok := itemMap["need_optimize"].(bool); ok && needOpt {
			builder.WriteString("⚠ [NEEDS OPTIMIZATION] ")
		}
		if viewCount, ok := itemMap["view_count"].(float64); ok {
			builder.WriteString(fmt.Sprintf("Views: %d", int(viewCount)))
		}
		if matchScore, ok := itemMap["match_score"].(float64); ok && matchScore > 0 {
			builder.WriteString(fmt.Sprintf(" | Score: %.1f", matchScore))
		}
		builder.WriteString("\n\n")
	}

	return builder.String()
}

func FormatCategoryTree(resp *protocol.Response) string {
	var builder strings.Builder

	if !resp.Success {
		return FormatResponse(resp)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return FormatResponse(resp)
	}

	categories, ok := dataMap["categories"].([]interface{})
	if !ok {
		return FormatResponse(resp)
	}

	builder.WriteString("✓ Category Tree:\n\n")

	for _, cat := range categories {
		catMap, _ := cat.(map[string]interface{})
		printCategory(&builder, catMap, 0)
	}

	return builder.String()
}

func printCategory(builder *strings.Builder, catMap map[string]interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	name, _ := catMap["name"].(string)
	id, _ := catMap["id"].(string)

	builder.WriteString(fmt.Sprintf("%s├── %s (%s)\n", prefix, name, id))

	children, ok := catMap["children"].([]interface{})
	if ok && len(children) > 0 {
		for _, child := range children {
			childMap, _ := child.(map[string]interface{})
			printCategory(builder, childMap, indent+1)
		}
	}
}

func FormatStatistics(resp *protocol.Response) string {
	var builder strings.Builder

	if !resp.Success {
		return FormatResponse(resp)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return FormatResponse(resp)
	}

	builder.WriteString("✓ FAQ System Statistics\n\n")
	builder.WriteString("  ┌─────────────────────────────────┐\n")

	totalFAQ, _ := data["total_faq"].(float64)
	builder.WriteString(fmt.Sprintf("  │ Total FAQs:         %10d │\n", int(totalFAQ)))

	enabledFAQ, _ := data["enabled_faq"].(float64)
	builder.WriteString(fmt.Sprintf("  │ Enabled FAQs:       %10d │\n", int(enabledFAQ)))

	hotFAQ, _ := data["hot_faq"].(float64)
	builder.WriteString(fmt.Sprintf("  │ 🔥 Hot FAQs:        %10d │\n", int(hotFAQ)))

	needOptimize, _ := data["need_optimize"].(float64)
	builder.WriteString(fmt.Sprintf("  │ ⚠ Needs Optimization: %8d │\n", int(needOptimize)))

	totalViews, _ := data["total_views"].(float64)
	builder.WriteString(fmt.Sprintf("  │ Total Views:        %10d │\n", int(totalViews)))

	totalClicks, _ := data["total_clicks"].(float64)
	builder.WriteString(fmt.Sprintf("  │ Total Clicks:       %10d │\n", int(totalClicks)))

	builder.WriteString("  └─────────────────────────────────┘\n")

	return builder.String()
}

func FormatBatchImportResult(resp *protocol.Response) string {
	var builder strings.Builder

	if !resp.Success {
		return FormatResponse(resp)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return FormatResponse(resp)
	}

	successCount, _ := data["success_count"].(float64)
	failedCount, _ := data["failed_count"].(float64)

	builder.WriteString(fmt.Sprintf("✓ Batch Import Completed\n\n"))
	builder.WriteString(fmt.Sprintf("  Success: %d\n", int(successCount)))
	builder.WriteString(fmt.Sprintf("  Failed:  %d\n", int(failedCount)))

	if failedCount > 0 {
		builder.WriteString("\n  Failed Items:\n")
		failedItems, ok := data["failed_items"].([]interface{})
		if ok {
			for i, item := range failedItems {
				itemMap, _ := item.(map[string]interface{})
				index, _ := itemMap["index"].(float64)
				errMsg, _ := itemMap["error"].(string)
				builder.WriteString(fmt.Sprintf("    [%d] Index %d: %s\n", i+1, int(index), errMsg))
			}
		}
	}

	return builder.String()
}

func FormatFAQDetail(resp *protocol.Response) string {
	var builder strings.Builder

	if !resp.Success {
		return FormatResponse(resp)
	}

	faq, ok := resp.Data.(map[string]interface{})
	if !ok {
		return FormatResponse(resp)
	}

	builder.WriteString("✓ FAQ Detail\n\n")

	if id, ok := faq["id"].(string); ok {
		builder.WriteString(fmt.Sprintf("  ID: %s\n", id))
	}
	if catName, ok := faq["category_name"].(string); ok {
		builder.WriteString(fmt.Sprintf("  Category: %s\n", catName))
	}
	if sortWeight, ok := faq["sort_weight"].(float64); ok {
		builder.WriteString(fmt.Sprintf("  Sort Weight: %d\n", int(sortWeight)))
	}

	builder.WriteString("  Status: ")
	if isEnabled, ok := faq["is_enabled"].(bool); ok && isEnabled {
		builder.WriteString("[Enabled] ")
	} else {
		builder.WriteString("[Disabled] ")
	}
	if isPinned, ok := faq["is_pinned"].(bool); ok && isPinned {
		builder.WriteString("[Pinned] ")
	}
	if isHot, ok := faq["is_hot"].(bool); ok && isHot {
		builder.WriteString("🔥 [Hot] ")
	}
	if needOpt, ok := faq["need_optimize"].(bool); ok && needOpt {
		builder.WriteString("⚠ [Needs Optimization]")
	}
	builder.WriteString("\n")

	if viewCount, ok := faq["view_count"].(float64); ok {
		builder.WriteString(fmt.Sprintf("  Views: %d\n", int(viewCount)))
	}

	content, ok := faq["content"].(map[string]interface{})
	if ok {
		builder.WriteString("\n  Content:\n")
		for lang, langContent := range content {
			langMap, _ := langContent.(map[string]interface{})
			question, _ := langMap["question"].(string)
			answer, _ := langMap["answer"].(string)

			builder.WriteString(fmt.Sprintf("\n  ┌─ [%s] ──────────────────\n", strings.ToUpper(lang)))
			builder.WriteString(fmt.Sprintf("  │ Q: %s\n", question))
			builder.WriteString(fmt.Sprintf("  │ A: %s\n", answer))
			builder.WriteString("  └──────────────────────────\n")
		}
	}

	return builder.String()
}
