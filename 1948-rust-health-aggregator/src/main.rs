use actix_web::{web, App, HttpServer, Responder, HttpResponse};
use chrono::{DateTime, Utc};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum HealthStatus {
    Healthy,
    Unhealthy,
}

impl Default for HealthStatus {
    fn default() -> Self {
        HealthStatus::Unhealthy
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackendConfig {
    pub name: String,
    pub health_url: String,
    #[serde(with = "serde_humantime")]
    pub interval: Duration,
    #[serde(with = "serde_humantime")]
    pub timeout: Duration,
}

#[derive(Debug, Clone)]
pub struct BackendState {
    pub config: BackendConfig,
    pub current_status: HealthStatus,
    pub last_check_time: Option<DateTime<Utc>>,
    pub consecutive_failures: u32,
}

impl BackendState {
    fn new(config: BackendConfig) -> Self {
        Self {
            config,
            current_status: HealthStatus::Unhealthy,
            last_check_time: None,
            consecutive_failures: 0,
        }
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct BackendStatusInfo {
    pub name: String,
    pub status: HealthStatus,
    #[serde(rename = "lastCheckTime")]
    pub last_check_time: Option<DateTime<Utc>>,
    #[serde(rename = "consecutiveFailures")]
    pub consecutive_failures: u32,
}

#[derive(Debug, Clone, Serialize)]
pub struct AggregatedHealthResponse {
    pub status: String,
    pub backends: Vec<BackendStatusInfo>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct SubscriptionRequest {
    #[serde(rename = "callbackUrl")]
    pub callback_url: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct StatusChangeNotification {
    pub backend: String,
    #[serde(rename = "oldStatus")]
    pub old_status: HealthStatus,
    #[serde(rename = "newStatus")]
    pub new_status: HealthStatus,
    #[serde(rename = "changedAt")]
    pub changed_at: DateTime<Utc>,
}

pub type Subscriptions = HashMap<Uuid, String>;

pub struct AppState {
    pub backends: RwLock<HashMap<String, BackendState>>,
    pub subscriptions: RwLock<Subscriptions>,
    pub http_client: reqwest::Client,
}

async fn check_backend_health(
    http_client: &reqwest::Client,
    config: &BackendConfig,
) -> (HealthStatus, Option<DateTime<Utc>>) {
    let now = Utc::now();
    let result = http_client
        .get(&config.health_url)
        .timeout(config.timeout)
        .send()
        .await;

    match result {
        Ok(response) => {
            if response.status().is_success() {
                match response.json::<serde_json::Value>().await {
                    Ok(json) => {
                        if let Some(status) = json.get("status") {
                            if let Some(status_str) = status.as_str() {
                                if status_str == "healthy" {
                                    return (HealthStatus::Healthy, Some(now));
                                }
                            }
                        }
                        (HealthStatus::Unhealthy, Some(now))
                    }
                    Err(_) => (HealthStatus::Unhealthy, Some(now)),
                }
            } else {
                (HealthStatus::Unhealthy, Some(now))
            }
        }
        Err(_) => (HealthStatus::Unhealthy, Some(now)),
    }
}

async fn notify_subscribers(
    app_state: Arc<AppState>,
    notification: &StatusChangeNotification,
) {
    let callbacks: Vec<String> = {
        let subs = app_state.subscriptions.read();
        subs.values().cloned().collect()
    };

    for callback_url in callbacks {
        let app_state_clone = app_state.clone();
        let notification_clone = notification.clone();
        tokio::spawn(async move {
            let retry_delays = [Duration::from_secs(1), Duration::from_secs(3)];
            let mut attempts = 0;
            loop {
                let result = app_state_clone
                    .http_client
                    .post(&callback_url)
                    .timeout(Duration::from_secs(10))
                    .json(&notification_clone)
                    .send()
                    .await;

                match result {
                    Ok(response) => {
                        if response.status().is_success() {
                            break;
                        }
                    }
                    Err(_) => {}
                }

                if attempts >= retry_delays.len() {
                    break;
                }
                tokio::time::sleep(retry_delays[attempts]).await;
                attempts += 1;
            }
        });
    }
}

async fn health_check_worker(app_state: Arc<AppState>, backend_name: String) {
    let interval = {
        let backends = app_state.backends.read();
        backends
            .get(&backend_name)
            .map(|b| b.config.interval)
            .unwrap_or_else(|| Duration::from_secs(10))
    };

    let mut interval = tokio::time::interval(interval);

    loop {
        interval.tick().await;

        let (config, old_status) = {
            let backends = app_state.backends.read();
            match backends.get(&backend_name) {
                Some(state) => (state.config.clone(), state.current_status),
                None => return,
            }
        };

        let (new_status, check_time) = check_backend_health(&app_state.http_client, &config).await;

        {
            let mut backends = app_state.backends.write();
            if let Some(state) = backends.get_mut(&backend_name) {
                state.last_check_time = check_time;
                if new_status == HealthStatus::Healthy {
                    state.consecutive_failures = 0;
                } else {
                    state.consecutive_failures += 1;
                }

                if old_status != new_status {
                    state.current_status = new_status;
                    let notification = StatusChangeNotification {
                        backend: backend_name.clone(),
                        old_status,
                        new_status,
                        changed_at: Utc::now(),
                    };
                    let app_state_for_notify = app_state.clone();
                    let notification_cloned = notification.clone();
                    tokio::spawn(async move {
                        notify_subscribers(app_state_for_notify, &notification_cloned).await
                    });
                }
            }
        }
    }
}

async fn get_health(app_state: web::Data<Arc<AppState>>) -> impl Responder {
    let backends = app_state.backends.read();
    let backend_infos: Vec<BackendStatusInfo> = backends
        .values()
        .map(|state| BackendStatusInfo {
            name: state.config.name.clone(),
            status: state.current_status,
            last_check_time: state.last_check_time,
            consecutive_failures: state.consecutive_failures,
        })
        .collect();

    let healthy_count = backend_infos
        .iter()
        .filter(|b| b.status == HealthStatus::Healthy)
        .count();
    let total_count = backend_infos.len();

    let overall_status = if total_count == 0 {
        "healthy".to_string()
    } else if healthy_count == total_count {
        "healthy".to_string()
    } else if healthy_count == 0 {
        "unhealthy".to_string()
    } else {
        "degraded".to_string()
    };

    HttpResponse::Ok().json(AggregatedHealthResponse {
        status: overall_status,
        backends: backend_infos,
    })
}

async fn create_subscription(
    app_state: web::Data<Arc<AppState>>,
    req: web::Json<SubscriptionRequest>,
) -> impl Responder {
    let id = Uuid::new_v4();
    {
        let mut subs = app_state.subscriptions.write();
        subs.insert(id, req.callback_url.clone());
    }
    HttpResponse::Created().json(serde_json::json!({ "id": id }))
}

fn load_config() -> Vec<BackendConfig> {
    let config_path =
        std::env::var("HEALTH_CONFIG").unwrap_or_else(|_| "config.json".to_string());

    if let Ok(content) = std::fs::read_to_string(&config_path) {
        if let Ok(configs) = serde_json::from_str::<Vec<BackendConfig>>(&content) {
            return configs;
        }
    }

    vec![
        BackendConfig {
            name: "service-a".to_string(),
            health_url: "http://localhost:8081/health".to_string(),
            interval: Duration::from_secs(10),
            timeout: Duration::from_secs(5),
        },
        BackendConfig {
            name: "service-b".to_string(),
            health_url: "http://localhost:8082/health".to_string(),
            interval: Duration::from_secs(15),
            timeout: Duration::from_secs(5),
        },
    ]
}

mod serde_humantime {
    use serde::{Deserialize, Deserializer, Serializer};
    use std::time::Duration;

    pub fn serialize<S>(duration: &Duration, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        serializer.serialize_u64(duration.as_secs())
    }

    pub fn deserialize<'de, D>(deserializer: D) -> Result<Duration, D::Error>
    where
        D: Deserializer<'de>,
    {
        let secs = u64::deserialize(deserializer)?;
        Ok(Duration::from_secs(secs))
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .unwrap_or(8080);

    let configs = load_config();

    let http_client = reqwest::Client::builder()
        .timeout(Duration::from_secs(30))
        .build()
        .unwrap();

    let mut backends_map = HashMap::new();
    for config in configs.clone() {
        backends_map.insert(config.name.clone(), BackendState::new(config));
    }

    let app_state = Arc::new(AppState {
        backends: RwLock::new(backends_map),
        subscriptions: RwLock::new(HashMap::new()),
        http_client,
    });

    for config in &configs {
        let app_state_clone = app_state.clone();
        let name = config.name.clone();
        tokio::spawn(async move {
            health_check_worker(app_state_clone, name).await;
        });
    }

    let app_state_for_server = app_state.clone();
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state_for_server.clone()))
            .route("/health", web::get().to(get_health))
            .route("/subscriptions", web::post().to(create_subscription))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
