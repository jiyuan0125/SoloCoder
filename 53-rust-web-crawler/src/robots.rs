use std::collections::HashMap;
use url::Url;

#[derive(Debug, Clone, Default)]
pub struct RobotsTxt {
    disallow_paths: HashMap<String, Vec<String>>,
    allow_paths: HashMap<String, Vec<String>>,
}

impl RobotsTxt {
    pub fn new() -> Self {
        Self {
            disallow_paths: HashMap::new(),
            allow_paths: HashMap::new(),
        }
    }

    pub fn parse(content: &str) -> Self {
        let mut robots = RobotsTxt::new();
        let mut current_user_agent = String::new();

        for line in content.lines() {
            let line = line.trim();
            if line.is_empty() || line.starts_with('#') {
                continue;
            }

            let lower_line = line.to_lowercase();
            if lower_line.starts_with("user-agent:") {
                let ua = line["user-agent:".len()..].trim().to_string();
                current_user_agent = ua;
            } else if lower_line.starts_with("disallow:") {
                let path = line["disallow:".len()..].trim().to_string();
                if !path.is_empty() {
                    robots.disallow_paths
                        .entry(current_user_agent.clone())
                        .or_default()
                        .push(path);
                }
            } else if lower_line.starts_with("allow:") {
                let path = line["allow:".len()..].trim().to_string();
                if !path.is_empty() {
                    robots.allow_paths
                        .entry(current_user_agent.clone())
                        .or_default()
                        .push(path);
                }
            }
        }

        robots
    }

    pub fn is_allowed(&self, url: &Url, user_agent: &str) -> bool {
        let path = url.path().to_string();
        
        let uas_to_check = [user_agent, "*"];
        
        for ua in uas_to_check.iter() {
            if let Some(disallow) = self.disallow_paths.get(*ua) {
                for disallow_path in disallow {
                    if path.starts_with(disallow_path) {
                        if let Some(allow) = self.allow_paths.get(*ua) {
                            for allow_path in allow {
                                if path.starts_with(allow_path) && allow_path.len() > disallow_path.len() {
                                    return true;
                                }
                            }
                        }
                        return false;
                    }
                }
            }
        }
        
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_robots_txt() {
        let content = r#"
User-agent: *
Disallow: /admin
Disallow: /private

User-agent: Googlebot
Allow: /public
Disallow: /
"#;
        let robots = RobotsTxt::parse(content);
        
        assert_eq!(robots.disallow_paths.get("*").unwrap().len(), 2);
        assert!(robots.disallow_paths.get("*").unwrap().contains(&"/admin".to_string()));
        assert!(robots.disallow_paths.get("*").unwrap().contains(&"/private".to_string()));
    }

    #[test]
    fn test_is_allowed() {
        let content = r#"
User-agent: *
Disallow: /admin
Disallow: /private
"#;
        let robots = RobotsTxt::parse(content);
        
        let url1 = Url::parse("http://example.com/page").unwrap();
        assert!(robots.is_allowed(&url1, "my-crawler"));
        
        let url2 = Url::parse("http://example.com/admin/login").unwrap();
        assert!(!robots.is_allowed(&url2, "my-crawler"));
        
        let url3 = Url::parse("http://example.com/private/secret").unwrap();
        assert!(!robots.is_allowed(&url3, "my-crawler"));
    }
}
