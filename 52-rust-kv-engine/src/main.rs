mod storage;
mod index;
mod compaction;
mod repl;

use repl::{KVEngine, run_repl};
use std::env;
use std::path::PathBuf;

fn main() {
    let args: Vec<String> = env::args().collect();
    
    let data_path = if args.len() > 1 {
        PathBuf::from(&args[1])
    } else {
        PathBuf::from("data.db")
    };

    println!("Data file: {:?}", data_path);

    match KVEngine::new(data_path) {
        Ok(mut engine) => {
            if let Err(e) = run_repl(&mut engine) {
                eprintln!("Error running REPL: {}", e);
                std::process::exit(1);
            }
        }
        Err(e) => {
            eprintln!("Failed to initialize KV engine: {}", e);
            std::process::exit(1);
        }
    }
}
