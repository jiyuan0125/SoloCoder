use std::collections::{HashMap, HashSet, VecDeque};
use std::sync::Arc;
use tokio::sync::{Mutex, OwnedSemaphorePermit, Semaphore};
use url::Url;

#[derive(Debug, Clone)]
pub struct UrlItem {
    pub url: Url,
    pub depth: usize,
}

impl UrlItem {
    pub fn new(url: Url, depth: usize) -> Self {
        Self { url, depth }
    }
}

pub struct UrlQueue {
    queue: Mutex<VecDeque<UrlItem>>,
    visited: Mutex<HashSet<String>>,
    max_depth: usize,
    domain_semaphores: Mutex<HashMap<String, Arc<Semaphore>>>,
    max_concurrent_per_domain: usize,
}

impl UrlQueue {
    pub fn new(max_depth: usize, max_concurrent_per_domain: usize) -> Self {
        Self {
            queue: Mutex::new(VecDeque::new()),
            visited: Mutex::new(HashSet::new()),
            max_depth,
            domain_semaphores: Mutex::new(HashMap::new()),
            max_concurrent_per_domain,
        }
    }

    pub async fn add_url(&self, url: Url, depth: usize) -> bool {
        if depth > self.max_depth {
            return false;
        }

        let url_str = url.to_string();
        let mut visited = self.visited.lock().await;
        
        if visited.contains(&url_str) {
            return false;
        }
        
        visited.insert(url_str);
        drop(visited);

        let mut queue = self.queue.lock().await;
        queue.push_back(UrlItem::new(url, depth));
        
        true
    }

    pub async fn add_urls(&self, urls: &[Url], depth: usize) -> usize {
        let mut added = 0;
        for url in urls {
            if self.add_url(url.clone(), depth).await {
                added += 1;
            }
        }
        added
    }

    pub async fn next(&self) -> Option<UrlItem> {
        let mut queue = self.queue.lock().await;
        queue.pop_front()
    }

    pub async fn is_empty(&self) -> bool {
        let queue = self.queue.lock().await;
        queue.is_empty()
    }

    pub async fn len(&self) -> usize {
        let queue = self.queue.lock().await;
        queue.len()
    }

    pub async fn is_visited(&self, url: &Url) -> bool {
        let visited = self.visited.lock().await;
        visited.contains(&url.to_string())
    }

    pub async fn acquire_domain_permit(&self, domain: &str) -> OwnedSemaphorePermit {
        let mut semaphores = self.domain_semaphores.lock().await;
        
        let semaphore = semaphores
            .entry(domain.to_string())
            .or_insert_with(|| Arc::new(Semaphore::new(self.max_concurrent_per_domain)));
        
        let sem = semaphore.clone();
        drop(semaphores);
        
        sem.acquire_owned().await.unwrap()
    }

    pub fn get_domain(&self, url: &Url) -> String {
        url.host_str().unwrap_or("unknown").to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_add_and_next() {
        let queue = UrlQueue::new(3, 2);
        
        let url1 = Url::parse("http://example.com/page1").unwrap();
        let url2 = Url::parse("http://example.com/page2").unwrap();
        
        assert!(queue.add_url(url1.clone(), 1).await);
        assert!(queue.add_url(url2.clone(), 1).await);
        
        let item1 = queue.next().await.unwrap();
        assert_eq!(item1.url, url1);
        assert_eq!(item1.depth, 1);
        
        let item2 = queue.next().await.unwrap();
        assert_eq!(item2.url, url2);
        assert_eq!(item2.depth, 1);
        
        assert!(queue.next().await.is_none());
    }

    #[tokio::test]
    async fn test_duplicate_url() {
        let queue = UrlQueue::new(3, 2);
        
        let url = Url::parse("http://example.com/page").unwrap();
        
        assert!(queue.add_url(url.clone(), 1).await);
        assert!(!queue.add_url(url.clone(), 1).await);
        
        assert_eq!(queue.len().await, 1);
    }

    #[tokio::test]
    async fn test_depth_limit() {
        let queue = UrlQueue::new(2, 2);
        
        let url1 = Url::parse("http://example.com/page1").unwrap();
        let url2 = Url::parse("http://example.com/page2").unwrap();
        let url3 = Url::parse("http://example.com/page3").unwrap();
        
        assert!(queue.add_url(url1, 1).await);
        assert!(queue.add_url(url2, 2).await);
        assert!(!queue.add_url(url3, 3).await);
        
        assert_eq!(queue.len().await, 2);
    }

    #[tokio::test]
    async fn test_is_visited() {
        let queue = UrlQueue::new(3, 2);
        
        let url = Url::parse("http://example.com/page").unwrap();
        
        assert!(!queue.is_visited(&url).await);
        queue.add_url(url.clone(), 1).await;
        assert!(queue.is_visited(&url).await);
    }

    #[tokio::test]
    async fn test_get_domain() {
        let queue = UrlQueue::new(3, 2);
        
        let url1 = Url::parse("http://example.com/page").unwrap();
        assert_eq!(queue.get_domain(&url1), "example.com");
        
        let url2 = Url::parse("https://sub.example.org/path").unwrap();
        assert_eq!(queue.get_domain(&url2), "sub.example.org");
    }
}
