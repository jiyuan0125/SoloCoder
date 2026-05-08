package mdparser

import (
	"strings"
)

type listItemInfo struct {
	line       string
	indent     int
	ordered    bool
	startNum   int
	content    string
	isCheckbox bool
	checked    bool
}

func (p *Parser) parseListBlock() *Block {
	startLine := p.lines[p.index]
	
	var isOrdered bool
	var startNum int
	var firstIndent int
	
	if indent, num, content := isOrderedList(startLine); indent >= 0 {
		isOrdered = true
		startNum = num
		firstIndent = indent
		_ = content
	} else if indent, content := isUnorderedList(startLine); indent >= 0 {
		isOrdered = false
		startNum = 1
		firstIndent = indent
		_ = content
	}
	
	var listItems []*Block
	
	for p.index < len(p.lines) {
		item := p.parseListItem(firstIndent, isOrdered)
		if item == nil {
			break
		}
		listItems = append(listItems, item)
	}
	
	return &Block{
		Type:     BlockList,
		Ordered:  isOrdered,
		StartNum: startNum,
		Children: listItems,
	}
}

func (p *Parser) parseListItem(baseIndent int, parentOrdered bool) *Block {
	if p.index >= len(p.lines) {
		return nil
	}
	
	line := p.lines[p.index]
	
	if isBlank(line) {
		p.index++
		if p.index >= len(p.lines) {
			return nil
		}
		line = p.lines[p.index]
	}
	
	var isOrdered bool
	var itemIndent int
	var startNum int
	var content string
	
	if indent, num, c := isOrderedList(line); indent >= 0 {
		isOrdered = true
		itemIndent = indent
		startNum = num
		content = c
	} else if indent, c := isUnorderedList(line); indent >= 0 {
		isOrdered = false
		itemIndent = indent
		startNum = 1
		content = c
	} else {
		return nil
	}
	
	if isOrdered != parentOrdered && itemIndent <= baseIndent {
		return nil
	}
	
	var itemContentLines []string
	hasCheckbox, checked, checkboxContent := isCheckbox(content)
	if hasCheckbox {
		content = checkboxContent
	}
	if content != "" {
		itemContentLines = append(itemContentLines, content)
	}
	p.index++
	
	for p.index < len(p.lines) {
		nextLine := p.lines[p.index]
		
		if isBlank(nextLine) {
			if p.index+1 >= len(p.lines) {
				break
			}
			peekLine := p.lines[p.index+1]
			
			if isBlank(peekLine) {
				break
			}
			
			if nIndent, _, _ := isOrderedList(peekLine); nIndent >= 0 {
				if nIndent <= baseIndent {
					break
				}
			}
			if nIndent, _ := isUnorderedList(peekLine); nIndent >= 0 {
				if nIndent <= baseIndent {
					break
				}
			}
			itemContentLines = append(itemContentLines, "")
			p.index++
			continue
		}
		
		nIndent, _, _ := isOrderedList(nextLine)
		uIndent, _ := isUnorderedList(nextLine)
		
		if nIndent >= 0 {
			if nIndent <= baseIndent {
				break
			}
			if nIndent > baseIndent {
				innerContent := p.parseNestedContent(itemContentLines, baseIndent)
				return &Block{
					Type:        BlockListItem,
					Content:     innerContent,
					Children:    nil,
					Ordered:     isOrdered,
					StartNum:    startNum,
					HasCheckbox: hasCheckbox,
					Checked:     checked,
				}
			}
		}
		
		if uIndent >= 0 {
			if uIndent <= baseIndent {
				break
			}
			if uIndent > baseIndent {
				innerContent := p.parseNestedContent(itemContentLines, baseIndent)
				return &Block{
					Type:        BlockListItem,
					Content:     innerContent,
					Children:    nil,
					Ordered:     isOrdered,
					StartNum:    startNum,
					HasCheckbox: hasCheckbox,
					Checked:     checked,
				}
			}
		}
		
		currentIndent := countLeadingSpaces(nextLine)
		if currentIndent > baseIndent {
			trimmed := nextLine
			if len(nextLine) > baseIndent {
				trimmed = nextLine[baseIndent:]
			}
			itemContentLines = append(itemContentLines, trimmed)
			p.index++
		} else if currentIndent == baseIndent {
			if nIndent >= 0 || uIndent >= 0 {
				break
			}
			itemContentLines = append(itemContentLines, strings.TrimLeft(nextLine, " \t"))
			p.index++
		} else {
			break
		}
	}
	
	itemContent := p.parseNestedContent(itemContentLines, baseIndent)
	
	return &Block{
		Type:        BlockListItem,
		Content:     itemContent,
		Children:    nil,
		Ordered:     isOrdered,
		StartNum:    startNum,
		HasCheckbox: hasCheckbox,
		Checked:     checked,
	}
}

func (p *Parser) parseNestedContent(lines []string, baseIndent int) string {
	var contentParts []string
	var i int
	
	for i < len(lines) {
		line := lines[i]
		
		if isFenceFlag, _ := isFence(strings.TrimLeft(line, " \t")); isFenceFlag {
			codeLines := []string{strings.TrimLeft(line, " \t")}
			i++
			for i < len(lines) {
				cl := lines[i]
				codeLines = append(codeLines, cl)
				if isFenceFlag2, _ := isFence(strings.TrimLeft(cl, " \t")); isFenceFlag2 {
					i++
					break
				}
				i++
			}
			contentParts = append(contentParts, strings.Join(codeLines, "\n"))
			continue
		}
		
		if isQ, _ := isQuote(strings.TrimLeft(line, " \t")); isQ {
			quoteLines := []string{strings.TrimLeft(line, " \t")}
			i++
			for i < len(lines) {
				ql := lines[i]
				trimmedQL := strings.TrimLeft(ql, " \t")
				if isQ2, _ := isQuote(trimmedQL); isQ2 {
					quoteLines = append(quoteLines, ql)
					i++
				} else if isBlank(trimmedQL) {
					quoteLines = append(quoteLines, ql)
					i++
				} else {
					break
				}
			}
			contentParts = append(contentParts, strings.Join(quoteLines, "\n"))
			continue
		}
		
		if indent, _, _ := isOrderedList(strings.TrimLeft(line, " \t")); indent >= 0 {
			nestedLines := []string{strings.TrimLeft(line, " \t")}
			i++
			for i < len(lines) {
				nl := lines[i]
				trimmedNL := strings.TrimLeft(nl, " \t")
				if nIndent, _, _ := isOrderedList(trimmedNL); nIndent >= 0 {
					nestedLines = append(nestedLines, trimmedNL)
					i++
				} else if uIndent, _ := isUnorderedList(trimmedNL); uIndent >= 0 {
					nestedLines = append(nestedLines, trimmedNL)
					i++
				} else if isBlank(trimmedNL) {
					if i+1 < len(lines) {
						peek := strings.TrimLeft(lines[i+1], " \t")
						if pIndent, _, _ := isOrderedList(peek); pIndent >= 0 {
							nestedLines = append(nestedLines, trimmedNL)
							i++
						} else if pIndent, _ := isUnorderedList(peek); pIndent >= 0 {
							nestedLines = append(nestedLines, trimmedNL)
							i++
						} else {
							break
						}
					} else {
						break
					}
				} else {
					break
				}
			}
			contentParts = append(contentParts, strings.Join(nestedLines, "\n"))
			continue
		}
		
		if indent, _ := isUnorderedList(strings.TrimLeft(line, " \t")); indent >= 0 {
			nestedLines := []string{strings.TrimLeft(line, " \t")}
			i++
			for i < len(lines) {
				nl := lines[i]
				trimmedNL := strings.TrimLeft(nl, " \t")
				if nIndent, _, _ := isOrderedList(trimmedNL); nIndent >= 0 {
					nestedLines = append(nestedLines, trimmedNL)
					i++
				} else if uIndent, _ := isUnorderedList(trimmedNL); uIndent >= 0 {
					nestedLines = append(nestedLines, trimmedNL)
					i++
				} else if isBlank(trimmedNL) {
					if i+1 < len(lines) {
						peek := strings.TrimLeft(lines[i+1], " \t")
						if pIndent, _, _ := isOrderedList(peek); pIndent >= 0 {
							nestedLines = append(nestedLines, trimmedNL)
							i++
						} else if pIndent, _ := isUnorderedList(peek); pIndent >= 0 {
							nestedLines = append(nestedLines, trimmedNL)
							i++
						} else {
							break
						}
					} else {
						break
					}
				} else {
					break
				}
			}
			contentParts = append(contentParts, strings.Join(nestedLines, "\n"))
			continue
		}
		
		if !isBlank(line) {
			contentParts = append(contentParts, line)
		}
		i++
	}
	
	return strings.Join(contentParts, "\n")
}
