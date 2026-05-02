use scraper::{Html, Selector};
use url::Url;

use crate::path_mapper::PathMapper;

#[derive(Debug, Clone)]
pub struct RewriteResult {
    pub rewritten_html: String,
    pub extracted_urls: Vec<Url>,
}

pub struct LinkRewriter {
    base_url: Url,
    path_mapper: PathMapper,
}

impl LinkRewriter {
    pub fn new(base_url: Url, path_mapper: PathMapper) -> Self {
        Self {
            base_url,
            path_mapper,
        }
    }

    pub fn rewrite_html(&self, html: &str) -> RewriteResult {
        let document = Html::parse_document(html);
        let mut extracted_urls = Vec::new();

        let link_selectors = [
            ("a", "href"),
            ("link", "href"),
            ("script", "src"),
            ("img", "src"),
            ("audio", "src"),
            ("video", "src"),
            ("source", "src"),
            ("form", "action"),
            ("iframe", "src"),
        ];

        let mut rewritten_html = html.to_string();

        for (tag, attr) in link_selectors {
            let selector_str = format!("{}[{}]", tag, attr);
            let selector = match Selector::parse(&selector_str) {
                Ok(s) => s,
                Err(_) => continue,
            };

            for element in document.select(&selector) {
                if let Some(attr_value) = element.value().attr(attr) {
                    if let Some((resolved_url, relative_path)) = self.resolve_url(attr_value) {
                        let is_same_domain = self.is_same_domain(&resolved_url);

                        if is_same_domain {
                            let normalized_url = PathMapper::normalize_url(&resolved_url);
                            extracted_urls.push(normalized_url);

                            let search_pattern = format!("{}=\"{}\"", attr, attr_value);
                            let replace_pattern = format!("{}=\"{}\"", attr, relative_path);
                            rewritten_html = rewritten_html.replace(&search_pattern, &replace_pattern);

                            let search_pattern_single = format!("{}='{}'", attr, attr_value);
                            let replace_pattern_single = format!("{}='{}'", attr, relative_path);
                            rewritten_html = rewritten_html.replace(&search_pattern_single, &replace_pattern_single);
                        }
                    }
                }
            }
        }

        extracted_urls.sort();
        extracted_urls.dedup();

        RewriteResult {
            rewritten_html,
            extracted_urls,
        }
    }

    fn resolve_url(&self, url_str: &str) -> Option<(Url, String)> {
        if url_str.starts_with("mailto:")
            || url_str.starts_with("tel:")
            || url_str.starts_with("javascript:")
            || url_str.starts_with("#")
            || url_str.starts_with("data:")
        {
            return None;
        }

        let resolved_url = match self.base_url.join(url_str) {
            Ok(url) => url,
            Err(_) => return None,
        };

        let is_same_domain = self.is_same_domain(&resolved_url);
        if !is_same_domain {
            return None;
        }

        let normalized_url = PathMapper::normalize_url(&resolved_url);
        let (target_path, _) = self.path_mapper.url_to_local_path(&normalized_url);
        let (base_path, _) = self.path_mapper.url_to_local_path(&self.base_url);

        let relative_path = match PathMapper::relative_path(&base_path, &target_path) {
            Some(p) => p,
            None => return Some((resolved_url, url_str.to_string())),
        };

        let relative_str = relative_path.to_string_lossy().replace("\\", "/");

        Some((resolved_url, relative_str))
    }

    fn is_same_domain(&self, url: &Url) -> bool {
        self.base_url.host_str() == url.host_str()
            && self.base_url.scheme() == url.scheme()
            && self.base_url.port() == url.port()
    }

    pub fn extract_urls_from_html(html: &str, base_url: &Url) -> Vec<Url> {
        let document = Html::parse_document(html);
        let mut urls = Vec::new();

        let link_selectors = [
            ("a", "href"),
            ("link", "href"),
            ("script", "src"),
            ("img", "src"),
            ("audio", "src"),
            ("video", "src"),
            ("source", "src"),
            ("form", "action"),
            ("iframe", "src"),
        ];

        for (tag, attr) in link_selectors {
            let selector_str = format!("{}[{}]", tag, attr);
            let selector = match Selector::parse(&selector_str) {
                Ok(s) => s,
                Err(_) => continue,
            };

            for element in document.select(&selector) {
                if let Some(attr_value) = element.value().attr(attr) {
                    if attr_value.starts_with("mailto:")
                        || attr_value.starts_with("tel:")
                        || attr_value.starts_with("javascript:")
                        || attr_value.starts_with("#")
                        || attr_value.starts_with("data:")
                    {
                        continue;
                    }

                    if let Ok(resolved_url) = base_url.join(attr_value) {
                        urls.push(PathMapper::normalize_url(&resolved_url));
                    }
                }
            }
        }

        urls.sort();
        urls.dedup();
        urls
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    #[test]
    fn test_rewrite_same_domain_link() {
        let base_url = Url::parse("https://example.com/about/team.html").unwrap();
        let path_mapper = PathMapper::new(PathBuf::from("/output"));
        let rewriter = LinkRewriter::new(base_url, path_mapper);

        let html = r#"<a href="https://example.com/contact">Contact</a>"#;
        let result = rewriter.rewrite_html(html);

        assert!(result.rewritten_html.contains("../contact.html"));
        assert_eq!(result.extracted_urls.len(), 1);
    }

    #[test]
    fn test_rewrite_relative_link() {
        let base_url = Url::parse("https://example.com/about/team.html").unwrap();
        let path_mapper = PathMapper::new(PathBuf::from("/output"));
        let rewriter = LinkRewriter::new(base_url, path_mapper);

        let html = r#"<a href="/contact">Contact</a>"#;
        let result = rewriter.rewrite_html(html);

        assert!(result.rewritten_html.contains("../contact.html"));
    }

    #[test]
    fn test_external_link_unchanged() {
        let base_url = Url::parse("https://example.com/about/team.html").unwrap();
        let path_mapper = PathMapper::new(PathBuf::from("/output"));
        let rewriter = LinkRewriter::new(base_url, path_mapper);

        let html = r#"<a href="https://other.com/page">External</a>"#;
        let result = rewriter.rewrite_html(html);

        assert!(result.rewritten_html.contains("https://other.com/page"));
        assert_eq!(result.extracted_urls.len(), 0);
    }

    #[test]
    fn test_static_resources() {
        let base_url = Url::parse("https://example.com/page.html").unwrap();
        let path_mapper = PathMapper::new(PathBuf::from("/output"));
        let rewriter = LinkRewriter::new(base_url, path_mapper);

        let html = r#"
            <link rel="stylesheet" href="/css/style.css">
            <script src="/js/app.js"></script>
            <img src="/images/logo.png">
        "#;
        let result = rewriter.rewrite_html(html);

        assert!(result.rewritten_html.contains("css/style.css"));
        assert!(result.rewritten_html.contains("js/app.js"));
        assert!(result.rewritten_html.contains("images/logo.png"));
        assert_eq!(result.extracted_urls.len(), 3);
    }

    #[test]
    fn test_extract_urls() {
        let base_url = Url::parse("https://example.com/page.html").unwrap();
        let html = r#"
            <a href="/about">About</a>
            <a href="https://other.com/external">External</a>
            <img src="/logo.png">
        "#;
        let urls = LinkRewriter::extract_urls_from_html(html, &base_url);

        assert_eq!(urls.len(), 2);
    }
}
