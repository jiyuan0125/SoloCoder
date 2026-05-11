use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use uuid::Uuid;

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
    Health,
    
    Live {
        #[command(subcommand)]
        action: LiveCommands,
    },

    User {
        #[command(subcommand)]
        action: UserCommands,
    },

    Product {
        #[command(subcommand)]
        action: ProductCommands,
    },

    Order {
        #[command(subcommand)]
        action: OrderCommands,
    },
}

#[derive(Subcommand, Debug)]
enum LiveCommands {
    Status,
    Start,
    End,
}

#[derive(Subcommand, Debug)]
enum UserCommands {
    Create {
        name: String,
    },
    List,
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ProductCommands {
    Create {
        name: String,
        #[arg(long)]
        original_price: f64,
        #[arg(long)]
        live_price: f64,
        #[arg(long)]
        total_stock: u32,
        #[arg(long)]
        purchase_limit: Option<u32>,
    },
    List,
    Get {
        id: Uuid,
    },
    AddStock {
        #[arg(long)]
        product_id: Uuid,
        #[arg(long)]
        additional_stock: u32,
    },
    SetSale {
        #[arg(long)]
        product_id: Uuid,
        #[arg(long)]
        on_sale: bool,
    },
}

#[derive(Subcommand, Debug)]
enum OrderCommands {
    Create {
        #[arg(long)]
        user_id: Uuid,
        #[arg(long)]
        product_id: Uuid,
        #[arg(long)]
        quantity: u32,
    },
    List,
    Get {
        id: Uuid,
    },
    Pay {
        order_id: Uuid,
    },
    Cancel {
        order_id: Uuid,
    },
    Ship {
        order_id: Uuid,
    },
    Complete {
        order_id: Uuid,
    },
    Refund {
        order_id: Uuid,
        #[arg(long)]
        is_return: bool,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateUserRequest {
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateProductRequest {
    name: String,
    original_price: f64,
    live_price: f64,
    total_stock: u32,
    purchase_limit: Option<u32>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateOrderRequest {
    user_id: Uuid,
    product_id: Uuid,
    quantity: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct PayOrderRequest {
    order_id: Uuid,
}

#[derive(Debug, Serialize, Deserialize)]
struct CancelOrderRequest {
    order_id: Uuid,
}

#[derive(Debug, Serialize, Deserialize)]
struct RefundOrderRequest {
    order_id: Uuid,
    is_return: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct UpdateProductStockRequest {
    product_id: Uuid,
    additional_stock: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct SetProductSaleStatusRequest {
    product_id: Uuid,
    is_on_sale: bool,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "client=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let cli = Cli::parse();
    let client = Client::new();

    match cli.command {
        Commands::Health => {
            let url = format!("{}/health", cli.server);
            let response = client.get(&url).send().await?;
            println!("Status: {}", response.status());
        }

        Commands::Live { action } => {
            match action {
                LiveCommands::Status => {
                    let url = format!("{}/live/status", cli.server);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
                LiveCommands::Start => {
                    let url = format!("{}/live/start", cli.server);
                    let response = client.post(&url).send().await?;
                    println!("Status: {}", response.status());
                }
                LiveCommands::End => {
                    let url = format!("{}/live/end", cli.server);
                    let response = client.post(&url).send().await?;
                    println!("Status: {}", response.status());
                }
            }
        }

        Commands::User { action } => {
            match action {
                UserCommands::Create { name } => {
                    let url = format!("{}/users", cli.server);
                    let request = CreateUserRequest { name };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
                UserCommands::List => {
                    let url = format!("{}/users", cli.server);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
                UserCommands::Get { id } => {
                    let url = format!("{}/users/{}", cli.server, id);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
            }
        }

        Commands::Product { action } => {
            match action {
                ProductCommands::Create { name, original_price, live_price, total_stock, purchase_limit } => {
                    let url = format!("{}/products", cli.server);
                    let request = CreateProductRequest {
                        name,
                        original_price,
                        live_price,
                        total_stock,
                        purchase_limit,
                    };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
                ProductCommands::List => {
                    let url = format!("{}/products", cli.server);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
                ProductCommands::Get { id } => {
                    let url = format!("{}/products/{}", cli.server, id);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
                ProductCommands::AddStock { product_id, additional_stock } => {
                    let url = format!("{}/products/stock", cli.server);
                    let request = UpdateProductStockRequest { product_id, additional_stock };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
                ProductCommands::SetSale { product_id, on_sale } => {
                    let url = format!("{}/products/sale", cli.server);
                    let request = SetProductSaleStatusRequest { product_id, is_on_sale: on_sale };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
            }
        }

        Commands::Order { action } => {
            match action {
                OrderCommands::Create { user_id, product_id, quantity } => {
                    let url = format!("{}/orders", cli.server);
                    let request = CreateOrderRequest { user_id, product_id, quantity };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::List => {
                    let url = format!("{}/orders", cli.server);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::Get { id } => {
                    let url = format!("{}/orders/{}", cli.server, id);
                    let response = client.get(&url).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::Pay { order_id } => {
                    let url = format!("{}/orders/pay", cli.server);
                    let request = PayOrderRequest { order_id };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::Cancel { order_id } => {
                    let url = format!("{}/orders/cancel", cli.server);
                    let request = CancelOrderRequest { order_id };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::Ship { order_id } => {
                    let url = format!("{}/orders/{}/ship", cli.server, order_id);
                    let response = client.post(&url).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::Complete { order_id } => {
                    let url = format!("{}/orders/{}/complete", cli.server, order_id);
                    let response = client.post(&url).send().await?;
                    handle_response(response).await?;
                }
                OrderCommands::Refund { order_id, is_return } => {
                    let url = format!("{}/orders/refund", cli.server);
                    let request = RefundOrderRequest { order_id, is_return };
                    let response = client.post(&url).json(&request).send().await?;
                    handle_response(response).await?;
                }
            }
        }
    }

    Ok(())
}

async fn handle_response(response: reqwest::Response) -> Result<(), Box<dyn std::error::Error>> {
    let status = response.status();
    let body = response.text().await?;
    
    if status.is_success() {
        if !body.is_empty() {
            let json: serde_json::Value = serde_json::from_str(&body)?;
            println!("{}", serde_json::to_string_pretty(&json)?);
        } else {
            println!("Success: {}", status);
        }
    } else {
        eprintln!("Error {}: {}", status, body);
    }

    Ok(())
}
