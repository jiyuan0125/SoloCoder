package mdparser

import (
	"strings"
)

func Convert(markdown string) string {
	if markdown == "" {
		return ""
	}
	
	p := &Parser{
		lines: strings.Split(markdown, "\n"),
		index: 0,
	}
	
	blocks := p.parseBlocks()
	return renderBlocks(blocks)
}

func (p *Parser) parseBlocks() []*Block {
	var blocks []*Block
	
	for p.index < len(p.lines) {
		line := p.lines[p.index]
		
		if isBlank(line) {
			p.index++
			continue
		}
		
		if level, content := isHeading(line); level > 0 {
			blocks = append(blocks, &Block{
				Type:    BlockHeading,
				Level:   level,
				Content: content,
			})
			p.index++
			continue
		}
		
		if isFence, _ := isFence(line); isFence {
			codeBlock := p.parseCodeBlock()
			if codeBlock != nil {
				blocks = append(blocks, codeBlock)
			}
			continue
		}
		
		if isQuote, _ := isQuote(line); isQuote {
			quoteBlock := p.parseQuoteBlock()
			if quoteBlock != nil {
				blocks = append(blocks, quoteBlock)
			}
			continue
		}
		
		if indent, _ := isUnorderedList(line); indent >= 0 {
			listBlock := p.parseListBlock()
			if listBlock != nil {
				blocks = append(blocks, listBlock)
			}
			continue
		}
		
		if indent, _, _ := isOrderedList(line); indent >= 0 {
			listBlock := p.parseListBlock()
			if listBlock != nil {
				blocks = append(blocks, listBlock)
			}
			continue
		}
		
		if tableBlock := p.tryParseTable(); tableBlock != nil {
			blocks = append(blocks, tableBlock)
			continue
		}
		
		paraBlock := p.parseParagraph()
		if paraBlock != nil {
			blocks = append(blocks, paraBlock)
		}
	}
	
	return blocks
}

func (p *Parser) parseParagraph() *Block {
	var lines []string
	
	for p.index < len(p.lines) {
		line := p.lines[p.index]
		
		if isBlank(line) {
			if len(lines) > 0 {
				break
			}
			p.index++
			continue
		}
		
		if level, _ := isHeading(line); level > 0 {
			break
		}
		
		if isFence, _ := isFence(line); isFence {
			break
		}
		
		if isQuote, _ := isQuote(line); isQuote {
			break
		}
		
		if indent, _ := isUnorderedList(line); indent >= 0 {
			break
		}
		
		if indent, _, _ := isOrderedList(line); indent >= 0 {
			break
		}
		
		lines = append(lines, line)
		p.index++
	}
	
	if len(lines) == 0 {
		return nil
	}
	
	return &Block{
		Type:    BlockParagraph,
		Content: strings.Join(lines, "\n"),
	}
}

func (p *Parser) parseCodeBlock() *Block {
	startLine := p.lines[p.index]
	_, language := isFence(startLine)
	
	var contentLines []string
	p.index++
	
	for p.index < len(p.lines) {
		line := p.lines[p.index]
		if isFence, _ := isFence(line); isFence {
			p.index++
			break
		}
		contentLines = append(contentLines, line)
		p.index++
	}
	
	content := strings.Join(contentLines, "\n")
	
	return &Block{
		Type:     BlockCode,
		Content:  content,
		Language: language,
	}
}

func (p *Parser) parseQuoteBlock() *Block {
	var contentLines []string
	
	for p.index < len(p.lines) {
		line := p.lines[p.index]
		
		if isBlank(line) {
			break
		}
		
		if isQ, qContent := isQuote(line); isQ {
			contentLines = append(contentLines, qContent)
			p.index++
		} else {
			break
		}
	}
	
	if len(contentLines) == 0 {
		return nil
	}
	
	content := strings.Join(contentLines, "\n")
	innerParser := &Parser{
		lines: strings.Split(content, "\n"),
		index: 0,
	}
	children := innerParser.parseBlocks()
	
	return &Block{
		Type:     BlockQuote,
		Children: children,
	}
}

func (p *Parser) tryParseTable() *Block {
	if p.index+2 >= len(p.lines) {
		return nil
	}
	
	headerLine := p.lines[p.index]
	sepLine := p.lines[p.index+1]
	
	if !tableRowRegex.MatchString(headerLine) {
		return nil
	}
	
	if !tableRowRegex.MatchString(sepLine) {
		return nil
	}
	
	if !tableSepRegex.MatchString(trimSpaces(sepLine)) {
		return nil
	}
	
	headerCells := splitTableRow(headerLine)
	
	hasNonEmptyHeader := false
	for _, cell := range headerCells {
		if trimSpaces(cell) != "" {
			hasNonEmptyHeader = true
			break
		}
	}
	
	if !hasNonEmptyHeader {
		return nil
	}
	
	sepCells := splitTableRow(sepLine)
	alignments := make([]string, len(headerCells))
	for i := range alignments {
		alignments[i] = "left"
	}
	
	for i, sepCell := range sepCells {
		if i >= len(headerCells) {
			break
		}
		trimmed := trimSpaces(sepCell)
		left := strings.HasPrefix(trimmed, ":")
		right := strings.HasSuffix(trimmed, ":")
		
		if left && right {
			alignments[i] = "center"
		} else if right {
			alignments[i] = "right"
		} else {
			alignments[i] = "left"
		}
	}
	
	headerBlock := &Block{
		Type:       BlockTableHeader,
		Children:   make([]*Block, len(headerCells)),
		Alignments: alignments,
	}
	
	for i, cellContent := range headerCells {
		headerBlock.Children[i] = &Block{
			Type:      BlockTableCell,
			Content:   cellContent,
			IsHeader:  true,
			CellAlign: alignments[i],
		}
	}
	
	rows := []*Block{headerBlock}
	
	p.index += 2
	
	for p.index < len(p.lines) {
		line := p.lines[p.index]
		
		if isBlank(line) {
			p.index++
			continue
		}
		
		if !tableRowRegex.MatchString(line) {
			break
		}
		
		dataCells := splitTableRow(line)
		rowBlock := &Block{
			Type:     BlockTableRow,
			Children: make([]*Block, len(headerCells)),
		}
		
		for i := range headerCells {
			cellContent := ""
			if i < len(dataCells) {
				cellContent = dataCells[i]
			}
			rowBlock.Children[i] = &Block{
				Type:      BlockTableCell,
				Content:   cellContent,
				IsHeader:  false,
				CellAlign: alignments[i],
			}
		}
		
		rows = append(rows, rowBlock)
		p.index++
	}
	
	return &Block{
		Type:       BlockTable,
		Children:   rows,
		Alignments: alignments,
	}
}

func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	
	if line == "" {
		return []string{""}
	}
	
	parts := strings.Split(line, "|")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = trimSpaces(part)
	}
	return result
}
