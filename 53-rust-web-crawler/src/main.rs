mod path_mapper;
mod robots_parser;
mod html_link_rewriter;
mod downloader;

use url::Url;
use std::path::Path;
use std::env;
use anyhow::{Result, anyhow};
use crate::downloader::Downloader;

fn print_usage() {
    println!("用法: web_crawler <种子 URL> [输出目录] [深度限制]");
    println!();
    println!("参数:");
    println!("  <种子 URL>       要下载的网站起始 URL（必需）");
    println!("  [输出目录]         保存文件的本地目录（默认: ./output）");
    println!("  [深度限制]         爬取深度限制（默认: 3）");
    println!();
    println!("示例:");
    println!("  web_crawler https://example.com/");
    println!("  web_crawler https://example.com/ ./my_site 5");
}

fn main() -> Result<()> {
    let args: Vec<String> = env::args().collect();
    
    if args.len() < 2 {
        print_usage();
        return Err(anyhow!("缺少种子 URL 是必需的参数"));
    }
    
    if args[1] == "--help" || args[1] == "-h" {
        print_usage();
        return Ok(());
    }
    
    let seed_url_str = &args[1];
    let seed_url = Url::parse(seed_url_str)
        .map_err(|e| anyhow!("无效的 URL: {}", e))?;
    
    let output_dir = if args.len() >= 3 {
        Path::new(&args[2])
    } else {
        Path::new("./output")
    };
    
    let max_depth: u32 = if args.len() >= 4 {
        args[3].parse().unwrap_or(3)
    } else {
        3
    };
    
    let max_concurrent = 3;
    
    println!("========================================");
    println!("站点镜像工具启动");
    println!("========================================");
    println!("种子 URL: {}", seed_url);
    println!("输出目录: {:?}", output_dir);
    println!("深度限制: {}", max_depth);
    println!("并发限制: {}", max_concurrent);
    println!("========================================");
    println!();
    
    let mut downloader = Downloader::new(
        &seed_url,
        output_dir,
        max_depth,
        max_concurrent,
    )?;
    
    let stats = downloader.download(&seed_url);
    
    stats.print_summary();
    
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use url::Url;
    use std::path::Path;

    #[test]
    fn test_url_parsing() {
        let url = Url::parse("https://example.com/about/team").unwrap();
        assert_eq!(url.domain(), Some("example.com"));
    }

    #[test]
    fn test_output_dir_creation() {
        let output_dir = Path::new("/tmp/test_crawler_output");
        assert!(!output_dir.exists());
    }
}
