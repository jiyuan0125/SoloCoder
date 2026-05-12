mod cache;
mod handler;
mod diag;

use std::sync::Arc;
use std::env;
use tokio::sync::Mutex;

use cache::CacheService;
use diag::DiagLogger;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(3000);

    let cache_size: usize = env::var("CACHE_SIZE")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(10000);

    let cache = Arc::new(Mutex::new(CacheService::new(cache_size)));
    let logger = Arc::new(Mutex::new(DiagLogger::new()));

    let logger_clone = Arc::clone(&logger);
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(std::time::Duration::from_secs(3600));
        loop {
            interval.tick().await;
            let mut log = logger_clone.lock().await;
            log.cleanup_old_logs(7);
        }
    });

    let app = handler::create_router(cache, logger);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
