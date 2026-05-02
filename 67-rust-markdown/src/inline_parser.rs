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
    let mut chars = text.chars().peekable();

    while let Some(c) = chars.peek() {
        match *c {
            '`' => {
                if !current_text.is_empty() {
                    elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                }
                chars.next();
                match parse_inline_code(&mut chars) {
                    Ok(code) => {
                        elements.push(InlineElement::Code(code));
                    }
                    Err(consumed) => {
                        current_text.push('`');
                        current_text.push_str(&consumed);
                    }
                }
            }
            '*' => {
                if !current_text.is_empty() {
                    elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                }
                chars.next();
                if chars.peek() == Some(&'*') {
                    chars.next();
                    match parse_bold(&mut chars) {
                        Ok(bold) => {
                            elements.push(InlineElement::Bold(bold));
                        }
                        Err(consumed) => {
                            current_text.push_str("**");
                            current_text.push_str(&consumed);
                        }
                    }
                } else {
                    match parse_italic(&mut chars) {
                        Ok(italic) => {
                            elements.push(InlineElement::Italic(italic));
                        }
                        Err(consumed) => {
                            current_text.push('*');
                            current_text.push_str(&consumed);
                        }
                    }
                }
            }
            '[' => {
                if !current_text.is_empty() {
                    elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                }
                chars.next();
                match parse_link_or_image(&mut chars) {
                    Ok(link) => {
                        elements.push(link);
                    }
                    Err(consumed) => {
                        current_text.push('[');
                        current_text.push_str(&consumed);
                    }
                }
            }
            '!' => {
                chars.next();
                if chars.peek() == Some(&'[') {
                    if !current_text.is_empty() {
                        elements.push(InlineElement::Text(std::mem::take(&mut current_text)));
                    }
                    chars.next();
                    match parse_link_or_image(&mut chars) {
                        Ok(link) => {
                            match link {
                                InlineElement::Link { text, url } => {
                                    elements.push(InlineElement::Image { alt: text, url });
                                }
                                _ => unreachable!(),
                            }
                        }
                        Err(consumed) => {
                            current_text.push_str("![");
                            current_text.push_str(&consumed);
                        }
                    }
                } else {
                    current_text.push('!');
                }
            }
            _ => {
                current_text.push(chars.next().unwrap());
            }
        }
    }

    if !current_text.is_empty() {
        elements.push(InlineElement::Text(current_text));
    }

    elements
}

fn parse_inline_code<I: Iterator<Item = char>>(
    chars: &mut std::iter::Peekable<I>,
) -> Result<String, String> {
    let mut code = String::new();
    while let Some(c) = chars.next() {
        if c == '`' {
            return Ok(code);
        }
        code.push(c);
    }
    Err(code)
}

fn parse_bold<I: Iterator<Item = char>>(
    chars: &mut std::iter::Peekable<I>,
) -> Result<Vec<InlineElement>, String> {
    let mut content = String::new();
    let mut asterisk_count = 0;

    while let Some(c) = chars.peek() {
        match *c {
            '*' => {
                asterisk_count += 1;
                chars.next();
                if asterisk_count >= 2 {
                    let inner = parse_inline(&content);
                    return Ok(inner);
                }
            }
            _ => {
                if asterisk_count > 0 {
                    content.push_str(&"*".repeat(asterisk_count));
                    asterisk_count = 0;
                }
                content.push(chars.next().unwrap());
            }
        }
    }

    if asterisk_count > 0 {
        content.push_str(&"*".repeat(asterisk_count));
    }
    Err(content)
}

fn parse_italic<I: Iterator<Item = char>>(
    chars: &mut std::iter::Peekable<I>,
) -> Result<Vec<InlineElement>, String> {
    let mut content = String::new();

    while let Some(c) = chars.peek() {
        match *c {
            '*' => {
                chars.next();
                let inner = parse_inline(&content);
                return Ok(inner);
            }
            _ => {
                content.push(chars.next().unwrap());
            }
        }
    }

    Err(content)
}

fn parse_link_or_image<I: Iterator<Item = char>>(
    chars: &mut std::iter::Peekable<I>,
) -> Result<InlineElement, String> {
    let mut consumed = String::new();
    let mut text = String::new();
    let mut bracket_count = 0;

    while let Some(c) = chars.peek() {
        match *c {
            '[' => {
                bracket_count += 1;
                let c = chars.next().unwrap();
                text.push(c);
                consumed.push(c);
            }
            ']' => {
                if bracket_count == 0 {
                    let c = chars.next().unwrap();
                    consumed.push(c);
                    break;
                } else {
                    bracket_count -= 1;
                    let c = chars.next().unwrap();
                    text.push(c);
                    consumed.push(c);
                }
            }
            _ => {
                let c = chars.next().unwrap();
                text.push(c);
                consumed.push(c);
            }
        }
    }

    if chars.peek() != Some(&'(') {
        return Err(consumed);
    }
    consumed.push('(');
    chars.next();

    let mut url = String::new();
    while let Some(c) = chars.peek() {
        match *c {
            ')' => {
                let c = chars.next().unwrap();
                consumed.push(c);
                return Ok(InlineElement::Link { text, url });
            }
            _ => {
                let c = chars.next().unwrap();
                url.push(c);
                consumed.push(c);
            }
        }
    }

    Err(consumed)
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
}
