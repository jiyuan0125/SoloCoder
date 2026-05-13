use axum::{
    routing::{get, post},
    Router,
};
use std::env;
use std::net::SocketAddr;

mod models;
mod validator;
mod storage;
mod handlers;

use storage::ConfigStore;

#[tokio::main]
async fn main() {
    let store = ConfigStore::new();
    store.init_sample_data();
    
    let app = Router::new()
        .route("/health", get(handlers::health_check))
        .route("/configs", post(handlers::create_config))
        .route("/configs/:key/impact", get(handlers::get_config_impact))
        .route("/configs/:key/versions", get(handlers::get_config_versions))
        .route("/configs/:key/deploy", post(handlers::deploy_config))
        .route("/configs/:key/rollback", post(handlers::rollback_config))
        .route("/configs/:key/deployments", get(handlers::get_deploy_status))
        .with_state(store);
    
    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse()
        .expect("PORT must be a valid number");
    
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    
    println!("Config Validator server running on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    
    axum::serve(listener, app).await.unwrap();
}
