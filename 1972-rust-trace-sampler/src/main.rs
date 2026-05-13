pub mod models;
pub mod sampler;
pub mod config;
pub mod stats;

use actix_web::{web, App, HttpServer, Responder, HttpResponse};
use models::{Span, SamplerConfig};
use config::ConfigManager;
use sampler::SamplerEngine;
use stats::StatsCollector;
use std::sync::Arc;

struct AppState {
    config: ConfigManager,
    sampler: SamplerEngine,
    stats: StatsCollector,
    storage_endpoint: String,
    client: reqwest::Client,
}

async fn handle_spans(
    spans: web::Json<Vec<Span>>,
    state: web::Data<Arc<AppState>>,
) -> impl Responder {
    for span in spans.iter() {
        state.stats.record_received();
        let sample = state.sampler.should_sample(
            span,
            &state.config.get_config(),
            &state.stats,
        );
        match sample {
            sampler::SampleDecision::Keep(_) => {
                state.stats.record_kept();
                let service = span.service_name.as_deref().unwrap_or("unknown");
                let is_error = match (span.http_status_code, span.error) {
                    (Some(code), _) => code >= 500,
                    (_, Some(true)) => true,
                    _ => false,
                };
                state.stats.record_service(service, true, is_error);
                forward_span(span, state.clone()).await;
            }
            sampler::SampleDecision::Drop(..) => {
                state.stats.record_dropped();
                let service = span.service_name.as_deref().unwrap_or("unknown");
                let is_error = match (span.http_status_code, span.error) {
                    (Some(code), _) => code >= 500,
                    (_, Some(true)) => true,
                    _ => false,
                };
                state.stats.record_service(service, false, is_error);
            }
        }
    }
    HttpResponse::Ok().finish()
}

async fn forward_span(span: &Span, state: web::Data<Arc<AppState>>) {
    if state.storage_endpoint.is_empty() {
        return;
    }
    let client = state.client.clone();
    let endpoint = state.storage_endpoint.clone();
    let span_clone = span.clone();
    tokio::spawn(async move {
        let _ = client.post(&endpoint)
            .json(&span_clone)
            .timeout(std::time::Duration::from_secs(5))
            .send()
            .await;
    });
}

async fn get_config(state: web::Data<Arc<AppState>>) -> impl Responder {
    HttpResponse::Ok().json(state.config.get_config())
}

async fn put_config(
    new_config: web::Json<SamplerConfig>,
    state: web::Data<Arc<AppState>>,
) -> impl Responder {
    if let Err(e) = state.config.set_config(new_config.into_inner()) {
        return HttpResponse::BadRequest().json(serde_json::json!({ "error": e }));
    }
    HttpResponse::Ok().json(state.config.get_config())
}

async fn get_stats(state: web::Data<Arc<AppState>>) -> impl Responder {
    HttpResponse::Ok().json(state.stats.get_stats(&state.sampler, &state.config.get_config()))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .unwrap_or(8080);

    let storage_endpoint = std::env::var("STORAGE_ENDPOINT")
        .unwrap_or_default();

    let initial_config = SamplerConfig::default();
    let config = ConfigManager::new(initial_config);
    let sampler = SamplerEngine::new();
    let stats = StatsCollector::new();
    let client = reqwest::Client::new();

    let app_state = Arc::new(AppState {
        config,
        sampler,
        stats,
        storage_endpoint,
        client,
    });

    println!("Trace Sampler listening on 0.0.0.0:{}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            .route("/spans", web::post().to(handle_spans))
            .route("/sampler/config", web::get().to(get_config))
            .route("/sampler/config", web::put().to(put_config))
            .route("/sampler/stats", web::get().to(get_stats))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
