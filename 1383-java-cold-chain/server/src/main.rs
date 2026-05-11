use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;

mod handlers;
mod routes;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "COLD_CHAIN_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "COLD_CHAIN_HOST", default_value = "127.0.0.1")]
    host: String,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let storage = cold_chain_core::InMemoryStorage::new();
    let service = cold_chain_core::ColdChainService::new(Arc::new(storage));
    
    let app = routes::create_routes(Arc::new(service));
    
    let addr = format!("{}:{}", args.host, args.port)
        .parse::<SocketAddr>()
        .expect("Invalid host or port");
    
    println!("Cold Chain Server running on http://{}", addr);
    
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
