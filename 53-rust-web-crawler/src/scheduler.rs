use std::collections::HashSet;
use std::path::PathBuf;
use std::sync::Arc;
use std::time::Instant;

use tokio::sync::{Semaphore, Mutex, mpsc};
use url::Url;
use reqwest::Client;

use crate::link_rewriter::LinkRewriter;
use crate::path_mapper::PathMapper;
use crate::robots_parser::RobotsTxt;

#[derive(Debug, Clone)]
pub struct CrawlStats {
    pub pages_downloaded: u32,
    pub resources_downloaded: u32,
    pub failed_requests: u32,
    pub start_time: Instant,
}

impl CrawlStats {
    pub fn new() -> Self {
        Self {
            pages_downloaded: 0,
            resources_downloaded: 0,
            failed_requests: 0,
            start_time: Instant::now(),
        }
    }

    pub fn elapsed_seconds(&self) -> u64 {
        self.start_time.elapsed().as_secs()
    }
}

impl Default for CrawlStats {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug, Clone)]
pub struct CrawlTask {
    pub url: Url,
    pub depth: u32,
}

pub struct Scheduler {
    client: Client,
    base_url: Url,
    output_dir: PathBuf,
    max_depth: u32,
    max_concurrent: usize,
    path_mapper: PathMapper,
    visited_urls: Arc<Mutex<HashSet<Url>>>,
    robots_txt: Arc<Mutex<Option<RobotsTxt>>>,
    stats: Arc<Mutex<CrawlStats>>,
}

impl Scheduler {
    pub fn new(
        base_url: Url,
        output_dir: PathBuf,
        max_depth: u32,
        max_concurrent: usize,
    ) -> Self {
        let client = Client::builder()
            .user_agent("WebMirror/0.1")
            .build()
            .unwrap_or_default();

        Self {
            client,
            base_url: base_url.clone(),
            output_dir: output_dir.clone(),
            max_depth,
            max_concurrent,
            path_mapper: PathMapper::new(output_dir),
            visited_urls: Arc::new(Mutex::new(HashSet::new())),
            robots_txt: Arc::new(Mutex::new(None)),
            stats: Arc::new(Mutex::new(CrawlStats::new())),
        }
    }

    pub async fn run(&self) -> CrawlStats {
        self.fetch_robots_txt().await;

        let semaphore = Arc::new(Semaphore::new(self.max_concurrent));
        let (tx, mut rx) = mpsc::unbounded_channel::<CrawlTask>();

        let initial_task = CrawlTask {
            url: PathMapper::normalize_url(&self.base_url),
            depth: 0,
        };

        tx.send(initial_task).ok();

        let mut pending_tasks: Vec<CrawlTask> = Vec::new();
        let mut active_tasks = 0;
        let (done_tx, mut done_rx) = mpsc::unbounded_channel::<()>();

        loop {
            tokio::select! {
                Some(task) = rx.recv() => {
                    let should_process = {
                        let mut visited = self.visited_urls.lock().await;
                        if visited.contains(&task.url) || task.depth > self.max_depth {
                            false
                        } else {
                            visited.insert(task.url.clone());
                            true
                        }
                    };

                    if should_process {
                        if !self.is_allowed_by_robots(&task.url).await {
                            continue;
                        }
                        if !self.is_same_domain(&task.url) {
                            continue;
                        }
                        pending_tasks.push(task);
                    }
                }

                Some(_) = done_rx.recv() => {
                    active_tasks -= 1;
                }

                else => break,
            }

            while let Some(task) = pending_tasks.pop() {
                let permit = match semaphore.clone().try_acquire_owned() {
                    Ok(p) => p,
                    Err(_) => {
                        pending_tasks.push(task);
                        break;
                    }
                };

                let client = self.client.clone();
                let path_mapper = self.path_mapper.clone();
                let base_url = self.base_url.clone();
                let tx = tx.clone();
                let visited = self.visited_urls.clone();
                let stats = self.stats.clone();
                let max_depth = self.max_depth;
                let done_tx = done_tx.clone();

                tokio::spawn(async move {
                    let _permit = permit;

                    let (local_path, is_html) = path_mapper.url_to_local_path(&task.url);

                    let result = Self::fetch_and_save(
                        &client,
                        &task.url,
                        &local_path,
                        is_html,
                        &path_mapper,
                        &base_url,
                        task.depth,
                        max_depth,
                        &tx,
                        &visited,
                    ).await;

                    let mut stats_lock = stats.lock().await;
                    match result {
                        Ok(was_html) => {
                            if was_html {
                                stats_lock.pages_downloaded += 1;
                            } else {
                                stats_lock.resources_downloaded += 1;
                            }
                        }
                        Err(_) => {
                            stats_lock.failed_requests += 1;
                        }
                    }

                    done_tx.send(()).ok();
                });

                active_tasks += 1;
            }

            if pending_tasks.is_empty() && active_tasks == 0 && rx.is_empty() {
                break;
            }
        }

        let stats = self.stats.lock().await;
        stats.clone()
    }

    async fn fetch_robots_txt(&self) {
        let robots_url = match self.base_url.join("/robots.txt") {
            Ok(url) => url,
            Err(_) => return,
        };

        let response = match self.client.get(robots_url).send().await {
            Ok(resp) => resp,
            Err(_) => return,
        };

        let content = match response.text().await {
            Ok(text) => text,
            Err(_) => return,
        };

        let robots = RobotsTxt::parse(&content);
        let mut robots_lock = self.robots_txt.lock().await;
        *robots_lock = Some(robots);
    }

    async fn is_allowed_by_robots(&self, url: &Url) -> bool {
        let robots_lock = self.robots_txt.lock().await;
        match &*robots_lock {
            Some(robots) => robots.is_allowed(url),
            None => true,
        }
    }

    fn is_same_domain(&self, url: &Url) -> bool {
        self.base_url.host_str() == url.host_str()
            && self.base_url.scheme() == url.scheme()
            && self.base_url.port() == url.port()
    }

    #[allow(clippy::too_many_arguments)]
    async fn fetch_and_save(
        client: &Client,
        url: &Url,
        local_path: &PathBuf,
        is_html: bool,
        path_mapper: &PathMapper,
        _base_url: &Url,
        current_depth: u32,
        max_depth: u32,
        tx: &mpsc::UnboundedSender<CrawlTask>,
        visited: &Arc<Mutex<HashSet<Url>>>,
    ) -> Result<bool, reqwest::Error> {
        let response = client.get(url.clone()).send().await?;

        let content_type = response.headers()
            .get(reqwest::header::CONTENT_TYPE)
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");

        let is_response_html = content_type.contains("text/html")
            || content_type.contains("application/xhtml+xml")
            || is_html;

        let bytes = response.bytes().await?;

        if is_response_html {
            let html_str = String::from_utf8_lossy(&bytes);
            let rewriter = LinkRewriter::new(url.clone(), path_mapper.clone());
            let result = rewriter.rewrite_html(&html_str);

            if current_depth < max_depth {
                for extracted_url in result.extracted_urls {
                    let should_add = {
                        let mut v = visited.lock().await;
                        if v.contains(&extracted_url) {
                            false
                        } else {
                            v.insert(extracted_url.clone());
                            true
                        }
                    };

                    if should_add {
                        tx.send(CrawlTask {
                            url: extracted_url,
                            depth: current_depth + 1,
                        }).ok();
                    }
                }
            }

            if let Some(parent) = local_path.parent() {
                std::fs::create_dir_all(parent).ok();
            }
            std::fs::write(local_path, result.rewritten_html).ok();
        } else {
            if let Some(parent) = local_path.parent() {
                std::fs::create_dir_all(parent).ok();
            }
            std::fs::write(local_path, bytes).ok();
        }

        Ok(is_response_html)
    }
}

pub struct Config {
    pub seed_url: Url,
    pub output_dir: PathBuf,
    pub max_depth: u32,
    pub max_concurrent: usize,
}

impl Config {
    pub fn from_args(args: &[String]) -> Result<Self, String> {
        let mut seed_url = None;
        let mut output_dir = None;
        let mut max_depth = 3;
        let mut max_concurrent = 3;

        let mut i = 1;
        while i < args.len() {
            match args[i].as_str() {
                "-h" | "--help" => {
                    return Err("help".to_string());
                }
                "-u" | "--url" => {
                    if i + 1 < args.len() {
                        seed_url = Some(args[i + 1].clone());
                        i += 2;
                    } else {
                        return Err("缺少 URL 参数值".to_string());
                    }
                }
                "-o" | "--output" => {
                    if i + 1 < args.len() {
                        output_dir = Some(args[i + 1].clone());
                        i += 2;
                    } else {
                        return Err("缺少输出目录参数值".to_string());
                    }
                }
                "-d" | "--depth" => {
                    if i + 1 < args.len() {
                        max_depth = args[i + 1].parse().map_err(|_| "深度必须是整数".to_string())?;
                        i += 2;
                    } else {
                        return Err("缺少深度参数值".to_string());
                    }
                }
                "-c" | "--concurrency" => {
                    if i + 1 < args.len() {
                        max_concurrent = args[i + 1].parse().map_err(|_| "并发数必须是整数".to_string())?;
                        i += 2;
                    } else {
                        return Err("缺少并发数参数值".to_string());
                    }
                }
                _ => {
                    if seed_url.is_none() {
                        seed_url = Some(args[i].clone());
                    } else if output_dir.is_none() {
                        output_dir = Some(args[i].clone());
                    }
                    i += 1;
                }
            }
        }

        let seed_url = seed_url.ok_or_else(|| "请提供种子 URL".to_string())?;
        let seed_url = Url::parse(&seed_url).map_err(|e| format!("无效的 URL: {}", e))?;

        let output_dir = output_dir.unwrap_or_else(|| {
            let host = seed_url.host_str().unwrap_or("mirror");
            format!("./{}", host)
        });
        let output_dir = PathBuf::from(output_dir);

        Ok(Self {
            seed_url,
            output_dir,
            max_depth,
            max_concurrent,
        })
    }

    pub fn print_help() {
        println!("用法: web_mirror [选项] <URL> [输出目录]");
        println!();
        println!("选项:");
        println!("  -u, --url <URL>       种子 URL");
        println!("  -o, --output <目录>    输出目录");
        println!("  -d, --depth <深度>     最大爬取深度 (默认: 3)");
        println!("  -c, --concurrency <数> 最大并发数 (默认: 3)");
        println!("  -h, --help            显示帮助信息");
    }
}
