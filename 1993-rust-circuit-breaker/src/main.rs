mod breaker;
mod models;
mod handlers;
mod notifier;

use std::env;
use std::sync::Arc;
use actix_web::{web, App, HttpServer};
use breaker::CircuitBreakerRegistry;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    
    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "10005".to_string())
        .parse()
        .expect("PORT must be a valid number");
    
    let registry = Arc::new(CircuitBreakerRegistry::new());
    
    log::info!("Starting circuit breaker server on port {}", port);
    
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(registry.clone()))
            .service(handlers::create_breaker)
            .service(handlers::get_breaker)
            .service(handlers::list_breakers)
            .service(handlers::update_config)
            .service(handlers::force_open)
            .service(handlers::force_close)
            .service(handlers::delete_breaker)
            .service(handlers::proxy_request)
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
