use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde::de::DeserializeOwned;
use std::time::Duration;
use uuid::Uuid;

use lab_core::models::{
    ConfirmBookingRequest, ConsumableUsage, CreateBookingRequest, CreateConsumableRequest,
    CreateEquipmentRequest, CreateUserRequest, UpdateConsumableStockRequest, UserRole,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    User {
        #[command(subcommand)]
        command: UserCommands,
    },
    Equipment {
        #[command(subcommand)]
        command: EquipmentCommands,
    },
    Consumable {
        #[command(subcommand)]
        command: ConsumableCommands,
    },
    Booking {
        #[command(subcommand)]
        command: BookingCommands,
    },
    Alerts,
}

#[derive(Subcommand, Debug)]
enum UserCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long, value_enum)]
        role: UserRoleArg,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(clap::ValueEnum, Debug, Clone, Copy)]
enum UserRoleArg {
    Researcher,
    Admin,
}

impl From<UserRoleArg> for UserRole {
    fn from(val: UserRoleArg) -> Self {
        match val {
            UserRoleArg::Researcher => UserRole::Researcher,
            UserRoleArg::Admin => UserRole::Admin,
        }
    }
}

#[derive(Subcommand, Debug)]
enum EquipmentCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        description: Option<String>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ConsumableCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long, default_value_t = 0)]
        initial_stock: u32,
        #[arg(short, long, default_value_t = 10)]
        safety_stock: u32,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Restock {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        quantity: u32,
    },
}

#[derive(Subcommand, Debug)]
enum BookingCommands {
    Create {
        #[arg(short, long)]
        equipment_id: Uuid,
        #[arg(short, long)]
        researcher_id: Uuid,
        #[arg(long)]
        start: String,
        #[arg(long)]
        end: String,
        #[arg(long, value_delimiter = ',')]
        consumables: Vec<String>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Confirm {
        #[arg(short, long)]
        booking_id: Uuid,
        #[arg(short, long)]
        admin_id: Uuid,
    },
    Start {
        #[arg(short, long)]
        id: Uuid,
    },
    Complete {
        #[arg(short, long)]
        id: Uuid,
    },
    Cancel {
        #[arg(short, long)]
        id: Uuid,
    },
}

struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("Failed to create HTTP client");
        Self { client, base_url }
    }

    fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        let resp = self
            .client
            .get(&url)
            .send()
            .map_err(|e| format!("Request failed: {}", e))?;
        if !resp.status().is_success() {
            let body = resp.text().map_err(|e| format!("Failed to read response: {}", e))?;
            return Err(format!("API error: {}", body));
        }
        resp.json().map_err(|e| format!("Failed to parse response: {}", e))
    }

    fn post<T: DeserializeOwned, B: serde::Serialize>(&self, path: &str, body: &B) -> Result<T, String> {
        let url = format!("{}{}", self.base_url, path);
        let resp = self
            .client
            .post(&url)
            .json(body)
            .send()
            .map_err(|e| format!("Request failed: {}", e))?;
        if !resp.status().is_success() {
            let body = resp.text().map_err(|e| format!("Failed to read response: {}", e))?;
            return Err(format!("API error: {}", body));
        }
        resp.json().map_err(|e| format!("Failed to parse response: {}", e))
    }
}

fn parse_consumables(inputs: &[String]) -> Result<Vec<ConsumableUsage>, String> {
    inputs
        .iter()
        .map(|s| {
            let parts: Vec<&str> = s.split(':').collect();
            if parts.len() != 2 {
                return Err(format!("Invalid consumable format: {}. Expected id:quantity", s));
            }
            let id = parts[0]
                .parse::<Uuid>()
                .map_err(|e| format!("Invalid UUID: {}", e))?;
            let quantity = parts[1]
                .parse::<u32>()
                .map_err(|e| format!("Invalid quantity: {}", e))?;
            Ok(ConsumableUsage {
                consumable_id: id,
                quantity,
            })
        })
        .collect()
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let api = ApiClient::new(args.server);

    match args.command {
        Commands::User { command } => match command {
            UserCommands::Create { name, role } => {
                let req = CreateUserRequest {
                    name,
                    role: role.into(),
                };
                let result: serde_json::Value = api.post("/api/users", &req)?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            UserCommands::List => {
                let result: serde_json::Value = api.get("/api/users")?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            UserCommands::Get { id } => {
                let result: serde_json::Value = api.get(&format!("/api/users/{}", id))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
        },
        Commands::Equipment { command } => match command {
            EquipmentCommands::Create { name, description } => {
                let req = CreateEquipmentRequest { name, description };
                let result: serde_json::Value = api.post("/api/equipment", &req)?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            EquipmentCommands::List => {
                let result: serde_json::Value = api.get("/api/equipment")?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            EquipmentCommands::Get { id } => {
                let result: serde_json::Value = api.get(&format!("/api/equipment/{}", id))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
        },
        Commands::Consumable { command } => match command {
            ConsumableCommands::Create {
                name,
                initial_stock,
                safety_stock,
            } => {
                let req = CreateConsumableRequest {
                    name,
                    initial_stock,
                    safety_stock,
                };
                let result: serde_json::Value = api.post("/api/consumables", &req)?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            ConsumableCommands::List => {
                let result: serde_json::Value = api.get("/api/consumables")?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            ConsumableCommands::Get { id } => {
                let result: serde_json::Value = api.get(&format!("/api/consumables/{}", id))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            ConsumableCommands::Restock { id, quantity } => {
                let req = UpdateConsumableStockRequest { quantity };
                let result: serde_json::Value = api.post(&format!("/api/consumables/{}/restock", id), &req)?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
        },
        Commands::Booking { command } => match command {
            BookingCommands::Create {
                equipment_id,
                researcher_id,
                start,
                end,
                consumables,
            } => {
                let start_time = chrono::DateTime::parse_from_rfc3339(&start)
                    .map_err(|e| format!("Invalid start time: {}", e))?
                    .with_timezone(&chrono::Utc);
                let end_time = chrono::DateTime::parse_from_rfc3339(&end)
                    .map_err(|e| format!("Invalid end time: {}", e))?
                    .with_timezone(&chrono::Utc);
                let consumables = parse_consumables(&consumables)?;
                let req = CreateBookingRequest {
                    equipment_id,
                    researcher_id,
                    start_time,
                    end_time,
                    consumables,
                };
                let result: serde_json::Value = api.post("/api/bookings", &req)?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            BookingCommands::List => {
                let result: serde_json::Value = api.get("/api/bookings")?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            BookingCommands::Get { id } => {
                let result: serde_json::Value = api.get(&format!("/api/bookings/{}", id))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            BookingCommands::Confirm { booking_id, admin_id } => {
                let req = ConfirmBookingRequest { admin_id };
                let result: serde_json::Value = api.post(&format!("/api/bookings/{}/confirm", booking_id), &req)?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            BookingCommands::Start { id } => {
                let result: serde_json::Value = api.post(&format!("/api/bookings/{}/start", id), &serde_json::json!({}))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            BookingCommands::Complete { id } => {
                let result: serde_json::Value = api.post(&format!("/api/bookings/{}/complete", id), &serde_json::json!({}))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
            BookingCommands::Cancel { id } => {
                let result: serde_json::Value = api.post(&format!("/api/bookings/{}/cancel", id), &serde_json::json!({}))?;
                println!("{}", serde_json::to_string_pretty(&result).unwrap());
            }
        },
        Commands::Alerts => {
            let result: serde_json::Value = api.get("/api/alerts/low-stock")?;
            println!("{}", serde_json::to_string_pretty(&result).unwrap());
        }
    }
    Ok(())
}
