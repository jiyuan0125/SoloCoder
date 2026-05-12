mod channel;
mod gateway;
mod handler;

use actix_web::{App, HttpServer};
use std::env;
use std::sync::Arc;
use tokio::sync::Mutex;

use channel::ChannelManager;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let port = env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let addr = format!("0.0.0.0:{}", port);

    let channel_manager = Arc::new(Mutex::new(ChannelManager::new()));

    tracing::info!("Starting server on {}", addr);

    HttpServer::new(move || {
        App::new()
            .app_data(actix_web::web::Data::new(channel_manager.clone()))
            .service(handler::list_channels)
            .service(handler::create_channel)
            .service(handler::get_channel)
            .service(handler::configure_channel)
            .service(handler::reset_channel)
            .service(handler::get_channel_stats)
            .service(handler::gateway_request)
    })
    .bind(&addr)?
    .run()
    .await
}
