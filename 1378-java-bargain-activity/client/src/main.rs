use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::Serialize;
use serde_json::Value;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "BARGAIN_SERVER", default_value = "http://localhost:8080")]
    server: String,
    
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateProduct {
        #[arg(short, long)]
        name: String,
        #[arg(short = 'o', long)]
        original_price: f64,
        #[arg(short = 'f', long)]
        floor_price: f64,
    },
    ListProducts,
    GetProduct {
        #[arg(short, long)]
        id: Uuid,
    },
    CreateUser {
        #[arg(short, long)]
        name: String,
    },
    StartBargain {
        #[arg(short = 'p', long)]
        product_id: Uuid,
        #[arg(short = 'u', long)]
        user_id: Uuid,
    },
    HelpBargain {
        #[arg(short = 'a', long)]
        activity_id: Uuid,
        #[arg(short = 'u', long)]
        user_id: Uuid,
    },
    GetActivity {
        #[arg(short, long)]
        id: Uuid,
    },
    Purchase {
        #[arg(short, long)]
        id: Uuid,
    },
    Abandon {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Serialize)]
struct CreateProductRequest {
    name: String,
    original_price: f64,
    floor_price: f64,
}

#[derive(Serialize)]
struct CreateUserRequest {
    name: String,
}

#[derive(Serialize)]
struct StartBargainRequest {
    product_id: Uuid,
    user_id: Uuid,
}

#[derive(Serialize)]
struct HelpBargainRequest {
    activity_id: Uuid,
    user_id: Uuid,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/');
    
    match args.command {
        Commands::CreateProduct { name, original_price, floor_price } => {
            let resp = client.post(format!("{}/api/products", base_url))
                .json(&CreateProductRequest { name, original_price, floor_price })
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::ListProducts => {
            let resp = client.get(format!("{}/api/products", base_url))
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::GetProduct { id } => {
            let resp = client.get(format!("{}/api/products/{}", base_url, id))
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::CreateUser { name } => {
            let resp = client.post(format!("{}/api/users", base_url))
                .json(&CreateUserRequest { name })
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::StartBargain { product_id, user_id } => {
            let resp = client.post(format!("{}/api/bargain/start", base_url))
                .json(&StartBargainRequest { product_id, user_id })
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::HelpBargain { activity_id, user_id } => {
            let resp = client.post(format!("{}/api/bargain/help", base_url))
                .json(&HelpBargainRequest { activity_id, user_id })
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::GetActivity { id } => {
            let resp = client.get(format!("{}/api/bargain/{}", base_url, id))
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::Purchase { id } => {
            let resp = client.post(format!("{}/api/bargain/{}/purchase", base_url, id))
                .send()
                .await?;
            handle_response(resp).await;
        }
        Commands::Abandon { id } => {
            let resp = client.post(format!("{}/api/bargain/{}/abandon", base_url, id))
                .send()
                .await?;
            handle_response(resp).await;
        }
    }
    
    Ok(())
}

async fn handle_response(resp: reqwest::Response) {
    let status = resp.status();
    match resp.json::<Value>().await {
        Ok(body) => {
            if status.is_success() {
                println!("{}", serde_json::to_string_pretty(&body).unwrap());
            } else {
                eprintln!("Error ({}): {}", status, serde_json::to_string_pretty(&body).unwrap());
            }
        }
        Err(e) => {
            eprintln!("Failed to parse response: {}", e);
        }
    }
}
