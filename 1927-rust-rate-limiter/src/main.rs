use actix_web::{web, App, HttpResponse, HttpServer, Responder, middleware::Logger};
use chrono::{DateTime, Utc};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::time::{self, Duration};
use uuid::Uuid;

const DEFAULT_SYNC_INTERVAL: u64 = 5;
const MAX_MISSING_PINGS: u64 = 3;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceInfo {
    pub address: String,
    pub last_seen: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KeyConfig {
    pub rate_per_second: u64,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncPayload {
    pub instance_id: String,
    pub counters: HashMap<String, u64>,
    pub window_start: DateTime<Utc>,
    pub configs: HashMap<String, KeyConfig>,
    pub instances: HashMap<String, InstanceInfo>,
}

#[derive(Debug, Clone)]
pub struct RemoteCounter {
    pub count: u64,
    pub window_start: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone)]
pub struct SmoothDistribution {
    pub remaining: u64,
    pub per_sync: u64,
}

pub struct AppState {
    pub self_address: String,
    pub instance_id: String,
    pub sync_interval: Duration,
    pub instances: Arc<RwLock<HashMap<String, InstanceInfo>>>,
    pub local_counters: Arc<RwLock<HashMap<String, u64>>>,
    pub remote_counters: Arc<RwLock<HashMap<String, HashMap<String, RemoteCounter>>>>,
    pub window_start: Arc<RwLock<DateTime<Utc>>>,
    pub configs: Arc<RwLock<HashMap<String, KeyConfig>>>,
    pub throttled_count: Arc<RwLock<u64>>,
    pub http_client: reqwest::Client,
    pub smooth_distribution: Arc<RwLock<HashMap<String, SmoothDistribution>>>,
}

impl AppState {
    pub fn new(address: String, sync_interval_secs: u64) -> Self {
        AppState {
            self_address: address.clone(),
            instance_id: Uuid::new_v4().to_string(),
            sync_interval: Duration::from_secs(sync_interval_secs),
            instances: Arc::new(RwLock::new(HashMap::new())),
            local_counters: Arc::new(RwLock::new(HashMap::new())),
            remote_counters: Arc::new(RwLock::new(HashMap::new())),
            window_start: Arc::new(RwLock::new(Utc::now())),
            configs: Arc::new(RwLock::new(HashMap::new())),
            throttled_count: Arc::new(RwLock::new(0)),
            http_client: reqwest::Client::builder()
                .timeout(Duration::from_secs(5))
                .build()
                .expect("Failed to build HTTP client"),
            smooth_distribution: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn get_global_count(&self, key: &str) -> u64 {
        let local = *self.local_counters.read().get(key).unwrap_or(&0);
        let mut total = local;
        
        let remote = self.remote_counters.read();
        for counters in remote.values() {
            if let Some(counter) = counters.get(key) {
                total += counter.count;
            }
        }
        
        let smooth = self.smooth_distribution.read();
        if let Some(dist) = smooth.get(key) {
            total += dist.remaining;
        }
        
        total
    }

    pub fn allow_request(&self, key: &str) -> bool {
        let configs = self.configs.read();
        let config = match configs.get(key) {
            Some(c) => c.clone(),
            None => return false,
        };
        
        let window_start = *self.window_start.read();
        let now = Utc::now();
        let window_elapsed = (now - window_start).num_seconds() as f64;
        
        if window_elapsed >= self.sync_interval.as_secs_f64() {
            let mut start = self.window_start.write();
            *start = now;
            let mut counters = self.local_counters.write();
            counters.clear();
            let mut remote = self.remote_counters.write();
            remote.clear();
        }
        
        let limit = (config.rate_per_second as f64 * self.sync_interval.as_secs_f64()) as u64;
        let global_count = self.get_global_count(key);
        
        if global_count >= limit {
            *self.throttled_count.write() += 1;
            return false;
        }
        
        *self.local_counters.write().entry(key.to_string()).or_insert(0) += 1;
        true
    }

    pub fn estimate_global_qps(&self) -> f64 {
        let now = Utc::now();
        let window_start = *self.window_start.read();
        let elapsed = (now - window_start).num_seconds() as f64;
        if elapsed <= 0.0 {
            return 0.0;
        }
        
        let mut total_requests = 0;
        let local = self.local_counters.read();
        for &count in local.values() {
            total_requests += count;
        }
        
        let remote = self.remote_counters.read();
        for counters in remote.values() {
            for counter in counters.values() {
                total_requests += counter.count;
            }
        }
        
        total_requests as f64 / elapsed
    }

    pub fn process_sync(&self, payload: SyncPayload) {
        let mut instances = self.instances.write();
        for (id, info) in payload.instances {
            if id == self.instance_id {
                continue;
            }
            instances.entry(id.clone())
                .and_modify(|existing| {
                    if info.last_seen > existing.last_seen {
                        *existing = info.clone();
                    }
                })
                .or_insert(info);
        }
        
        if payload.instance_id != self.instance_id {
            let mut remote = self.remote_counters.write();
            let key_counters: HashMap<String, RemoteCounter> = payload.counters
                .into_iter()
                .map(|(k, v)| (k, RemoteCounter {
                    count: v,
                    window_start: payload.window_start,
                    updated_at: Utc::now(),
                }))
                .collect();
            remote.insert(payload.instance_id.clone(), key_counters);
            
            instances.entry(payload.instance_id.clone())
                .and_modify(|existing| {
                    existing.last_seen = Utc::now();
                })
                .or_insert(InstanceInfo {
                    address: payload.instance_id.clone(),
                    last_seen: Utc::now(),
                });
        }
        
        let mut configs = self.configs.write();
        for (key, incoming_config) in payload.configs {
            configs.entry(key)
                .and_modify(|existing| {
                    if incoming_config.updated_at > existing.updated_at {
                        *existing = incoming_config.clone();
                    }
                })
                .or_insert(incoming_config);
        }
    }

    pub fn apply_smooth_distribution(&self) {
        let mut smooth = self.smooth_distribution.write();
        let keys: Vec<String> = smooth.keys().cloned().collect();
        
        for key in keys {
            if let Some(dist) = smooth.get_mut(&key) {
                if dist.remaining <= dist.per_sync {
                    dist.per_sync = dist.remaining;
                    dist.remaining = 0;
                } else {
                    dist.remaining -= dist.per_sync;
                }
            }
            if let Some(dist) = smooth.get(&key) {
                if dist.remaining == 0 {
                    smooth.remove(&key);
                }
            }
        }
    }

    pub fn check_offline_instances(&self) {
        let now = Utc::now();
        let max_age = self.sync_interval * MAX_MISSING_PINGS as u32;
        
        let mut instances = self.instances.write();
        let mut remote = self.remote_counters.write();
        
        let to_remove: Vec<String> = instances.iter()
            .filter(|(_, info)| (now - info.last_seen).num_seconds() as u64 >= max_age.as_secs())
            .map(|(id, _)| id.clone())
            .collect();
        
        let active_count = instances.len().saturating_sub(to_remove.len());
        let sync_periods_for_smooth = 5u64;
        
        for id in to_remove {
            instances.remove(&id);
            if let Some(counters) = remote.remove(&id) {
                let mut smooth = self.smooth_distribution.write();
                for (key, counter) in counters {
                    let total_share = counter.count;
                    if total_share > 0 && active_count > 0 {
                        let per_sync = total_share / sync_periods_for_smooth.max(1);
                        let entry = smooth.entry(key.clone()).or_insert(SmoothDistribution {
                            remaining: 0,
                            per_sync: 0,
                        });
                        entry.remaining += total_share;
                        entry.per_sync = entry.per_sync.max(per_sync);
                    }
                }
            }
        }
    }

    pub fn create_sync_payload(&self) -> SyncPayload {
        SyncPayload {
            instance_id: self.instance_id.clone(),
            counters: self.local_counters.read().clone(),
            window_start: *self.window_start.read(),
            configs: self.configs.read().clone(),
            instances: self.instances.read().clone(),
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct RegisterInstanceRequest {
    pub address: String,
}

#[derive(Debug, Deserialize)]
pub struct UpdateConfigRequest {
    pub rate_per_second: u64,
}

#[derive(Debug, Serialize)]
pub struct StatsResponse {
    pub global_qps_estimate: f64,
    pub throttled_count: u64,
    pub known_instances: Vec<String>,
    pub instance_id: String,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

async fn register_instance(
    req: web::Json<RegisterInstanceRequest>,
    state: web::Data<AppState>,
) -> impl Responder {
    let address = req.address.clone();
    
    {
        let mut instances = state.instances.write();
        instances.insert(address.clone(), InstanceInfo {
            address: address.clone(),
            last_seen: Utc::now(),
        });
    }
    
    let client = state.http_client.clone();
    let sync_url = format!("{}/sync", address);
    let payload = state.create_sync_payload();
    
    let state_clone = state.clone();
    tokio::spawn(async move {
        if let Ok(response) = client.post(&sync_url)
            .json(&payload)
            .send()
            .await
        {
            if let Ok(sync_data) = response.json::<SyncPayload>().await {
                state_clone.process_sync(sync_data);
            }
        }
    });
    
    HttpResponse::Ok().json(serde_json::json!({
        "status": "registered",
        "address": address
    }))
}

async fn update_key_config(
    kid: web::Path<String>,
    req: web::Json<UpdateConfigRequest>,
    state: web::Data<AppState>,
) -> impl Responder {
    let key_id = kid.into_inner();
    let config = KeyConfig {
        rate_per_second: req.rate_per_second,
        updated_at: Utc::now(),
    };
    
    {
        let mut configs = state.configs.write();
        configs.insert(key_id.clone(), config);
    }
    
    HttpResponse::Ok().json(serde_json::json!({
        "key": key_id,
        "rate_per_second": req.rate_per_second
    }))
}

async fn check_rate_limit(
    query: web::Query<HashMap<String, String>>,
    state: web::Data<AppState>,
) -> impl Responder {
    let key = match query.get("key") {
        Some(k) => k.clone(),
        None => return HttpResponse::BadRequest().json(ErrorResponse {
            error: "Missing 'key' query parameter".to_string()
        }),
    };
    
    {
        let configs = state.configs.read();
        if !configs.contains_key(&key) {
            return HttpResponse::NotFound().json(ErrorResponse {
                error: format!("Key '{}' not found", key)
            });
        }
    }
    
    if state.allow_request(&key) {
        HttpResponse::Ok().json(serde_json::json!({
            "status": "allowed",
            "key": key
        }))
    } else {
        HttpResponse::TooManyRequests().json(ErrorResponse {
            error: "Rate limit exceeded".to_string()
        })
    }
}

async fn sync(
    payload: web::Json<SyncPayload>,
    state: web::Data<AppState>,
) -> impl Responder {
    state.process_sync(payload.into_inner());
    
    let response_payload = state.create_sync_payload();
    HttpResponse::Ok().json(response_payload)
}

async fn get_stats(state: web::Data<AppState>) -> impl Responder {
    let instances: Vec<String> = state.instances.read().values().map(|i| i.address.clone()).collect();
    
    HttpResponse::Ok().json(StatsResponse {
        global_qps_estimate: state.estimate_global_qps(),
        throttled_count: *state.throttled_count.read(),
        known_instances: instances,
        instance_id: state.instance_id.clone(),
    })
}

async fn sync_task(state: web::Data<AppState>) {
    let mut interval = time::interval(state.sync_interval);
    
    loop {
        interval.tick().await;
        
        state.apply_smooth_distribution();
        state.check_offline_instances();
        
        let payload = state.create_sync_payload();
        let instances: Vec<String> = {
            state.instances.read().values()
                .map(|info| info.address.clone())
                .collect()
        };
        
        let client = state.http_client.clone();
        let state_clone = state.clone();
        
        for addr in instances {
            let url = format!("{}/sync", addr);
            let p = payload.clone();
            let c = client.clone();
            let sc = state_clone.clone();
            
            tokio::spawn(async move {
                if let Ok(response) = c.post(&url).json(&p).send().await {
                    if let Ok(sync_data) = response.json::<SyncPayload>().await {
                        sc.process_sync(sync_data);
                    }
                }
            });
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    
    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let sync_interval = std::env::var("SYNC_INTERVAL")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(DEFAULT_SYNC_INTERVAL);
    
    let self_address = format!("http://127.0.0.1:{}", port);
    let state = web::Data::new(AppState::new(self_address.clone(), sync_interval));
    
    let sync_state = state.clone();
    tokio::spawn(async move {
        sync_task(sync_state).await;
    });
    
    println!("Starting distributed rate limiter on {}", self_address);
    println!("Instance ID: {}", state.instance_id);
    
    HttpServer::new(move || {
        App::new()
            .wrap(Logger::default())
            .app_data(state.clone())
            .route("/instances", web::post().to(register_instance))
            .route("/keys/{kid}/config", web::put().to(update_key_config))
            .route("/stats", web::get().to(get_stats))
            .route("/sync", web::post().to(sync))
            .route("/check", web::get().to(check_rate_limit))
    })
    .bind(format!("0.0.0.0:{}", port))?
    .run()
    .await
}
