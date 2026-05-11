use clap::{Parser, Subcommand};
use flash_sale_core::models::{ActivityStatistics, FlashSaleActivity, Order};
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateActivity {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        price: f64,
        #[arg(short, long)]
        stock: u32,
        #[arg(short, long)]
        start: chrono::DateTime<chrono::Utc>,
        #[arg(short, long)]
        end: chrono::DateTime<chrono::Utc>,
    },
    ListActivities,
    GetActivity {
        id: String,
    },
    CreateOrder {
        activity_id: String,
        user_id: String,
    },
    ListOrders,
    GetOrder {
        id: String,
    },
    PayOrder {
        id: String,
    },
    CancelOrder {
        id: String,
    },
    GetStatistics {
        activity_id: String,
    },
}

#[derive(Debug, Serialize)]
struct CreateActivityRequest {
    product_name: String,
    flash_price: f64,
    stock: u32,
    start_time: chrono::DateTime<chrono::Utc>,
    end_time: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Serialize)]
struct CreateOrderRequest {
    user_id: String,
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = reqwest::Client::new();

    match cli.command {
        Commands::CreateActivity { name, price, stock, start, end } => {
            let req = CreateActivityRequest {
                product_name: name,
                flash_price: price,
                stock,
                start_time: start,
                end_time: end,
            };
            let resp = client
                .post(&format!("{}/api/activities", cli.server))
                .json(&req)
                .send()
                .await;
            handle_response::<FlashSaleActivity>(resp).await;
        }
        Commands::ListActivities => {
            let resp = client
                .get(&format!("{}/api/activities", cli.server))
                .send()
                .await;
            handle_response::<Vec<FlashSaleActivity>>(resp).await;
        }
        Commands::GetActivity { id } => {
            let resp = client
                .get(&format!("{}/api/activities/{}", cli.server, id))
                .send()
                .await;
            handle_response::<FlashSaleActivity>(resp).await;
        }
        Commands::CreateOrder { activity_id, user_id } => {
            let req = CreateOrderRequest { user_id };
            let resp = client
                .post(&format!("{}/api/activities/{}/orders", cli.server, activity_id))
                .json(&req)
                .send()
                .await;
            handle_response::<Order>(resp).await;
        }
        Commands::ListOrders => {
            let resp = client
                .get(&format!("{}/api/orders", cli.server))
                .send()
                .await;
            handle_response::<Vec<Order>>(resp).await;
        }
        Commands::GetOrder { id } => {
            let resp = client
                .get(&format!("{}/api/orders/{}", cli.server, id))
                .send()
                .await;
            handle_response::<Order>(resp).await;
        }
        Commands::PayOrder { id } => {
            let resp = client
                .post(&format!("{}/api/orders/{}/pay", cli.server, id))
                .send()
                .await;
            handle_response::<Order>(resp).await;
        }
        Commands::CancelOrder { id } => {
            let resp = client
                .post(&format!("{}/api/orders/{}/cancel", cli.server, id))
                .send()
                .await;
            handle_response::<Order>(resp).await;
        }
        Commands::GetStatistics { activity_id } => {
            let resp = client
                .get(&format!("{}/api/activities/{}/statistics", cli.server, activity_id))
                .send()
                .await;
            handle_response::<ActivityStatistics>(resp).await;
        }
    }
}

async fn handle_response<T: serde::de::DeserializeOwned + serde::Serialize>(
    result: reqwest::Result<reqwest::Response>,
) {
    match result {
        Ok(resp) => {
            if resp.status().is_success() {
                let data: T = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&data).unwrap());
            } else {
                let err: ErrorResponse = resp.json().await.unwrap();
                eprintln!("Error: {}", err.error);
            }
        }
        Err(e) => {
            eprintln!("Request failed: {}", e);
        }
    }
}
