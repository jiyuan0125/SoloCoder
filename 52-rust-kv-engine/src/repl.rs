use crate::compaction::compact_if_needed;
use crate::index::Index;
use crate::storage::{append_record, open_data_file, read_record, Record, RecordType, MAX_VALUE_SIZE};
use std::fs::File;
use std::io::{self, BufRead, Write};
use std::path::PathBuf;

pub struct KVEngine {
    data_path: PathBuf,
    file: File,
    index: Index,
}

impl KVEngine {
    pub fn new(data_path: PathBuf) -> io::Result<Self> {
        let mut file = open_data_file(&data_path)?;
        let index = Index::build_from_file(&mut file)?;
        
        Ok(KVEngine {
            data_path,
            file,
            index,
        })
    }

    pub fn put(&mut self, key: String, value: Vec<u8>) -> io::Result<()> {
        if value.len() as u64 > MAX_VALUE_SIZE {
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                format!("Value size exceeds maximum limit of {} bytes", MAX_VALUE_SIZE),
            ));
        }

        let record = Record::new_put(key, value);
        let offset = append_record(&mut self.file, &record)?;
        let record_size = record.serialized_size();
        
        self.index.update(&record, offset, record_size);

        if compact_if_needed(&self.data_path, &mut self.file, &mut self.index)? {
            self.file = open_data_file(&self.data_path)?;
        }

        Ok(())
    }

    pub fn get(&mut self, key: &str) -> io::Result<Option<Vec<u8>>> {
        let entry = match self.index.get(key) {
            Some(e) => e,
            None => return Ok(None),
        };

        if !entry.is_alive {
            return Ok(None);
        }

        let (record, _) = match read_record(&mut self.file, entry.offset)? {
            Some(r) => r,
            None => return Ok(None),
        };

        match record.op_type {
            RecordType::Put => Ok(record.value),
            RecordType::Delete => Ok(None),
        }
    }

    pub fn delete(&mut self, key: String) -> io::Result<()> {
        let record = Record::new_delete(key);
        let offset = append_record(&mut self.file, &record)?;
        let record_size = record.serialized_size();
        
        self.index.update(&record, offset, record_size);

        if compact_if_needed(&self.data_path, &mut self.file, &mut self.index)? {
            self.file = open_data_file(&self.data_path)?;
        }

        Ok(())
    }

    pub fn scan_prefix(&self, prefix: &str) -> Vec<String> {
        self.index.scan_prefix(prefix)
    }

    pub fn count(&self) -> usize {
        self.index.count()
    }
}

pub enum Command {
    Put { key: String, value: String },
    Get { key: String },
    Delete { key: String },
    Scan { prefix: String },
    Count,
    Quit,
    Empty,
    Unknown,
}

pub fn parse_command(input: &str) -> Command {
    let trimmed = input.trim();
    if trimmed.is_empty() {
        return Command::Empty;
    }

    let parts: Vec<&str> = trimmed.splitn(3, |c: char| c.is_whitespace()).collect();

    match parts[0].to_lowercase().as_str() {
        "put" => {
            if parts.len() < 3 {
                Command::Unknown
            } else {
                Command::Put {
                    key: parts[1].to_string(),
                    value: parts[2].to_string(),
                }
            }
        }
        "get" => {
            if parts.len() < 2 {
                Command::Unknown
            } else {
                Command::Get {
                    key: parts[1].to_string(),
                }
            }
        }
        "delete" => {
            if parts.len() < 2 {
                Command::Unknown
            } else {
                Command::Delete {
                    key: parts[1].to_string(),
                }
            }
        }
        "scan" => {
            let prefix = if parts.len() >= 2 { parts[1] } else { "" };
            Command::Scan {
                prefix: prefix.to_string(),
            }
        }
        "count" => Command::Count,
        "quit" => Command::Quit,
        _ => Command::Unknown,
    }
}

pub fn run_repl(engine: &mut KVEngine) -> io::Result<()> {
    println!("KV Storage Engine REPL");
    println!("Available commands: put, get, delete, scan, count, quit");
    println!();

    let stdin = io::stdin();
    let mut stdout = io::stdout();

    loop {
        print!("> ");
        stdout.flush()?;

        let mut line = String::new();
        match stdin.lock().read_line(&mut line) {
            Ok(0) => break,
            Ok(_) => {}
            Err(e) => {
                eprintln!("Error reading input: {}", e);
                continue;
            }
        }

        match parse_command(&line) {
            Command::Put { key, value } => {
                match engine.put(key.clone(), value.into_bytes()) {
                    Ok(_) => println!("OK"),
                    Err(e) => println!("Error: {}", e),
                }
            }
            Command::Get { key } => {
                match engine.get(&key) {
                    Ok(Some(value)) => {
                        match String::from_utf8(value) {
                            Ok(s) => println!("{}", s),
                            Err(_) => println!("[binary data]"),
                        }
                    }
                    Ok(None) => println!("Key not found"),
                    Err(e) => println!("Error: {}", e),
                }
            }
            Command::Delete { key } => {
                match engine.delete(key) {
                    Ok(_) => println!("OK"),
                    Err(e) => println!("Error: {}", e),
                }
            }
            Command::Scan { prefix } => {
                let keys = engine.scan_prefix(&prefix);
                if keys.is_empty() {
                    println!("(empty)");
                } else {
                    for key in keys {
                        println!("{}", key);
                    }
                }
            }
            Command::Count => {
                println!("{}", engine.count());
            }
            Command::Quit => {
                println!("Bye!");
                break;
            }
            Command::Empty => {
                continue;
            }
            Command::Unknown => {
                println!("Unknown command. Available commands: put, get, delete, scan, count, quit");
            }
        }
    }

    Ok(())
}
