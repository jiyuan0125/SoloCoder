mod app_state;
mod backend;
mod config;
mod handlers;
mod health;
mod proxy;
mod stats;

use actix_web::{middleware, web, App, HttpServer};
use app_state::AppState;
use config::Config;
use std::env;
use tokio::time::Duration;
use tracing::info;
use tracing_subscriber::{fmt, EnvFilter};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let config = Config::from_env();
    let port = env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let addr = format!("0.0.0.0:{}", port);

    info!("Starting reverse proxy on {}", addr);

    let state = AppState::new(config.clone());
    let health_state = state.clone();

    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(15));
        loop {
            interval.tick().await;
            health_state.check_health().await;
        }
    });

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(state.clone()))
            .wrap(middleware::Logger::default())
            .service(
                web::scope("/backends")
                    .route("", web::post().to(handlers::add_backend))
                    .route("/{id}", web::delete().to(handlers::delete_backend))
                    .route("", web::get().to(handlers::list_backends)),
            )
            .route("/stats", web::get().to(handlers::get_stats))
            .route("/config/timeout", web::post().to(handlers::set_timeout))
            .default_service(web::route().to(proxy::handle_request))
    })
    .bind(&addr)?
    .run()
    .await
}
