mod handlers;
mod health;
mod models;
mod proxy;
mod state;

use crate::handlers::*;
use crate::health::process_health_check_result;
use crate::state::{AppState, HEALTH_CHECK_INTERVAL_SECS};
use actix_web::{web, App, HttpServer};
use log::info;
use reqwest::Client;
use std::env;
use std::time::Duration;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    
    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .expect("PORT must be a valid port number");
    
    let app_state = AppState::default();
    let client = Client::builder()
        .timeout(Duration::from_secs(30))
        .build()
        .expect("Failed to create HTTP client");
    
    {
        let state = app_state.clone();
        let client = client.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(HEALTH_CHECK_INTERVAL_SECS));
            loop {
                interval.tick().await;
                let backends = state.get_all_backends().await;
                for backend in backends {
                    process_health_check_result(&state, &client, backend).await;
                }
            }
        });
    }
    
    let state_for_server = app_state.clone();
    let client_for_server = client.clone();
    
    info!("Gateway starting on port {}", port);
    
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(state_for_server.clone()))
            .app_data(web::Data::new(client_for_server.clone()))
            .service(
                web::resource("/routes")
                    .route(web::post().to(create_route))
                    .route(web::get().to(get_routes)),
            )
            .service(
                web::resource("/routes/{id}")
                    .route(web::delete().to(delete_route)),
            )
            .service(
                web::resource("/keys")
                    .route(web::post().to(create_api_key))
                    .route(web::get().to(get_api_keys)),
            )
            .service(
                web::resource("/services")
                    .route(web::get().to(get_services))
                    .route(web::post().to(register_backend)),
            )
            .service(
                web::resource("/services/{name}/offline")
                    .route(web::post().to(set_backend_offline)),
            )
            .service(
                web::resource("/stats")
                    .route(web::get().to(get_stats)),
            )
            .default_service(web::route().to(proxy_handler))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
