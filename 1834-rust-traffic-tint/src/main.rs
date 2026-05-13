use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::{Mutex, mpsc};
use rand::Rng;
use chrono::Utc;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Version {
    name: String,
    weight: f64,
    request_count: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct WeightHistory {
    version_name: String,
    from_weight: f64,
    to_weight: f64,
    timestamp: chrono::DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct RegisterVersion {
    name: String,
    initial_weight: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct UpdateWeight {
    new_weight: f64,
}

#[derive(Debug, Clone, Serialize)]
struct PendingAdjustment {
    version_name: String,
    target_weight: f64,
    step_size: f64,
    remaining_steps: usize,
}

#[derive(Debug, Clone, Serialize)]
struct VersionInfo {
    name: String,
    current_weight: f64,
    target_weight: Option<f64>,
    request_count: u64,
}

struct SharedState {
    versions: Mutex<HashMap<String, Version>>,
    history: Mutex<Vec<WeightHistory>>,
    pending: Mutex<HashMap<String, PendingAdjustment>>,
}

const MAX_WEIGHT_CHANGE: f64 = 5.0;
const ADJUSTMENT_INTERVAL_SECS: u64 = 1;

async fn register_version(
    state: web::Data<Arc<SharedState>>,
    version: web::Json<RegisterVersion>,
) -> impl Responder {
    let mut versions = state.versions.lock().await;
    if versions.contains_key(&version.name) {
        return HttpResponse::Conflict().json(serde_json::json!({
            "error": "Version already exists"
        }));
    }
    versions.insert(version.name.clone(), Version {
        name: version.name.clone(),
        weight: version.initial_weight,
        request_count: 0,
    });
    HttpResponse::Created().json(serde_json::json!({
        "message": "Version registered successfully",
        "version": version.name,
        "initial_weight": version.initial_weight
    }))
}

async fn update_weight(
    state: web::Data<Arc<SharedState>>,
    path: web::Path<String>,
    weight: web::Json<UpdateWeight>,
) -> impl Responder {
    let version_name = path.into_inner();
    let new_weight = weight.new_weight;
    
    let versions = state.versions.lock().await;
    let version = match versions.get(&version_name) {
        Some(v) => v,
        None => return HttpResponse::NotFound().json(serde_json::json!({
            "error": "Version not found"
        })),
    };
    
    let current_weight = version.weight;
    let diff = new_weight - current_weight;
    let abs_diff = diff.abs();
    
    if abs_diff <= MAX_WEIGHT_CHANGE {
        drop(versions);
        let mut versions = state.versions.lock().await;
        let mut history = state.history.lock().await;
        
        if let Some(v) = versions.get_mut(&version_name) {
            history.push(WeightHistory {
                version_name: version_name.clone(),
                from_weight: current_weight,
                to_weight: new_weight,
                timestamp: Utc::now(),
            });
            v.weight = new_weight;
        }
        
        return HttpResponse::Ok().json(serde_json::json!({
            "message": "Weight updated",
            "version": version_name,
            "from": current_weight,
            "to": new_weight,
            "steps": 1,
            "completed_instantly": true
        }));
    }
    
    drop(versions);
    
    let mut pending = state.pending.lock().await;
    if pending.contains_key(&version_name) {
        return HttpResponse::Conflict().json(serde_json::json!({
            "error": "Version already has an ongoing adjustment",
            "message": "Please wait for the current adjustment to complete"
        }));
    }
    
    let steps = (abs_diff / MAX_WEIGHT_CHANGE).ceil() as usize;
    let step_size = if diff > 0.0 { MAX_WEIGHT_CHANGE } else { -MAX_WEIGHT_CHANGE };
    
    let adjustment = PendingAdjustment {
        version_name: version_name.clone(),
        target_weight: new_weight,
        step_size,
        remaining_steps: steps,
    };
    
    pending.insert(version_name.clone(), adjustment);
    drop(pending);
    
    let mut versions = state.versions.lock().await;
    let mut history = state.history.lock().await;
    
    if let Some(v) = versions.get_mut(&version_name) {
        let first_step_weight = v.weight + step_size;
        history.push(WeightHistory {
            version_name: version_name.clone(),
            from_weight: v.weight,
            to_weight: first_step_weight,
            timestamp: Utc::now(),
        });
        v.weight = first_step_weight;
    }
    
    HttpResponse::Ok().json(serde_json::json!({
        "message": "Weight adjustment started",
        "version": version_name,
        "from": current_weight,
        "to": new_weight,
        "total_steps": steps,
        "steps_completed": 1,
        "steps_remaining": steps - 1,
        "interval_seconds": ADJUSTMENT_INTERVAL_SECS,
        "max_change_per_step": MAX_WEIGHT_CHANGE
    }))
}

async fn get_versions(state: web::Data<Arc<SharedState>>) -> impl Responder {
    let versions = state.versions.lock().await;
    let history = state.history.lock().await;
    let pending = state.pending.lock().await;
    
    let total_weight: f64 = versions.values().map(|v| v.weight).sum();
    
    let versions_info: Vec<VersionInfo> = versions
        .values()
        .map(|v| {
            let pending = pending.get(&v.name);
            VersionInfo {
                name: v.name.clone(),
                current_weight: v.weight,
                target_weight: pending.map(|p| p.target_weight),
                request_count: v.request_count,
            }
        })
        .collect();
    
    let pending_info: Vec<serde_json::Value> = pending
        .values()
        .map(|p| serde_json::json!({
            "version": p.version_name,
            "target_weight": p.target_weight,
            "remaining_steps": p.remaining_steps,
            "interval_seconds": ADJUSTMENT_INTERVAL_SECS
        }))
        .collect();
    
    HttpResponse::Ok().json(serde_json::json!({
        "versions": versions_info,
        "total_weight": total_weight,
        "pending_adjustments": pending_info,
        "history": history.clone()
    }))
}

async fn delete_version(
    state: web::Data<Arc<SharedState>>,
    path: web::Path<String>,
) -> impl Responder {
    let version_name = path.into_inner();
    let versions = state.versions.lock().await;
    let pending = state.pending.lock().await;
    
    if pending.contains_key(&version_name) {
        return HttpResponse::BadRequest().json(serde_json::json!({
            "error": "Cannot delete version with ongoing adjustment"
        }));
    }
    
    match versions.get(&version_name) {
        Some(v) => {
            if v.request_count > 0 {
                return HttpResponse::BadRequest().json(serde_json::json!({
                    "error": "Cannot delete version with active traffic",
                    "request_count": v.request_count
                }));
            }
        },
        None => return HttpResponse::NotFound().json(serde_json::json!({
            "error": "Version not found"
        })),
    }
    
    drop(versions);
    drop(pending);
    
    let mut versions = state.versions.lock().await;
    versions.remove(&version_name);
    
    HttpResponse::Ok().json(serde_json::json!({
        "message": "Version deleted successfully",
        "version": version_name
    }))
}

async fn handle_traffic(state: web::Data<Arc<SharedState>>) -> impl Responder {
    let mut versions = state.versions.lock().await;
    let total_weight: f64 = versions.values().map(|v| v.weight).sum();
    
    if total_weight <= 0.0 {
        return HttpResponse::ServiceUnavailable().json(serde_json::json!({
            "error": "No available versions"
        }));
    }
    
    let mut rng = rand::thread_rng();
    let mut random = rng.gen_range(0.0..total_weight);
    
    let mut selected = None;
    for version in versions.values_mut() {
        random -= version.weight;
        if random <= 0.0 {
            version.request_count += 1;
            selected = Some(version.name.clone());
            break;
        }
    }
    
    match selected {
        Some(name) => HttpResponse::Ok().json(serde_json::json!({
            "message": "Request routed",
            "version": name
        })),
        None => HttpResponse::ServiceUnavailable().json(serde_json::json!({
            "error": "Failed to route request"
        })),
    }
}

async fn adjustment_loop(
    state: Arc<SharedState>,
    mut shutdown_rx: mpsc::Receiver<()>,
) {
    let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(ADJUSTMENT_INTERVAL_SECS));
    
    loop {
        tokio::select! {
            _ = interval.tick() => {
                process_pending_adjustments(&state).await;
            }
            _ = shutdown_rx.recv() => {
                println!("Adjustment loop shutting down");
                break;
            }
        }
    }
}

async fn process_pending_adjustments(state: &Arc<SharedState>) {
    let mut pending = state.pending.lock().await;
    
    let versions_to_adjust: Vec<String> = pending.keys().cloned().collect();
    
    for version_name in versions_to_adjust {
        let should_remove = {
            let adjustment = pending.get_mut(&version_name).unwrap();
            
            if adjustment.remaining_steps <= 1 {
                true
            } else {
                adjustment.remaining_steps -= 1;
                false
            }
        };
        
        let mut versions = state.versions.lock().await;
        let mut history = state.history.lock().await;
        
        if let Some(v) = versions.get_mut(&version_name) {
            let adjustment = pending.get(&version_name).unwrap();
            let next_weight = if should_remove {
                adjustment.target_weight
            } else {
                v.weight + adjustment.step_size
            };
            
            history.push(WeightHistory {
                version_name: version_name.clone(),
                from_weight: v.weight,
                to_weight: next_weight,
                timestamp: Utc::now(),
            });
            
            v.weight = next_weight;
        }
        
        if should_remove {
            pending.remove(&version_name);
            println!("Adjustment completed for version: {}", version_name);
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let port: u16 = port.parse().expect("PORT must be a valid port number");
    
    let state = Arc::new(SharedState {
        versions: Mutex::new(HashMap::new()),
        history: Mutex::new(Vec::new()),
        pending: Mutex::new(HashMap::new()),
    });
    
    let (shutdown_tx, shutdown_rx) = mpsc::channel::<()>(1);
    
    let adjustment_state = state.clone();
    let adjustment_handle = tokio::spawn(adjustment_loop(adjustment_state, shutdown_rx));
    
    println!("Gray Gateway starting on port {}", port);
    println!("Max weight change per step: {}%", MAX_WEIGHT_CHANGE);
    println!("Adjustment interval: {} seconds", ADJUSTMENT_INTERVAL_SECS);
    
    let server = HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(state.clone()))
            .route("/versions", web::post().to(register_version))
            .route("/versions/{name}/weight", web::put().to(update_weight))
            .route("/versions", web::get().to(get_versions))
            .route("/versions/{name}", web::delete().to(delete_version))
            .route("/", web::get().to(handle_traffic))
    })
    .bind(("127.0.0.1", port))?
    .run();
    
    let server_result = server.await;
    
    let _ = shutdown_tx.send(()).await;
    let _ = adjustment_handle.await;
    
    server_result
}
