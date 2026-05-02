use scraper::{Html, Selector, ElementRef};
use std::collections::HashSet;
use std::path::{Path, PathBuf};
use url::Url;

use crate::path_mapper::PathMapper;

#[derive(Debug, Clone)]
pub struct RewriteResult {
    pub rewritten_html: String,
    pub extracted_links: HashSet<Url>,
}

pub struct HtmlRewriter {
    base_url: Url,
    path_mapper: PathMapper,
}

impl HtmlRewriter {
    pub fn new(base_url: Url, output_dir: impl Into<PathBuf>) -> Self {
        Self {
            base_url: base_url.clone(),
            path_mapper: PathMapper::new(output_dir),
        }
    }

    pub fn rewrite(&self, html_content: &str, current_url: &Url) -> RewriteResult {
        let document = Html::parse_document(html_content);
        let mut extracted_links = HashSet::new();
        let mut rewritten_html = html_content.to_string();

        let link_selectors = vec![
            ("a", "href", LinkType::Html),
            ("link", "href", LinkType::Resource),
            ("script", "src", LinkType::Resource),
            ("img", "src", LinkType::Resource),
            ("video", "src", LinkType::Resource),
            ("audio", "src", LinkType::Resource),
            ("source", "src", LinkType::Resource),
            ("object", "data", LinkType::Resource),
            ("embed", "src", LinkType::Resource),
            ("iframe", "src", LinkType::Resource),
        ];

        let mut replacements: Vec<(String, String)> = Vec::new();

        for (tag, attr, link_type) in link_selectors {
            let selector = Selector::parse(&format!("{}[{}]", tag, attr)).unwrap();
            
            for element in document.select(&selector) {
                if let Some(attr_value) = element.value().attr(attr) {
                    if let Some((original_url, local_path)) = self.process_link(
                        attr_value, 
                        current_url, 
                        link_type, 
                        &mut extracted_links
                    ) {
                        let original_pattern = format!(r#"{}="{}""#, attr, original_url);
                        let new_pattern = format!(r#"{}="{}""#, attr, local_path);
                        replacements.push((original_pattern, new_pattern));

                        let original_pattern_single = format!("{}='{}'", attr, original_url);
                        let new_pattern_single = format!("{}='{}'", attr, local_path);
                        replacements.push((original_pattern_single, new_pattern_single));
                    }
                }
            }
        }

        for (old, new) in replacements {
            rewritten_html = rewritten_html.replace(&old, &new);
        }

        RewriteResult {
            rewritten_html,
            extracted_links,
        }
    }

    fn process_link(
        &self,
        link: &str,
        current_url: &Url,
        link_type: LinkType,
        extracted_links: &mut HashSet<Url>,
    ) -> Option<(String, String)> {
        let link = link.trim();
        if link.is_empty() {
            return None;
        }

        if link.starts_with("mailto:") || 
           link.starts_with("tel:") || 
           link.starts_with("javascript:") ||
           link.starts_with("data:") ||
           link.starts_with("#") {
            return None;
        }

        let parsed_url = match current_url.join(link) {
            Ok(url) => url,
            Err(_) => return None,
        };

        if !PathMapper::is_same_domain(&parsed_url, &self.base_url) {
            return None;
        }

        let is_html = link_type == LinkType::Html && PathMapper::is_html_url(&parsed_url);
        
        if is_html || link_type == LinkType::Resource {
            extracted_links.insert(parsed_url.clone());
        }

        let local_path = self.path_mapper.url_to_local_path(&parsed_url, is_html);
        let current_local_path = self.path_mapper.url_to_local_path(current_url, PathMapper::is_html_url(current_url));

        let from_dir = current_local_path.parent().unwrap_or_else(|| Path::new(""));
        let to_dir = local_path.parent().unwrap_or_else(|| Path::new(""));

        let relative = match PathMapper::relative_path(from_dir, to_dir) {
            Some(r) => r,
            None => return None,
        };

        let file_name = local_path.file_name().unwrap_or_default();
        let mut relative_path = relative;
        relative_path.push(file_name);

        let relative_str = relative_path.to_string_lossy().replace('\\', "/");
        
        Some((link.to_string(), relative_str))
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum LinkType {
    Html,
    Resource,
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    #[test]
    fn test_rewrite_internal_link() {
        let temp_dir = TempDir::new().unwrap();
        let base_url = Url::parse("https://example.com/").unwrap();
        let rewriter = HtmlRewriter::new(base_url, temp_dir.path());

        let current_url = Url::parse("https://example.com/about/team.html").unwrap();
        let html = r#"<html><head><link rel="stylesheet" href="/css/style.css"></head><body><a href="https://example.com/contact">Contact</a><img src="/images/logo.png"></body></html>"#;

        let result = rewriter.rewrite(html, &current_url);

        assert!(result.rewritten_html.contains("../contact.html"));
        assert!(result.rewritten_html.contains("../css/style.css"));
        assert!(result.rewritten_html.contains("../images/logo.png"));

        assert!(result.extracted_links.contains(
            &Url::parse("https://example.com/contact").unwrap()
        ));
        assert!(result.extracted_links.contains(
            &Url::parse("https://example.com/css/style.css").unwrap()
        ));
        assert!(result.extracted_links.contains(
            &Url::parse("https://example.com/images/logo.png").unwrap()
        ));
    }

    #[test]
    fn test_external_link_unchanged() {
        let temp_dir = TempDir::new().unwrap();
        let base_url = Url::parse("https://example.com/").unwrap();
        let rewriter = HtmlRewriter::new(base_url, temp_dir.path());

        let current_url = Url::parse("https://example.com/page.html").unwrap();
        let html = r#"<a href="https://external.com/page">External</a>"#;

        let result = rewriter.rewrite(html, &current_url);

        assert!(result.rewritten_html.contains("https://external.com/page"));
        assert!(!result.extracted_links.contains(
            &Url::parse("https://external.com/page").unwrap()
        ));
    }

    #[test]
    fn test_special_links_unchanged() {
        let temp_dir = TempDir::new().unwrap();
        let base_url = Url::parse("https://example.com/").unwrap();
        let rewriter = HtmlRewriter::new(base_url, temp_dir.path());

        let current_url = Url::parse("https://example.com/page.html").unwrap();
        let html = r#"
            <a href="mailto:test@example.com">Email</a>
            <a href="tel:+123456">Phone</a>
            <a href="javascript:alert(1)">JS</a>
            <a href="#section">Section</a>
        "#;

        let result = rewriter.rewrite(html, &current_url);

        assert!(result.rewritten_html.contains("mailto:test@example.com"));
        assert!(result.rewritten_html.contains("tel:+123456"));
        assert!(result.rewritten_html.contains("javascript:alert(1)"));
        assert!(result.rewritten_html.contains("#section"));
    }

    #[test]
    fn test_relative_link_resolution() {
        let temp_dir = TempDir::new().unwrap();
        let base_url = Url::parse("https://example.com/").unwrap();
        let rewriter = HtmlRewriter::new(base_url, temp_dir.path());

        let current_url = Url::parse("https://example.com/about/team.html").unwrap();
        let html = r#"<a href="../contact">Contact</a>"#;

        let result = rewriter.rewrite(html, &current_url);

        assert!(result.rewritten_html.contains("../contact.html"));
    }

    #[test]
    fn test_query_and_fragment_stripping() {
        let temp_dir = TempDir::new().unwrap();
        let base_url = Url::parse("https://example.com/").unwrap();
        let rewriter = HtmlRewriter::new(base_url, temp_dir.path());

        let current_url = Url::parse("https://example.com/page.html").unwrap();
        let html = r#"<a href="/search?q=test#section">Search</a>"#;

        let result = rewriter.rewrite(html, &current_url);

        assert!(result.rewritten_html.contains("search.html"));
        assert!(!result.rewritten_html.contains("q=test"));
        assert!(!result.rewritten_html.contains("section"));
    }
}
