use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;
use uuid::Uuid;

const DEFAULT_UNHEALTHY_TIMEOUT: Duration = Duration::from_secs(60);

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum InstanceStatus {
    Registering,
    Healthy,
    Unhealthy,
    Deregistered,
}

#[derive(Debug, Clone)]
pub struct ServiceInstance {
    pub instance_id: Uuid,
    pub service_name: String,
    pub ip: String,
    pub port: u16,
    pub metadata: HashMap<String, String>,
    pub status: InstanceStatus,
    pub heartbeat_interval: Duration,
    pub last_heartbeat: Instant,
    pub unhealthy_since: Option<Instant>,
    pub registered_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ServiceInstanceView {
    pub instance_id: Uuid,
    pub service_name: String,
    pub ip: String,
    pub port: u16,
    pub metadata: HashMap<String, String>,
    pub status: InstanceStatus,
    pub heartbeat_interval_secs: u64,
    pub registered_at: DateTime<Utc>,
}

impl From<&ServiceInstance> for ServiceInstanceView {
    fn from(instance: &ServiceInstance) -> Self {
        ServiceInstanceView {
            instance_id: instance.instance_id,
            service_name: instance.service_name.clone(),
            ip: instance.ip.clone(),
            port: instance.port,
            metadata: instance.metadata.clone(),
            status: instance.status.clone(),
            heartbeat_interval_secs: instance.heartbeat_interval.as_secs(),
            registered_at: instance.registered_at,
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct RegisterRequest {
    pub service_name: String,
    pub ip: String,
    pub port: u16,
    #[serde(default)]
    pub metadata: HashMap<String, String>,
    #[serde(default = "default_heartbeat_interval_secs")]
    pub heartbeat_interval_secs: u64,
}

fn default_heartbeat_interval_secs() -> u64 {
    15
}

#[derive(Debug, Clone, Serialize)]
pub struct RegisterResponse {
    pub instance_id: Uuid,
}

#[derive(Debug, Clone, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub struct RegistryState {
    pub instances: HashMap<Uuid, ServiceInstance>,
    pub service_index: HashMap<String, Vec<Uuid>>,
}

impl RegistryState {
    fn new() -> Self {
        RegistryState {
            instances: HashMap::new(),
            service_index: HashMap::new(),
        }
    }
}

pub type SharedRegistry = Arc<Mutex<RegistryState>>;

fn metadata_matches(
    instance: &ServiceInstance,
    filters: &HashMap<String, String>,
) -> bool {
    for (key, value) in filters {
        match instance.metadata.get(key) {
            Some(v) if v == value => continue,
            _ => return false,
        }
    }
    true
}

async fn register(
    registry: web::Data<SharedRegistry>,
    req: web::Json<RegisterRequest>,
) -> impl Responder {
    let req = req.into_inner();
    let instance_id = Uuid::new_v4();
    let heartbeat_interval = Duration::from_secs(req.heartbeat_interval_secs);
    
    let instance = ServiceInstance {
        instance_id,
        service_name: req.service_name.clone(),
        ip: req.ip,
        port: req.port,
        metadata: req.metadata,
        status: InstanceStatus::Registering,
        heartbeat_interval,
        last_heartbeat: Instant::now(),
        unhealthy_since: None,
        registered_at: Utc::now(),
    };
    
    let mut state = registry.lock().await;
    state.instances.insert(instance_id, instance);
    state.service_index
        .entry(req.service_name)
        .or_default()
        .push(instance_id);
    
    HttpResponse::Ok().json(RegisterResponse { instance_id })
}

async fn heartbeat(
    registry: web::Data<SharedRegistry>,
    instance_id: web::Path<Uuid>,
) -> impl Responder {
    let instance_id = instance_id.into_inner();
    let mut state = registry.lock().await;
    
    let instance = state.instances.get_mut(&instance_id);
    
    match instance {
        Some(instance) => {
            instance.last_heartbeat = Instant::now();
            instance.unhealthy_since = None;
            
            match instance.status {
                InstanceStatus::Registering => {
                    instance.status = InstanceStatus::Healthy;
                }
                InstanceStatus::Unhealthy => {
                    instance.status = InstanceStatus::Healthy;
                }
                _ => {}
            }
            
            HttpResponse::Ok().finish()
        }
        None => {
            HttpResponse::NotFound().json(ErrorResponse {
                error: format!("Instance not found: {}", instance_id),
            })
        }
    }
}

async fn deregister(
    registry: web::Data<SharedRegistry>,
    instance_id: web::Path<Uuid>,
) -> impl Responder {
    let instance_id = instance_id.into_inner();
    let mut state = registry.lock().await;
    
    if let Some(instance) = state.instances.get_mut(&instance_id) {
        instance.status = InstanceStatus::Deregistered;
        return HttpResponse::Ok().finish();
    }
    
    HttpResponse::NotFound().json(ErrorResponse {
        error: format!("Instance not found: {}", instance_id),
    })
}

async fn get_service(
    registry: web::Data<SharedRegistry>,
    service_name: web::Path<String>,
    query: web::Query<HashMap<String, String>>,
) -> impl Responder {
    let service_name = service_name.into_inner();
    let state = registry.lock().await;
    
    let filters: HashMap<String, String> = query
        .into_inner()
        .into_iter()
        .filter(|(key, _)| key != "service_name")
        .collect();
    
    let instance_ids = state.service_index.get(&service_name);
    
    let instances: Vec<ServiceInstanceView> = match instance_ids {
        Some(ids) => ids
            .iter()
            .filter_map(|id| state.instances.get(id))
            .filter(|instance| {
                instance.status == InstanceStatus::Healthy
                    && metadata_matches(instance, &filters)
            })
            .map(ServiceInstanceView::from)
            .collect(),
        None => Vec::new(),
    };
    
    HttpResponse::Ok().json(instances)
}

async fn get_all_instances(registry: web::Data<SharedRegistry>) -> impl Responder {
    let state = registry.lock().await;
    let instances: Vec<ServiceInstanceView> = state
        .instances
        .values()
        .map(ServiceInstanceView::from)
        .collect();
    HttpResponse::Ok().json(instances)
}

async fn health_check_task(registry: SharedRegistry) {
    loop {
        tokio::time::sleep(Duration::from_secs(5)).await;
        let now = Instant::now();
        
        let mut to_remove = Vec::new();
        let mut to_mark_unhealthy = Vec::new();
        
        {
            let state = registry.lock().await;
            
            for (id, instance) in &state.instances {
                if instance.status == InstanceStatus::Deregistered {
                    continue;
                }
                
                let heartbeat_timeout = instance.heartbeat_interval * 2;
                let since_last_heartbeat = now.duration_since(instance.last_heartbeat);
                
                if since_last_heartbeat >= heartbeat_timeout {
                    match instance.status {
                        InstanceStatus::Healthy | InstanceStatus::Registering => {
                            to_mark_unhealthy.push(*id);
                        }
                        InstanceStatus::Unhealthy => {
                            if let Some(unhealthy_since) = instance.unhealthy_since {
                                let since_unhealthy = now.duration_since(unhealthy_since);
                                if since_unhealthy >= DEFAULT_UNHEALTHY_TIMEOUT {
                                    to_remove.push(*id);
                                }
                            }
                        }
                        _ => {}
                    }
                }
            }
        }
        
        let mut state = registry.lock().await;
        
        for id in to_mark_unhealthy {
            if let Some(instance) = state.instances.get_mut(&id) {
                instance.status = InstanceStatus::Unhealthy;
                if instance.unhealthy_since.is_none() {
                    instance.unhealthy_since = Some(now);
                }
            }
        }
        
        for id in to_remove {
            if let Some(instance) = state.instances.remove(&id) {
                if let Some(ids) = state.service_index.get_mut(&instance.service_name) {
                    ids.retain(|&x| x != id);
                    if ids.is_empty() {
                        state.service_index.remove(&instance.service_name);
                    }
                }
            }
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let registry = Arc::new(Mutex::new(RegistryState::new()));
    
    {
        let registry_clone = Arc::clone(&registry);
        tokio::spawn(async move {
            health_check_task(registry_clone).await;
        });
    }
    
    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid port number");
    
    println!("Service Registry starting on port {}", port);
    
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(Arc::clone(&registry)))
            .route("/register", web::post().to(register))
            .route("/heartbeat/{instance_id}", web::put().to(heartbeat))
            .route("/deregister/{instance_id}", web::delete().to(deregister))
            .route("/services/{name}", web::get().to(get_service))
            .route("/admin/instances", web::get().to(get_all_instances))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
