mod models;
mod probes;
mod scheduler;
mod server;
mod state;

use std::env;

use state::AppState;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .expect("PORT must be a valid port number");

    let state = AppState::new();
    state.start_all_schedulers();

    server::run_server(port, state).await
}
