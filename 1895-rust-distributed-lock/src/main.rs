mod handlers;
mod lock_manager;
mod models;

use crate::handlers::{
    acquire_lock, force_release_lock, health_check, list_locks, release_lock, renew_lock,
};
use crate::lock_manager::LockManager;
use actix_web::{web, App, HttpServer};
use std::env;
use std::time::Duration;
use tokio::time::interval;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid number");

    let lock_manager = LockManager::new();
    let cleanup_manager = lock_manager.clone();

    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(1));
        loop {
            interval.tick().await;
            cleanup_manager.cleanup_expired().await;
        }
    });

    println!("Distributed Lock Service starting on port {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(lock_manager.clone()))
            .route("/health", web::get().to(health_check))
            .route("/locks", web::get().to(list_locks))
            .route("/locks/{name}/acquire", web::post().to(acquire_lock))
            .route("/locks/{name}/release", web::post().to(release_lock))
            .route("/locks/{name}/renew", web::post().to(renew_lock))
            .route(
                "/admin/locks/{name}/force-release",
                web::post().to(force_release_lock),
            )
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
