mod routes;
mod handlers;
mod state;

use clap::Parser;
use std::net::SocketAddr;
use axum::Router;
use tower::ServiceBuilder;
use tower::layer::Layer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use state::AppState;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=debug,tower_http=debug,axum::rejection=trace".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();
    
    let store = cert_core::InMemoryStore::new();
    let service = cert_core::CertificationService::new(store.clone());
    let state = AppState::new(service);

    let app = Router::new()
        .merge(routes::create_routes(state))
        .layer(ServiceBuilder::new().layer(axum::middleware::map_response(
            |mut response: axum::response::Response| async {
                response.headers_mut().insert(
                    axum::http::header::CONTENT_TYPE,
                    axum::http::HeaderValue::from_static("application/json"),
                );
                response
            },
        )));

    let addr = SocketAddr::from(([127, 0, 0, 1], args.port));
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
