use html5ever::driver::{parse_document, ParseOpts};
use html5ever::interface::Attribute;
use html5ever::serialize::{serialize, SerializeOpts};
use html5ever::tendril::TendrilSink;
use markup5ever_rcdom::{Handle, NodeData, RcDom, SerializableHandle};
use url::Url;
use std::path::Path;
use crate::path_mapper::PathMapper;

pub struct HtmlLinkRewriter {
    path_mapper: PathMapper,
    current_url: Url,
    current_local_path: std::path::PathBuf,
}

impl HtmlLinkRewriter {
    pub fn new(
        path_mapper: PathMapper,
        current_url: &Url,
        current_local_path: &Path,
    ) -> Self {
        HtmlLinkRewriter {
            path_mapper,
            current_url: current_url.clone(),
            current_local_path: current_local_path.to_path_buf(),
        }
    }

    pub fn rewrite(&self, html: &str) -> Result<String, anyhow::Error> {
        let dom = parse_document(RcDom::default(), ParseOpts::default())
            .from_utf8()
            .read_from(&mut html.as_bytes())?;
        
        self.rewrite_node(&dom.document);
        
        let mut output = Vec::new();
        let serializable = SerializableHandle::from(dom.document.clone());
        serialize(&mut output, &serializable, SerializeOpts::default())?;
        
        Ok(String::from_utf8_lossy(&output).to_string())
    }

    fn rewrite_node(&self, node: &Handle) {
        match node.data {
            NodeData::Element { ref name, ref attrs, .. } => {
                let local_name = name.local.as_ref();
                
                match local_name {
                    "a" => {
                        self.rewrite_attribute(node, attrs, "href", true);
                    }
                    "link" => {
                        self.rewrite_attribute(node, attrs, "href", false);
                    }
                    "script" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                    }
                    "img" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                    }
                    "video" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                        self.rewrite_attribute(node, attrs, "poster", false);
                    }
                    "audio" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                    }
                    "source" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                        self.rewrite_attribute(node, attrs, "srcset", false);
                    }
                    "track" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                    }
                    "iframe" => {
                        self.rewrite_attribute(node, attrs, "src", false);
                    }
                    "form" => {
                        self.rewrite_attribute(node, attrs, "action", true);
                    }
                    "area" => {
                        self.rewrite_attribute(node, attrs, "href", true);
                    }
                    "base" => {
                        self.rewrite_attribute(node, attrs, "href", true);
                    }
                    _ => {}
                }
            }
            _ => {}
        }
        
        for child in node.children.borrow().iter() {
            self.rewrite_node(child);
        }
    }

    fn rewrite_attribute(
        &self,
        _node: &Handle,
        attrs: &std::cell::RefCell<Vec<Attribute>>,
        attr_name: &str,
        _is_html: bool,
    ) {
        let mut attrs_borrow = attrs.borrow_mut();
        
        for attr in attrs_borrow.iter_mut() {
            if attr.name.local.as_ref() == attr_name {
                let original_url = attr.value.to_string();
                
                if let Ok(resolved_url) = self.current_url.join(&original_url) {
                    if self.path_mapper.is_same_domain(&resolved_url) {
                        let canonical_url = self.path_mapper.canonicalize_url(&resolved_url);
                        let target_local_path = self.path_mapper.url_to_local_path(&canonical_url);
                        
                        let relative_path = self.path_mapper.relative_path(
                            &self.current_local_path,
                            &target_local_path,
                        );
                        
                        let relative_str = relative_path.to_string_lossy().replace("\\", "/");
                        attr.value = tendril::StrTendril::from(relative_str);
                    }
                }
            }
        }
    }

    pub fn extract_links(html: &str, base_url: &Url) -> Vec<Url> {
        let dom = parse_document(RcDom::default(), ParseOpts::default())
            .from_utf8()
            .read_from(&mut html.as_bytes())
            .unwrap_or_else(|_| RcDom::default());
        
        let mut links = Vec::new();
        Self::extract_links_from_node(&dom.document, base_url, &mut links);
        links
    }

    fn extract_links_from_node(node: &Handle, base_url: &Url, links: &mut Vec<Url>) {
        match node.data {
            NodeData::Element { ref name, ref attrs, .. } => {
                let local_name = name.local.as_ref();
                
                let mut attr_names = Vec::new();
                
                match local_name {
                    "a" => attr_names.push("href"),
                    "link" => attr_names.push("href"),
                    "script" => attr_names.push("src"),
                    "img" => attr_names.push("src"),
                    "video" => {
                        attr_names.push("src");
                        attr_names.push("poster");
                    }
                    "audio" => attr_names.push("src"),
                    "source" => {
                        attr_names.push("src");
                        attr_names.push("srcset");
                    }
                    "track" => attr_names.push("src"),
                    "iframe" => attr_names.push("src"),
                    "form" => attr_names.push("action"),
                    "area" => attr_names.push("href"),
                    "base" => attr_names.push("href"),
                    _ => {}
                }
                
                let attrs_borrow = attrs.borrow();
                for attr_name in attr_names {
                    for attr in attrs_borrow.iter() {
                        if attr.name.local.as_ref() == attr_name {
                            let attr_value = attr.value.to_string();
                            
                            if let Ok(url) = base_url.join(&attr_value) {
                                links.push(url);
                            }
                        }
                    }
                }
            }
            _ => {}
        }
        
        for child in node.children.borrow().iter() {
            Self::extract_links_from_node(child, base_url, links);
        }
    }

    pub fn is_html_content(content_type: &str) -> bool {
        let lower = content_type.to_lowercase();
        lower.contains("text/html") || lower.contains("application/xhtml+xml")
    }

    pub fn is_static_resource(content_type: &str) -> bool {
        let lower = content_type.to_lowercase();
        lower.contains("text/css") ||
        lower.contains("application/javascript") ||
        lower.contains("text/javascript") ||
        lower.contains("image/") ||
        lower.contains("font/") ||
        lower.contains("application/font")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::Path;

    #[test]
    fn test_rewrite_internal_links() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let path_mapper = PathMapper::new(&base_url, output_dir);
        
        let current_url = Url::parse("https://example.com/about/team").unwrap();
        let current_local_path = path_mapper.url_to_local_path(&current_url);
        
        let rewriter = HtmlLinkRewriter::new(path_mapper, &current_url, &current_local_path);
        
        let html = r#"<html><head></head><body>
            <a href="https://example.com/contact">Contact</a>
            <a href="https://other.com/external">External</a>
            <img src="/images/logo.png" />
            <link rel="stylesheet" href="/css/style.css" />
            <script src="/js/app.js"></script>
        </body></html>"#;
        
        let result = rewriter.rewrite(html).unwrap();
        
        assert!(result.contains("contact.html"));
        assert!(result.contains("external"));
        assert!(result.contains("logo.png"));
        assert!(result.contains("style.css"));
        assert!(result.contains("app.js"));
    }

    #[test]
    fn test_extract_links() {
        let base_url = Url::parse("https://example.com/").unwrap();
        
        let html = r#"<html><head>
            <link rel="stylesheet" href="/css/style.css" />
        </head><body>
            <a href="/about">About</a>
            <a href="https://other.com/external">External</a>
            <img src="/images/logo.png" />
        </body></html>"#;
        
        let links = HtmlLinkRewriter::extract_links(html, &base_url);
        
        assert!(links.len() >= 3);
        assert!(links.iter().any(|u| u.path() == "/about"));
        assert!(links.iter().any(|u| u.path() == "/css/style.css"));
        assert!(links.iter().any(|u| u.path() == "/images/logo.png"));
    }

    #[test]
    fn test_is_html_content() {
        assert!(HtmlLinkRewriter::is_html_content("text/html"));
        assert!(HtmlLinkRewriter::is_html_content("text/html; charset=utf-8"));
        assert!(HtmlLinkRewriter::is_html_content("application/xhtml+xml"));
        assert!(!HtmlLinkRewriter::is_html_content("text/css"));
        assert!(!HtmlLinkRewriter::is_html_content("image/png"));
    }

    #[test]
    fn test_is_static_resource() {
        assert!(HtmlLinkRewriter::is_static_resource("text/css"));
        assert!(HtmlLinkRewriter::is_static_resource("application/javascript"));
        assert!(HtmlLinkRewriter::is_static_resource("image/png"));
        assert!(HtmlLinkRewriter::is_static_resource("image/jpeg"));
        assert!(HtmlLinkRewriter::is_static_resource("font/woff2"));
        assert!(!HtmlLinkRewriter::is_static_resource("text/html"));
    }

    #[test]
    fn test_relative_path_rewrite() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let path_mapper = PathMapper::new(&base_url, output_dir);
        
        let current_url = Url::parse("https://example.com/docs/api/v1/guide").unwrap();
        let current_local_path = path_mapper.url_to_local_path(&current_url);
        
        let rewriter = HtmlLinkRewriter::new(path_mapper, &current_url, &current_local_path);
        
        let html = r#"<html><body>
            <a href="/home">Home</a>
            <a href="../index">Parent</a>
            <img src="/images/logo.png" />
        </body></html>"#;
        
        let result = rewriter.rewrite(html).unwrap();
        
        assert!(result.contains("../../../home.html"));
        assert!(result.contains("../../../images/logo.png"));
    }
}
