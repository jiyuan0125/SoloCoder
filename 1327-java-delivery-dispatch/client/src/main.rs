use clap::{Parser, Subcommand};
use delivery_dispatch_core::{Location, Order, Rider, Merchant};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Rider {
        #[command(subcommand)]
        subcommand: RiderCommands,
    },
    Merchant {
        #[command(subcommand)]
        subcommand: MerchantCommands,
    },
    Order {
        #[command(subcommand)]
        subcommand: OrderCommands,
    },
    Stats,
    Queue,
}

#[derive(Subcommand, Debug)]
enum RiderCommands {
    Add {
        name: String,
        #[arg(long)]
        lat: f64,
        #[arg(long)]
        lng: f64,
    },
    List,
    Get {
        id: Uuid,
    },
    UpdateLocation {
        id: Uuid,
        #[arg(long)]
        lat: f64,
        #[arg(long)]
        lng: f64,
    },
}

#[derive(Subcommand, Debug)]
enum MerchantCommands {
    Add {
        name: String,
        #[arg(long)]
        lat: f64,
        #[arg(long)]
        lng: f64,
    },
    List,
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum OrderCommands {
    Create {
        merchant_id: Uuid,
    },
    List,
    Get {
        id: Uuid,
    },
    Complete {
        id: Uuid,
    },
}

#[derive(Debug, Deserialize, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

#[derive(Debug, Deserialize, Serialize)]
struct CreateRiderRequest {
    name: String,
    location: Location,
}

#[derive(Debug, Deserialize, Serialize)]
struct CreateMerchantRequest {
    name: String,
    location: Location,
}

#[derive(Debug, Deserialize, Serialize)]
struct CreateOrderRequest {
    merchant_id: Uuid,
}

#[derive(Debug, Deserialize, Serialize)]
struct UpdateRiderLocationRequest {
    location: Location,
}

async fn handle_rider_commands(
    client: &Client,
    base_url: &str,
    subcommand: RiderCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match subcommand {
        RiderCommands::Add { name, lat, lng } => {
            let url = format!("{}/riders", base_url);
            let payload = CreateRiderRequest {
                name,
                location: Location {
                    latitude: lat,
                    longitude: lng,
                },
            };
            let response = client
                .post(&url)
                .json(&payload)
                .send()
                .await?
                .json::<ApiResponse<Rider>>()
                .await?;
            if response.success {
                if let Some(rider) = response.data {
                    println!("Rider created:");
                    println!("  ID: {}", rider.id);
                    println!("  Name: {}", rider.name);
                    println!("  Location: ({}, {})", rider.location.latitude, rider.location.longitude);
                    println!("  Status: {:?}", rider.status);
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        RiderCommands::List => {
            let url = format!("{}/riders", base_url);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Vec<Rider>>>()
                .await?;
            if response.success {
                if let Some(riders) = response.data {
                    println!("Riders ({} total):", riders.len());
                    for rider in riders {
                        println!("  - {} ({}) - {:?} - {} orders", rider.name, rider.id, rider.status, rider.current_orders.len());
                    }
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        RiderCommands::Get { id } => {
            let url = format!("{}/riders/{}", base_url, id);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Rider>>()
                .await?;
            if response.success {
                if let Some(rider) = response.data {
                    println!("Rider:");
                    println!("  ID: {}", rider.id);
                    println!("  Name: {}", rider.name);
                    println!("  Location: ({}, {})", rider.location.latitude, rider.location.longitude);
                    println!("  Status: {:?}", rider.status);
                    println!("  Current orders: {}", rider.current_orders.len());
                    for order_id in rider.current_orders {
                        println!("    - {}", order_id);
                    }
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        RiderCommands::UpdateLocation { id, lat, lng } => {
            let url = format!("{}/riders/{}", base_url, id);
            let payload = UpdateRiderLocationRequest {
                location: Location {
                    latitude: lat,
                    longitude: lng,
                },
            };
            let response = client
                .put(&url)
                .json(&payload)
                .send()
                .await?
                .json::<ApiResponse<Rider>>()
                .await?;
            if response.success {
                println!("Rider location updated.");
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
    }
    Ok(())
}

async fn handle_merchant_commands(
    client: &Client,
    base_url: &str,
    subcommand: MerchantCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match subcommand {
        MerchantCommands::Add { name, lat, lng } => {
            let url = format!("{}/merchants", base_url);
            let payload = CreateMerchantRequest {
                name,
                location: Location {
                    latitude: lat,
                    longitude: lng,
                },
            };
            let response = client
                .post(&url)
                .json(&payload)
                .send()
                .await?
                .json::<ApiResponse<Merchant>>()
                .await?;
            if response.success {
                if let Some(merchant) = response.data {
                    println!("Merchant created:");
                    println!("  ID: {}", merchant.id);
                    println!("  Name: {}", merchant.name);
                    println!("  Location: ({}, {})", merchant.location.latitude, merchant.location.longitude);
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        MerchantCommands::List => {
            let url = format!("{}/merchants", base_url);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Vec<Merchant>>>()
                .await?;
            if response.success {
                if let Some(merchants) = response.data {
                    println!("Merchants ({} total):", merchants.len());
                    for merchant in merchants {
                        println!("  - {} ({})", merchant.name, merchant.id);
                    }
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        MerchantCommands::Get { id } => {
            let url = format!("{}/merchants/{}", base_url, id);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Merchant>>()
                .await?;
            if response.success {
                if let Some(merchant) = response.data {
                    println!("Merchant:");
                    println!("  ID: {}", merchant.id);
                    println!("  Name: {}", merchant.name);
                    println!("  Location: ({}, {})", merchant.location.latitude, merchant.location.longitude);
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
    }
    Ok(())
}

async fn handle_order_commands(
    client: &Client,
    base_url: &str,
    subcommand: OrderCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match subcommand {
        OrderCommands::Create { merchant_id } => {
            let url = format!("{}/orders", base_url);
            let payload = CreateOrderRequest { merchant_id };
            let response = client
                .post(&url)
                .json(&payload)
                .send()
                .await?
                .json::<ApiResponse<Order>>()
                .await?;
            if response.success {
                if let Some(order) = response.data {
                    println!("Order created:");
                    println!("  ID: {}", order.id);
                    println!("  Merchant ID: {}", order.merchant_id);
                    println!("  Status: {:?}", order.status);
                    if let Some(rider_id) = order.assigned_rider_id {
                        println!("  Assigned to: {}", rider_id);
                    } else {
                        println!("  Assigned to: Waiting in queue");
                    }
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        OrderCommands::List => {
            let url = format!("{}/orders", base_url);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Vec<Order>>>()
                .await?;
            if response.success {
                if let Some(orders) = response.data {
                    println!("Orders ({} total):", orders.len());
                    for order in orders {
                        let rider_info = match order.assigned_rider_id {
                            Some(id) => format!("assigned to {}", id),
                            None => "waiting".to_string(),
                        };
                        println!("  - {} ({}) - {:?} - {}", order.id, order.merchant_id, order.status, rider_info);
                    }
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        OrderCommands::Get { id } => {
            let url = format!("{}/orders/{}", base_url, id);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Order>>()
                .await?;
            if response.success {
                if let Some(order) = response.data {
                    println!("Order:");
                    println!("  ID: {}", order.id);
                    println!("  Merchant ID: {}", order.merchant_id);
                    println!("  Status: {:?}", order.status);
                    if let Some(rider_id) = order.assigned_rider_id {
                        println!("  Assigned to: {}", rider_id);
                    } else {
                        println!("  Assigned to: Waiting in queue");
                    }
                    println!("  Created at: {}", order.created_at);
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        OrderCommands::Complete { id } => {
            let url = format!("{}/orders/{}", base_url, id);
            let response = client
                .post(&url)
                .send()
                .await?
                .json::<ApiResponse<Order>>()
                .await?;
            if response.success {
                println!("Order completed.");
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
    }
    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Rider { subcommand } => {
            handle_rider_commands(&client, &base_url, subcommand).await?;
        }
        Commands::Merchant { subcommand } => {
            handle_merchant_commands(&client, &base_url, subcommand).await?;
        }
        Commands::Order { subcommand } => {
            handle_order_commands(&client, &base_url, subcommand).await?;
        }
        Commands::Stats => {
            let url = format!("{}/stats", base_url);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<serde_json::Value>>()
                .await?;
            if response.success {
                if let Some(stats) = response.data {
                    println!("System Statistics:");
                    println!("{}", serde_json::to_string_pretty(&stats)?);
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
        Commands::Queue => {
            let url = format!("{}/waiting-queue", base_url);
            let response = client
                .get(&url)
                .send()
                .await?
                .json::<ApiResponse<Vec<Uuid>>>()
                .await?;
            if response.success {
                if let Some(queue) = response.data {
                    println!("Waiting queue ({} orders):", queue.len());
                    for (i, order_id) in queue.iter().enumerate() {
                        println!("  {}. {}", i + 1, order_id);
                    }
                }
            } else {
                eprintln!("Error: {}", response.message.unwrap_or_else(|| "Unknown error".to_string()));
            }
        }
    }

    Ok(())
}
