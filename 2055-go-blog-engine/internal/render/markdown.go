package render

import (
	"bytes"
	"html"
	"regexp"
	"strings"
)

func MarkdownToHTML(markdown string) string {
	var buf bytes.Buffer
	lines := strings.Split(markdown, "\n")

	inCodeBlock := false
	inList := false
	listType := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				buf.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				buf.WriteString("<pre><code>")
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			buf.WriteString(html.EscapeString(line))
			buf.WriteString("\n")
			continue
		}

		if strings.TrimSpace(line) == "" {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			continue
		}

		line = processInlineElements(line)

		if h1 := regexp.MustCompile(`^#\s+(.*)$`).FindStringSubmatch(line); h1 != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<h1>" + h1[1] + "</h1>\n")
		} else if h2 := regexp.MustCompile(`^##\s+(.*)$`).FindStringSubmatch(line); h2 != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<h2>" + h2[1] + "</h2>\n")
		} else if h3 := regexp.MustCompile(`^###\s+(.*)$`).FindStringSubmatch(line); h3 != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<h3>" + h3[1] + "</h3>\n")
		} else if h4 := regexp.MustCompile(`^####\s+(.*)$`).FindStringSubmatch(line); h4 != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<h4>" + h4[1] + "</h4>\n")
		} else if h5 := regexp.MustCompile(`^#####\s+(.*)$`).FindStringSubmatch(line); h5 != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<h5>" + h5[1] + "</h5>\n")
		} else if h6 := regexp.MustCompile(`^######\s+(.*)$`).FindStringSubmatch(line); h6 != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<h6>" + h6[1] + "</h6>\n")
		} else if ul := regexp.MustCompile(`^[\*\-]\s+(.*)$`).FindStringSubmatch(line); ul != nil {
			if !inList || listType != "ul" {
				if inList {
					buf.WriteString("</ul>\n")
				}
				buf.WriteString("<ul>\n")
				inList = true
				listType = "ul"
			}
			buf.WriteString("<li>" + ul[1] + "</li>\n")
		} else if ol := regexp.MustCompile(`^\d+\.\s+(.*)$`).FindStringSubmatch(line); ol != nil {
			if !inList || listType != "ol" {
				if inList {
					buf.WriteString("</ol>\n")
				}
				buf.WriteString("<ol>\n")
				inList = true
				listType = "ol"
			}
			buf.WriteString("<li>" + ol[1] + "</li>\n")
		} else if quote := regexp.MustCompile(`^>\s+(.*)$`).FindStringSubmatch(line); quote != nil {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<blockquote><p>" + quote[1] + "</p></blockquote>\n")
		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "***") {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<hr />\n")
		} else {
			if inList {
				buf.WriteString("</ul>\n")
				inList = false
			}
			buf.WriteString("<p>" + line + "</p>\n")
		}
	}

	if inList {
		if listType == "ul" {
			buf.WriteString("</ul>\n")
		} else {
			buf.WriteString("</ol>\n")
		}
	}

	return buf.String()
}

func processInlineElements(text string) string {
	text = regexp.MustCompile(`\*\*\*(.*?)\*\*\*`).ReplaceAllString(text, "<strong><em>$1</em></strong>")
	text = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(text, "<strong>$1</strong>")
	text = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(text, "<em>$1</em>")
	text = regexp.MustCompile(`___(.*?)___`).ReplaceAllString(text, "<strong><em>$1</em></strong>")
	text = regexp.MustCompile(`__(.*?)__`).ReplaceAllString(text, "<strong>$1</strong>")
	text = regexp.MustCompile(`_(.*?)_`).ReplaceAllString(text, "<em>$1</em>")
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "<code>$1</code>")
	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, `<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>`)
	text = regexp.MustCompile(`!\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, `<img src="$2" alt="$1" />`)
	text = regexp.MustCompile(`~~(.*?)~~`).ReplaceAllString(text, "<del>$1</del>")
	return text
}

func GenerateExcerpt(htmlContent string, maxLength int) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, "")
	text = strings.TrimSpace(text)

	if len(text) <= maxLength {
		return text
	}

	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}

	return string(runes[:maxLength]) + "..."
}
