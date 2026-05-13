use std::env;
use std::path::PathBuf;
use std::sync::Arc;

use tokio::net::TcpListener;
use tracing_subscriber::{EnvFilter, fmt};

use stream_mq::{AppState, TopicManager, create_router};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();
    
    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse()
        .expect("PORT must be a valid port number");
    
    let data_dir = env::var("DATA_DIR")
        .map(PathBuf::from)
        .unwrap_or_else(|_| {
            env::current_dir().unwrap().join("data")
        });
    
    tracing::info!("Starting stream-mq server on port {}", port);
    tracing::info!("Data directory: {:?}", data_dir);
    
    let topic_manager = TopicManager::new(data_dir)?;
    let app_state = AppState {
        topic_manager: Arc::new(topic_manager),
    };
    
    let router = create_router(app_state);
    let address = format!("0.0.0.0:{}", port);
    let listener = TcpListener::bind(&address).await?;
    
    tracing::info!("Server listening on {}", address);
    
    axum::serve(listener, router).await?;
    
    Ok(())
}
