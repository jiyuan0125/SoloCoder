use crate::block_parser::*;
use crate::inline_parser::*;
use crate::utils::escape_html;

pub fn generate_html(blocks: &[BlockElement]) -> String {
    let mut html = String::new();
    html.push_str("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n<title>Markdown Output</title>\n</head>\n<body>\n");

    for block in blocks {
        html.push_str(&generate_block(block));
    }

    html.push_str("</body>\n</html>\n");
    html
}

fn generate_block(block: &BlockElement) -> String {
    match block {
        BlockElement::Heading { level, content } => {
            let tag = format!("h{}", level);
            let inner = generate_inlines(content);
            format!("<{}>{}</{}>\n", tag, inner, tag)
        }
        BlockElement::Paragraph(content) => {
            let inner = generate_inlines(content);
            format!("<p>{}</p>\n", inner)
        }
        BlockElement::UnorderedList(items) => {
            let mut html = String::from("<ul>\n");
            for item in items {
                html.push_str(&generate_list_item(item));
            }
            html.push_str("</ul>\n");
            html
        }
        BlockElement::OrderedList(items) => {
            let mut html = String::from("<ol>\n");
            for item in items {
                html.push_str(&generate_list_item(item));
            }
            html.push_str("</ol>\n");
            html
        }
        BlockElement::CodeBlock { language: _language, content } => {
            format!("<pre><code>{}</code></pre>\n", content)
        }
        BlockElement::Table { headers, rows } => {
            let mut html = String::from("<table>\n");
            html.push_str("<thead>\n<tr>\n");
            for header in headers {
                html.push_str(&format!("<th>{}</th>\n", generate_inlines(header)));
            }
            html.push_str("</tr>\n</thead>\n");

            if !rows.is_empty() {
                html.push_str("<tbody>\n");
                for row in rows {
                    html.push_str("<tr>\n");
                    for cell in row {
                        html.push_str(&format!("<td>{}</td>\n", generate_inlines(cell)));
                    }
                    html.push_str("</tr>\n");
                }
                html.push_str("</tbody>\n");
            }

            html.push_str("</table>\n");
            html
        }
        BlockElement::HorizontalRule => {
            String::from("<hr>\n")
        }
    }
}

fn generate_list_item(item: &ListItem) -> String {
    let mut html = String::from("<li>");
    html.push_str(&generate_inlines(&item.content));

    if let Some(nested) = &item.nested {
        for block in nested {
            html.push_str(&generate_block(block));
        }
    }

    html.push_str("</li>\n");
    html
}

fn generate_inlines(inlines: &[InlineElement]) -> String {
    let mut html = String::new();
    for inline in inlines {
        html.push_str(&generate_inline(inline));
    }
    html
}

fn generate_inline(inline: &InlineElement) -> String {
    match inline {
        InlineElement::Text(text) => escape_html(text),
        InlineElement::Code(code) => format!("<code>{}</code>", code),
        InlineElement::Bold(content) => {
            format!("<strong>{}</strong>", generate_inlines(content))
        }
        InlineElement::Italic(content) => {
            format!("<em>{}</em>", generate_inlines(content))
        }
        InlineElement::Link { text, url } => {
            format!("<a href=\"{}\">{}</a>", escape_html(url), escape_html(text))
        }
        InlineElement::Image { alt, url } => {
            format!("<img src=\"{}\" alt=\"{}\">", escape_html(url), escape_html(alt))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_heading() {
        let blocks = vec![BlockElement::Heading {
            level: 1,
            content: vec![InlineElement::Text("Hello".to_string())],
        }];
        let html = generate_html(&blocks);
        assert!(html.contains("<h1>Hello</h1>"));
    }

    #[test]
    fn test_generate_paragraph() {
        let blocks = vec![BlockElement::Paragraph(vec![
            InlineElement::Text("Hello ".to_string()),
            InlineElement::Bold(vec![InlineElement::Text("world".to_string())]),
        ])];
        let html = generate_html(&blocks);
        assert!(html.contains("<p>Hello <strong>world</strong></p>"));
    }

    #[test]
    fn test_generate_code_block() {
        let blocks = vec![BlockElement::CodeBlock {
            language: Some("rust".to_string()),
            content: "fn main() { < > & }".to_string(),
        }];
        let html = generate_html(&blocks);
        assert!(html.contains("<pre><code>fn main() { < > & }</code></pre>"));
    }

    #[test]
    fn test_escape_html_in_text() {
        let blocks = vec![BlockElement::Paragraph(vec![
            InlineElement::Text("<div> & </div>".to_string()),
        ])];
        let html = generate_html(&blocks);
        assert!(html.contains("&lt;div&gt; &amp; &lt;/div&gt;"));
    }

    #[test]
    fn test_escape_html_not_in_code() {
        let blocks = vec![BlockElement::Paragraph(vec![
            InlineElement::Code("<div> & </div>".to_string()),
        ])];
        let html = generate_html(&blocks);
        assert!(html.contains("<code><div> & </div></code>"));
    }

    #[test]
    fn test_full_html_document() {
        let blocks = vec![BlockElement::Heading {
            level: 1,
            content: vec![InlineElement::Text("Test".to_string())],
        }];
        let html = generate_html(&blocks);
        assert!(html.starts_with("<!DOCTYPE html>"));
        assert!(html.contains("<html>"));
        assert!(html.contains("<body>"));
        assert!(html.contains("</body>"));
        assert!(html.contains("</html>"));
    }
}
