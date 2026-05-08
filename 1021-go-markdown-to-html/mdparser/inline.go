package mdparser

import (
	"strings"
)

func parseInline(text string) string {
	nodes := tokenizeInline(text)
	return renderInlineNodes(nodes)
}

func tokenizeInline(text string) []*InlineNode {
	var nodes []*InlineNode
	runes := []rune(text)
	i := 0
	
	for i < len(runes) {
		if runes[i] == '\\' && i+1 < len(runes) && escapeChars[runes[i+1]] {
			nodes = append(nodes, &InlineNode{
				Type:    InlineText,
				Content: string(runes[i+1]),
			})
			i += 2
			continue
		}
		
		if i+2 <= len(runes) && runes[i] == '`' && runes[i+1] == '`' {
			end := findDoubleBacktickCode(runes, i+2)
			if end > i+2 {
				content := string(runes[i+2 : end])
				content = strings.Trim(content, " ")
				nodes = append(nodes, &InlineNode{
					Type:    InlineCode,
					Content: content,
				})
				i = end + 2
				continue
			}
		}
		
		if runes[i] == '`' {
			end := findSingleBacktickCode(runes, i+1)
			if end > i+1 {
				content := string(runes[i+1 : end])
				nodes = append(nodes, &InlineNode{
					Type:    InlineCode,
					Content: content,
				})
				i = end + 1
				continue
			}
		}
		
		if i+2 <= len(runes) && runes[i] == '*' && runes[i+1] == '*' {
			end := findMatchingDelimiter(runes, i+2, '*', 2)
			if end > i+2 {
				content := string(runes[i+2 : end])
				nodes = append(nodes, &InlineNode{
					Type:     InlineBold,
					Content:  content,
					Children: tokenizeInline(content),
				})
				i = end + 2
				continue
			}
		}
		
		if runes[i] == '*' {
			end := findMatchingDelimiter(runes, i+1, '*', 1)
			if end > i+1 {
				content := string(runes[i+1 : end])
				nodes = append(nodes, &InlineNode{
					Type:     InlineItalic,
					Content:  content,
					Children: tokenizeInline(content),
				})
				i = end + 1
				continue
			}
		}
		
		if i+2 <= len(runes) && runes[i] == '_' && runes[i+1] == '_' {
			end := findMatchingDelimiter(runes, i+2, '_', 2)
			if end > i+2 {
				content := string(runes[i+2 : end])
				nodes = append(nodes, &InlineNode{
					Type:     InlineBold,
					Content:  content,
					Children: tokenizeInline(content),
				})
				i = end + 2
				continue
			}
		}
		
		if runes[i] == '_' {
			end := findMatchingDelimiter(runes, i+1, '_', 1)
			if end > i+1 {
				content := string(runes[i+1 : end])
				nodes = append(nodes, &InlineNode{
					Type:     InlineItalic,
					Content:  content,
					Children: tokenizeInline(content),
				})
				i = end + 1
				continue
			}
		}
		
		if i+2 <= len(runes) && runes[i] == '~' && runes[i+1] == '~' {
			end := findStrikethroughEnd(runes, i+2)
			if end > i+2 {
				content := string(runes[i+2 : end])
				nodes = append(nodes, &InlineNode{
					Type:     InlineStrikethrough,
					Content:  content,
					Children: tokenizeInline(content),
				})
				i = end + 2
				continue
			}
		}
		
		if runes[i] == '[' {
			linkResult := parseLinkOrImage(runes, i)
			if linkResult != nil {
				nodes = append(nodes, linkResult.node)
				i = linkResult.end
				continue
			}
		}
		
		nodes = append(nodes, &InlineNode{
			Type:    InlineText,
			Content: string(runes[i]),
		})
		i++
	}
	
	return nodes
}

func findDoubleBacktickCode(runes []rune, start int) int {
	for i := start; i < len(runes)-1; i++ {
		if runes[i] == '`' && runes[i+1] == '`' {
			return i
		}
	}
	return -1
}

func findSingleBacktickCode(runes []rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == '`' {
			if i+1 < len(runes) && runes[i+1] == '`' {
				continue
			}
			return i
		}
	}
	return -1
}

func findMatchingDelimiter(runes []rune, start int, delim rune, count int) int {
	depth := 0
	i := start
	
	for i < len(runes) {
		if runes[i] == '`' {
			i = findSingleBacktickCode(runes, i+1)
			if i < 0 {
				return -1
			}
			i++
			continue
		}
		
		if runes[i] == '[' {
			depth++
		} else if runes[i] == ']' {
			depth--
		}
		
		if depth == 0 {
			if count == 2 && i+1 < len(runes) && runes[i] == delim && runes[i+1] == delim {
				if i+2 < len(runes) && !isWordChar(runes[i+2]) {
					return i
				}
				if i+2 >= len(runes) {
					return i
				}
			}
			
			if count == 1 && runes[i] == delim {
				if (i+1 < len(runes) && !isWordChar(runes[i+1])) || i+1 >= len(runes) {
					if i == start {
						i++
						continue
					}
					return i
				}
			}
		}
		i++
	}
	return -1
}

func findStrikethroughEnd(runes []rune, start int) int {
	for i := start; i < len(runes)-1; i++ {
		if runes[i] == '~' && runes[i+1] == '~' {
			return i
		}
	}
	return -1
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

type linkParseResult struct {
	node *InlineNode
	end  int
}

func parseLinkOrImage(runes []rune, start int) *linkParseResult {
	isImage := false
	i := start
	
	if i > 0 && runes[i-1] == '!' {
		isImage = true
		if i > 1 {
			start = i - 1
		}
	}
	
	textStart := i + 1
	textEnd := -1
	bracketCount := 1
	
	for j := textStart; j < len(runes); j++ {
		if runes[j] == '\\' && j+1 < len(runes) {
			j++
			continue
		}
		if runes[j] == '[' {
			bracketCount++
		} else if runes[j] == ']' {
			bracketCount--
			if bracketCount == 0 {
				textEnd = j
				break
			}
		}
	}
	
	if textEnd < 0 || textEnd+1 >= len(runes) {
		return nil
	}
	
	if runes[textEnd+1] != '(' {
		return nil
	}
	
	textContent := string(runes[textStart:textEnd])
	
	urlStart := textEnd + 2
	urlEnd := -1
	inTitle := false
	titleStart := -1
	titleEnd := -1
	parenCount := 1
	
	for j := urlStart; j < len(runes); j++ {
		if runes[j] == '\\' && j+1 < len(runes) {
			j++
			continue
		}
		if inTitle {
			if runes[j] == '"' || runes[j] == '\'' {
				titleEnd = j
				inTitle = false
			}
			continue
		}
		if runes[j] == '(' {
			parenCount++
		} else if runes[j] == ')' {
			parenCount--
			if parenCount == 0 {
				urlEnd = j
				break
			}
		} else if (runes[j] == '"' || runes[j] == '\'') && j > urlStart {
			inTitle = true
			titleStart = j
		}
	}
	
	if urlEnd < 0 {
		return nil
	}
	
	urlContent := string(runes[urlStart:urlEnd])
	urlContent = strings.TrimSpace(urlContent)
	
	title := ""
	if titleStart > 0 && titleEnd > 0 {
		title = string(runes[titleStart+1 : titleEnd])
		urlContent = strings.TrimSpace(string(runes[urlStart:titleStart]))
	}
	
	node := &InlineNode{
		Type:    InlineLink,
		Content: textContent,
		URL:     urlContent,
		Title:   title,
	}
	
	if isImage {
		node.Type = InlineImage
		node.Alt = textContent
	}
	
	return &linkParseResult{
		node: node,
		end:  urlEnd + 1,
	}
}
