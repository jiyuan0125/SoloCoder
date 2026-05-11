use clap::{Parser, Subcommand};
use colored::*;
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use serde_json::Value;

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
    Order {
        #[command(subcommand)]
        action: OrderCommands,
    },
    Promoter {
        #[command(subcommand)]
        action: PromoterCommands,
    },
    Settlement {
        #[command(subcommand)]
        action: SettlementCommands,
    },
    OrderTypes,
}

#[derive(Subcommand, Debug)]
enum OrderCommands {
    Create {
        #[arg(short, long)]
        order_type: String,
        #[arg(short, long)]
        amount: u64,
        #[arg(short, long)]
        merchant_id: Uuid,
        #[arg(short, long)]
        promoter_id: Option<Uuid>,
    },
    Complete {
        #[arg(short, long)]
        order_id: Uuid,
        #[arg(short, long)]
        completed_date: String,
    },
    Get {
        #[arg(short, long)]
        order_id: Uuid,
    },
    List,
}

#[derive(Subcommand, Debug)]
enum PromoterCommands {
    Register {
        #[arg(short, long)]
        name: String,
    },
    BindAccount {
        #[arg(short, long)]
        promoter_id: Uuid,
    },
    Get {
        #[arg(short, long)]
        promoter_id: Uuid,
    },
    List,
}

#[derive(Subcommand, Debug)]
enum SettlementCommands {
    Process {
        #[arg(short, long)]
        settlement_date: String,
    },
    Get {
        #[arg(short, long)]
        settlement_id: Uuid,
    },
    List,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

#[derive(Debug, Serialize)]
struct CreateOrderRequest {
    order_type: String,
    amount: u64,
    merchant_id: Uuid,
    promoter_id: Option<Uuid>,
}

#[derive(Debug, Serialize)]
struct CompleteOrderRequest {
    order_id: Uuid,
    completed_date: String,
}

#[derive(Debug, Serialize)]
struct RegisterPromoterRequest {
    name: String,
}

#[derive(Debug, Serialize)]
struct BindAccountRequest {
    promoter_id: Uuid,
}

#[derive(Debug, Serialize)]
struct ProcessSettlementRequest {
    settlement_date: String,
}

fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();
    
    match cli.command {
        Commands::Order { action } => handle_order_commands(&client, &base_url, action),
        Commands::Promoter { action } => handle_promoter_commands(&client, &base_url, action),
        Commands::Settlement { action } => handle_settlement_commands(&client, &base_url, action),
        Commands::OrderTypes => handle_order_types(&client, &base_url),
    }
}

fn handle_order_commands(client: &Client, base_url: &str, action: OrderCommands) -> anyhow::Result<()> {
    match action {
        OrderCommands::Create { order_type, amount, merchant_id, promoter_id } => {
            let request = CreateOrderRequest {
                order_type,
                amount,
                merchant_id,
                promoter_id,
            };
            
            let response: ApiResponse<Value> = client
                .post(format!("{}/api/orders", base_url))
                .json(&request)
                .send()?
                .json()?;
            
            print_response(response);
        }
        OrderCommands::Complete { order_id, completed_date } => {
            let request = CompleteOrderRequest {
                order_id,
                completed_date,
            };
            
            let response: ApiResponse<Value> = client
                .post(format!("{}/api/orders/complete", base_url))
                .json(&request)
                .send()?
                .json()?;
            
            print_response(response);
        }
        OrderCommands::Get { order_id } => {
            let response: ApiResponse<Value> = client
                .get(format!("{}/api/orders/{}", base_url, order_id))
                .send()?
                .json()?;
            
            print_response(response);
        }
        OrderCommands::List => {
            let response: ApiResponse<Value> = client
                .get(format!("{}/api/orders", base_url))
                .send()?
                .json()?;
            
            print_response(response);
        }
    }
    Ok(())
}

fn handle_promoter_commands(client: &Client, base_url: &str, action: PromoterCommands) -> anyhow::Result<()> {
    match action {
        PromoterCommands::Register { name } => {
            let request = RegisterPromoterRequest { name };
            
            let response: ApiResponse<Value> = client
                .post(format!("{}/api/promoters", base_url))
                .json(&request)
                .send()?
                .json()?;
            
            print_response(response);
        }
        PromoterCommands::BindAccount { promoter_id } => {
            let request = BindAccountRequest { promoter_id };
            
            let response: ApiResponse<Value> = client
                .post(format!("{}/api/promoters/bind-account", base_url))
                .json(&request)
                .send()?
                .json()?;
            
            print_response(response);
        }
        PromoterCommands::Get { promoter_id } => {
            let response: ApiResponse<Value> = client
                .get(format!("{}/api/promoters/{}", base_url, promoter_id))
                .send()?
                .json()?;
            
            print_response(response);
        }
        PromoterCommands::List => {
            let response: ApiResponse<Value> = client
                .get(format!("{}/api/promoters", base_url))
                .send()?
                .json()?;
            
            print_response(response);
        }
    }
    Ok(())
}

fn handle_settlement_commands(client: &Client, base_url: &str, action: SettlementCommands) -> anyhow::Result<()> {
    match action {
        SettlementCommands::Process { settlement_date } => {
            let request = ProcessSettlementRequest { settlement_date };
            
            let response: ApiResponse<Value> = client
                .post(format!("{}/api/settlements", base_url))
                .json(&request)
                .send()?
                .json()?;
            
            print_response(response);
        }
        SettlementCommands::Get { settlement_id } => {
            let response: ApiResponse<Value> = client
                .get(format!("{}/api/settlements/{}", base_url, settlement_id))
                .send()?
                .json()?;
            
            print_response(response);
        }
        SettlementCommands::List => {
            let response: ApiResponse<Value> = client
                .get(format!("{}/api/settlements", base_url))
                .send()?
                .json()?;
            
            print_response(response);
        }
    }
    Ok(())
}

fn handle_order_types(client: &Client, base_url: &str) -> anyhow::Result<()> {
    let response: ApiResponse<Value> = client
        .get(format!("{}/api/order-types", base_url))
        .send()?
        .json()?;
    
    print_response(response);
    Ok(())
}

fn print_response<T: Serialize>(response: ApiResponse<T>) {
    if response.success {
        if let Some(data) = response.data {
            println!("{}", "Success!".green().bold());
            println!("{}", serde_json::to_string_pretty(&data).unwrap());
        } else {
            println!("{}", "Success!".green().bold());
        }
    } else {
        if let Some(error) = response.error {
            eprintln!("{} {}", "Error:".red().bold(), error);
        } else {
            eprintln!("{}", "Unknown error occurred".red().bold());
        }
    }
}
