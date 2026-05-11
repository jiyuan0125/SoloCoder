package mdparser

import (
	"strconv"
	"strings"
)

func renderBlocks(blocks []*Block) string {
	var html strings.Builder
	
	for _, block := range blocks {
		html.WriteString(renderBlock(block))
	}
	
	return html.String()
}

func renderBlock(block *Block) string {
	switch block.Type {
	case BlockHeading:
		return renderHeading(block)
	case BlockParagraph:
		return renderParagraph(block)
	case BlockCode:
		return renderCode(block)
	case BlockQuote:
		return renderQuote(block)
	case BlockList:
		return renderList(block)
	case BlockListItem:
		return renderListItem(block)
	case BlockTable:
		return renderTable(block)
	case BlockTableHeader:
		return renderTableHeader(block)
	case BlockTableRow:
		return renderTableRow(block)
	case BlockTableCell:
		return renderTableCell(block)
	default:
		return ""
	}
}

func renderHeading(block *Block) string {
	content := parseInline(block.Content)
	return "<h" + strconv.Itoa(block.Level) + ">" + content + "</h" + strconv.Itoa(block.Level) + ">\n"
}

func renderParagraph(block *Block) string {
	content := strings.ReplaceAll(block.Content, "\n", " ")
	content = parseInline(content)
	return "<p>" + content + "</p>\n"
}

func renderCode(block *Block) string {
	content := escapeHTML(block.Content)
	lang := block.Language
	if lang != "" {
		return "<pre><code class=\"language-" + lang + "\">" + content + "</code></pre>\n"
	}
	return "<pre><code>" + content + "</code></pre>\n"
}

func renderQuote(block *Block) string {
	var html strings.Builder
	html.WriteString("<blockquote>\n")
	for _, child := range block.Children {
		html.WriteString(renderBlock(child))
	}
	html.WriteString("</blockquote>\n")
	return html.String()
}

func renderList(block *Block) string {
	var html strings.Builder
	
	tag := "ul"
	attrs := ""
	if block.Ordered {
		tag = "ol"
		if block.StartNum != 1 {
			attrs = " start=\"" + strconv.Itoa(block.StartNum) + "\""
		}
	}
	
	html.WriteString("<" + tag + attrs + ">\n")
	for _, child := range block.Children {
		html.WriteString(renderListItem(child))
	}
	html.WriteString("</" + tag + ">\n")
	
	return html.String()
}

func renderListItem(block *Block) string {
	var html strings.Builder
	html.WriteString("<li>")
	
	if block.HasCheckbox {
		checked := ""
		if block.Checked {
			checked = " checked"
		}
		html.WriteString("<input type=\"checkbox\"" + checked + " disabled> ")
	}
	
	content := block.Content
	if content != "" {
		innerParser := &Parser{
			lines: strings.Split(content, "\n"),
			index: 0,
		}
		innerBlocks := innerParser.parseBlocks()
		
		if len(innerBlocks) == 1 && innerBlocks[0].Type == BlockParagraph {
			para := innerBlocks[0]
			paraContent := strings.ReplaceAll(para.Content, "\n", " ")
			paraContent = parseInline(paraContent)
			html.WriteString(paraContent)
		} else {
			for _, innerBlock := range innerBlocks {
				html.WriteString(renderBlock(innerBlock))
			}
		}
	}
	
	for _, nestedBlock := range block.Children {
		html.WriteString(renderBlock(nestedBlock))
	}
	
	html.WriteString("</li>\n")
	return html.String()
}

func renderTable(block *Block) string {
	var html strings.Builder
	html.WriteString("<table>\n")
	
	for i, row := range block.Children {
		if i == 0 {
			html.WriteString("<thead>\n")
			html.WriteString(renderTableRow(row))
			html.WriteString("</thead>\n")
			html.WriteString("<tbody>\n")
		} else {
			html.WriteString(renderTableRow(row))
		}
	}
	
	if len(block.Children) > 1 {
		html.WriteString("</tbody>\n")
	}
	html.WriteString("</table>\n")
	
	return html.String()
}

func renderTableHeader(block *Block) string {
	return ""
}

func renderTableRow(block *Block) string {
	var html strings.Builder
	html.WriteString("<tr>\n")
	for _, cell := range block.Children {
		html.WriteString(renderTableCell(cell))
	}
	html.WriteString("</tr>\n")
	return html.String()
}

func renderTableCell(block *Block) string {
	var html strings.Builder
	tag := "td"
	if block.IsHeader {
		tag = "th"
	}
	
	align := ""
	if block.CellAlign == "center" {
		align = " style=\"text-align:center\""
	} else if block.CellAlign == "right" {
		align = " style=\"text-align:right\""
	}
	
	content := parseInline(block.Content)
	html.WriteString("<" + tag + align + ">" + content + "</" + tag + ">\n")
	return html.String()
}

func renderInlineNodes(nodes []*InlineNode) string {
	var html strings.Builder
	
	for _, node := range nodes {
		html.WriteString(renderInlineNode(node))
	}
	
	return html.String()
}

func renderInlineNode(node *InlineNode) string {
	switch node.Type {
	case InlineText:
		return escapeHTML(node.Content)
	case InlineBold:
		return "<strong>" + renderInlineNodes(node.Children) + "</strong>"
	case InlineItalic:
		return "<em>" + renderInlineNodes(node.Children) + "</em>"
	case InlineLink:
		url := escapeHTML(node.URL)
		title := ""
		if node.Title != "" {
			title = " title=\"" + escapeHTML(node.Title) + "\""
		}
		content := renderInlineNodes(tokenizeInline(node.Content))
		return "<a href=\"" + url + "\"" + title + ">" + content + "</a>"
	case InlineCode:
		return "<code>" + escapeHTML(node.Content) + "</code>"
	case InlineStrikethrough:
		return "<del>" + renderInlineNodes(node.Children) + "</del>"
	case InlineImage:
		alt := escapeHTML(node.Alt)
		url := escapeHTML(node.URL)
		title := ""
		if node.Title != "" {
			title = " title=\"" + escapeHTML(node.Title) + "\""
		}
		return "<img src=\"" + url + "\" alt=\"" + alt + "\"" + title + ">"
	default:
		return ""
	}
}
