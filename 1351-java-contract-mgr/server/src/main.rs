mod routes;

use clap::Parser;
use std::sync::Arc;
use axum::Router;
use contract_core::{ContractStore, ContractService};
use tower_http::cors::{Any, CorsLayer};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value = "3000")]
    port: u16,

    #[arg(short = 'H', long, env = "HOST", default_value = "0.0.0.0")]
    host: String,
}

#[derive(Clone)]
pub struct AppState {
    pub service: Arc<ContractService>,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();

    let store = ContractStore::new();
    let service = Arc::new(ContractService::new(store));
    let state = AppState { service };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .merge(routes::routes())
        .layer(cors)
        .with_state(state);

    let addr = format!("{}:{}", args.host, args.port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();

    tracing::info!("Contract Manager Server running on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
