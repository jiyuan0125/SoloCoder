mod api;
mod state;

use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::net::TcpListener;
use axum::Router;
use tower_http::cors::{Any, CorsLayer};

use state::AppState;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "GRADING_SERVER_PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(short, long, env = "GRADING_SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::DEBUG)
        .init();

    let args = Args::parse();
    
    let state = Arc::new(AppState::new());
    
    let app = Router::new()
        .merge(api::routes())
        .with_state(state)
        .layer(
            CorsLayer::new()
                .allow_origin(Any)
                .allow_methods(Any)
                .allow_headers(Any),
        );

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse()?;
    tracing::info!("Server listening on {}", addr);

    let listener = TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
