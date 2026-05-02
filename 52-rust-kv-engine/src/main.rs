use kv_engine::{StorageEngine, Repl};

const DEFAULT_DATA_FILE: &str = "kv_data.dat";

fn main() {
    let data_file = std::env::args().nth(1).unwrap_or_else(|| DEFAULT_DATA_FILE.to_string());
    
    println!("KV Storage Engine");
    println!("Data file: {}", data_file);
    println!();

    let engine = match StorageEngine::new(&data_file) {
        Ok(e) => e,
        Err(e) => {
            eprintln!("Failed to initialize storage engine: {}", e);
            std::process::exit(1);
        }
    };

    let mut repl = Repl::new(engine);
    repl.run();
}
