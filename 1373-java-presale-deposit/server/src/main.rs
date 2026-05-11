mod handlers;
mod middleware;
mod routes;
mod state;

use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::time::{interval, Duration};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use state::AppState;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Config {
    #[arg(short, long, env = "PRESALE_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "PRESALE_HOST", default_value = "0.0.0.0")]
    host: String,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "presale_server=debug,tower_http=debug,axum=trace".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = Config::parse();

    let app_state = Arc::new(AppState::new());

    start_overdue_checker(app_state.clone());

    let app = routes::create_router().with_state(app_state);

    let addr: SocketAddr = format!("{}:{}", config.host, config.port)
        .parse()
        .expect("Invalid address");

    tracing::info!("Presale server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

fn start_overdue_checker(state: Arc<AppState>) {
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(60));
        loop {
            interval.tick().await;
            let now = chrono::Utc::now();
            let expired = state.service.process_overdue_orders(now);
            if !expired.is_empty() {
                tracing::info!("Processed {} overdue orders", expired.len());
            }
        }
    });
}
