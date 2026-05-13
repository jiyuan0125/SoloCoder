use actix_web::{
    dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform},
    web, App, Error, HttpResponse, HttpServer, Responder,
};
use futures::future::LocalBoxFuture;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::future::{ready, Ready};
use std::sync::Mutex;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KeyConfig {
    pub burst: u64,
    pub refill_rate: u64,
}

#[derive(Debug, Clone)]
pub struct TokenBucket {
    pub config: KeyConfig,
    pub tokens: u64,
    pub last_refill: Instant,
    pub throttle_count: u64,
}

impl TokenBucket {
    pub fn new(config: KeyConfig) -> Self {
        let burst = config.burst;
        Self {
            config,
            tokens: burst,
            last_refill: Instant::now(),
            throttle_count: 0,
        }
    }

    pub fn refill(&mut self) {
        let now = Instant::now();
        let elapsed = now.duration_since(self.last_refill).as_secs();
        if elapsed > 0 {
            let new_tokens = elapsed.saturating_mul(self.config.refill_rate);
            self.tokens = std::cmp::min(self.tokens.saturating_add(new_tokens), self.config.burst);
            self.last_refill = now;
        }
    }

    pub fn try_consume(&mut self) -> Result<(), u64> {
        self.refill();
        if self.tokens > 0 {
            self.tokens -= 1;
            Ok(())
        } else {
            self.throttle_count = self.throttle_count.saturating_add(1);
            let wait = if self.config.refill_rate > 0 {
                (self.config.burst as f64 / self.config.refill_rate as f64).ceil() as u64
            } else {
                60
            };
            Err(std::cmp::max(wait, 1))
        }
    }

    pub fn seconds_since_last_refill(&self) -> u64 {
        Instant::now().duration_since(self.last_refill).as_secs()
    }
}

#[derive(Debug, Clone)]
pub struct CooldownRecord {
    pub ip: String,
    pub start_time: Instant,
    pub duration: Duration,
}

impl CooldownRecord {
    pub fn is_expired(&self) -> bool {
        Instant::now().duration_since(self.start_time) >= self.duration
    }

    pub fn remaining_seconds(&self) -> u64 {
        let elapsed = Instant::now().duration_since(self.start_time);
        if elapsed >= self.duration {
            0
        } else {
            self.duration.saturating_sub(elapsed).as_secs()
        }
    }

    pub fn trigger_timestamp(&self) -> u64 {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();
        now.saturating_sub(self.start_time.elapsed().as_secs())
    }
}

#[derive(Debug, Clone)]
pub struct IpThrottleTracker {
    pub keys: HashMap<String, Instant>,
}

impl IpThrottleTracker {
    pub fn new() -> Self {
        Self {
            keys: HashMap::new(),
        }
    }

    pub fn add_throttle(&mut self, key: &str) {
        self.keys.insert(key.to_string(), Instant::now());
    }

    pub fn active_throttle_count(&self) -> usize {
        self.keys
            .values()
            .filter(|t| t.elapsed().as_secs() < 60)
            .count()
    }

    pub fn cleanup(&mut self) {
        let now = Instant::now();
        self.keys.retain(|_, t| now.duration_since(*t).as_secs() < 60);
    }
}

pub struct AppStateInner {
    pub keys: HashMap<String, TokenBucket>,
    pub cooldowns: HashMap<String, CooldownRecord>,
    pub ip_throttles: HashMap<String, IpThrottleTracker>,
}

pub type AppState = web::Data<Mutex<AppStateInner>>;

impl AppStateInner {
    pub fn new() -> Self {
        Self {
            keys: HashMap::new(),
            cooldowns: HashMap::new(),
            ip_throttles: HashMap::new(),
        }
    }

    pub fn cleanup_expired_cooldowns(&mut self) {
        self.cooldowns.retain(|_, c| !c.is_expired());
    }

    pub fn add_cooldown(&mut self, ip: String) {
        self.cooldowns.insert(
            ip.clone(),
            CooldownRecord {
                ip: ip.clone(),
                start_time: Instant::now(),
                duration: Duration::from_secs(60),
            },
        );
        self.ip_throttles.remove(&ip);
    }

    pub fn is_in_cooldown(&mut self, ip: &str) -> Option<u64> {
        self.cleanup_expired_cooldowns();
        self.cooldowns.get(ip).map(|c| c.remaining_seconds())
    }

    pub fn record_throttle_for_ip(&mut self, ip: &str, key: &str) -> bool {
        let tracker = self
            .ip_throttles
            .entry(ip.to_string())
            .or_insert_with(IpThrottleTracker::new);
        tracker.cleanup();
        tracker.add_throttle(key);
        tracker.active_throttle_count() >= 3
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct KeyStatsResponse {
    pub remaining_tokens: u64,
    pub last_refill_seconds_ago: u64,
    pub total_throttled: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct CooldownInfo {
    pub ip: String,
    pub remaining_seconds: u64,
    pub trigger_timestamp: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct CooldownsResponse {
    pub cooldowns: Vec<CooldownInfo>,
}

pub struct RateLimitMiddleware;

impl<S, B> Transform<S, ServiceRequest> for RateLimitMiddleware
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error> + 'static,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type InitError = ();
    type Transform = RateLimitMiddlewareService<S>;
    type Future = Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        ready(Ok(RateLimitMiddlewareService { service }))
    }
}

pub struct RateLimitMiddlewareService<S> {
    service: S,
}

fn get_client_ip(req: &ServiceRequest) -> String {
    if let Some(forwarded) = req.headers().get("x-forwarded-for") {
        if let Ok(s) = forwarded.to_str() {
            return s.split(',').next().unwrap_or("").trim().to_string();
        }
    }
    if let Some(real_ip) = req.headers().get("x-real-ip") {
        if let Ok(s) = real_ip.to_str() {
            return s.to_string();
        }
    }
    req.peer_addr()
        .map(|addr| addr.ip().to_string())
        .unwrap_or_else(|| "unknown".to_string())
}

fn get_api_key(req: &ServiceRequest) -> Option<String> {
    if let Some(auth) = req.headers().get("x-api-key") {
        auth.to_str().ok().map(|s| s.to_string())
    } else if let Some(auth) = req.headers().get("authorization") {
        auth.to_str()
            .ok()
            .and_then(|s| s.strip_prefix("Bearer "))
            .map(|s| s.to_string())
    } else {
        None
    }
}

impl<S, B> Service<ServiceRequest> for RateLimitMiddlewareService<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error> + 'static,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;

    forward_ready!(service);

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let path = req.path().to_string();
        let _method = req.method().clone();

        let is_management = path.starts_with("/keys") || path == "/cooldowns";

        if is_management {
            let fut = self.service.call(req);
            return Box::pin(async move { fut.await });
        }

        let api_key = match get_api_key(&req) {
            Some(k) => k,
            None => {
                let (_req, _) = req.into_parts();
                return Box::pin(async move {
                    Err(actix_web::error::InternalError::from_response(
                        "",
                        HttpResponse::Unauthorized().body("Missing API Key"),
                    )
                    .into())
                });
            }
        };

        let client_ip = get_client_ip(&req);

        let state = match req.app_data::<AppState>() {
            Some(s) => s.clone(),
            None => {
                let (_req, _) = req.into_parts();
                return Box::pin(async move {
                    Err(actix_web::error::InternalError::from_response(
                        "",
                        HttpResponse::InternalServerError().body("Internal error"),
                    )
                    .into())
                });
            }
        };

        let cooldown_remaining = {
            let mut inner = state.lock().unwrap();
            inner.is_in_cooldown(&client_ip)
        };

        if let Some(remaining) = cooldown_remaining {
            let (_req, _) = req.into_parts();
            return Box::pin(async move {
                Err(actix_web::error::InternalError::from_response(
                    "",
                    HttpResponse::TooManyRequests()
                        .insert_header(("Retry-After", remaining.to_string()))
                        .body("IP is in cooldown"),
                )
                .into())
            });
        }

        let (result, should_cooldown) = {
            let mut inner = state.lock().unwrap();
            match inner.keys.get_mut(&api_key) {
                Some(bucket) => {
                    if let Err(wait) = bucket.try_consume() {
                        let need_cooldown = inner.record_throttle_for_ip(&client_ip, &api_key);
                        (Err(wait), need_cooldown)
                    } else {
                        (Ok(()), false)
                    }
                }
                None => (Err(401u64), false),
            }
        };

        match result {
            Ok(_) => {
                let fut = self.service.call(req);
                Box::pin(async move { fut.await })
            }
            Err(401) => {
                let (_req, _) = req.into_parts();
                Box::pin(async move {
                    Err(actix_web::error::InternalError::from_response(
                        "",
                        HttpResponse::Unauthorized().body("Unknown API Key"),
                    )
                    .into())
                })
            }
            Err(wait) => {
                if should_cooldown {
                    let mut inner = state.lock().unwrap();
                    inner.add_cooldown(client_ip.clone());
                }
                let (_req, _) = req.into_parts();
                Box::pin(async move {
                    Err(actix_web::error::InternalError::from_response(
                        "",
                        HttpResponse::TooManyRequests()
                            .insert_header(("Retry-After", wait.to_string()))
                            .body("Too Many Requests"),
                    )
                    .into())
                })
            }
        }
    }
}

pub async fn create_key(
    state: AppState,
    key: web::Json<KeyConfig>,
) -> impl Responder {
    let key_id = format!("key-{}", uuid_v4());
    let mut inner = state.lock().unwrap();
    let bucket = TokenBucket::new(key.into_inner());
    inner.keys.insert(key_id.clone(), bucket);
    HttpResponse::Created().json(serde_json::json!({ "key_id": key_id }))
}

pub async fn delete_key(state: AppState, path: web::Path<String>) -> impl Responder {
    let key_id = path.into_inner();
    let mut inner = state.lock().unwrap();
    if inner.keys.remove(&key_id).is_some() {
        HttpResponse::Ok().body("Key deleted")
    } else {
        HttpResponse::NotFound().body("Key not found")
    }
}

pub async fn update_config(
    state: AppState,
    path: web::Path<String>,
    config: web::Json<KeyConfig>,
) -> impl Responder {
    let key_id = path.into_inner();
    let mut inner = state.lock().unwrap();
    match inner.keys.get_mut(&key_id) {
        Some(bucket) => {
            bucket.config = config.into_inner();
            bucket.tokens = std::cmp::min(bucket.tokens, bucket.config.burst);
            HttpResponse::Ok().json(serde_json::json!({ "status": "updated" }))
        }
        None => HttpResponse::NotFound().body("Key not found"),
    }
}

pub async fn get_key_stats(state: AppState, path: web::Path<String>) -> impl Responder {
    let key_id = path.into_inner();
    let mut inner = state.lock().unwrap();
    match inner.keys.get_mut(&key_id) {
        Some(bucket) => {
            bucket.refill();
            let response = KeyStatsResponse {
                remaining_tokens: bucket.tokens,
                last_refill_seconds_ago: bucket.seconds_since_last_refill(),
                total_throttled: bucket.throttle_count,
            };
            HttpResponse::Ok().json(response)
        }
        None => HttpResponse::NotFound().body("Key not found"),
    }
}

pub async fn get_cooldowns(state: AppState) -> impl Responder {
    let mut inner = state.lock().unwrap();
    inner.cleanup_expired_cooldowns();

    let cooldowns: Vec<CooldownInfo> = inner
        .cooldowns
        .values()
        .map(|c| CooldownInfo {
            ip: c.ip.clone(),
            remaining_seconds: c.remaining_seconds(),
            trigger_timestamp: c.trigger_timestamp(),
        })
        .collect();

    HttpResponse::Ok().json(CooldownsResponse { cooldowns })
}

pub async fn protected_endpoint() -> impl Responder {
    HttpResponse::Ok().json(serde_json::json!({ "status": "ok", "message": "Request successful" }))
}

fn uuid_v4() -> String {
    use rand::Rng;
    let mut rng = rand::thread_rng();
    let bytes: [u8; 16] = rng.gen();
    format!(
        "{:02x}{:02x}{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}{:02x}{:02x}{:02x}{:02x}",
        bytes[0], bytes[1], bytes[2], bytes[3],
        bytes[4], bytes[5],
        bytes[6], bytes[7],
        bytes[8], bytes[9],
        bytes[10], bytes[11], bytes[12], bytes[13], bytes[14], bytes[15]
    )
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8900".to_string())
        .parse::<u16>()
        .unwrap_or(8900);

    let app_state = web::Data::new(Mutex::new(AppStateInner::new()));

    println!("Starting rate limiter server on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .wrap(RateLimitMiddleware)
            .route("/keys", web::post().to(create_key))
            .route("/keys/{kid}", web::delete().to(delete_key))
            .route("/keys/{kid}/config", web::put().to(update_config))
            .route("/keys/{kid}/stats", web::get().to(get_key_stats))
            .route("/cooldowns", web::get().to(get_cooldowns))
            .route("/api/protected", web::get().to(protected_endpoint))
            .route("/api/protected", web::post().to(protected_endpoint))
            .default_service(web::to(|| async {
                HttpResponse::NotFound().body("Not found")
            }))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
