use actix_web::{delete, get, post, web, App, HttpResponse, HttpServer, Responder};
use serde::Deserialize;
use uuid::Uuid;

use crate::models::{CreateProbe, Probe};
use crate::state::AppState;

#[derive(Deserialize)]
struct ProbeIdPath {
    id: Uuid,
}

#[post("/probes")]
async fn create_probe(
    state: web::Data<AppState>,
    create: web::Json<CreateProbe>,
) -> impl Responder {
    let probe = Probe::from_create(create.into_inner());
    let _id = state.add_probe(probe.clone()).await;
    HttpResponse::Created().json(probe)
}

#[get("/probes")]
async fn list_probes(state: web::Data<AppState>) -> impl Responder {
    let probes = state.list_probes().await;
    HttpResponse::Ok().json(probes)
}

#[get("/probes/{id}/status")]
async fn get_probe_status(
    state: web::Data<AppState>,
    path: web::Path<ProbeIdPath>,
) -> impl Responder {
    match state.get_result(&path.id).await {
        Some(result) => HttpResponse::Ok().json(result),
        None => HttpResponse::NotFound().json(serde_json::json!({
            "error": "No result found for this probe"
        })),
    }
}

#[delete("/probes/{id}")]
async fn delete_probe(
    state: web::Data<AppState>,
    path: web::Path<ProbeIdPath>,
) -> impl Responder {
    if state.remove_probe(&path.id).await {
        HttpResponse::NoContent().finish()
    } else {
        HttpResponse::NotFound().json(serde_json::json!({
            "error": "Probe not found"
        }))
    }
}

#[get("/health")]
async fn health(state: web::Data<AppState>) -> impl Responder {
    let aggregated = state.aggregate_health().await;
    HttpResponse::Ok().json(aggregated)
}

pub async fn run_server(port: u16, state: AppState) -> std::io::Result<()> {
    let server = HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(state.clone()))
            .service(create_probe)
            .service(list_probes)
            .service(get_probe_status)
            .service(delete_probe)
            .service(health)
    })
    .bind(("0.0.0.0", port))?
    .run();

    println!("Server running on port {}", port);
    server.await
}
