use clap::{Parser, Subcommand};
use reqwest::Client;
use uuid::Uuid;

use charging_core::models::{
    Charger, CreateChargerRequest, CreateOrderRequest, Order,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "CHARGING_SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    ListChargers,
    GetCharger { id: Uuid },
    AddCharger { name: String, power: u32 },
    ListOrders,
    GetOrder { id: Uuid },
    StartCharging { charger_id: Uuid, user_id: Uuid },
    EndCharging { order_id: Uuid },
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/');

    match cli.command {
        Commands::Health => {
            let response = client
                .get(format!("{}/health", base_url))
                .send()
                .await?;
            let text = response.text().await?;
            println!("Server health: {}", text);
        }
        Commands::ListChargers => {
            let chargers: Vec<Charger> = client
                .get(format!("{}/chargers", base_url))
                .send()
                .await?
                .json()
                .await?;
            println!("Chargers ({} total):", chargers.len());
            for charger in chargers {
                println!(
                    "  ID: {}, Name: {}, Power: {}kW, Status: {:?}",
                    charger.id, charger.name, charger.power_kw, charger.status
                );
            }
        }
        Commands::GetCharger { id } => {
            let response = client
                .get(format!("{}/chargers/{}", base_url, id))
                .send()
                .await?;
            
            if response.status().is_success() {
                let charger: Charger = response.json().await?;
                println!("Charger:");
                println!("  ID: {}", charger.id);
                println!("  Name: {}", charger.name);
                println!("  Power: {}kW", charger.power_kw);
                println!("  Status: {:?}", charger.status);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("Error: {}", error);
            }
        }
        Commands::AddCharger { name, power } => {
            let request = CreateChargerRequest { name, power_kw: power };
            let response = client
                .post(format!("{}/chargers", base_url))
                .json(&request)
                .send()
                .await?;
            
            if response.status().is_success() {
                let charger: Charger = response.json().await?;
                println!("Added charger:");
                println!("  ID: {}", charger.id);
                println!("  Name: {}", charger.name);
                println!("  Power: {}kW", charger.power_kw);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("Error: {}", error);
            }
        }
        Commands::ListOrders => {
            let orders: Vec<Order> = client
                .get(format!("{}/orders", base_url))
                .send()
                .await?
                .json()
                .await?;
            println!("Orders ({} total):", orders.len());
            for order in orders {
                print_order(&order);
            }
        }
        Commands::GetOrder { id } => {
            let response = client
                .get(format!("{}/orders/{}", base_url, id))
                .send()
                .await?;
            
            if response.status().is_success() {
                let order: Order = response.json().await?;
                print_order(&order);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("Error: {}", error);
            }
        }
        Commands::StartCharging { charger_id, user_id } => {
            let request = CreateOrderRequest { charger_id, user_id };
            let response = client
                .post(format!("{}/orders", base_url))
                .json(&request)
                .send()
                .await?;
            
            if response.status().is_success() {
                let order: Order = response.json().await?;
                println!("Started charging:");
                print_order(&order);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("Error: {}", error);
            }
        }
        Commands::EndCharging { order_id } => {
            let response = client
                .post(format!("{}/orders/{}/end", base_url, order_id))
                .send()
                .await?;
            
            if response.status().is_success() {
                let order: Order = response.json().await?;
                println!("Ended charging:");
                print_order(&order);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("Error: {}", error);
            }
        }
    }

    Ok(())
}

fn print_order(order: &Order) {
    println!("Order:");
    println!("  ID: {}", order.id);
    println!("  Charger ID: {}", order.charger_id);
    println!("  User ID: {}", order.user_id);
    println!("  Status: {:?}", order.status);
    println!("  Start Time: {}", order.start_time);
    if let Some(end_time) = order.end_time {
        println!("  End Time: {}", end_time);
    }
    if let Some(duration) = order.duration_minutes {
        println!("  Duration: {} minutes", duration);
    }
    if let Some(energy) = order.energy_kwh {
        println!("  Energy: {:.4} kWh", energy);
    }
    if let Some(cost) = order.total_cost {
        println!("  Total Cost: {:.2} yuan", cost);
    }
}
