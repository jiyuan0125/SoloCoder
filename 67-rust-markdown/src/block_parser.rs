use crate::inline_parser::parse_inline;
use crate::inline_parser::InlineElement;

#[derive(Debug, Clone, PartialEq)]
pub enum BlockElement {
    Heading { level: u8, content: Vec<InlineElement> },
    Paragraph(Vec<InlineElement>),
    UnorderedList(Vec<ListItem>),
    OrderedList(Vec<ListItem>),
    CodeBlock { language: Option<String>, content: String },
    Table { headers: Vec<Vec<InlineElement>>, rows: Vec<Vec<Vec<InlineElement>>> },
    HorizontalRule,
}

#[derive(Debug, Clone, PartialEq)]
pub struct ListItem {
    pub content: Vec<InlineElement>,
    pub nested: Option<Vec<BlockElement>>,
}

pub fn parse_blocks(input: &str) -> Vec<BlockElement> {
    let mut blocks = Vec::new();
    let lines: Vec<String> = input.lines().map(|s| s.to_string()).collect();
    let mut lines_iter = lines.iter().peekable();
    let mut paragraph_lines: Vec<String> = Vec::new();

    while let Some(line) = lines_iter.next() {
        let trimmed = line.trim();

        if trimmed.is_empty() {
            if !paragraph_lines.is_empty() {
                blocks.push(parse_paragraph(&paragraph_lines));
                paragraph_lines.clear();
            }
            continue;
        }

        if let Some(heading) = parse_heading(line) {
            if !paragraph_lines.is_empty() {
                blocks.push(parse_paragraph(&paragraph_lines));
                paragraph_lines.clear();
            }
            blocks.push(heading);
            continue;
        }

        if let Some(hr) = parse_horizontal_rule(trimmed) {
            if !paragraph_lines.is_empty() {
                blocks.push(parse_paragraph(&paragraph_lines));
                paragraph_lines.clear();
            }
            blocks.push(hr);
            continue;
        }

        if let Some((language, code_content)) = parse_code_block(line, &mut lines_iter) {
            if !paragraph_lines.is_empty() {
                blocks.push(parse_paragraph(&paragraph_lines));
                paragraph_lines.clear();
            }
            blocks.push(BlockElement::CodeBlock { language, content: code_content });
            continue;
        }

        if let Some(table) = parse_table(line, &mut lines_iter) {
            if !paragraph_lines.is_empty() {
                blocks.push(parse_paragraph(&paragraph_lines));
                paragraph_lines.clear();
            }
            blocks.push(table);
            continue;
        }

        if is_list_item(line) {
            if !paragraph_lines.is_empty() {
                blocks.push(parse_paragraph(&paragraph_lines));
                paragraph_lines.clear();
            }
            if let Some(list) = parse_list(line, &mut lines_iter) {
                blocks.push(list);
            }
            continue;
        }

        paragraph_lines.push(line.to_string());
    }

    if !paragraph_lines.is_empty() {
        blocks.push(parse_paragraph(&paragraph_lines));
    }

    blocks
}

fn parse_paragraph(lines: &[String]) -> BlockElement {
    let text = lines.join("\n");
    BlockElement::Paragraph(parse_inline(&text))
}

fn parse_heading(line: &str) -> Option<BlockElement> {
    let trimmed = line.trim_start();
    if !trimmed.starts_with('#') {
        return None;
    }

    let mut level: u8 = 0;
    for c in trimmed.chars() {
        if c == '#' {
            level += 1;
        } else {
            break;
        }
    }

    if level > 6 {
        return None;
    }

    let content_str = trimmed[level as usize..].trim_start();
    if content_str.is_empty() {
        return Some(BlockElement::Heading {
            level,
            content: Vec::new(),
        });
    }

    Some(BlockElement::Heading {
        level,
        content: parse_inline(content_str),
    })
}

fn parse_horizontal_rule(line: &str) -> Option<BlockElement> {
    let trimmed = line.trim();
    if trimmed.len() < 3 {
        return None;
    }

    let all_dashes = trimmed.chars().all(|c| c == '-');
    let all_asterisks = trimmed.chars().all(|c| c == '*');

    if all_dashes || all_asterisks {
        Some(BlockElement::HorizontalRule)
    } else {
        None
    }
}

fn parse_code_block<'a, I: Iterator<Item = &'a String>>(
    first_line: &str,
    lines: &mut std::iter::Peekable<I>,
) -> Option<(Option<String>, String)> {
    let trimmed = first_line.trim();
    if !trimmed.starts_with("```") {
        return None;
    }

    let language: Option<String> = if trimmed.len() > 3 {
        Some(trimmed[3..].trim().to_string())
    } else {
        None
    };

    let mut content = String::new();
    for line in lines.by_ref() {
        if line.trim() == "```" {
            return Some((language, content.trim_end_matches('\n').to_string()));
        }
        if !content.is_empty() {
            content.push('\n');
        }
        content.push_str(line);
    }

    None
}

fn parse_table<'a, I: Iterator<Item = &'a String>>(
    first_line: &str,
    lines: &mut std::iter::Peekable<I>,
) -> Option<BlockElement> {
    if !first_line.contains('|') {
        return None;
    }

    let headers = parse_table_row(first_line);
    if headers.is_empty() {
        return None;
    }

    let column_count = headers.len();

    match lines.peek() {
        Some(line) if is_table_separator(line, column_count) => {
            lines.next();
        }
        _ => return None,
    };

    let mut rows = Vec::new();
    while let Some(line) = lines.peek() {
        if line.trim().is_empty() {
            break;
        }
        if !line.contains('|') {
            break;
        }
        let row = parse_table_row_with_columns(line, column_count);
        rows.push(row);
        lines.next();
    }

    Some(BlockElement::Table { headers, rows })
}

fn parse_table_row(line: &str) -> Vec<Vec<InlineElement>> {
    let trimmed = line.trim();
    let without_borders = if trimmed.starts_with('|') && trimmed.ends_with('|') {
        &trimmed[1..trimmed.len() - 1]
    } else {
        trimmed
    };

    without_borders
        .split('|')
        .map(|cell| parse_inline(cell.trim()))
        .collect()
}

fn parse_table_row_with_columns(line: &str, column_count: usize) -> Vec<Vec<InlineElement>> {
    let trimmed = line.trim();
    let without_borders = if trimmed.starts_with('|') && trimmed.ends_with('|') {
        &trimmed[1..trimmed.len() - 1]
    } else {
        trimmed
    };

    if column_count <= 1 {
        return vec![parse_inline(without_borders.trim())];
    }

    let mut cells = Vec::new();
    let mut remaining = without_borders;

    for _ in 0..column_count - 1 {
        if let Some(pipe_pos) = remaining.find('|') {
            let cell = &remaining[..pipe_pos];
            cells.push(parse_inline(cell.trim()));
            remaining = &remaining[pipe_pos + 1..];
        } else {
            cells.push(Vec::new());
        }
    }

    cells.push(parse_inline(remaining.trim()));

    cells
}

fn is_table_separator(line: &str, expected_columns: usize) -> bool {
    let trimmed = line.trim();
    if !trimmed.contains('|') {
        return false;
    }

    let without_borders = if trimmed.starts_with('|') && trimmed.ends_with('|') {
        &trimmed[1..trimmed.len() - 1]
    } else {
        trimmed
    };

    let cells: Vec<&str> = without_borders.split('|').collect();
    if cells.len() != expected_columns {
        return false;
    }

    for cell in &cells {
        let cell_trimmed = cell.trim();
        if !cell_trimmed.chars().all(|c| c == '-' || c == ':') {
            return false;
        }
        if cell_trimmed.is_empty() {
            return false;
        }
    }

    true
}

fn is_list_item(line: &str) -> bool {
    let trimmed = line.trim_start();
    if trimmed.starts_with("- ") || trimmed.starts_with("* ") {
        return true;
    }

    if let Some(dot_pos) = trimmed.find('.') {
        let num_part = &trimmed[..dot_pos];
        if num_part.chars().all(|c| c.is_ascii_digit()) && trimmed[dot_pos..].starts_with(". ") {
            return true;
        }
    }

    false
}

fn get_list_indent(line: &str) -> usize {
    line.chars().take_while(|c| c.is_whitespace()).count()
}

fn parse_list<'a, I: Iterator<Item = &'a String>>(
    first_line: &str,
    lines: &mut std::iter::Peekable<I>,
) -> Option<BlockElement> {
    let is_ordered = is_ordered_list_item(first_line);
    let mut items = Vec::new();
    let base_indent = get_list_indent(first_line);

    let mut current_item_lines = vec![first_line.to_string()];

    while let Some(line) = lines.peek() {
        if line.trim().is_empty() {
            lines.next();
            continue;
        }

        let line_indent = get_list_indent(line);

        if line_indent < base_indent {
            break;
        }

        if line_indent == base_indent {
            if is_list_item(line) {
                if let Some(item) = parse_list_item(&current_item_lines, base_indent) {
                    items.push(item);
                }
                current_item_lines = vec![(*line).clone()];
                lines.next();
            } else {
                break;
            }
        } else {
            current_item_lines.push((*line).clone());
            lines.next();
        }
    }

    if let Some(item) = parse_list_item(&current_item_lines, base_indent) {
        items.push(item);
    }

    if items.is_empty() {
        None
    } else if is_ordered {
        Some(BlockElement::OrderedList(items))
    } else {
        Some(BlockElement::UnorderedList(items))
    }
}

fn is_ordered_list_item(line: &str) -> bool {
    let trimmed = line.trim_start();
    if let Some(dot_pos) = trimmed.find('.') {
        let num_part = &trimmed[..dot_pos];
        num_part.chars().all(|c| c.is_ascii_digit()) && trimmed[dot_pos..].starts_with(". ")
    } else {
        false
    }
}

fn parse_list_item(lines: &[String], base_indent: usize) -> Option<ListItem> {
    if lines.is_empty() {
        return None;
    }

    let first_line = &lines[0];
    let trimmed = first_line.trim_start();

    let content_start: usize;
    if trimmed.starts_with("- ") || trimmed.starts_with("* ") {
        content_start = first_line.len() - trimmed.len() + 2;
    } else if let Some(dot_pos) = trimmed.find('.') {
        content_start = first_line.len() - trimmed.len() + dot_pos + 2;
    } else {
        return None;
    }

    let mut content_lines = vec![first_line[content_start..].to_string()];
    let mut nested_lines = Vec::new();

    for line in &lines[1..] {
        let line_indent = get_list_indent(line);
        if line_indent >= base_indent + 2 && line_indent <= base_indent + 6 {
            nested_lines.push(line[base_indent..].to_string());
        } else {
            content_lines.push(line.trim_start().to_string());
        }
    }

    let content_text = content_lines.join("\n");
    let content = parse_inline(&content_text);

    let nested = if !nested_lines.is_empty() {
        let nested_input = nested_lines.join("\n");
        let nested_blocks = parse_blocks(&nested_input);
        if !nested_blocks.is_empty() {
            Some(nested_blocks)
        } else {
            None
        }
    } else {
        None
    };

    Some(ListItem { content, nested })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_heading() {
        let result = parse_blocks("# Heading 1");
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::Heading { level, content } => {
                assert_eq!(*level, 1);
                assert_eq!(content, &vec![InlineElement::Text("Heading 1".to_string())]);
            }
            _ => panic!("Expected heading"),
        }
    }

    #[test]
    fn test_parse_heading_level_6() {
        let result = parse_blocks("###### Heading 6");
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::Heading { level, .. } => {
                assert_eq!(*level, 6);
            }
            _ => panic!("Expected heading"),
        }
    }

    #[test]
    fn test_parse_heading_level_7_is_paragraph() {
        let result = parse_blocks("####### Too deep");
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::Paragraph(_) => {}
            _ => panic!("Expected paragraph for 7+ #"),
        }
    }

    #[test]
    fn test_horizontal_rule() {
        let result = parse_blocks("---");
        assert_eq!(result, vec![BlockElement::HorizontalRule]);

        let result = parse_blocks("***");
        assert_eq!(result, vec![BlockElement::HorizontalRule]);
    }

    #[test]
    fn test_code_block() {
        let input = "```rust\nfn main() {}\n```";
        let result = parse_blocks(input);
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::CodeBlock { language, content } => {
                assert_eq!(language, &Some("rust".to_string()));
                assert_eq!(content, "fn main() {}");
            }
            _ => panic!("Expected code block"),
        }
    }

    #[test]
    fn test_table() {
        let input = "| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |";
        let result = parse_blocks(input);
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::Table { headers, rows } => {
                assert_eq!(headers.len(), 2);
                assert_eq!(rows.len(), 1);
                assert_eq!(rows[0].len(), 2);
            }
            _ => panic!("Expected table"),
        }
    }

    #[test]
    fn test_table_with_pipe_in_cell() {
        let input = "| Name | Age | Info |\n|------|-----|------|\n| Alice | 30 | First | Second |";
        let result = parse_blocks(input);
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::Table { headers, rows } => {
                assert_eq!(headers.len(), 3);
                assert_eq!(rows.len(), 1);
                assert_eq!(rows[0].len(), 3);
                assert_eq!(rows[0][2], vec![InlineElement::Text("First | Second".to_string())]);
            }
            _ => panic!("Expected table"),
        }
    }

    #[test]
    fn test_unordered_list() {
        let input = "- Item 1\n- Item 2";
        let result = parse_blocks(input);
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::UnorderedList(items) => {
                assert_eq!(items.len(), 2);
            }
            _ => panic!("Expected unordered list"),
        }
    }

    #[test]
    fn test_ordered_list() {
        let input = "1. Item 1\n2. Item 2";
        let result = parse_blocks(input);
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::OrderedList(items) => {
                assert_eq!(items.len(), 2);
            }
            _ => panic!("Expected ordered list"),
        }
    }

    #[test]
    fn test_paragraph() {
        let result = parse_blocks("Hello world\nThis is a paragraph.");
        assert_eq!(result.len(), 1);
        match &result[0] {
            BlockElement::Paragraph(content) => {
                assert!(!content.is_empty());
            }
            _ => panic!("Expected paragraph"),
        }
    }
}
