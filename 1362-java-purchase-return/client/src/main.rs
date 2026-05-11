use purchase_return_client::{run_client, Args};
use clap::Parser;

fn main() {
    let args = Args::parse();
    if let Err(e) = run_client(&args) {
        eprintln!("Error: {}", e);
        std::process::exit(1);
    }
}
