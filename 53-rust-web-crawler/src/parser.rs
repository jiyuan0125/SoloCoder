use scraper::{Html, Selector};
use url::Url;

#[derive(Debug, Clone)]
pub struct ParsedPage {
    pub title: Option<String>,
    pub links: Vec<Url>,
}

pub fn parse_html(html_content: &str, base_url: &Url) -> ParsedPage {
    let document = Html::parse_document(html_content);
    
    let title = extract_title(&document);
    let links = extract_links(&document, base_url);
    
    ParsedPage { title, links }
}

fn extract_title(document: &Html) -> Option<String> {
    let title_selector = Selector::parse("title").ok()?;
    
    document.select(&title_selector).next().map(|element| {
        element.text().collect::<Vec<_>>().join(" ").trim().to_string()
    })
}

fn extract_links(document: &Html, base_url: &Url) -> Vec<Url> {
    let mut links = Vec::new();
    
    let a_selector = match Selector::parse("a[href]") {
        Ok(s) => s,
        Err(_) => return links,
    };
    
    for element in document.select(&a_selector) {
        if let Some(href) = element.value().attr("href") {
            if let Ok(url) = base_url.join(href) {
                if url.scheme() == "http" || url.scheme() == "https" {
                    links.push(url);
                }
            }
        }
    }
    
    links
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_title() {
        let html = r#"
<html>
<head>
    <title>Test Page Title</title>
</head>
<body>
    <h1>Hello</h1>
</body>
</html>
"#;
        let base_url = Url::parse("http://example.com").unwrap();
        let parsed = parse_html(html, &base_url);
        
        assert_eq!(parsed.title, Some("Test Page Title".to_string()));
    }

    #[test]
    fn test_extract_title_no_title() {
        let html = r#"
<html>
<head>
</head>
<body>
    <h1>Hello</h1>
</body>
</html>
"#;
        let base_url = Url::parse("http://example.com").unwrap();
        let parsed = parse_html(html, &base_url);
        
        assert_eq!(parsed.title, None);
    }

    #[test]
    fn test_extract_links_absolute() {
        let html = r#"
<html>
<body>
    <a href="http://example.com/page1">Page 1</a>
    <a href="https://example.org/page2">Page 2</a>
    <a href="javascript:void(0)">JS</a>
</body>
</html>
"#;
        let base_url = Url::parse("http://example.com").unwrap();
        let parsed = parse_html(html, &base_url);
        
        assert_eq!(parsed.links.len(), 2);
        assert!(parsed.links.iter().any(|u| u.as_str() == "http://example.com/page1"));
        assert!(parsed.links.iter().any(|u| u.as_str() == "https://example.org/page2"));
    }

    #[test]
    fn test_extract_links_relative() {
        let html = r#"
<html>
<body>
    <a href="/about">About</a>
    <a href="contact.html">Contact</a>
    <a href="../products">Products</a>
</body>
</html>
"#;
        let base_url = Url::parse("http://example.com/blog/").unwrap();
        let parsed = parse_html(html, &base_url);
        
        assert_eq!(parsed.links.len(), 3);
        assert!(parsed.links.iter().any(|u| u.as_str() == "http://example.com/about"));
        assert!(parsed.links.iter().any(|u| u.as_str() == "http://example.com/blog/contact.html"));
        assert!(parsed.links.iter().any(|u| u.as_str() == "http://example.com/products"));
    }

    #[test]
    fn test_extract_links_empty() {
        let html = r#"
<html>
<body>
    <h1>No Links</h1>
    <p>Just some text</p>
</body>
</html>
"#;
        let base_url = Url::parse("http://example.com").unwrap();
        let parsed = parse_html(html, &base_url);
        
        assert_eq!(parsed.links.len(), 0);
    }
}
