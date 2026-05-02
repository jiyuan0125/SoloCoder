#[derive(Debug, Clone, PartialEq)]
pub enum InlineElement {
    Text(String),
    Code(String),
    Bold(Vec<InlineElement>),
    Italic(Vec<InlineElement>),
    Link { text: String, url: String },
    Image { alt: String, url: String },
}

pub fn parse_inline(text: &str) -> Vec<InlineElement> {
    let mut elements = Vec::new();
    let mut current_text = String::new();
    let chars: Vec<char> = text.chars().collect();
    let mut i = 0;

    while i < chars.len() {
        match chars[i] {
            '`' => {
                if !current_text.is_empty() {
                    elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                }
                match parse_inline_code(&chars, i + 1) {
                    Some((code, end_idx)) => {
                        elements.push(InlineElement::Code(code));
                        i = end_idx;
                    }
                    None => {
                        current_text.push('`');
                        i += 1;
                    }
                }
            }
            '*' => {
                if !current_text.is_empty() {
                    elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                }

                let is_bold = i + 1 < chars.len() && chars[i + 1] == '*';

                if is_bold {
                    match parse_bold_content(&chars, i + 2) {
                        Some((content, end_idx)) => {
                            elements.push(InlineElement::Bold(content));
                            i = end_idx;
                        }
                        None => {
                            current_text.push_str("**");
                            i += 2;
                        }
                    }
                } else {
                    match parse_italic_content(&chars, i + 1) {
                        Some((content, end_idx)) => {
                            elements.push(InlineElement::Italic(content));
                            i = end_idx;
                        }
                        None => {
                            current_text.push('*');
                            i += 1;
                        }
                    }
                }
            }
            '[' => {
                if !current_text.is_empty() {
                    elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                }
                match parse_link_or_image(&chars, i + 1, false) {
                    Some((link, end_idx)) => {
                        elements.push(link);
                        i = end_idx;
                    }
                    None => {
                        current_text.push('[');
                        i += 1;
                    }
                }
            }
            '!' => {
                if i + 1 < chars.len() && chars[i + 1] == '[' {
                    if !current_text.is_empty() {
                        elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                    }
                    match parse_link_or_image(&chars, i + 2, true) {
                        Some((link, end_idx)) => {
                            match link {
                                InlineElement::Link { text, url } => {
                                    elements.push(InlineElement::Image { alt: text, url });
                                }
                                _ => unreachable!(),
                            }
                            i = end_idx;
                        }
                        None => {
                            current_text.push_str("![");
                            i += 2;
                        }
                    }
                } else {
                    current_text.push('!');
                    i += 1;
                }
            }
            _ => {
                current_text.push(chars[i]);
                i += 1;
            }
        }
    }

    if !current_text.is_empty() {
        elements.push(InlineElement::Text(current_text));
    }

    elements
}

fn parse_inline_code(chars: &[char], start_idx: usize) -> Option<(String, usize)> {
    let mut code = String::new();
    let mut i = start_idx;

    while i < chars.len() {
        if chars[i] == '`' {
            return Some((code, i + 1));
        }
        code.push(chars[i]);
        i += 1;
    }

    None
}

fn parse_bold_content(chars: &[char], start_idx: usize) -> Option<(Vec<InlineElement>, usize)> {
    let mut i = start_idx;

    while i < chars.len() {
        if chars[i] == '*' && i + 1 < chars.len() && chars[i + 1] == '*' {
            let content: String = chars[start_idx..i].iter().collect();
            let parsed = parse_inline(&content);
            return Some((parsed, i + 2));
        }
        i += 1;
    }

    None
}

fn parse_italic_content(chars: &[char], start_idx: usize) -> Option<(Vec<InlineElement>, usize)> {
    let mut i = start_idx;

    while i < chars.len() {
        if chars[i] == '*' {
            let is_double_star = i + 1 < chars.len() && chars[i + 1] == '*';
            let is_prev_star = i > 0 && chars[i - 1] == '*';

            if !is_double_star && !is_prev_star {
                let content: String = chars[start_idx..i].iter().collect();
                let parsed = parse_inline(&content);
                return Some((parsed, i + 1));
            }

            if is_double_star {
                i += 2;
                continue;
            }
        }
        i += 1;
    }

    None
}

fn parse_link_or_image(chars: &[char], start_idx: usize, _is_image: bool) -> Option<(InlineElement, usize)> {
    let mut i = start_idx;
    let mut bracket_count = 0;
    let mut text = String::new();

    while i < chars.len() {
        match chars[i] {
            '[' => {
                bracket_count += 1;
                text.push('[');
                i += 1;
            }
            ']' => {
                if bracket_count == 0 {
                    i += 1;
                    break;
                } else {
                    bracket_count -= 1;
                    text.push(']');
                    i += 1;
                }
            }
            _ => {
                text.push(chars[i]);
                i += 1;
            }
        }
    }

    if i >= chars.len() || chars[i - 1] != ']' {
        return None;
    }

    while i < chars.len() && chars[i].is_whitespace() {
        i += 1;
    }

    if i >= chars.len() || chars[i] != '(' {
        return None;
    }
    i += 1;

    while i < chars.len() && chars[i].is_whitespace() {
        i += 1;
    }

    let mut url = String::new();
    let mut paren_count = 0;

    while i < chars.len() {
        match chars[i] {
            '(' => {
                paren_count += 1;
                url.push('(');
                i += 1;
            }
            ')' => {
                if paren_count == 0 {
                    i += 1;
                    break;
                } else {
                    paren_count -= 1;
                    url.push(')');
                    i += 1;
                }
            }
            _ => {
                url.push(chars[i]);
                i += 1;
            }
        }
    }

    url = url.trim_end().to_string();

    Some((InlineElement::Link { text, url }, i))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_text() {
        let result = parse_inline("hello world");
        assert_eq!(result, vec![InlineElement::Text("hello world".to_string())]);
    }

    #[test]
    fn test_parse_inline_code() {
        let result = parse_inline("`code`");
        assert_eq!(result, vec![InlineElement::Code("code".to_string())]);
    }

    #[test]
    fn test_parse_bold() {
        let result = parse_inline("**bold**");
        assert_eq!(result, vec![InlineElement::Bold(vec![InlineElement::Text("bold".to_string())])]);
    }

    #[test]
    fn test_parse_italic() {
        let result = parse_inline("*italic*");
        assert_eq!(result, vec![InlineElement::Italic(vec![InlineElement::Text("italic".to_string())])]);
    }

    #[test]
    fn test_parse_link() {
        let result = parse_inline("[text](http://example.com)");
        assert_eq!(result, vec![InlineElement::Link {
            text: "text".to_string(),
            url: "http://example.com".to_string(),
        }]);
    }

    #[test]
    fn test_parse_image() {
        let result = parse_inline("![alt](http://example.com/img.png)");
        assert_eq!(result, vec![InlineElement::Image {
            alt: "alt".to_string(),
            url: "http://example.com/img.png".to_string(),
        }]);
    }

    #[test]
    fn test_unclosed_code() {
        let result = parse_inline("`unclosed");
        assert_eq!(result, vec![InlineElement::Text("`unclosed".to_string())]);
    }

    #[test]
    fn test_unclosed_bold() {
        let result = parse_inline("**bold");
        assert_eq!(result, vec![InlineElement::Text("**bold".to_string())]);
    }

    #[test]
    fn test_unclosed_italic() {
        let result = parse_inline("*italic");
        assert_eq!(result, vec![InlineElement::Text("*italic".to_string())]);
    }

    #[test]
    fn test_unclosed_link() {
        let result = parse_inline("[text");
        assert_eq!(result, vec![InlineElement::Text("[text".to_string())]);
    }

    #[test]
    fn test_link_without_url() {
        let result = parse_inline("[text]");
        assert_eq!(result, vec![InlineElement::Text("[text]".to_string())]);
    }

    #[test]
    fn test_nested_bold_in_italic() {
        let result = parse_inline("*italic with **bold** inside*");
        assert_eq!(
            result,
            vec![InlineElement::Italic(vec![
                InlineElement::Text("italic with ".to_string()),
                InlineElement::Bold(vec![InlineElement::Text("bold".to_string())]),
                InlineElement::Text(" inside".to_string()),
            ])]
        );
    }

    #[test]
    fn test_nested_italic_in_bold() {
        let result = parse_inline("**bold with *italic* inside**");
        assert_eq!(
            result,
            vec![InlineElement::Bold(vec![
                InlineElement::Text("bold with ".to_string()),
                InlineElement::Italic(vec![InlineElement::Text("italic".to_string())]),
                InlineElement::Text(" inside".to_string()),
            ])]
        );
    }

    #[test]
    fn test_multiple_stars() {
        let result = parse_inline("*test **inner** test*");
        assert_eq!(
            result,
            vec![InlineElement::Italic(vec![
                InlineElement::Text("test ".to_string()),
                InlineElement::Bold(vec![InlineElement::Text("inner".to_string())]),
                InlineElement::Text(" test".to_string()),
            ])]
        );
    }
}
