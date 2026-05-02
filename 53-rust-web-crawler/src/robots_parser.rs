use url::Url;

#[derive(Debug, Clone)]
pub struct RobotsTxt {
    disallow_patterns: Vec<String>,
    allow_patterns: Vec<String>,
}

impl RobotsTxt {
    pub fn new() -> Self {
        Self {
            disallow_patterns: Vec::new(),
            allow_patterns: Vec::new(),
        }
    }

    pub fn parse(content: &str) -> Self {
        let mut disallow_patterns = Vec::new();
        let mut allow_patterns = Vec::new();

        let mut in_our_agent = false;

        for line in content.lines() {
            let line = line.trim();
            if line.is_empty() || line.starts_with('#') {
                continue;
            }

            let lower_line = line.to_lowercase();
            if lower_line.starts_with("user-agent:") {
                let agent = line["user-agent:".len()..].trim();
                in_our_agent = agent == "*" || agent.to_lowercase().contains("webmirror");
                continue;
            }

            if !in_our_agent {
                continue;
            }

            if lower_line.starts_with("disallow:") {
                let pattern = line["disallow:".len()..].trim();
                if !pattern.is_empty() {
                    disallow_patterns.push(pattern.to_string());
                }
            } else if lower_line.starts_with("allow:") {
                let pattern = line["allow:".len()..].trim();
                if !pattern.is_empty() {
                    allow_patterns.push(pattern.to_string());
                }
            }
        }

        Self {
            disallow_patterns,
            allow_patterns,
        }
    }

    pub fn is_allowed(&self, url: &Url) -> bool {
        let path = url.path();

        for allow in &self.allow_patterns {
            if self.matches_pattern(path, allow) {
                return true;
            }
        }

        for disallow in &self.disallow_patterns {
            if self.matches_pattern(path, disallow) {
                return false;
            }
        }

        true
    }

    fn matches_pattern(&self, path: &str, pattern: &str) -> bool {
        if pattern == "/" {
            return true;
        }

        let pattern = if pattern.ends_with('$') {
            &pattern[..pattern.len() - 1]
        } else {
            pattern
        };

        if pattern.contains('*') {
            self.wildcard_match(path, pattern)
        } else {
            path.starts_with(pattern)
        }
    }

    fn wildcard_match(&self, path: &str, pattern: &str) -> bool {
        let pattern_parts: Vec<&str> = pattern.split('*').collect();
        let mut current_pos = 0;

        for (i, part) in pattern_parts.iter().enumerate() {
            if part.is_empty() {
                continue;
            }

            let rest = &path[current_pos..];
            match rest.find(part) {
                Some(pos) => {
                    if i == 0 && pos != 0 {
                        return false;
                    }
                    current_pos += pos + part.len();
                }
                None => return false,
            }
        }

        true
    }
}

impl Default for RobotsTxt {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_disallow() {
        let content = r#"
User-agent: *
Disallow: /private/
Disallow: /admin
Allow: /admin/public
"#;
        let robots = RobotsTxt::parse(content);
        assert_eq!(robots.disallow_patterns, vec!["/private/", "/admin"]);
        assert_eq!(robots.allow_patterns, vec!["/admin/public"]);
    }

    #[test]
    fn test_is_allowed() {
        let content = r#"
User-agent: *
Disallow: /private/
Disallow: /admin
Allow: /admin/public
"#;
        let robots = RobotsTxt::parse(content);

        assert!(robots.is_allowed(&Url::parse("https://example.com/").unwrap()));
        assert!(robots.is_allowed(&Url::parse("https://example.com/page.html").unwrap()));
        assert!(!robots.is_allowed(&Url::parse("https://example.com/private/secret").unwrap()));
        assert!(!robots.is_allowed(&Url::parse("https://example.com/admin").unwrap()));
        assert!(!robots.is_allowed(&Url::parse("https://example.com/admin/secret").unwrap()));
        assert!(robots.is_allowed(&Url::parse("https://example.com/admin/public").unwrap()));
    }

    #[test]
    fn test_wildcard_match() {
        let robots = RobotsTxt::new();
        assert!(robots.wildcard_match("/images/test.jpg", "/*.jpg$"));
        assert!(robots.wildcard_match("/files/test.txt", "/*"));
    }
}
