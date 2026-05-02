mod cli;
mod gitignore;
mod mmap_reader;
mod parallel;
mod regex_matcher;

use std::sync::Arc;

use cli::Config;
use gitignore::{find_gitignore_files, GitignoreMatcher};
use parallel::{collect_files, search_file_parallel};
use regex_matcher::Matcher;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    
    let config = match Config::from_args(&args) {
        Ok(c) => c,
        Err(e) => {
            eprintln!("错误: {}", e);
            std::process::exit(1);
        }
    };
    
    let matcher = match Matcher::new(&config.pattern, config.is_regex, config.ignore_case) {
        Ok(m) => Arc::new(m),
        Err(e) => {
            eprintln!("正则表达式错误: {}", e);
            std::process::exit(1);
        }
    };
    
    let mut gitignore_matchers: Vec<GitignoreMatcher> = Vec::new();
    for path in &config.paths {
        let gitignore_files = find_gitignore_files(path);
        for gitignore_path in gitignore_files {
            if let Some(parent) = gitignore_path.parent() {
                let mut matcher = GitignoreMatcher::new(parent);
                let _ = matcher.load_from_file(&gitignore_path);
                gitignore_matchers.push(matcher);
            }
        }
    }
    
    let files = collect_files(&config.paths, &config.exclude_patterns, &gitignore_matchers);
    
    let config = Arc::new(config);
    
    let mut matched_files_count = 0;
    let mut total_matches = 0;
    
    for file_path in files {
        let result = search_file_parallel(
            &file_path,
            Arc::clone(&matcher),
            Arc::clone(&config),
        );
        
        if let Some(file_match) = result {
            if config.files_with_matches {
                println!("{}", file_match.path.display());
                matched_files_count += 1;
                total_matches += file_match.total_matches;
                continue;
            }
            
            if config.count {
                println!("{}:{}", file_match.path.display(), file_match.total_matches);
                if file_match.total_matches > 0 {
                    matched_files_count += 1;
                }
                total_matches += file_match.total_matches;
                continue;
            }
            
            if file_match.total_matches > 0 {
                matched_files_count += 1;
                total_matches += file_match.total_matches;
                
                for line_match in &file_match.matches {
                    if config.line_number {
                        println!(
                            "{}:{}:{}",
                            file_match.path.display(),
                            line_match.line_number,
                            line_match.content
                        );
                    } else {
                        println!("{}:{}", file_match.path.display(), line_match.content);
                    }
                }
            }
        }
    }
    
    println!("{} 个文件，{} 个匹配", matched_files_count, total_matches);
}
