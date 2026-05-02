use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::{mpsc, Mutex, Semaphore};
use url::Url;

use crate::parser::parse_html;
use crate::robots::RobotsTxt;
use crate::url_queue::{UrlItem, UrlQueue};

#[derive(Debug, Clone, serde::Serialize)]
pub struct CrawlResult {
    pub url: String,
    pub title: Option<String>,
    pub status_code: u16,
    pub depth: usize,
}

pub struct CrawlerConfig {
    pub seed_urls: Vec<Url>,
    pub max_depth: usize,
    pub max_concurrent_per_domain: usize,
    pub max_total_concurrent: usize,
    pub user_agent: String,
    pub output_file: String,
}

impl Default for CrawlerConfig {
    fn default() -> Self {
        Self {
            seed_urls: Vec::new(),
            max_depth: 3,
            max_concurrent_per_domain: 2,
            max_total_concurrent: 10,
            user_agent: "WebCrawler/1.0".to_string(),
            output_file: "crawl_results.json".to_string(),
        }
    }
}

pub struct CrawlerScheduler {
    config: CrawlerConfig,
    url_queue: Arc<UrlQueue>,
    http_client: reqwest::Client,
    robots_cache: Arc<Mutex<HashMap<String, RobotsTxt>>>,
    results: Arc<Mutex<Vec<CrawlResult>>>,
    shutdown_signal: tokio::sync::watch::Sender<bool>,
    shutdown_receiver: tokio::sync::watch::Receiver<bool>,
}

impl CrawlerScheduler {
    pub fn new(config: CrawlerConfig) -> Self {
        let (shutdown_tx, shutdown_rx) = tokio::sync::watch::channel(false);
        
        let http_client = reqwest::Client::builder()
            .user_agent(&config.user_agent)
            .timeout(std::time::Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client");
        
        let url_queue = Arc::new(UrlQueue::new(
            config.max_depth,
            config.max_concurrent_per_domain,
        ));
        
        Self {
            config,
            url_queue,
            http_client,
            robots_cache: Arc::new(Mutex::new(HashMap::new())),
            results: Arc::new(Mutex::new(Vec::new())),
            shutdown_signal: shutdown_tx,
            shutdown_receiver: shutdown_rx,
        }
    }

    pub async fn run(&self) -> anyhow::Result<()> {
        for seed_url in &self.config.seed_urls {
            self.url_queue.add_url(seed_url.clone(), 0).await;
        }

        let total_semaphore = Arc::new(Semaphore::new(self.config.max_total_concurrent));
        
        let (result_tx, mut result_rx) = mpsc::unbounded_channel();
        
        let results_clone = self.results.clone();
        let result_handler = tokio::spawn(async move {
            while let Some(result) = result_rx.recv().await {
                let mut results = results_clone.lock().await;
                results.push(result);
            }
        });

        let mut active_tasks = Vec::new();

        loop {
            let mut rx = self.shutdown_receiver.clone();
            
            tokio::select! {
                _ = rx.changed() => {
                    if *rx.borrow() {
                        break;
                    }
                }
                
                url_item = self.url_queue.next() => {
                    if let Some(item) = url_item {
                        if item.depth > self.config.max_depth {
                            continue;
                        }

                        let permit = total_semaphore.clone().acquire_owned().await.unwrap();
                        
                        let http_client = self.http_client.clone();
                        let url_queue = self.url_queue.clone();
                        let robots_cache = self.robots_cache.clone();
                        let result_tx = result_tx.clone();
                        let user_agent = self.config.user_agent.clone();
                        let max_depth = self.config.max_depth;
                        
                        let task = tokio::spawn(async move {
                            let _permit = permit;
                            
                            Self::process_url(
                                http_client,
                                url_queue,
                                robots_cache,
                                result_tx,
                                item,
                                &user_agent,
                                max_depth,
                            ).await;
                        });
                        
                        active_tasks.push(task);
                    } else {
                        if active_tasks.is_empty() {
                            break;
                        }
                        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
                    }
                }
            }
            
            active_tasks.retain(|task| !task.is_finished());
        }

        drop(result_tx);
        
        for task in active_tasks {
            let _ = task.await;
        }
        
        result_handler.await?;
        
        self.save_results().await?;
        
        Ok(())
    }

    async fn process_url(
        http_client: reqwest::Client,
        url_queue: Arc<UrlQueue>,
        robots_cache: Arc<Mutex<HashMap<String, RobotsTxt>>>,
        result_tx: mpsc::UnboundedSender<CrawlResult>,
        item: UrlItem,
        user_agent: &str,
        max_depth: usize,
    ) {
        let url = item.url.clone();
        let domain = url_queue.get_domain(&url);
        
        let is_allowed = Self::check_robots(&http_client, &robots_cache, &url, user_agent).await;
        
        if !is_allowed {
            let result = CrawlResult {
                url: url.to_string(),
                title: None,
                status_code: 403,
                depth: item.depth,
            };
            let _ = result_tx.send(result);
            return;
        }
        
        let _permit = url_queue.acquire_domain_permit(&domain).await;
        
        let response_result = http_client.get(url.clone()).send().await;
        
        match response_result {
            Ok(response) => {
                let status = response.status().as_u16();
                let mut title = None;
                
                if status == 200 {
                    if let Ok(text) = response.text().await {
                        let parsed = parse_html(&text, &url);
                        title = parsed.title.clone();
                        
                        if item.depth < max_depth {
                            let next_depth = item.depth + 1;
                            for link in parsed.links {
                                url_queue.add_url(link, next_depth).await;
                            }
                        }
                    }
                }
                
                let result = CrawlResult {
                    url: url.to_string(),
                    title,
                    status_code: status,
                    depth: item.depth,
                };
                let _ = result_tx.send(result);
            }
            Err(e) => {
                eprintln!("Error fetching {}: {}", url, e);
                let result = CrawlResult {
                    url: url.to_string(),
                    title: None,
                    status_code: 0,
                    depth: item.depth,
                };
                let _ = result_tx.send(result);
            }
        }
    }

    async fn check_robots(
        http_client: &reqwest::Client,
        robots_cache: &Arc<Mutex<HashMap<String, RobotsTxt>>>,
        url: &Url,
        user_agent: &str,
    ) -> bool {
        let domain = match url.host_str() {
            Some(d) => d.to_string(),
            None => return true,
        };
        
        let scheme = url.scheme();
        let port = url.port().map(|p| format!(":{}", p)).unwrap_or_default();
        let robots_url_str = format!("{}://{}{}/robots.txt", scheme, domain, port);
        
        let robots_url = match Url::parse(&robots_url_str) {
            Ok(u) => u,
            Err(_) => return true,
        };
        
        let cache = robots_cache.lock().await;
        
        if let Some(robots) = cache.get(&domain) {
            return robots.is_allowed(url, user_agent);
        }
        
        drop(cache);
        
        let robots = match http_client.get(robots_url).send().await {
            Ok(response) => {
                if response.status().is_success() {
                    match response.text().await {
                        Ok(text) => RobotsTxt::parse(&text),
                        Err(_) => RobotsTxt::new(),
                    }
                } else {
                    RobotsTxt::new()
                }
            }
            Err(_) => RobotsTxt::new(),
        };
        
        let is_allowed = robots.is_allowed(url, user_agent);
        
        let mut cache = robots_cache.lock().await;
        cache.insert(domain, robots);
        
        is_allowed
    }

    pub async fn save_results(&self) -> anyhow::Result<()> {
        let results = self.results.lock().await;
        
        let file = std::fs::File::create(&self.config.output_file)?;
        serde_json::to_writer_pretty(file, &*results)?;
        
        println!("Saved {} results to {}", results.len(), self.config.output_file);
        
        Ok(())
    }

    pub fn shutdown(&self) {
        let _ = self.shutdown_signal.send(true);
    }

    pub async fn get_results(&self) -> Vec<CrawlResult> {
        self.results.lock().await.clone()
    }
}

pub async fn setup_signal_handler(scheduler: Arc<CrawlerScheduler>) {
    tokio::spawn(async move {
        match tokio::signal::ctrl_c().await {
            Ok(_) => {
                println!("\nReceived SIGINT, initiating graceful shutdown...");
                println!("Waiting for in-flight requests to complete...");
                scheduler.shutdown();
            }
            Err(e) => {
                eprintln!("Failed to listen for ctrl-c signal: {}", e);
            }
        }
    });
}
