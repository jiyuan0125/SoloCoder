use regex::Regex;

#[derive(Debug, Clone)]
pub struct MatchResult {
    pub line_number: usize,
    pub content: String,
}

pub struct Matcher {
    regex: Regex,
}

impl Matcher {
    pub fn new(pattern: &str, is_regex: bool, ignore_case: bool) -> Result<Self, regex::Error> {
        let regex_pattern = if is_regex {
            pattern.to_string()
        } else {
            regex::escape(pattern)
        };
        
        let regex = if ignore_case {
            Regex::new(&format!("(?i){}", regex_pattern))
        } else {
            Regex::new(&regex_pattern)
        }?;
        
        Ok(Matcher { regex })
    }

    pub fn find_matches_in_buffer(
        &self,
        buffer: &[u8],
        start_offset: usize,
    ) -> Vec<MatchResult> {
        let mut results = Vec::new();
        let mut line_number = 1;
        let mut _current_offset = 0;
        
        let text = String::from_utf8_lossy(buffer);
        
        for line in text.lines() {
            if self.regex.is_match(line) {
                let _byte_offset = start_offset + _current_offset;
                results.push(MatchResult {
                    line_number,
                    content: line.to_string(),
                });
            }
            _current_offset += line.len() + 1;
            line_number += 1;
        }
        
        results
    }
}
