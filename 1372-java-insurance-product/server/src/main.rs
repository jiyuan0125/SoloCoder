mod handlers;

use clap::Parser;
use insurance_core::InMemoryStore;
use std::net::SocketAddr;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "INSURANCE_PORT", default_value = "8080")]
    port: u16,

    #[arg(long, env = "INSURANCE_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
pub struct AppState {
    pub store: InMemoryStore,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let args = Args::parse();

    let state = AppState {
        store: InMemoryStore::new(),
    };

    let app = handlers::create_app(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("无效的地址");

    tracing::info!("服务启动在 {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .expect("服务启动失败");
}
