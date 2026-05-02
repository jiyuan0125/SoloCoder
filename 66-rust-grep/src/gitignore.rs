use std::path::{Path, PathBuf};
use std::fs::File;
use std::io::{BufRead, BufReader};

pub struct GitignoreMatcher {
    patterns: Vec<(String, bool)>,
    base_path: PathBuf,
}

impl GitignoreMatcher {
    pub fn new(base_path: &Path) -> Self {
        GitignoreMatcher {
            patterns: Vec::new(),
            base_path: base_path.to_path_buf(),
        }
    }

    pub fn load_from_file(&mut self, path: &Path) -> std::io::Result<()> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        
        for line in reader.lines() {
            let line = line?;
            let trimmed = line.trim();
            
            if trimmed.is_empty() || trimmed.starts_with('#') {
                continue;
            }
            
            let is_negation = trimmed.starts_with('!');
            let pattern = if is_negation {
                trimmed[1..].to_string()
            } else {
                trimmed.to_string()
            };
            
            self.patterns.push((pattern, is_negation));
        }
        
        Ok(())
    }

    pub fn matches(&self, path: &Path) -> bool {
        let relative_path = match path.strip_prefix(&self.base_path) {
            Ok(p) => p,
            Err(_) => return false,
        };
        
        let path_str = relative_path.to_string_lossy();
        let mut result = false;
        
        for (pattern, is_negation) in &self.patterns {
            if self.pattern_matches(pattern, &path_str) {
                result = !is_negation;
            }
        }
        
        result
    }

    fn pattern_matches(&self, pattern: &str, path_str: &str) -> bool {
        let normalized_pattern = pattern.replace("\\\\", "/");
        let normalized_path = path_str.replace("\\\\", "/");
        
        let pattern_components: Vec<&str> = normalized_pattern.split('/').collect();
        let path_components: Vec<&str> = normalized_path.split('/').collect();
        
        if pattern_components.contains(&"**") {
            self.globstar_match(&pattern_components, &path_components)
        } else {
            self.simple_match(&pattern_components, &path_components)
        }
    }

    fn simple_match(&self, pattern_components: &[&str], path_components: &[&str]) -> bool {
        if pattern_components.len() != path_components.len() {
            return false;
        }
        
        for (p, pc) in pattern_components.iter().zip(path_components.iter()) {
            if !self.fnmatch(p, pc) {
                return false;
            }
        }
        
        true
    }

    fn globstar_match(&self, pattern_components: &[&str], path_components: &[&str]) -> bool {
        let mut pi = 0;
        let mut pj = 0;
        
        while pi < pattern_components.len() && pj < path_components.len() {
            match pattern_components[pi] {
                "**" => {
                    pi += 1;
                    if pi >= pattern_components.len() {
                        return true;
                    }
                    
                    let target = pattern_components[pi];
                    while pj < path_components.len() {
                        if self.fnmatch(target, path_components[pj]) {
                            pi += 1;
                            pj += 1;
                            break;
                        }
                        pj += 1;
                    }
                }
                p => {
                    if !self.fnmatch(p, path_components[pj]) {
                        return false;
                    }
                    pi += 1;
                    pj += 1;
                }
            }
        }
        
        pi == pattern_components.len() && pj == path_components.len()
    }

    fn fnmatch(&self, pattern: &str, s: &str) -> bool {
        let mut p_chars = pattern.chars().peekable();
        let mut s_chars = s.chars().peekable();
        
        while let (Some(p), Some(c)) = (p_chars.peek(), s_chars.peek()) {
            match p {
                '*' => {
                    p_chars.next();
                    if p_chars.peek().is_none() {
                        return true;
                    }
                    
                    let next_p = *p_chars.peek().unwrap();
                    while let Some(sc) = s_chars.peek() {
                        if *sc == next_p || next_p == '?' {
                            break;
                        }
                        s_chars.next();
                    }
                }
                '?' => {
                    p_chars.next();
                    s_chars.next();
                }
                _ => {
                    if *p != *c {
                        return false;
                    }
                    p_chars.next();
                    s_chars.next();
                }
            }
        }
        
        while let Some('*') = p_chars.peek() {
            p_chars.next();
        }
        
        p_chars.peek().is_none() && s_chars.peek().is_none()
    }
}

pub fn find_gitignore_files(start_path: &Path) -> Vec<PathBuf> {
    let mut gitignore_files = Vec::new();
    let mut current = start_path.to_path_buf();
    
    loop {
        let gitignore = current.join(".gitignore");
        if gitignore.exists() {
            gitignore_files.push(gitignore);
        }
        
        if !current.pop() {
            break;
        }
    }
    
    gitignore_files.reverse();
    gitignore_files
}

pub fn should_exclude_by_exclude_pattern(
    path: &Path,
    exclude_patterns: &[String],
) -> bool {
    let file_name = match path.file_name() {
        Some(name) => name.to_string_lossy(),
        None => return false,
    };
    
    for pattern in exclude_patterns {
        if simple_glob_match(pattern, &file_name) {
            return true;
        }
    }
    
    false
}

fn simple_glob_match(pattern: &str, s: &str) -> bool {
    let mut p_chars = pattern.chars().peekable();
    let mut s_chars = s.chars().peekable();
    
    while let (Some(p), Some(c)) = (p_chars.peek(), s_chars.peek()) {
        match p {
            '*' => {
                p_chars.next();
                if p_chars.peek().is_none() {
                    return true;
                }
                
                let next_p = *p_chars.peek().unwrap();
                while let Some(sc) = s_chars.peek() {
                    if *sc == next_p || next_p == '?' {
                        break;
                    }
                    s_chars.next();
                }
            }
            '?' => {
                p_chars.next();
                s_chars.next();
            }
            _ => {
                if *p != *c {
                    return false;
                }
                p_chars.next();
                s_chars.next();
            }
        }
    }
    
    while let Some('*') = p_chars.peek() {
        p_chars.next();
    }
    
    p_chars.peek().is_none() && s_chars.peek().is_none()
}
