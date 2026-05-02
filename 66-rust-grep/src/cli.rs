use std::path::PathBuf;

#[derive(Debug, Clone)]
pub struct Config {
    pub pattern: String,
    pub is_regex: bool,
    pub paths: Vec<PathBuf>,
    pub ignore_case: bool,
    pub files_with_matches: bool,
    pub count: bool,
    pub line_number: bool,
    pub exclude_patterns: Vec<String>,
}

impl Config {
    pub fn new() -> Self {
        Config {
            pattern: String::new(),
            is_regex: false,
            paths: Vec::new(),
            ignore_case: false,
            files_with_matches: false,
            count: false,
            line_number: true,
            exclude_patterns: Vec::new(),
        }
    }

    pub fn from_args(args: &[String]) -> Result<Self, String> {
        let mut config = Config::new();
        let mut pattern_set = false;
        let mut i = 1;

        while i < args.len() {
            match args[i].as_str() {
                "-e" => {
                    if i + 1 >= args.len() {
                        return Err("-e 需要一个正则表达式参数".to_string());
                    }
                    config.pattern = args[i + 1].clone();
                    config.is_regex = true;
                    pattern_set = true;
                    i += 2;
                }
                "-i" => {
                    config.ignore_case = true;
                    i += 1;
                }
                "-l" => {
                    config.files_with_matches = true;
                    i += 1;
                }
                "-c" => {
                    config.count = true;
                    i += 1;
                }
                "-n" => {
                    config.line_number = true;
                    i += 1;
                }
                "--exclude" => {
                    if i + 1 >= args.len() {
                        return Err("--exclude 需要一个模式参数".to_string());
                    }
                    config.exclude_patterns.push(args[i + 1].clone());
                    i += 2;
                }
                arg => {
                    if !pattern_set {
                        if arg.starts_with('-') {
                            return Err(format!("未知选项: {}", arg));
                        }
                        config.pattern = arg.to_string();
                        pattern_set = true;
                    } else {
                        config.paths.push(PathBuf::from(arg));
                    }
                    i += 1;
                }
            }
        }

        if !pattern_set {
            return Err("缺少搜索模式".to_string());
        }

        if config.paths.is_empty() {
            config.paths.push(PathBuf::from("."));
        }

        Ok(config)
    }
}
