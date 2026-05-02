use std::path::{Path, PathBuf};
use url::Url;

#[derive(Debug, Clone)]
pub struct PathMapper {
    output_dir: PathBuf,
}

impl PathMapper {
    pub fn new(output_dir: PathBuf) -> Self {
        Self { output_dir }
    }

    pub fn url_to_local_path(&self, url: &Url) -> (PathBuf, bool) {
        let path = url.path();
        let normalized = Self::normalize_path(path);

        let (file_path, is_html) = Self::path_to_file_path(&normalized);

        let path_to_join = if file_path.starts_with('/') {
            &file_path[1..]
        } else {
            &file_path
        };

        let local_path = self.output_dir.join(path_to_join);

        (local_path, is_html)
    }

    fn normalize_path(path: &str) -> String {
        if path.is_empty() {
            return "/".to_string();
        }

        let parts: Vec<&str> = path.split('/').collect();
        let mut result = Vec::new();

        for part in parts {
            match part {
                "" | "." => continue,
                ".." => {
                    result.pop();
                }
                _ => result.push(part),
            }
        }

        if path.starts_with('/') {
            format!("/{}", result.join("/"))
        } else {
            result.join("/")
        }
    }

    fn path_to_file_path(path: &str) -> (String, bool) {
        if path.ends_with('/') || path.is_empty() || path == "/" {
            let base = if path.is_empty() || path == "/" {
                "/index"
            } else {
                &path[..path.len() - 1]
            };
            (format!("{}.html", base), true)
        } else {
            let ext = Self::get_extension(path);
            match ext.as_deref() {
                Some("html") | Some("htm") => (path.to_string(), true),
                Some(_) => (path.to_string(), false),
                None => (format!("{}.html", path), true),
            }
        }
    }

    fn get_extension(path: &str) -> Option<String> {
        let last_segment = path.rsplit('/').next()?;
        if !last_segment.contains('.') {
            return None;
        }
        let ext = last_segment.rsplit('.').next()?;
        if ext.is_empty() || ext.contains('/') {
            return None;
        }
        Some(ext.to_lowercase())
    }

    pub fn relative_path(from: &Path, to: &Path) -> Option<PathBuf> {
        let from_dir = if from.is_file() {
            from.parent()?
        } else {
            from
        };

        let mut from_components: Vec<_> = from_dir.components().collect();
        let mut to_components: Vec<_> = to.components().collect();

        let mut common = 0;
        while common < from_components.len() && common < to_components.len() {
            if from_components[common] == to_components[common] {
                common += 1;
            } else {
                break;
            }
        }

        from_components.drain(..common);
        to_components.drain(..common);

        let mut result = PathBuf::new();
        for _ in &from_components {
            result.push("..");
        }
        for comp in to_components {
            result.push(comp);
        }

        if result.components().next().is_none() {
            Some(PathBuf::from("."))
        } else {
            Some(result)
        }
    }

    pub fn normalize_url(url: &Url) -> Url {
        let mut normalized = url.clone();
        normalized.set_query(None);
        normalized.set_fragment(None);
        normalized
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::Path;

    #[test]
    fn test_url_to_local_path_basic() {
        let mapper = PathMapper::new(PathBuf::from("/output"));
        let url = Url::parse("https://example.com/about/team").unwrap();
        let (path, is_html) = mapper.url_to_local_path(&url);
        assert_eq!(path, Path::new("/output/about/team.html"));
        assert!(is_html);
    }

    #[test]
    fn test_url_to_local_path_with_extension() {
        let mapper = PathMapper::new(PathBuf::from("/output"));
        let url = Url::parse("https://example.com/style.css").unwrap();
        let (path, is_html) = mapper.url_to_local_path(&url);
        assert_eq!(path, Path::new("/output/style.css"));
        assert!(!is_html);
    }

    #[test]
    fn test_url_to_local_path_with_query() {
        let mapper = PathMapper::new(PathBuf::from("/output"));
        let url = Url::parse("https://example.com/search?q=test").unwrap();
        let normalized = PathMapper::normalize_url(&url);
        let (path, is_html) = mapper.url_to_local_path(&normalized);
        assert_eq!(path, Path::new("/output/search.html"));
        assert!(is_html);
    }

    #[test]
    fn test_normalize_path_with_dot_dot() {
        assert_eq!(PathMapper::normalize_path("/a/../b"), "/b");
        assert_eq!(PathMapper::normalize_path("/a/b/../../c"), "/c");
        assert_eq!(PathMapper::normalize_path("/a/./b"), "/a/b");
    }

    #[test]
    fn test_path_to_file_path_trailing_slash() {
        assert_eq!(PathMapper::path_to_file_path("/about/"), ("/about.html".to_string(), true));
        assert_eq!(PathMapper::path_to_file_path("/"), ("/index.html".to_string(), true));
    }

    #[test]
    fn test_relative_path() {
        let from = Path::new("/output/about/team.html");
        let to = Path::new("/output/contact.html");
        let relative = PathMapper::relative_path(from, to);
        assert_eq!(relative, Some(PathBuf::from("../contact.html")));

        let from = Path::new("/output/index.html");
        let to = Path::new("/output/about/team.html");
        let relative = PathMapper::relative_path(from, to);
        assert_eq!(relative, Some(PathBuf::from("about/team.html")));
    }

    #[test]
    fn test_normalize_url() {
        let url = Url::parse("https://example.com/page?q=1#section").unwrap();
        let normalized = PathMapper::normalize_url(&url);
        assert_eq!(normalized.as_str(), "https://example.com/page");
    }
}
