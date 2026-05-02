mod path_mapper;
mod link_rewriter;
mod robots_parser;
mod scheduler;

use scheduler::{Config, Scheduler};

#[tokio::main]
async fn main() {
    let args: Vec<String> = std::env::args().collect();

    let config = match Config::from_args(&args) {
        Ok(c) => c,
        Err(e) => {
            if e == "help" {
                Config::print_help();
                std::process::exit(0);
            }
            eprintln!("错误: {}", e);
            println!();
            Config::print_help();
            std::process::exit(1);
        }
    };

    println!("站点镜像工具启动");
    println!("种子 URL: {}", config.seed_url);
    println!("输出目录: {}", config.output_dir.display());
    println!("最大深度: {}", config.max_depth);
    println!("最大并发: {}", config.max_concurrent);
    println!();

    let scheduler = Scheduler::new(
        config.seed_url,
        config.output_dir,
        config.max_depth,
        config.max_concurrent,
    );

    let stats = scheduler.run().await;

    println!();
    println!("爬取完成!");
    println!("页面数: {}", stats.pages_downloaded);
    println!("资源数 (CSS/JS/图片): {}", stats.resources_downloaded);
    println!("失败数: {}", stats.failed_requests);
    println!("总耗时: {} 秒", stats.elapsed_seconds());
}
