use actix_web::{web, App, HttpServer};
use std::env;

use rust_limiter::app_state::AppState;
use rust_limiter::middleware::RateLimiterMiddleware;
use rust_limiter::routes::{
    get_sliding_window_config, get_webhook, health_check, list_token_bucket_rules,
    remove_token_bucket_rule, set_sliding_window_config, set_token_bucket_rule, set_webhook,
};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    let port = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .unwrap_or(8080);

    let app_state = AppState::new();

    let sliding_window_clone = app_state.sliding_window_limiter.clone();
    let token_bucket_clone = app_state.token_bucket_limiter.clone();
    let alert_manager_clone = app_state.alert_manager.clone();

    log::info!("Starting server on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            .wrap(RateLimiterMiddleware::new(
                sliding_window_clone.clone(),
                token_bucket_clone.clone(),
                alert_manager_clone.clone(),
            ))
            .route("/health", web::get().to(health_check))
            .route("/api/limiter/sliding-window", web::post().to(set_sliding_window_config))
            .route("/api/limiter/sliding-window", web::get().to(get_sliding_window_config))
            .route("/api/limiter/token-bucket", web::post().to(set_token_bucket_rule))
            .route("/api/limiter/token-bucket", web::delete().to(remove_token_bucket_rule))
            .route("/api/limiter/token-bucket/rules", web::get().to(list_token_bucket_rules))
            .route("/api/limiter/webhook", web::post().to(set_webhook))
            .route("/api/limiter/webhook", web::get().to(get_webhook))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
