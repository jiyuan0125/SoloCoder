pub mod error;
pub mod event;
pub mod lexer;
pub mod parser;

pub use error::{Error, Result};
pub use event::{Event, Number, Position};
pub use parser::{parse, parse_with_config, Config};

use std::collections::HashSet;
use std::fs::File;
use std::io::BufReader;

pub fn json_keys<P: AsRef<std::path::Path>>(path: P) -> Result<Vec<String>> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let mut keys = Vec::new();
    let mut seen = HashSet::new();

    parse(reader, |event| {
        if let Event::Key(key, _) = event {
            let key_str = key.to_string();
            if !seen.contains(&key_str) {
                seen.insert(key_str.clone());
                keys.push(key_str);
            }
        }
    })?;

    Ok(keys)
}

pub fn json_sum_numbers<P: AsRef<std::path::Path>>(path: P) -> Result<f64> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let mut sum = 0.0;

    parse(reader, |event| {
        if let Event::Number(num, _) = event {
            sum += num.to_f64();
        }
    })?;

    Ok(sum)
}
