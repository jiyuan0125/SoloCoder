use std::sync::Arc;
use std::time::Duration;
use tokio::sync::Semaphore;
use tokio::time::sleep;
use chrono::Utc;
use reqwest::Client;
use url::Url;
use tracing::{info, warn, error};

use crate::models::TaskStatus;
use crate::store::TaskStore;
use crate::persistence::PersistenceManager;

const DEFAULT_CONCURRENCY: usize = 100;
const CLEANUP_INTERVAL_SECS: u64 = 60;
const TASK_TTL_HOURS: i64 = 1;
const PERSISTENCE_INTERVAL_SECS: u64 = 5;

pub struct Scheduler {
    store: TaskStore,
    concurrency_limit: usize,
    client: Client,
    persistence: Option<Arc<PersistenceManager>>,
}

impl Scheduler {
    pub fn new(
        store: TaskStore,
        concurrency_limit: Option<usize>,
        persistence: Option<Arc<PersistenceManager>>,
    ) -> Self {
        Self {
            store,
            concurrency_limit: concurrency_limit.unwrap_or(DEFAULT_CONCURRENCY),
            client: Client::new(),
            persistence,
        }
    }

    pub fn validate_url(url_str: &str) -> bool {
        if url_str.is_empty() {
            return false;
        }
        Url::parse(url_str).is_ok()
    }

    pub async fn run(self: Arc<Self>) {
        let semaphore = Arc::new(Semaphore::new(self.concurrency_limit));
        
        let scheduler_self = self.clone();
        tokio::spawn(async move {
            scheduler_self.run_cleanup_loop().await;
        });

        let scheduler_self = self.clone();
        tokio::spawn(async move {
            scheduler_self.run_persistence_loop().await;
        });

        let scheduler_self = self.clone();
        let sem_pending = semaphore.clone();
        tokio::spawn(async move {
            scheduler_self.run_pending_loop(sem_pending).await;
        });

        let scheduler_self = self.clone();
        let sem_retry = semaphore.clone();
        tokio::spawn(async move {
            scheduler_self.run_retry_loop(sem_retry).await;
        });
    }

    async fn run_cleanup_loop(&self) {
        loop {
            sleep(Duration::from_secs(CLEANUP_INTERVAL_SECS)).await;
            let cutoff = Utc::now() - chrono::Duration::hours(TASK_TTL_HOURS);
            let removed = self.store.cleanup_expired(cutoff);
            if removed > 0 {
                info!("Cleaned up {} expired tasks", removed);
            }
        }
    }

    async fn run_persistence_loop(&self) {
        loop {
            sleep(Duration::from_secs(PERSISTENCE_INTERVAL_SECS)).await;
            
            if let Some(persistence) = &self.persistence {
                let tasks = self.store.all_tasks();
                if let Err(e) = persistence.save(&tasks) {
                    error!("Failed to persist tasks: {}", e);
                }
            }
        }
    }

    async fn run_pending_loop(self: Arc<Self>, semaphore: Arc<Semaphore>) {
        loop {
            let permit = semaphore.clone().acquire_owned().await;
            match permit {
                Ok(permit) => {
                    if let Some(task_id) = self.store.pop_pending() {
                        let scheduler = self.clone();
                        tokio::spawn(async move {
                            scheduler.execute_task(task_id).await;
                            drop(permit);
                        });
                    } else {
                        drop(permit);
                        sleep(Duration::from_millis(100)).await;
                    }
                }
                Err(e) => {
                    error!("Semaphore error: {}", e);
                    sleep(Duration::from_millis(100)).await;
                }
            }
        }
    }

    async fn run_retry_loop(self: Arc<Self>, semaphore: Arc<Semaphore>) {
        loop {
            if let Some(task_id) = self.store.pop_retry() {
                if let Some(task) = self.store.get(&task_id) {
                    let delay = task.get_retry_delay_ms();
                    let permit = semaphore.clone().acquire_owned().await;
                    match permit {
                        Ok(permit) => {
                            sleep(Duration::from_millis(delay)).await;
                            let scheduler = self.clone();
                            tokio::spawn(async move {
                                scheduler.execute_task(task_id).await;
                                drop(permit);
                            });
                        }
                        Err(e) => {
                            error!("Semaphore error: {}", e);
                            self.store.push_retry(task_id);
                            sleep(Duration::from_millis(100)).await;
                        }
                    }
                }
            } else {
                sleep(Duration::from_millis(100)).await;
            }
        }
    }

    async fn execute_task(&self, task_id: uuid::Uuid) {
        let mut task = match self.store.get(&task_id) {
            Some(t) => t,
            None => return,
        };

        if !Self::validate_url(&task.url) {
            warn!("Invalid URL for task {}: {}", task_id, task.url);
            self.store.update(&task_id, |t| {
                t.status = TaskStatus::Failed;
                t.last_error = Some("Invalid URL".to_string());
                t.completed_at = Some(Utc::now());
            });
            return;
        }

        if task.status == TaskStatus::Retrying {
            self.store.update(&task_id, |t| {
                t.status = TaskStatus::Running;
                t.current_attempt += 1;
            });
        } else {
            self.store.update(&task_id, |t| {
                t.status = TaskStatus::Running;
                t.current_attempt = 1;
                t.started_at = Some(Utc::now());
            });
        }

        self.store.update(&task_id, |t| {
            t.execution_started_at = Some(Utc::now());
        });

        task = self.store.get(&task_id).unwrap();

        let timeout = Duration::from_secs(task.http_timeout_seconds);
        let url = task.url.clone();
        let body = task.body.clone();

        let result = tokio::time::timeout(
            timeout,
            self.make_request(&url, body),
        ).await;

        match result {
            Ok(Ok((status, resp_body))) => {
                if status.is_success() {
                    info!("Task {} succeeded with status {}", task_id, status);
                    self.store.update(&task_id, |t| {
                        t.status = TaskStatus::Succeeded;
                        t.response_status = Some(status.as_u16());
                        t.response_body = resp_body;
                        t.completed_at = Some(Utc::now());
                    });
                } else {
                    warn!("Task {} failed with status {}", task_id, status);
                    self.handle_failure(task_id, format!("HTTP error: {}", status), Some(status.as_u16()), resp_body).await;
                }
            }
            Ok(Err(e)) => {
                warn!("Task {} request error: {}", task_id, e);
                self.handle_failure(task_id, e, None, None).await;
            }
            Err(_) => {
                warn!("Task {} timeout after {}s", task_id, task.http_timeout_seconds);
                self.handle_failure(task_id, "Request timeout".to_string(), None, None).await;
            }
        }
    }

    async fn handle_failure(&self, task_id: uuid::Uuid, error: String, status: Option<u16>, body: Option<serde_json::Value>) {
        let task = match self.store.get(&task_id) {
            Some(t) => t,
            None => return,
        };

        let should_retry = task.current_attempt <= task.max_retries;

        if should_retry {
            info!("Task {} will be retried (attempt {} of {})", task_id, task.current_attempt, task.max_retries);
            self.store.update(&task_id, |t| {
                t.status = TaskStatus::Retrying;
                t.last_error = Some(error);
                t.response_status = status;
                t.response_body = body;
            });
            self.store.push_retry(task_id);
        } else {
            error!("Task {} failed after {} attempts", task_id, task.current_attempt);
            self.store.update(&task_id, |t| {
                t.status = TaskStatus::Failed;
                t.last_error = Some(error);
                t.response_status = status;
                t.response_body = body;
                t.completed_at = Some(Utc::now());
            });
        }
    }

    async fn make_request(
        &self,
        url: &str,
        body: Option<serde_json::Value>,
    ) -> Result<(reqwest::StatusCode, Option<serde_json::Value>), String> {
        let request = if let Some(b) = body {
            self.client.post(url).json(&b)
        } else {
            self.client.post(url)
        };

        let response = request
            .send()
            .await
            .map_err(|e| e.to_string())?;

        let status = response.status();
        let resp_body: Option<serde_json::Value> = response
            .json()
            .await
            .ok();

        Ok((status, resp_body))
    }
}
