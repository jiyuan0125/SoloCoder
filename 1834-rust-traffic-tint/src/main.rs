use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
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
struct VersionInfo {
    name: String,
    current_weight: f64,
    request_count: u64,
}

struct AppState {
    versions: Mutex<HashMap<String, Version>>,
    history: Mutex<Vec<WeightHistory>>,
}

const MAX_WEIGHT_CHANGE: f64 = 5.0;

async fn register_version(
    state: web::Data<AppState>,
    version: web::Json<RegisterVersion>,
) -> impl Responder {
    let mut versions = state.versions.lock().unwrap();
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
    state: web::Data<AppState>,
    path: web::Path<String>,
    weight: web::Json<UpdateWeight>,
) -> impl Responder {
    let version_name = path.into_inner();
    let new_weight = weight.new_weight;
    
    let mut versions = state.versions.lock().unwrap();
    let version = match versions.get_mut(&version_name) {
        Some(v) => v,
        None => return HttpResponse::NotFound().json(serde_json::json!({
            "error": "Version not found"
        })),
    };
    
    let current_weight = version.weight;
    let mut history = state.history.lock().unwrap();
    
    let diff = new_weight - current_weight;
    let abs_diff = diff.abs();
    
    if abs_diff <= MAX_WEIGHT_CHANGE {
        history.push(WeightHistory {
            version_name: version_name.clone(),
            from_weight: current_weight,
            to_weight: new_weight,
            timestamp: Utc::now(),
        });
        version.weight = new_weight;
        return HttpResponse::Ok().json(serde_json::json!({
            "message": "Weight updated",
            "version": version_name,
            "from": current_weight,
            "to": new_weight,
            "steps": 1
        }));
    }
    
    let _steps = (abs_diff / MAX_WEIGHT_CHANGE).ceil() as usize;
    let step_size = if diff > 0.0 { MAX_WEIGHT_CHANGE } else { -MAX_WEIGHT_CHANGE };
    
    let mut current = current_weight;
    let mut step_count = 0;
    
    while (current - new_weight).abs() > MAX_WEIGHT_CHANGE {
        let next = current + step_size;
        history.push(WeightHistory {
            version_name: version_name.clone(),
            from_weight: current,
            to_weight: next,
            timestamp: Utc::now(),
        });
        current = next;
        step_count += 1;
    }
    
    history.push(WeightHistory {
        version_name: version_name.clone(),
        from_weight: current,
        to_weight: new_weight,
        timestamp: Utc::now(),
    });
    current = new_weight;
    step_count += 1;
    
    version.weight = current;
    
    HttpResponse::Ok().json(serde_json::json!({
        "message": "Weight updated in multiple steps",
        "version": version_name,
        "from": current_weight,
        "to": new_weight,
        "steps": step_count,
        "max_change_per_step": MAX_WEIGHT_CHANGE
    }))
}

async fn get_versions(state: web::Data<AppState>) -> impl Responder {
    let versions = state.versions.lock().unwrap();
    let history = state.history.lock().unwrap();
    
    let total_weight: f64 = versions.values().map(|v| v.weight).sum();
    
    let versions_info: Vec<VersionInfo> = versions
        .values()
        .map(|v| VersionInfo {
            name: v.name.clone(),
            current_weight: v.weight,
            request_count: v.request_count,
        })
        .collect();
    
    HttpResponse::Ok().json(serde_json::json!({
        "versions": versions_info,
        "total_weight": total_weight,
        "history": history.clone()
    }))
}

async fn delete_version(
    state: web::Data<AppState>,
    path: web::Path<String>,
) -> impl Responder {
    let version_name = path.into_inner();
    let mut versions = state.versions.lock().unwrap();
    
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
    
    versions.remove(&version_name);
    
    HttpResponse::Ok().json(serde_json::json!({
        "message": "Version deleted successfully",
        "version": version_name
    }))
}

async fn handle_traffic(state: web::Data<AppState>) -> impl Responder {
    let mut versions = state.versions.lock().unwrap();
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

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let port: u16 = port.parse().expect("PORT must be a valid port number");
    
    let state = web::Data::new(AppState {
        versions: Mutex::new(HashMap::new()),
        history: Mutex::new(Vec::new()),
    });
    
    println!("Gray Gateway starting on port {}", port);
    
    HttpServer::new(move || {
        App::new()
            .app_data(state.clone())
            .route("/versions", web::post().to(register_version))
            .route("/versions/{name}/weight", web::put().to(update_weight))
            .route("/versions", web::get().to(get_versions))
            .route("/versions/{name}", web::delete().to(delete_version))
            .route("/", web::get().to(handle_traffic))
    })
    .bind(("127.0.0.1", port))?
    .run()
    .await
}
