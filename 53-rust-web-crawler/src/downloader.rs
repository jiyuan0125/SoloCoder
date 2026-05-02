use url::Url;
use std::path::Path;
use std::collections::VecDeque;
use std::sync::{Arc, Mutex};
use std::time::Instant;
use reqwest::blocking::Client;
use hashbrown::HashSet as HashSetExt;

use crate::path_mapper::PathMapper;
use crate::robots_parser::{RobotsTxt, RobotsError};
use crate::html_link_rewriter::HtmlLinkRewriter;

#[derive(Debug, Clone)]
pub struct DownloadStats {
    pub pages_downloaded: u64,
    pub resources_downloaded: u64,
    pub failed_downloads: u64,
    pub start_time: Instant,
    pub elapsed_seconds: f64,
}

impl DownloadStats {
    pub fn new() -> Self {
        DownloadStats {
            pages_downloaded: 0,
            resources_downloaded: 0,
            failed_downloads: 0,
            start_time: Instant::now(),
            elapsed_seconds: 0.0,
        }
    }

    pub fn finish(&mut self) {
        self.elapsed_seconds = self.start_time.elapsed().as_secs_f64();
    }

    pub fn print_summary(&self) {
        println!("========================================");
        println!("下载完成统计：");
        println!("  HTML 页面数: {}", self.pages_downloaded);
        println!("  静态资源数: {}", self.resources_downloaded);
        println!("  下载失败数: {}", self.failed_downloads);
        println!("  总耗时: {:.2} 秒", self.elapsed_seconds);
        println!("========================================");
    }
}

struct DownloadTask {
    url: Url,
    depth: u32,
}

pub struct Downloader {
    path_mapper: Arc<PathMapper>,
    robots_txt: Arc<Mutex<RobotsTxt>>,
    client: Arc<Client>,
    visited_urls: Arc<Mutex<HashSetExt<String>>>,
    download_queue: Arc<Mutex<VecDeque<DownloadTask>>>,
    stats: Arc<Mutex<DownloadStats>>,
    max_depth: u32,
    max_concurrent: usize,
}

impl Downloader {
    pub fn new(
        base_url: &Url,
        output_dir: &Path,
        max_depth: u32,
        max_concurrent: usize,
    ) -> Result<Self, RobotsError> {
        let path_mapper = Arc::new(PathMapper::new(base_url, output_dir));
        
        let client = Arc::new(Client::builder()
            .user_agent("WebCrawler/1.0")
            .timeout(std::time::Duration::from_secs(30))
            .build()?);
        
        let mut robots_txt = RobotsTxt::new(base_url);
        robots_txt.fetch(&client)?;
        
        let robots_txt = Arc::new(Mutex::new(robots_txt));
        let visited_urls = Arc::new(Mutex::new(HashSetExt::new()));
        let download_queue = Arc::new(Mutex::new(VecDeque::new()));
        let stats = Arc::new(Mutex::new(DownloadStats::new()));
        
        Ok(Downloader {
            path_mapper,
            robots_txt,
            client,
            visited_urls,
            download_queue,
            stats,
            max_depth,
            max_concurrent,
        })
    }

    pub fn download(&mut self, seed_url: &Url) -> DownloadStats {
        let canonical_seed = self.path_mapper.canonicalize_url(seed_url);
        
        {
            let mut visited = self.visited_urls.lock().unwrap();
            visited.insert(canonical_seed.to_string());
        }
        
        {
            let mut queue = self.download_queue.lock().unwrap();
            queue.push_back(DownloadTask {
                url: canonical_seed,
                depth: 0,
            });
        }
        
        self.process_queue();
        
        let mut stats = self.stats.lock().unwrap();
        stats.finish();
        stats.clone()
    }

    fn process_queue(&mut self) {
        use std::thread;
        
        loop {
            let tasks_remaining = {
                let queue = self.download_queue.lock().unwrap();
                queue.len()
            };
            
            if tasks_remaining == 0 {
                break;
            }
            
            let mut handles = Vec::with_capacity(self.max_concurrent);
            
            for _ in 0..self.max_concurrent {
                let task = {
                    let mut queue = self.download_queue.lock().unwrap();
                    queue.pop_front()
                };
                
                if let Some(task) = task {
                    let path_mapper = Arc::clone(&self.path_mapper);
                    let robots_txt = Arc::clone(&self.robots_txt);
                    let client = Arc::clone(&self.client);
                    let visited_urls = Arc::clone(&self.visited_urls);
                    let download_queue = Arc::clone(&self.download_queue);
                    let stats = Arc::clone(&self.stats);
                    let max_depth = self.max_depth;
                    
                    let handle = thread::spawn(move || {
                        Self::download_task(
                            task,
                            path_mapper,
                            robots_txt,
                            client,
                            visited_urls,
                            download_queue,
                            stats,
                            max_depth,
                        );
                    });
                    
                    handles.push(handle);
                }
            }
            
            for handle in handles {
                let _ = handle.join();
            }
        }
    }

    fn download_task(
        task: DownloadTask,
        path_mapper: Arc<PathMapper>,
        robots_txt: Arc<Mutex<RobotsTxt>>,
        client: Arc<Client>,
        visited_urls: Arc<Mutex<HashSetExt<String>>>,
        download_queue: Arc<Mutex<VecDeque<DownloadTask>>>,
        stats: Arc<Mutex<DownloadStats>>,
        max_depth: u32,
    ) {
        let canonical_url = path_mapper.canonicalize_url(&task.url);
        
        let is_allowed = {
            let robots = robots_txt.lock().unwrap();
            robots.is_allowed(&canonical_url)
        };
        
        if !is_allowed {
            println!("已跳过 (robots.txt 禁止): {}", canonical_url);
            return;
        }
        
        let local_path = path_mapper.url_to_local_path(&canonical_url);
        
        println!("正在下载 [深度 {}]: {}", task.depth, canonical_url);
        
        let response = match client.get(canonical_url.as_str()).send() {
            Ok(resp) => resp,
            Err(e) => {
                eprintln!("下载失败 {}: {}", canonical_url, e);
                {
                    let mut stats = stats.lock().unwrap();
                    stats.failed_downloads += 1;
                }
                return;
            }
        };
        
        if !response.status().is_success() {
            eprintln!("下载失败 {}: HTTP 状态码 {}", canonical_url, response.status());
            {
                let mut stats = stats.lock().unwrap();
                stats.failed_downloads += 1;
            }
            return;
        }
        
        let content_type = response
            .headers()
            .get("content-type")
            .map(|v| v.to_str().unwrap_or(""))
            .unwrap_or("");
        
        let is_html = HtmlLinkRewriter::is_html_content(content_type);
        let is_resource = HtmlLinkRewriter::is_static_resource(content_type);
        
        let body = match response.bytes() {
            Ok(b) => b,
            Err(e) => {
                eprintln!("读取响应失败 {}: {}", canonical_url, e);
                {
                    let mut stats = stats.lock().unwrap();
                    stats.failed_downloads += 1;
                }
                return;
            }
        };
        
        let final_content: Vec<u8> = if is_html {
            let html_str = String::from_utf8_lossy(&body);
            let rewriter = HtmlLinkRewriter::new(
                (*path_mapper).clone(),
                &canonical_url,
                &local_path,
            );
            
            match rewriter.rewrite(&html_str) {
                Ok(rewritten) => rewritten.into_bytes(),
                Err(e) => {
                    eprintln!("HTML 重写失败 {}: {}", canonical_url, e);
                    body.to_vec()
                }
            }
        } else {
            body.to_vec()
        };
        
        if let Some(parent) = local_path.parent() {
            if let Err(e) = std::fs::create_dir_all(parent) {
                eprintln!("创建目录失败 {:?}: {}", parent, e);
                {
                    let mut stats = stats.lock().unwrap();
                    stats.failed_downloads += 1;
                }
                return;
            }
        }
        
        if let Err(e) = std::fs::write(&local_path, &final_content) {
            eprintln!("写入文件失败 {:?}: {}", local_path, e);
            {
                let mut stats = stats.lock().unwrap();
                stats.failed_downloads += 1;
            }
            return;
        }
        
        {
            let mut stats = stats.lock().unwrap();
            if is_html {
                stats.pages_downloaded += 1;
            } else if is_resource {
                stats.resources_downloaded += 1;
            } else {
                stats.pages_downloaded += 1;
            }
        }
        
        println!("已保存: {:?}", local_path);
        
        if is_html && task.depth < max_depth {
            let html_str = String::from_utf8_lossy(&body);
            let extracted_links = HtmlLinkRewriter::extract_links(&html_str, &canonical_url);
            
            for link_url in extracted_links {
                let canonical_link = path_mapper.canonicalize_url(&link_url);
                
                if !path_mapper.is_same_domain(&canonical_link) {
                    continue;
                }
                
                let url_str = canonical_link.to_string();
                let already_visited = {
                    let mut visited = visited_urls.lock().unwrap();
                    if visited.contains(&url_str) {
                        true
                    } else {
                        visited.insert(url_str);
                        false
                    }
                };
                
                if !already_visited {
                    let mut queue = download_queue.lock().unwrap();
                    queue.push_back(DownloadTask {
                        url: canonical_link,
                        depth: task.depth + 1,
                    });
                }
            }
        }
    }

    pub fn get_path_mapper(&self) -> &Arc<PathMapper> {
        &self.path_mapper
    }

    pub fn get_stats(&self) -> DownloadStats {
        let stats = self.stats.lock().unwrap();
        stats.clone()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use url::Url;
    use std::path::Path;

    #[test]
    fn test_downloader_creation() {
        let base_url = Url::parse("https://example.com/").unwrap();
        let output_dir = Path::new("/tmp/test_output");
        let max_depth = 3;
        let max_concurrent = 3;
        
        let downloader = Downloader::new(&base_url, output_dir, max_depth, max_concurrent);
        
        assert!(downloader.is_ok());
        
        if let Ok(d) = downloader {
            assert_eq!(d.max_depth, 3);
            assert_eq!(d.max_concurrent, 3);
        }
    }

    #[test]
    fn test_stats_initialization() {
        let stats = DownloadStats::new();
        
        assert_eq!(stats.pages_downloaded, 0);
        assert_eq!(stats.resources_downloaded, 0);
        assert_eq!(stats.failed_downloads, 0);
        assert_eq!(stats.elapsed_seconds, 0.0);
    }

    #[test]
    fn test_stats_finish() {
        let mut stats = DownloadStats::new();
        std::thread::sleep(std::time::Duration::from_millis(100));
        stats.finish();
        
        assert!(stats.elapsed_seconds >= 0.1);
    }
}
