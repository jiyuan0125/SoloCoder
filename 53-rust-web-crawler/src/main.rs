use std::sync::Arc;
use url::Url;

use web_crawler::scheduler::{setup_signal_handler, CrawlerConfig, CrawlerScheduler};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args: Vec<String> = std::env::args().collect();
    
    if args.len() < 2 {
        println!("Usage: {} <seed_url> [max_depth] [max_concurrent_per_domain] [output_file]", args[0]);
        println!("Example: {} https://example.com 3 2 results.json", args[0]);
        std::process::exit(1);
    }
    
    let seed_url_str = &args[1];
    let seed_url = match Url::parse(seed_url_str) {
        Ok(url) => url,
        Err(e) => {
            eprintln!("Invalid URL: {}", e);
            std::process::exit(1);
        }
    };
    
    let max_depth = args.get(2)
        .and_then(|s| s.parse::<usize>().ok())
        .unwrap_or(3);
    
    let max_concurrent_per_domain = args.get(3)
        .and_then(|s| s.parse::<usize>().ok())
        .unwrap_or(2);
    
    let output_file = args.get(4)
        .cloned()
        .unwrap_or_else(|| "crawl_results.json".to_string());
    
    println!("========================================");
    println!("Web Crawler Configuration");
    println!("========================================");
    println!("Seed URL: {}", seed_url);
    println!("Max Depth: {}", max_depth);
    println!("Max Concurrent per Domain: {}", max_concurrent_per_domain);
    println!("Output File: {}", output_file);
    println!("User Agent: WebCrawler/1.0");
    println!("========================================");
    println!("Press Ctrl+C to gracefully stop");
    println!();
    
    let config = CrawlerConfig {
        seed_urls: vec![seed_url],
        max_depth,
        max_concurrent_per_domain,
        max_total_concurrent: 10,
        user_agent: "WebCrawler/1.0".to_string(),
        output_file,
    };
    
    let scheduler = Arc::new(CrawlerScheduler::new(config));
    
    setup_signal_handler(scheduler.clone()).await;
    
    println!("Starting crawl...");
    
    match scheduler.run().await {
        Ok(_) => {
            let results = scheduler.get_results().await;
            println!();
            println!("========================================");
            println!("Crawl Completed");
            println!("========================================");
            println!("Total pages crawled: {}", results.len());
            
            let success_count = results.iter().filter(|r| r.status_code == 200).count();
            println!("Successful (200 OK): {}", success_count);
            
            let error_count = results.iter().filter(|r| r.status_code != 200 && r.status_code != 0).count();
            if error_count > 0 {
                println!("HTTP errors: {}", error_count);
            }
            
            let fail_count = results.iter().filter(|r| r.status_code == 0).count();
            if fail_count > 0 {
                println!("Network errors: {}", fail_count);
            }
        }
        Err(e) => {
            eprintln!("Crawl failed with error: {}", e);
            std::process::exit(1);
        }
    }
    
    Ok(())
}
