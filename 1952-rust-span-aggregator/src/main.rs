use actix_web::{web, App, HttpServer};
use std::env;

mod models;
mod storage;
mod handlers;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(8080);

    let p99_threshold_ms: u64 = env::var("P99_THRESHOLD_MS")
        .ok()
        .and_then(|t| t.parse().ok())
        .unwrap_or(2000);

    let app_state = web::Data::new(storage::AppState::new(p99_threshold_ms));

    println!(
        "Starting Span Aggregator on port {} with P99 threshold {}ms",
        port, p99_threshold_ms
    );

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .route("/spans", web::post().to(handlers::submit_span))
            .route("/traces", web::get().to(handlers::list_traces))
            .route("/traces/{trace_id}", web::get().to(handlers::get_trace))
            .route("/dependency-graph", web::get().to(handlers::get_dependency_graph))
            .route("/stats", web::get().to(handlers::get_stats))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
