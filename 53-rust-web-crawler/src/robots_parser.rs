use url::Url;
use std::collections::HashSet;
use reqwest::blocking::Client;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum RobotsError {
    #[error("HTTP request failed: {0}")]
    HttpError(#[from] reqwest::Error),
    
    #[error("Invalid URL: {0}")]
    UrlError(#[from] url::ParseError),
    
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}

pub struct RobotsTxt {
    base_url: Url,
    disallow_paths: HashSet<String>,
    allow_paths: HashSet<String>,
}

impl RobotsTxt {
    pub fn new(base_url: &Url) -> Self {
        RobotsTxt {
            base_url: base_url.clone(),
            disallow_paths: HashSet::new(),
            allow_paths: HashSet::new(),
        }
    }

    pub fn fetch(&mut self, client: &Client) -> Result<(), RobotsError> {
        let robots_url = self.get_robots_url()?;
        
        let response = match client.get(robots_url.as_str()).send() {
            Ok(resp) => resp,
            Err(_) => return Ok(()),
        };
        
        if !response.status().is_success() {
            return Ok(());
        }
        
        let content = response.text()?;
        self.parse(&content);
        
        Ok(())
    }

    fn get_robots_url(&self) -> Result<Url, url::ParseError> {
        let mut robots_url = self.base_url.clone();
        robots_url.set_path("/robots.txt");
        robots_url.set_query(None);
        robots_url.set_fragment(None);
        Ok(robots_url)
    }

    fn parse(&mut self, content: &str) {
        let mut in_user_agent_section = false;
        
        for line in content.lines() {
            let trimmed = line.trim();
            
            if trimmed.is_empty() || trimmed.starts_with('#') {
                continue;
            }
            
            let lower = trimmed.to_lowercase();
            
            if lower.starts_with("user-agent:") {
                let user_agent = trimmed["user-agent:".len()..].trim();
                in_user_agent_section = user_agent == "*" || 
                    user_agent.to_lowercase().contains("web_crawler");
                continue;
            }
            
            if !in_user_agent_section {
                continue;
            }
            
            if lower.starts_with("disallow:") {
                let path = trimmed["disallow:".len()..].trim();
                if !path.is_empty() {
                    self.disallow_paths.insert(path.to_string());
                }
                continue;
            }
            
            if lower.starts_with("allow:") {
                let path = trimmed["allow:".len()..].trim();
                if !path.is_empty() {
                    self.allow_paths.insert(path.to_string());
                }
            }
        }
    }

    pub fn is_allowed(&self, url: &Url) -> bool {
        if self.disallow_paths.is_empty() && self.allow_paths.is_empty() {
            return true;
        }
        
        let path = url.path();
        
        for allow_path in &self.allow_paths {
            if path.starts_with(allow_path) {
                return true;
            }
        }
        
        for disallow_path in &self.disallow_paths {
            if path.starts_with(disallow_path) {
                return false;
            }
        }
        
        true
    }

    pub fn get_base_url(&self) -> &Url {
        &self.base_url
    }

    pub fn get_disallow_paths(&self) -> &HashSet<String> {
        &self.disallow_paths
    }

    pub fn get_allow_paths(&self) -> &HashSet<String> {
        &self.allow_paths
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_robots_txt() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let mut robots = RobotsTxt::new(&base_url);
        
        let content = r#"
User-agent: *
Disallow: /admin
Disallow: /private
Allow: /public
"#;
        
        robots.parse(content);
        
        assert!(robots.disallow_paths.contains("/admin"));
        assert!(robots.disallow_paths.contains("/private"));
        assert!(robots.allow_paths.contains("/public"));
    }

    #[test]
    fn test_is_allowed() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let mut robots = RobotsTxt::new(&base_url);
        
        let content = r#"
User-agent: *
Disallow: /admin
Disallow: /private
Allow: /public
"#;
        
        robots.parse(content);
        
        let public_url = Url::parse("https://example.com/public/page").unwrap();
        let admin_url = Url::parse("https://example.com/admin/dashboard").unwrap();
        let allowed_url = Url::parse("https://example.com/about").unwrap();
        
        assert!(robots.is_allowed(&public_url));
        assert!(!robots.is_allowed(&admin_url));
        assert!(robots.is_allowed(&allowed_url));
    }

    #[test]
    fn test_empty_robots_allows_everything() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let robots = RobotsTxt::new(&base_url);
        
        let url = Url::parse("https://example.com/any/path").unwrap();
        assert!(robots.is_allowed(&url));
    }

    #[test]
    fn test_allow_overrides_disallow() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let mut robots = RobotsTxt::new(&base_url);
        
        let content = r#"
User-agent: *
Disallow: /api
Allow: /api/public
"#;
        
        robots.parse(content);
        
        let public_api_url = Url::parse("https://example.com/api/public/data").unwrap();
        let private_api_url = Url::parse("https://example.com/api/private").unwrap();
        
        assert!(robots.is_allowed(&public_api_url));
        assert!(!robots.is_allowed(&private_api_url));
    }

    #[test]
    fn test_specific_user_agent() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let mut robots = RobotsTxt::new(&base_url);
        
        let content = r#"
User-agent: Googlebot
Disallow: /

User-agent: *
Disallow: /admin
"#;
        
        robots.parse(content);
        
        let admin_url = Url::parse("https://example.com/admin").unwrap();
        let home_url = Url::parse("https://example.com/").unwrap();
        
        assert!(!robots.is_allowed(&admin_url));
        assert!(robots.is_allowed(&home_url));
    }
}
