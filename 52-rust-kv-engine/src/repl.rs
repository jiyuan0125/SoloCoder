use std::io::{self, BufRead, Write};
use crate::storage::StorageEngine;
use crate::storage::MAX_VALUE_SIZE;

pub struct Repl {
    engine: StorageEngine,
}

impl Repl {
    pub fn new(engine: StorageEngine) -> Self {
        Repl { engine }
    }

    pub fn run(&mut self) {
        println!("KV Storage Engine REPL");
        println!("Commands: put <key> <value>, get <key>, delete <key>, scan <prefix>, count, quit");
        println!();

        let stdin = io::stdin();
        let mut stdout = io::stdout();
        
        loop {
            print!("> ");
            if let Err(e) = stdout.flush() {
                eprintln!("Error: {}", e);
                continue;
            }

            let mut line = String::new();
            match stdin.lock().read_line(&mut line) {
                Ok(0) => {
                    println!();
                    break;
                }
                Ok(_) => {}
                Err(e) => {
                    eprintln!("Error reading input: {}", e);
                    continue;
                }
            }

            let trimmed = line.trim();
            if trimmed.is_empty() {
                continue;
            }

            if let Err(e) = self.execute_command(trimmed) {
                eprintln!("Error: {}", e);
            }
        }
    }

    fn execute_command(&mut self, line: &str) -> Result<(), String> {
        let parts: Vec<&str> = line.splitn(3, |c: char| c.is_whitespace())
            .filter(|s| !s.is_empty())
            .collect();

        if parts.is_empty() {
            return Ok(());
        }

        let command = parts[0].to_lowercase();

        match command.as_str() {
            "quit" | "exit" => {
                println!("Bye!");
                std::process::exit(0);
            }
            "put" => {
                if parts.len() < 3 {
                    return Err("Usage: put <key> <value>".to_string());
                }
                let key = parts[1].to_string();
                let value = parts[2..].join(" ").into_bytes();
                
                if value.len() > MAX_VALUE_SIZE {
                    return Err(format!("Value size exceeds maximum limit of {} bytes", MAX_VALUE_SIZE));
                }
                
                self.engine.put(key, value)
                    .map_err(|e| format!("Put failed: {}", e))?;
                println!("OK");
            }
            "get" => {
                if parts.len() < 2 {
                    return Err("Usage: get <key>".to_string());
                }
                let key = parts[1];
                match self.engine.get(key) {
                    Ok(Some(value)) => {
                        match String::from_utf8(value) {
                            Ok(s) => println!("{}", s),
                            Err(_) => println!("<binary data>"),
                        }
                    }
                    Ok(None) => {
                        println!("Key not found");
                    }
                    Err(e) => {
                        return Err(format!("Get failed: {}", e));
                    }
                }
            }
            "delete" => {
                if parts.len() < 2 {
                    return Err("Usage: delete <key>".to_string());
                }
                let key = parts[1].to_string();
                self.engine.delete(key)
                    .map_err(|e| format!("Delete failed: {}", e))?;
                println!("OK");
            }
            "scan" => {
                let prefix = if parts.len() >= 2 { parts[1] } else { "" };
                let keys = self.engine.scan(prefix);
                for key in keys {
                    println!("{}", key);
                }
            }
            "count" => {
                let count = self.engine.count();
                println!("{}", count);
            }
            _ => {
                println!("Unknown command: {}", command);
                println!("Available commands: put, get, delete, scan, count, quit");
            }
        }

        Ok(())
    }
}
