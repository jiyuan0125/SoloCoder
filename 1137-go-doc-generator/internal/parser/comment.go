package parser

import (
	"go/ast"
	"strings"
)

func extractComment(commentGroup *ast.CommentGroup) string {
	if commentGroup == nil {
		return ""
	}
	return commentGroup.Text()
}

func extractSummary(comment string) string {
	if comment == "" {
		return ""
	}

	lines := strings.Split(comment, "\n")
	var summaryLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		summaryLines = append(summaryLines, trimmed)
	}

	return strings.Join(summaryLines, " ")
}

func extractExamples(comment string) []Example {
	if comment == "" {
		return nil
	}

	var examples []Example
	lines := strings.Split(comment, "\n")

	var currentExample *Example
	var inExample bool

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(strings.TrimSpace(line), "Example") {
			if currentExample != nil {
				examples = append(examples, *currentExample)
			}
			title := strings.TrimSpace(line)
			currentExample = &Example{Title: title}
			inExample = false
			continue
		}

		if currentExample != nil {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				if !inExample {
					inExample = true
				}
				continue
			}

			if inExample {
				if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
					if currentExample.Code != "" {
						currentExample.Code += "\n"
					}
					currentExample.Code += strings.TrimLeft(line, " \t")
				} else {
					examples = append(examples, *currentExample)
					currentExample = nil
					inExample = false
				}
			}
		}
	}

	if currentExample != nil && currentExample.Code != "" {
		examples = append(examples, *currentExample)
	}

	examples = append(examples, extractIndentedExamples(comment)...)

	return dedupeExamples(examples)
}

func extractIndentedExamples(comment string) []Example {
	var examples []Example
	lines := strings.Split(comment, "\n")

	var currentCode []string
	var inCodeBlock bool

	for _, line := range lines {
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			if !inCodeBlock {
				inCodeBlock = true
				currentCode = []string{strings.TrimLeft(line, " \t")}
			} else {
				currentCode = append(currentCode, strings.TrimLeft(line, " \t"))
			}
		} else {
			if inCodeBlock && len(currentCode) > 0 {
				examples = append(examples, Example{
					Title: "Code Example",
					Code:  strings.Join(currentCode, "\n"),
				})
				inCodeBlock = false
				currentCode = nil
			}
		}
	}

	if inCodeBlock && len(currentCode) > 0 {
		examples = append(examples, Example{
			Title: "Code Example",
			Code:  strings.Join(currentCode, "\n"),
		})
	}

	return examples
}

func dedupeExamples(examples []Example) []Example {
	if len(examples) == 0 {
		return examples
	}

	seen := make(map[string]bool)
	var result []Example

	for _, ex := range examples {
		key := ex.Title + "|" + ex.Code
		if !seen[key] {
			seen[key] = true
			result = append(result, ex)
		}
	}

	return result
}

func parseLineComment(line string) string {
	idx := strings.Index(line, "//")
	if idx == -1 {
		return ""
	}
	return strings.TrimSpace(line[idx+2:])
}
