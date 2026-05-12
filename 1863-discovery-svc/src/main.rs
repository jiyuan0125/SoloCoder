use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use chrono::{DateTime, Utc};
use log::{error, info, warn};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

const DEFAULT_HEARTBEAT_INTERVAL_SECS: u64 = 15;
const MAX_NOTIFICATION_FAILURES: u32 = 5;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterRequest {
    pub name: String,
    pub ip: String,
    pub port: u16,
    #[serde(default)]
    pub metadata: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterResponse {
    pub instance_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Instance {
    pub id: Uuid,
    pub name: String,
    pub ip: String,
    pub port: u16,
    pub metadata: HashMap<String, String>,
    pub last_heartbeat: DateTime<Utc>,
    pub is_healthy: bool,
    pub heartbeat_interval_secs: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Subscription {
    pub id: Uuid,
    pub service_name: String,
    pub callback_url: String,
    pub consecutive_failures: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubscribeRequest {
    pub service_name: String,
    pub callback_url: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubscribeResponse {
    pub subscription_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "event_type")]
pub enum NotificationEvent {
    Registered { instance: Instance },
    Deregistered { instance: Instance },
    MetadataUpdated { instance: Instance },
}

pub struct AppState {
    pub instances: RwLock<HashMap<Uuid, Instance>>,
    pub subscriptions: RwLock<HashMap<Uuid, Subscription>>,
    pub http_client: reqwest::Client,
}

pub type AppStateData = web::Data<Arc<AppState>>;

async fn register_service(
    state: AppStateData,
    req: web::Json<RegisterRequest>,
) -> impl Responder {
    let heartbeat_interval = DEFAULT_HEARTBEAT_INTERVAL_SECS;
    let instance_id = Uuid::new_v4();
    let instance = Instance {
        id: instance_id,
        name: req.name.clone(),
        ip: req.ip.clone(),
        port: req.port,
        metadata: req.metadata.clone(),
        last_heartbeat: Utc::now(),
        is_healthy: true,
        heartbeat_interval_secs: heartbeat_interval,
    };

    {
        let mut instances = state.instances.write().await;
        instances.insert(instance_id, instance.clone());
    }

    info!("Registered service: {} ({})", req.name, instance_id);

    send_notification(&state, &instance, NotificationEvent::Registered { instance: instance.clone() }).await;

    HttpResponse::Ok().json(RegisterResponse { instance_id })
}

async fn heartbeat(state: AppStateData, instance_id: web::Path<Uuid>) -> impl Responder {
    let instance_id = instance_id.into_inner();
    let mut instances = state.instances.write().await;

    if let Some(instance) = instances.get_mut(&instance_id) {
        instance.last_heartbeat = Utc::now();
        instance.is_healthy = true;
        info!("Heartbeat received for instance: {}", instance_id);
        HttpResponse::Ok().finish()
    } else {
        drop(instances);
        let heartbeat_interval = DEFAULT_HEARTBEAT_INTERVAL_SECS;
        let instance = Instance {
            id: instance_id,
            name: "unknown".to_string(),
            ip: "unknown".to_string(),
            port: 0,
            metadata: HashMap::new(),
            last_heartbeat: Utc::now(),
            is_healthy: true,
            heartbeat_interval_secs: heartbeat_interval,
        };

        {
            let mut instances = state.instances.write().await;
            instances.insert(instance_id, instance.clone());
        }

        warn!("Heartbeat for unknown instance, re-registering: {}", instance_id);
        send_notification(&state, &instance, NotificationEvent::Registered { instance: instance.clone() }).await;

        HttpResponse::Ok().finish()
    }
}

async fn deregister(state: AppStateData, instance_id: web::Path<Uuid>) -> impl Responder {
    let instance_id = instance_id.into_inner();
    let instance = {
        let mut instances = state.instances.write().await;
        instances.remove(&instance_id)
    };

    if let Some(instance) = instance {
        info!("Deregistered instance: {}", instance_id);
        send_notification(&state, &instance, NotificationEvent::Deregistered { instance: instance.clone() }).await;
        HttpResponse::Ok().finish()
    } else {
        HttpResponse::NotFound().finish()
    }
}

async fn update_metadata(
    state: AppStateData,
    instance_id: web::Path<Uuid>,
    metadata: web::Json<HashMap<String, String>>,
) -> impl Responder {
    let instance_id = instance_id.into_inner();
    let mut instances = state.instances.write().await;

    if let Some(instance) = instances.get_mut(&instance_id) {
        for (k, v) in metadata.iter() {
            instance.metadata.insert(k.clone(), v.clone());
        }

        let updated_instance = instance.clone();
        drop(instances);

        info!("Updated metadata for instance: {}", instance_id);
        send_notification(&state, &updated_instance, NotificationEvent::MetadataUpdated { instance: updated_instance.clone() }).await;

        HttpResponse::Ok().json(updated_instance)
    } else {
        HttpResponse::NotFound().finish()
    }
}

async fn get_service_instances(
    state: AppStateData,
    service_name: web::Path<String>,
    query: web::Query<HashMap<String, String>>,
) -> impl Responder {
    let service_name = service_name.into_inner();
    let filter_metadata: HashMap<String, String> = query
        .into_inner()
        .into_iter()
        .filter(|(k, _)| !k.is_empty())
        .collect();

    let instances = state.instances.read().await;
    let result: Vec<Instance> = instances
        .values()
        .filter(|i| i.name == service_name && i.is_healthy)
        .filter(|i| {
            filter_metadata
                .iter()
                .all(|(k, v)| i.metadata.get(k).map_or(false, |val| val == v))
        })
        .cloned()
        .collect();

    HttpResponse::Ok().json(result)
}

async fn get_all_instances(state: AppStateData) -> impl Responder {
    let instances = state.instances.read().await;
    let result: Vec<Instance> = instances.values().cloned().collect();
    HttpResponse::Ok().json(result)
}

async fn subscribe(
    state: AppStateData,
    req: web::Json<SubscribeRequest>,
) -> impl Responder {
    let subscription_id = Uuid::new_v4();
    let subscription = Subscription {
        id: subscription_id,
        service_name: req.service_name.clone(),
        callback_url: req.callback_url.clone(),
        consecutive_failures: 0,
    };

    {
        let mut subscriptions = state.subscriptions.write().await;
        subscriptions.insert(subscription_id, subscription);
    }

    info!("Subscribed to service: {}, callback: {}", req.service_name, req.callback_url);
    HttpResponse::Ok().json(SubscribeResponse { subscription_id })
}

async fn send_notification(state: &Arc<AppState>, instance: &Instance, event: NotificationEvent) {
    let subscriptions = state.subscriptions.read().await;
    let service_name = &instance.name;

    let relevant_subs: Vec<Subscription> = subscriptions
        .values()
        .filter(|s| s.service_name == *service_name)
        .cloned()
        .collect();

    drop(subscriptions);

    for mut sub in relevant_subs {
        let event_clone = event.clone();
        let client = state.http_client.clone();
        let state_clone = Arc::clone(&state);

        tokio::spawn(async move {
            let result = client
                .post(&sub.callback_url)
                .json(&event_clone)
                .send()
                .await;

            let mut subscriptions = state_clone.subscriptions.write().await;

            match result {
                Ok(resp) if resp.status().is_success() => {
                    sub.consecutive_failures = 0;
                    subscriptions.insert(sub.id, sub);
                }
                _ => {
                    sub.consecutive_failures += 1;
                    warn!(
                        "Notification failed for subscription: {}, failures: {}",
                        sub.id, sub.consecutive_failures
                    );

                    if sub.consecutive_failures >= MAX_NOTIFICATION_FAILURES {
                        error!(
                            "Removing subscription {} due to {} consecutive failures",
                            sub.id, sub.consecutive_failures
                        );
                        subscriptions.remove(&sub.id);
                    } else {
                        subscriptions.insert(sub.id, sub);
                    }
                }
            }
        });
    }
}

async fn check_heartbeats(state: Arc<AppState>, check_interval_secs: u64) {
    loop {
        tokio::time::sleep(tokio::time::Duration::from_secs(check_interval_secs)).await;

        let now = Utc::now();
        let mut to_remove = Vec::new();
        let mut to_mark_unhealthy = Vec::new();

        {
            let instances = state.instances.read().await;
            for (id, instance) in instances.iter() {
                let elapsed = now.signed_duration_since(instance.last_heartbeat);
                let threshold = chrono::Duration::seconds(instance.heartbeat_interval_secs as i64);

                if elapsed > threshold * 2 {
                    to_remove.push(*id);
                } else if elapsed > threshold {
                    if instance.is_healthy {
                        to_mark_unhealthy.push(*id);
                    }
                }
            }
        }

        if !to_mark_unhealthy.is_empty() {
            let mut instances = state.instances.write().await;
            for id in &to_mark_unhealthy {
                if let Some(instance) = instances.get_mut(id) {
                    instance.is_healthy = false;
                    warn!("Instance marked unhealthy: {}", id);
                }
            }
        }

        if !to_remove.is_empty() {
            let mut instances = state.instances.write().await;
            for id in &to_remove {
                if let Some(instance) = instances.remove(id) {
                    warn!("Removing unhealthy instance: {}", id);
                    let event = NotificationEvent::Deregistered { instance: instance.clone() };
                    drop(instances);
                    send_notification(&state, &instance, event).await;
                    instances = state.instances.write().await;
                }
            }
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .unwrap_or(8080);

    let state = Arc::new(AppState {
        instances: RwLock::new(HashMap::new()),
        subscriptions: RwLock::new(HashMap::new()),
        http_client: reqwest::Client::new(),
    });

    let heartbeat_state = state.clone();
    tokio::spawn(async move {
        check_heartbeats(heartbeat_state, DEFAULT_HEARTBEAT_INTERVAL_SECS / 2).await;
    });

    info!("Starting discovery service on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(state.clone()))
            .route("/register", web::post().to(register_service))
            .route("/instances/{id}/heartbeat", web::post().to(heartbeat))
            .route("/instances/{id}", web::delete().to(deregister))
            .route("/instances/{id}/metadata", web::put().to(update_metadata))
            .route("/services/{name}", web::get().to(get_service_instances))
            .route("/admin/instances", web::get().to(get_all_instances))
            .route("/subscriptions", web::post().to(subscribe))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
