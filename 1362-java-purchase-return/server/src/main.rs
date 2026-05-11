use clap::Parser;
use purchase_return_server::run_server;
use std::net::SocketAddr;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_HOST", default_value = "127.0.0.1")]
    host: String,

    #[arg(short, long, env = "SERVER_PORT", default_value_t = 8080)]
    port: u16,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    
    println!("Starting purchase return server on {}", addr);
    run_server(addr).await;
}
