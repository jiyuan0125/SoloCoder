use std::env;
use std::fs;
use std::path::Path;

use markdown_renderer::block_parser::parse_blocks;
use markdown_renderer::html_generator::generate_html;

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() < 2 {
        eprintln!("Usage: {} <input.md> [output.html]", args[0]);
        std::process::exit(1);
    }

    let input_path = Path::new(&args[1]);
    let output_path = if args.len() >= 3 {
        Path::new(&args[2]).to_path_buf()
    } else {
        input_path.with_extension("html")
    };

    let markdown_content = match fs::read_to_string(input_path) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Error reading input file: {}", e);
            std::process::exit(1);
        }
    };

    let blocks = parse_blocks(&markdown_content);
    let html = generate_html(&blocks);

    if let Err(e) = fs::write(&output_path, html) {
        eprintln!("Error writing output file: {}", e);
        std::process::exit(1);
    }

    println!("Successfully converted {} to {}", input_path.display(), output_path.display());
}
