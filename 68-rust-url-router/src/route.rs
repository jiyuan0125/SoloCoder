use super::RouterError;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Segment {
    Static(String),
    Param(String),
    Wildcard,
    DoubleWildcard,
}

pub struct PathParser;

impl PathParser {
    pub fn validate_and_parse(path: &str) -> Result<Vec<Segment>, RouterError> {
        if path.is_empty() {
            return Err(RouterError::InvalidPathFormat("Path cannot be empty".to_string()));
        }

        if !path.starts_with('/') {
            return Err(RouterError::InvalidPathFormat("Path must start with '/'".to_string()));
        }

        if path.contains("//") {
            return Err(RouterError::InvalidPathFormat("Path contains consecutive '/'".to_string()));
        }

        let segments: Vec<&str> = path.split('/').skip(1).filter(|s| !s.is_empty()).collect();

        let mut parsed_segments = Vec::new();
        let mut has_wildcard = false;
        let mut has_double_wildcard = false;

        for (i, segment) in segments.iter().enumerate() {
            if has_double_wildcard {
                return Err(RouterError::InvalidPathFormat("Segments after double wildcard '**' are not allowed".to_string()));
            }

            if has_wildcard {
                return Err(RouterError::InvalidPathFormat("Segments after wildcard '*' are not allowed".to_string()));
            }

            let parsed = match *segment {
                "*" => {
                    if i != segments.len() - 1 {
                        return Err(RouterError::WildcardNotAtEnd);
                    }
                    has_wildcard = true;
                    Segment::Wildcard
                }
                "**" => {
                    if i != segments.len() - 1 {
                        return Err(RouterError::DoubleWildcardNotAtEnd);
                    }
                    has_double_wildcard = true;
                    Segment::DoubleWildcard
                }
                s if s.starts_with(':') => {
                    if s.len() == 1 {
                        return Err(RouterError::InvalidPathFormat("Parameter name cannot be empty (':' followed by '/')".to_string()));
                    }
                    if s.contains('*') {
                        return Err(RouterError::InvalidPathFormat("Parameter name cannot contain '*'".to_string()));
                    }
                    if s.contains(':') && s.matches(':').count() > 1 {
                        return Err(RouterError::InvalidPathFormat("Parameter name cannot contain multiple ':'".to_string()));
                    }
                    Segment::Param(s[1..].to_string())
                }
                s => {
                    if s.contains('*') {
                        let star_count = s.matches('*').count();
                        if star_count == 1 && s == "*" {
                        } else if star_count == 2 && s == "**" {
                        } else {
                            return Err(RouterError::InvalidPathFormat("Wildcards '*' and '**' must be standalone segments".to_string()));
                        }
                    }
                    if s.contains(':') {
                        return Err(RouterError::InvalidPathFormat("Static segments cannot contain ':'".to_string()));
                    }
                    Segment::Static(s.to_string())
                }
            };
            parsed_segments.push(parsed);
        }

        if has_wildcard && has_double_wildcard {
            return Err(RouterError::ConflictingWildcards);
        }

        Ok(parsed_segments)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_paths() {
        assert!(PathParser::validate_and_parse("/").is_ok());
        assert!(PathParser::validate_and_parse("/users").is_ok());
        assert!(PathParser::validate_and_parse("/users/:id").is_ok());
        assert!(PathParser::validate_and_parse("/users/:id/posts/:post_id").is_ok());
        assert!(PathParser::validate_and_parse("/files/*").is_ok());
        assert!(PathParser::validate_and_parse("/static/**").is_ok());
    }

    #[test]
    fn test_invalid_paths() {
        assert!(PathParser::validate_and_parse("").is_err());
        assert!(PathParser::validate_and_parse("no-start-slash").is_err());
        assert!(PathParser::validate_and_parse("//double-slash").is_err());
        assert!(PathParser::validate_and_parse("/users/:").is_err());
        assert!(PathParser::validate_and_parse("/files/*/more").is_err());
        assert!(PathParser::validate_and_parse("/static/**/more").is_err());
        assert!(PathParser::validate_and_parse("/mixed/*/**").is_err());
    }
}
