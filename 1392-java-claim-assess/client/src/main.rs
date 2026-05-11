use clap::{Parser, Subcommand};
use reqwest::Client;
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use claim_core::{Claim, ClaimStatus, RepairItemType, Vehicle};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Vehicle {
        #[command(subcommand)]
        cmd: VehicleCommands,
    },
    Claim {
        #[command(subcommand)]
        cmd: ClaimCommands,
    },
}

#[derive(Subcommand, Debug)]
enum VehicleCommands {
    List,
    Get {
        id: String,
    },
    GetByPlate {
        plate_number: String,
    },
    Create {
        #[arg(long)]
        plate_number: String,
        #[arg(long)]
        model: String,
        #[arg(long)]
        actual_value: Decimal,
    },
    UpdateValue {
        id: String,
        #[arg(long)]
        actual_value: Decimal,
    },
}

#[derive(Subcommand, Debug)]
enum ClaimCommands {
    List,
    ListByVehicle {
        vehicle_id: String,
    },
    Get {
        id: String,
    },
    Create {
        #[arg(long)]
        vehicle_id: String,
    },
    AddItem {
        claim_id: String,
        #[arg(long)]
        name: String,
        #[arg(long, value_parser = parse_item_type)]
        item_type: RepairItemType,
        #[arg(long)]
        cost: Decimal,
    },
    Settle {
        claim_id: String,
        #[arg(long)]
        salvage_value: Option<Decimal>,
    },
}

fn parse_item_type(s: &str) -> Result<RepairItemType, String> {
    match s.to_lowercase().as_str() {
        "part" | "partreplacement" => Ok(RepairItemType::PartReplacement),
        "labor" => Ok(RepairItemType::Labor),
        _ => Err(format!("Invalid item type: {}. Use 'part' or 'labor'", s)),
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateVehicleRequest {
    plate_number: String,
    model: String,
    actual_value: Decimal,
}

#[derive(Debug, Serialize, Deserialize)]
struct UpdateVehicleValueRequest {
    actual_value: Decimal,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateClaimRequest {
    vehicle_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AddRepairItemRequest {
    name: String,
    item_type: RepairItemType,
    cost: Decimal,
}

#[derive(Debug, Serialize, Deserialize)]
struct SettleClaimRequest {
    salvage_value: Option<Decimal>,
}

fn print_vehicle(v: &Vehicle) {
    println!("ID: {}", v.id);
    println!("  车牌: {}", v.plate_number);
    println!("  型号: {}", v.model);
    println!("  实际价值: {}", v.actual_value);
}

fn print_claim(c: &Claim) {
    let status_str = match c.status {
        ClaimStatus::Pending => "待处理",
        ClaimStatus::TotalLoss => "推定全损",
        ClaimStatus::Settled => "已结案",
    };

    println!("ID: {}", c.id);
    println!("  车辆ID: {}", c.vehicle_id);
    println!("  状态: {}", status_str);
    println!("  维修项目数: {}", c.repair_items.len());
    if !c.repair_items.is_empty() {
        let total = c.repair_items.iter().fold(Decimal::ZERO, |acc, i| acc + i.cost);
        println!("  总维修费: {}", total);
        for item in &c.repair_items {
            let type_str = match item.item_type {
                RepairItemType::PartReplacement => "配件",
                RepairItemType::Labor => "工时",
            };
            println!("    - {} ({}) : {}", item.name, type_str, item.cost);
        }
    }
    if let Some(salvage) = c.salvage_value {
        println!("  残值: {}", salvage);
    }
    if let Some(payout) = c.payout {
        println!("  赔付款: {}", payout);
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Vehicle { cmd } => match cmd {
            VehicleCommands::List => {
                let vehicles: Vec<Vehicle> = client
                    .get(format!("{}/vehicles", base_url))
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 车辆列表 ===");
                for v in &vehicles {
                    print_vehicle(v);
                    println!();
                }
            }
            VehicleCommands::Get { id } => {
                let vehicle: Vehicle = client
                    .get(format!("{}/vehicles/{}", base_url, id))
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 车辆详情 ===");
                print_vehicle(&vehicle);
            }
            VehicleCommands::GetByPlate { plate_number } => {
                let vehicle: Vehicle = client
                    .get(format!("{}/vehicles/plate/{}", base_url, plate_number))
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 车辆详情 ===");
                print_vehicle(&vehicle);
            }
            VehicleCommands::Create {
                plate_number,
                model,
                actual_value,
            } => {
                let vehicle: Vehicle = client
                    .post(format!("{}/vehicles", base_url))
                    .json(&CreateVehicleRequest {
                        plate_number,
                        model,
                        actual_value,
                    })
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 创建成功 ===");
                print_vehicle(&vehicle);
            }
            VehicleCommands::UpdateValue { id, actual_value } => {
                let vehicle: Vehicle = client
                    .post(format!("{}/vehicles/{}", base_url, id))
                    .json(&UpdateVehicleValueRequest { actual_value })
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 更新成功 ===");
                print_vehicle(&vehicle);
            }
        },
        Commands::Claim { cmd } => match cmd {
            ClaimCommands::List => {
                let claims: Vec<Claim> = client
                    .get(format!("{}/claims", base_url))
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 理赔单列表 ===");
                for c in &claims {
                    print_claim(c);
                    println!();
                }
            }
            ClaimCommands::ListByVehicle { vehicle_id } => {
                let claims: Vec<Claim> = client
                    .get(format!("{}/vehicles/{}/claims", base_url, vehicle_id))
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 车辆理赔单列表 ===");
                for c in &claims {
                    print_claim(c);
                    println!();
                }
            }
            ClaimCommands::Get { id } => {
                let claim: Claim = client
                    .get(format!("{}/claims/{}", base_url, id))
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 理赔单详情 ===");
                print_claim(&claim);
            }
            ClaimCommands::Create { vehicle_id } => {
                let claim: Claim = client
                    .post(format!("{}/claims", base_url))
                    .json(&CreateClaimRequest { vehicle_id })
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 创建成功 ===");
                print_claim(&claim);
            }
            ClaimCommands::AddItem {
                claim_id,
                name,
                item_type,
                cost,
            } => {
                let claim: Claim = client
                    .post(format!("{}/claims/{}/items", base_url, claim_id))
                    .json(&AddRepairItemRequest {
                        name,
                        item_type,
                        cost,
                    })
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 添加成功 ===");
                print_claim(&claim);
            }
            ClaimCommands::Settle {
                claim_id,
                salvage_value,
            } => {
                let claim: Claim = client
                    .post(format!("{}/claims/{}/settle", base_url, claim_id))
                    .json(&SettleClaimRequest { salvage_value })
                    .send()
                    .await?
                    .json()
                    .await?;
                println!("=== 结案成功 ===");
                print_claim(&claim);
            }
        },
    }

    Ok(())
}
