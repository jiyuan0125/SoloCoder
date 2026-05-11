use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use warehouse_core::models::{BatchOutboundItem, Material, MaterialInventory, InboundRecord, OutboundRecord, ReplenishmentOrder};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "WAREHOUSE_SERVER", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Material {
        #[command(subcommand)]
        action: MaterialCommands,
    },
    Inventory {
        #[command(subcommand)]
        action: InventoryCommands,
    },
    Inbound {
        material_id: Uuid,
        quantity: i32,
        supplier: String,
        batch_number: String,
    },
    Outbound {
        material_id: Uuid,
        quantity: i32,
    },
    BatchOutbound {
        items: String,
    },
    Records {
        #[command(subcommand)]
        action: RecordCommands,
    },
    Replenishment {
        #[command(subcommand)]
        action: ReplenishmentCommands,
    },
}

#[derive(Subcommand, Debug)]
enum MaterialCommands {
    List,
    Add {
        name: String,
        safety_stock: i32,
        lead_time_days: u32,
        unit: String,
    },
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum InventoryCommands {
    List,
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum RecordCommands {
    Inbound {
        #[arg(long)]
        material_id: Option<Uuid>,
    },
    Outbound {
        #[arg(long)]
        material_id: Option<Uuid>,
    },
}

#[derive(Subcommand, Debug)]
enum ReplenishmentCommands {
    List {
        #[arg(long)]
        material_id: Option<Uuid>,
        #[arg(long)]
        status: Option<String>,
    },
    Confirm {
        id: Uuid,
    },
    Cancel {
        id: Uuid,
    },
}

#[derive(Debug, Deserialize)]
struct BatchOutboundJsonItem {
    material_id: Uuid,
    quantity: i32,
}

fn print_json<T: Serialize>(data: &T) {
    println!("{}", serde_json::to_string_pretty(data).unwrap());
}

fn print_error(status: reqwest::StatusCode, body: &str) {
    eprintln!("请求失败: {}", status);
    if !body.is_empty() {
        eprintln!("{}", body);
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/');

    match &cli.command {
        Commands::Material { action } => match action {
            MaterialCommands::List => {
                let url = format!("{}/materials", base_url);
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let materials: Vec<Material> = resp.json().await.unwrap();
                    print_json(&materials);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
            MaterialCommands::Add {
                name,
                safety_stock,
                lead_time_days,
                unit,
            } => {
                #[derive(Serialize)]
                struct Req<'a> {
                    name: &'a str,
                    safety_stock: i32,
                    lead_time_days: u32,
                    unit: &'a str,
                }

                let url = format!("{}/materials", base_url);
                let req = Req {
                    name,
                    safety_stock: *safety_stock,
                    lead_time_days: *lead_time_days,
                    unit,
                };
                let resp = client.post(&url).json(&req).send().await.unwrap();
                if resp.status().is_success() {
                    let material: Material = resp.json().await.unwrap();
                    print_json(&material);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
            MaterialCommands::Get { id } => {
                let url = format!("{}/materials/{}", base_url, id);
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let material: Material = resp.json().await.unwrap();
                    print_json(&material);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
        },
        Commands::Inventory { action } => match action {
            InventoryCommands::List => {
                let url = format!("{}/inventories", base_url);
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let inventories: Vec<MaterialInventory> = resp.json().await.unwrap();
                    print_json(&inventories);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
            InventoryCommands::Get { id } => {
                let url = format!("{}/inventories/{}", base_url, id);
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let inventory: MaterialInventory = resp.json().await.unwrap();
                    print_json(&inventory);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
        },
        Commands::Inbound {
            material_id,
            quantity,
            supplier,
            batch_number,
        } => {
            #[derive(Serialize)]
            struct Req<'a> {
                material_id: Uuid,
                quantity: i32,
                supplier: &'a str,
                batch_number: &'a str,
            }

            let url = format!("{}/inbound", base_url);
            let req = Req {
                material_id: *material_id,
                quantity: *quantity,
                supplier,
                batch_number,
            };
            let resp = client.post(&url).json(&req).send().await.unwrap();
            if resp.status().is_success() {
                let record: InboundRecord = resp.json().await.unwrap();
                print_json(&record);
            } else {
                print_error(resp.status(), &resp.text().await.unwrap());
            }
        }
        Commands::Outbound {
            material_id,
            quantity,
        } => {
            #[derive(Serialize)]
            struct Req {
                material_id: Uuid,
                quantity: i32,
            }

            let url = format!("{}/outbound", base_url);
            let req = Req {
                material_id: *material_id,
                quantity: *quantity,
            };
            let resp = client.post(&url).json(&req).send().await.unwrap();
            if resp.status().is_success() {
                let record: OutboundRecord = resp.json().await.unwrap();
                print_json(&record);
            } else {
                print_error(resp.status(), &resp.text().await.unwrap());
            }
        }
        Commands::BatchOutbound { items } => {
            let parsed_items: Vec<BatchOutboundItem> =
                serde_json::from_str(items).expect("items 必须是有效的 JSON 数组");

            #[derive(Serialize)]
            struct Req {
                items: Vec<BatchOutboundItem>,
            }

            let url = format!("{}/outbound/batch", base_url);
            let req = Req { items: parsed_items };
            let resp = client.post(&url).json(&req).send().await.unwrap();
            if resp.status().is_success() {
                let records: Vec<OutboundRecord> = resp.json().await.unwrap();
                print_json(&records);
            } else {
                print_error(resp.status(), &resp.text().await.unwrap());
            }
        }
        Commands::Records { action } => match action {
            RecordCommands::Inbound { material_id } => {
                let mut url = format!("{}/records/inbound", base_url);
                if let Some(id) = material_id {
                    url = format!("{}?material_id={}", url, id);
                }
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let records: Vec<InboundRecord> = resp.json().await.unwrap();
                    print_json(&records);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
            RecordCommands::Outbound { material_id } => {
                let mut url = format!("{}/records/outbound", base_url);
                if let Some(id) = material_id {
                    url = format!("{}?material_id={}", url, id);
                }
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let records: Vec<OutboundRecord> = resp.json().await.unwrap();
                    print_json(&records);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
        },
        Commands::Replenishment { action } => match action {
            ReplenishmentCommands::List {
                material_id,
                status,
            } => {
                let mut url = format!("{}/replenishment", base_url);
                let mut params = Vec::new();
                if let Some(id) = material_id {
                    params.push(format!("material_id={}", id));
                }
                if let Some(s) = status {
                    params.push(format!("status={}", s));
                }
                if !params.is_empty() {
                    url = format!("{}?{}", url, params.join("&"));
                }
                let resp = client.get(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let orders: Vec<ReplenishmentOrder> = resp.json().await.unwrap();
                    print_json(&orders);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
            ReplenishmentCommands::Confirm { id } => {
                let url = format!("{}/replenishment/{}/confirm", base_url, id);
                let resp = client.post(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let order: ReplenishmentOrder = resp.json().await.unwrap();
                    print_json(&order);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
            ReplenishmentCommands::Cancel { id } => {
                let url = format!("{}/replenishment/{}/cancel", base_url, id);
                let resp = client.post(&url).send().await.unwrap();
                if resp.status().is_success() {
                    let order: ReplenishmentOrder = resp.json().await.unwrap();
                    print_json(&order);
                } else {
                    print_error(resp.status(), &resp.text().await.unwrap());
                }
            }
        },
    }
}
