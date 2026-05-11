package mdparser

import (
	"strings"
)

type listItemData struct {
	line     string
	indent   int
	ordered  bool
	startNum int
	content  string
}

func parseListItemData(line string) *listItemData {
	if indent, num, content := isOrderedList(line); indent >= 0 {
		return &listItemData{
			line:     line,
			indent:   indent,
			ordered:  true,
			startNum: num,
			content:  content,
		}
	}
	if indent, content := isUnorderedList(line); indent >= 0 {
		return &listItemData{
			line:     line,
			indent:   indent,
			ordered:  false,
			startNum: 1,
			content:  content,
		}
	}
	return nil
}

func (p *Parser) parseListBlock() *Block {
	if p.index >= len(p.lines) {
		return nil
	}
	
	firstItem := parseListItemData(p.lines[p.index])
	if firstItem == nil {
		return nil
	}
	
	baseIndent := firstItem.indent
	isOrdered := firstItem.ordered
	startNum := firstItem.startNum
	
	var items []*Block
	
	for p.index < len(p.lines) {
		item := p.parseListItemAtLevel(baseIndent, isOrdered)
		if item == nil {
			break
		}
		items = append(items, item)
	}
	
	return &Block{
		Type:     BlockList,
		Ordered:  isOrdered,
		StartNum: startNum,
		Children: items,
	}
}

func (p *Parser) parseListItemAtLevel(baseIndent int, parentOrdered bool) *Block {
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
	
	itemData := parseListItemData(line)
	if itemData == nil {
		return nil
	}
	
	if itemData.indent < baseIndent {
		return nil
	}
	
	if itemData.indent == baseIndent && itemData.ordered != parentOrdered {
		return nil
	}
	
	if itemData.indent > baseIndent {
		return nil
	}
	
	hasCheckbox, checked, checkboxContent := isCheckbox(itemData.content)
	
	var contentParts []string
	var nestedBlocks []*Block
	
	if !hasCheckbox || checkboxContent != "" {
		if hasCheckbox {
			contentParts = append(contentParts, checkboxContent)
		} else if itemData.content != "" {
			contentParts = append(contentParts, itemData.content)
		}
	}
	
	p.index++
	
	var pendingBlankLines []string
	
	for p.index < len(p.lines) {
		nextLine := p.lines[p.index]
		
		if isBlank(nextLine) {
			pendingBlankLines = append(pendingBlankLines, "")
			p.index++
			continue
		}
		
		nextIndent := countLeadingSpaces(nextLine)
		nextItem := parseListItemData(nextLine)
		
		if nextItem != nil {
			if nextItem.indent == baseIndent {
				if nextItem.ordered == parentOrdered {
					break
				}
			}
			
			if nextItem.indent > baseIndent {
				if len(pendingBlankLines) > 0 {
					contentParts = append(contentParts, pendingBlankLines...)
					pendingBlankLines = nil
				}
				
				sublist := p.parseSublistAtLevel(nextItem.indent, nextItem.ordered, nextItem.startNum)
				if sublist != nil {
					nestedBlocks = append(nestedBlocks, sublist)
				}
				continue
			}
			
			if nextItem.indent < baseIndent {
				break
			}
		}
		
		if nextIndent > baseIndent {
			if len(pendingBlankLines) > 0 {
				contentParts = append(contentParts, pendingBlankLines...)
				pendingBlankLines = nil
			}
			
			trimmed := nextLine
			if len(nextLine) > baseIndent {
				trimmed = nextLine[baseIndent:]
			}
			
			if isFenceFlag, _ := isFence(strings.TrimLeft(trimmed, " \t")); isFenceFlag {
				codeLines := []string{strings.TrimLeft(trimmed, " \t")}
				p.index++
				for p.index < len(p.lines) {
					cl := p.lines[p.index]
					trimmedCL := cl
					if len(cl) > baseIndent {
						trimmedCL = cl[baseIndent:]
					}
					codeLines = append(codeLines, trimmedCL)
					if isFenceFlag2, _ := isFence(strings.TrimLeft(trimmedCL, " \t")); isFenceFlag2 {
						p.index++
						break
					}
					p.index++
				}
				contentParts = append(contentParts, strings.Join(codeLines, "\n"))
				continue
			}
			
			if isQ, _ := isQuote(strings.TrimLeft(trimmed, " \t")); isQ {
				quoteLines := []string{strings.TrimLeft(trimmed, " \t")}
				p.index++
				for p.index < len(p.lines) {
					ql := p.lines[p.index]
					trimmedQL := ql
					if len(ql) > baseIndent {
						trimmedQL = ql[baseIndent:]
					}
					trimmedQL = strings.TrimLeft(trimmedQL, " \t")
					if isQ2, _ := isQuote(trimmedQL); isQ2 {
						quoteLines = append(quoteLines, trimmedQL)
						p.index++
					} else if isBlank(trimmedQL) {
						quoteLines = append(quoteLines, trimmedQL)
						p.index++
					} else {
						break
					}
				}
				contentParts = append(contentParts, strings.Join(quoteLines, "\n"))
				continue
			}
			
			contentParts = append(contentParts, trimmed)
			p.index++
			continue
		}
		
		if nextIndent == baseIndent {
			if isBlank(nextLine) {
				p.index++
				continue
			}
			if nextItem == nil {
				contentParts = append(contentParts, strings.TrimLeft(nextLine, " \t"))
				p.index++
				continue
			}
		}
		
		break
	}
	
	content := strings.Join(contentParts, "\n")
	
	return &Block{
		Type:        BlockListItem,
		Content:     content,
		Children:    nestedBlocks,
		Ordered:     itemData.ordered,
		StartNum:    itemData.startNum,
		HasCheckbox: hasCheckbox,
		Checked:     checked,
	}
}

func (p *Parser) parseSublistAtLevel(baseIndent int, isOrdered bool, startNum int) *Block {
	var items []*Block
	
	for p.index < len(p.lines) {
		item := p.parseListItemAtLevel(baseIndent, isOrdered)
		if item == nil {
			break
		}
		items = append(items, item)
	}
	
	if len(items) == 0 {
		return nil
	}
	
	return &Block{
		Type:     BlockList,
		Ordered:  isOrdered,
		StartNum: startNum,
		Children: items,
	}
}
