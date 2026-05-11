mod routes;

use std::net::SocketAddr;
use std::time::Duration;

use axum::Router;
use clap::Parser;
use tokio::time::interval;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use claim_audit_core::ClaimService;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_PORT", default_value_t = 8080)]
    port: u16,

    #[arg(short, long, env = "SERVER_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "claim_audit_server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let service = ClaimService::new();

    let cors = CorsLayer::new()
        .allow_methods(Any)
        .allow_origin(Any)
        .allow_headers(Any);

    let app = Router::new()
        .merge(routes::claim_routes(service.clone()))
        .layer(cors);

    let service_clone = service.clone();
    tokio::spawn(async move {
        let mut ticker = interval(Duration::from_secs(60));
        loop {
            ticker.tick().await;
            let results = service_clone.check_all_timeouts();
            for (id, _result) in results {
                tracing::info!("Claim {} timeout check triggered", id);
            }
        }
    });

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    tracing::info!("listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
