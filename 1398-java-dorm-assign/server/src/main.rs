mod handlers;
mod state;

use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use dorm_core::{DormService, Repository};
use state::AppState;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "DORM_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short = 'H', long, env = "DORM_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "dorm_server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let repo = Arc::new(Repository::new());
    let service = Arc::new(DormService::new(Arc::clone(&repo)));
    let state = Arc::new(AppState { service });

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = handlers::create_router(state).layer(cors);

    let addr = SocketAddr::from((
        args.host.parse::<std::net::IpAddr>().expect("Invalid host address"),
        args.port,
    ));
    
    tracing::debug!("listening on {}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
