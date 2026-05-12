mod model;
mod storage;
mod api;

use std::env;
use std::net::SocketAddr;
use std::sync::Arc;

use tokio::sync::Mutex;
use storage::TraceStorage;

#[tokio::main]
async fn main() {
    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(3000);

    let storage = Arc::new(Mutex::new(TraceStorage::new()));

    let app = api::create_router(storage);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("trace-chain server listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
