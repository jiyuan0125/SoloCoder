use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use community_core::{CreateOrderRequest, CreateOrderItemRequest, CreateProductRequest};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,
    
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreatePickupPoint {
        name: String,
        #[arg(long, default_value = "20:00")]
        cut_off_time: String,
    },
    ListPickupPoints,
    GetPickupPoint {
        id: Uuid,
    },
    CreateProduct {
        pickup_point_id: Uuid,
        name: String,
        #[arg(long)]
        unit_price: i64,
        #[arg(long)]
        stock: i64,
    },
    ListProducts {
        pickup_point_id: Uuid,
    },
    CreateOrder {
        pickup_point_id: Uuid,
        user_id: Uuid,
        #[arg(long, value_parser = parse_order_item, value_delimiter = ',', num_args = 1..)]
        items: Vec<CreateOrderItemRequest>,
    },
    GetOrder {
        id: Uuid,
    },
    ListUserOrders {
        user_id: Uuid,
    },
    PayOrder {
        order_id: Uuid,
    },
    CancelOrder {
        order_id: Uuid,
    },
    RefundOrder {
        order_id: Uuid,
    },
    PickupOrder {
        order_id: Uuid,
    },
    ProcessCutOff {
        pickup_point_id: Uuid,
    },
    ListSortingLists {
        pickup_point_id: Uuid,
    },
    ConfirmSorting {
        sorting_list_id: Uuid,
    },
}

fn parse_order_item(s: &str) -> Result<CreateOrderItemRequest, String> {
    let parts: Vec<&str> = s.split(':').collect();
    if parts.len() != 2 {
        return Err(format!("Invalid item format: {}. Expected product_id:quantity", s));
    }
    let product_id: Uuid = parts[0].parse().map_err(|e| format!("Invalid UUID: {}", e))?;
    let quantity: i64 = parts[1].parse().map_err(|e| format!("Invalid quantity: {}", e))?;
    Ok(CreateOrderItemRequest { product_id, quantity })
}

#[derive(Serialize, Deserialize, Debug)]
struct CreatePickupPointRequest {
    name: String,
    cut_off_time: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();
    
    match cli.command {
        Commands::CreatePickupPoint { name, cut_off_time } => {
            let req = CreatePickupPointRequest { name, cut_off_time };
            let resp = client.post(&format!("{}/api/pickup-points", base_url))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::ListPickupPoints => {
            let resp = client.get(&format!("{}/api/pickup-points", base_url))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::GetPickupPoint { id } => {
            let resp = client.get(&format!("{}/api/pickup-points/{}", base_url, id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::CreateProduct { pickup_point_id, name, unit_price, stock } => {
            let req = CreateProductRequest { name, unit_price, stock };
            let resp = client.post(&format!("{}/api/pickup-points/{}/products", base_url, pickup_point_id))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::ListProducts { pickup_point_id } => {
            let resp = client.get(&format!("{}/api/pickup-points/{}/products", base_url, pickup_point_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::CreateOrder { pickup_point_id, user_id, items } => {
            let req = CreateOrderRequest { user_id, items };
            let resp = client.post(&format!("{}/api/pickup-points/{}/orders", base_url, pickup_point_id))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::GetOrder { id } => {
            let resp = client.get(&format!("{}/api/orders/{}", base_url, id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::ListUserOrders { user_id } => {
            let resp = client.get(&format!("{}/api/users/{}/orders", base_url, user_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::PayOrder { order_id } => {
            let resp = client.post(&format!("{}/api/orders/{}/pay", base_url, order_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::CancelOrder { order_id } => {
            let resp = client.post(&format!("{}/api/orders/{}/cancel", base_url, order_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::RefundOrder { order_id } => {
            let resp = client.post(&format!("{}/api/orders/{}/refund", base_url, order_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::PickupOrder { order_id } => {
            let resp = client.post(&format!("{}/api/orders/{}/pickup", base_url, order_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::ProcessCutOff { pickup_point_id } => {
            let resp = client.post(&format!("{}/api/pickup-points/{}/cut-off", base_url, pickup_point_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::ListSortingLists { pickup_point_id } => {
            let resp = client.get(&format!("{}/api/pickup-points/{}/sorting-lists", base_url, pickup_point_id))
                .send()
                .await?;
            print_response(resp).await;
        }
        Commands::ConfirmSorting { sorting_list_id } => {
            let resp = client.post(&format!("{}/api/sorting-lists/{}/confirm", base_url, sorting_list_id))
                .send()
                .await?;
            print_response(resp).await;
        }
    }
    
    Ok(())
}

async fn print_response(resp: reqwest::Response) {
    let status = resp.status();
    let body = resp.text().await.unwrap_or_default();
    
    if status.is_success() {
        if let Ok(json) = serde_json::from_str::<serde_json::Value>(&body) {
            println!("{}", serde_json::to_string_pretty(&json).unwrap_or(body));
        } else {
            println!("{}", body);
        }
    } else {
        eprintln!("Error {}:", status);
        if let Ok(json) = serde_json::from_str::<serde_json::Value>(&body) {
            eprintln!("{}", serde_json::to_string_pretty(&json).unwrap_or(body));
        } else {
            eprintln!("{}", body);
        }
        std::process::exit(1);
    }
}
