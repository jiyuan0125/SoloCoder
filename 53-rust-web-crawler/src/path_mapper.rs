use url::Url;
use std::path::{Path, PathBuf, Component};

#[derive(Clone)]
pub struct PathMapper {
    base_url: Url,
    output_dir: PathBuf,
}

impl PathMapper {
    pub fn new(base_url: &Url, output_dir: &Path) -> Self {
        PathMapper {
            base_url: base_url.clone(),
            output_dir: output_dir.to_path_buf(),
        }
    }

    pub fn url_to_local_path(&self, url: &Url) -> PathBuf {
        let canonical_url = self.canonicalize_url(url);
        
        let path_str = canonical_url.path();
        let path = Path::new(path_str);
        
        let mut local_path = self.output_dir.clone();
        
        for component in path.components() {
            match component {
                Component::RootDir => continue,
                Component::Normal(seg) => local_path.push(seg),
                _ => continue,
            }
        }
        
        if Self::needs_html_extension(&local_path) {
            let file_name = local_path.file_name()
                .map(|n| n.to_string_lossy().to_string())
                .unwrap_or_else(|| "index".to_string());
            
            local_path.pop();
            local_path.push(format!("{}.html", file_name));
        }
        
        local_path
    }

    pub fn canonicalize_url(&self, url: &Url) -> Url {
        let mut result = url.clone();
        
        result.set_query(None);
        result.set_fragment(None);
        
        let path = result.path();
        let path_buf = PathBuf::from(path);
        let mut normalized_components = Vec::new();
        
        for component in path_buf.components() {
            match component {
                Component::ParentDir => {
                    if !normalized_components.is_empty() {
                        normalized_components.pop();
                    }
                }
                Component::CurDir => continue,
                _ => normalized_components.push(component),
            }
        }
        
        let normalized_path: PathBuf = normalized_components.into_iter().collect();
        let normalized_path_str = normalized_path.to_string_lossy().replace("\\", "/");
        
        result.set_path(&normalized_path_str);
        result
    }

    pub fn relative_path(&self, from: &Path, to: &Path) -> PathBuf {
        let from_dir = if from.extension().is_some() {
            from.parent().unwrap_or(Path::new(""))
        } else {
            from
        };
        
        let mut relative = PathBuf::new();
        let mut from_components = from_dir.components().skip_while(|c| *c == Component::RootDir);
        let mut to_components = to.components().skip_while(|c| *c == Component::RootDir);
        
        while let (Some(f), Some(t)) = (from_components.next(), to_components.next()) {
            if f != t {
                relative.push("..");
                let mut temp = PathBuf::new();
                temp.push(f);
                for fc in from_components.by_ref() {
                    temp.push(fc);
                    relative.push("..");
                }
                relative.push(t);
                for tc in to_components.by_ref() {
                    relative.push(tc);
                }
                return relative;
            }
        }
        
        for _ in from_components {
            relative.push("..");
        }
        
        for tc in to_components {
            relative.push(tc);
        }
        
        relative
    }

    fn needs_html_extension(path: &Path) -> bool {
        let extension = path.extension()
            .map(|e| e.to_string_lossy().to_lowercase());
        
        match extension {
            Some(ext) if Self::is_resource_extension(&ext) => false,
            Some(_) => false,
            None => true,
        }
    }

    fn is_resource_extension(ext: &str) -> bool {
        matches!(ext.as_ref(),
            "html" | "htm" | "css" | "js" | "jpg" | "jpeg" | "png" | "gif" |
            "svg" | "ico" | "woff" | "woff2" | "ttf" | "eot" | "pdf" | "txt"
        )
    }

    pub fn is_same_domain(&self, url: &Url) -> bool {
        self.base_url.domain() == url.domain() &&
        self.base_url.port_or_known_default() == url.port_or_known_default()
    }

    pub fn get_base_url(&self) -> &Url {
        &self.base_url
    }

    pub fn get_output_dir(&self) -> &PathBuf {
        &self.output_dir
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::Path;

    #[test]
    fn test_url_to_local_path_without_extension() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let url = Url::parse("https://example.com/about/team").unwrap();
        let result = mapper.url_to_local_path(&url);
        
        assert_eq!(result, PathBuf::from("/tmp/output/about/team.html"));
    }

    #[test]
    fn test_url_to_local_path_with_extension() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let url = Url::parse("https://example.com/css/style.css").unwrap();
        let result = mapper.url_to_local_path(&url);
        
        assert_eq!(result, PathBuf::from("/tmp/output/css/style.css"));
    }

    #[test]
    fn test_url_with_query_string() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let url = Url::parse("https://example.com/search?q=test&page=1").unwrap();
        let result = mapper.url_to_local_path(&url);
        
        assert_eq!(result, PathBuf::from("/tmp/output/search.html"));
    }

    #[test]
    fn test_url_with_fragment() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let url = Url::parse("https://example.com/docs#section").unwrap();
        let result = mapper.url_to_local_path(&url);
        
        assert_eq!(result, PathBuf::from("/tmp/output/docs.html"));
    }

    #[test]
    fn test_path_normalization() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let url = Url::parse("https://example.com/a/../b/c/./d").unwrap();
        let result = mapper.url_to_local_path(&url);
        
        assert_eq!(result, PathBuf::from("/tmp/output/b/c/d.html"));
    }

    #[test]
    fn test_relative_path() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let from = Path::new("/tmp/output/about/team.html");
        let to = Path::new("/tmp/output/contact.html");
        
        let result = mapper.relative_path(from, to);
        
        assert_eq!(result, PathBuf::from("../contact.html"));
    }

    #[test]
    fn test_is_same_domain() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/output");
        let mapper = PathMapper::new(&base_url, output_dir);
        
        let same_domain_url = Url::parse("https://example.com/about").unwrap();
        let different_domain_url = Url::parse("https://other.com/about").unwrap();
        
        assert!(mapper.is_same_domain(&same_domain_url));
        assert!(!mapper.is_same_domain(&different_domain_url));
    }
}
