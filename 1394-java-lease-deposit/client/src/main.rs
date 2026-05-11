use std::error::Error;

use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde::Serialize;
use rust_decimal::Decimal;
use chrono::NaiveDate;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "LEASE_SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(subcommand)]
    Tenant(TenantCommands),
    #[command(subcommand)]
    Room(RoomCommands),
    #[command(subcommand)]
    Contract(ContractCommands),
    Reminders,
    Health,
}

#[derive(Subcommand, Debug)]
enum TenantCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        phone: String,
        #[arg(short, long)]
        id_card: Option<String>,
    },
    List,
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum RoomCommands {
    Create {
        #[arg(short, long)]
        room_number: String,
        #[arg(short, long)]
        area: Decimal,
        #[arg(short = 'r', long)]
        default_rent: Decimal,
    },
    List,
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ContractCommands {
    Create {
        #[arg(short, long)]
        room_id: Uuid,
        #[arg(short, long)]
        tenant_id: Uuid,
        #[arg(short, long)]
        start_date: String,
        #[arg(short, long)]
        end_date: String,
        #[arg(short = 'r', long)]
        monthly_rent: Option<Decimal>,
    },
    List,
    Get {
        id: Uuid,
    },
    Checkin {
        contract_id: Uuid,
    },
    Checkout {
        contract_id: Uuid,
        #[arg(short, long)]
        checkout_date: String,
        #[arg(short, long)]
        damage_fee: Option<Decimal>,
    },
    Renew {
        contract_id: Uuid,
        #[arg(short, long)]
        new_end_date: String,
        #[arg(short = 'r', long)]
        new_monthly_rent: Option<Decimal>,
    },
    ProcessExpired {
        contract_id: Uuid,
    },
}

#[derive(Debug, Serialize)]
struct CreateTenantReq {
    name: String,
    phone: String,
    id_card: Option<String>,
}

#[derive(Debug, Serialize)]
struct CreateRoomReq {
    room_number: String,
    area: Decimal,
    default_monthly_rent: Decimal,
}

#[derive(Debug, Serialize)]
struct CreateContractReq {
    room_id: Uuid,
    tenant_id: Uuid,
    start_date: NaiveDate,
    end_date: NaiveDate,
    monthly_rent: Option<Decimal>,
}

#[derive(Debug, Serialize)]
struct CheckinReq {
    contract_id: Uuid,
}

#[derive(Debug, Serialize)]
struct CheckoutReq {
    contract_id: Uuid,
    checkout_date: NaiveDate,
    damage_fee: Option<Decimal>,
}

#[derive(Debug, Serialize)]
struct RenewContractReq {
    contract_id: Uuid,
    new_end_date: NaiveDate,
    new_monthly_rent: Option<Decimal>,
}

fn parse_date(s: &str) -> Result<NaiveDate, Box<dyn Error>> {
    Ok(NaiveDate::parse_from_str(s, "%Y-%m-%d")?)
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::Health => {
            let resp = client.get(format!("{}/health", base_url))
                .send()?;
            let body = resp.text()?;
            println!("{}", body);
        }
        Commands::Reminders => {
            let resp = client.get(format!("{}/reminders/renewal", base_url))
                .send()?;
            let body = resp.text()?;
            println!("{}", body);
        }
        Commands::Tenant(cmd) => match cmd {
            TenantCommands::Create { name, phone, id_card } => {
                let req = CreateTenantReq { name, phone, id_card };
                let resp = client.post(format!("{}/tenants", base_url))
                    .json(&req)
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            TenantCommands::List => {
                let resp = client.get(format!("{}/tenants", base_url))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            TenantCommands::Get { id } => {
                let resp = client.get(format!("{}/tenants/{}", base_url, id))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
        },
        Commands::Room(cmd) => match cmd {
            RoomCommands::Create { room_number, area, default_rent } => {
                let req = CreateRoomReq {
                    room_number,
                    area,
                    default_monthly_rent: default_rent,
                };
                let resp = client.post(format!("{}/rooms", base_url))
                    .json(&req)
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            RoomCommands::List => {
                let resp = client.get(format!("{}/rooms", base_url))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            RoomCommands::Get { id } => {
                let resp = client.get(format!("{}/rooms/{}", base_url, id))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
        },
        Commands::Contract(cmd) => match cmd {
            ContractCommands::Create { room_id, tenant_id, start_date, end_date, monthly_rent } => {
                let req = CreateContractReq {
                    room_id,
                    tenant_id,
                    start_date: parse_date(&start_date)?,
                    end_date: parse_date(&end_date)?,
                    monthly_rent,
                };
                let resp = client.post(format!("{}/contracts", base_url))
                    .json(&req)
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            ContractCommands::List => {
                let resp = client.get(format!("{}/contracts", base_url))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            ContractCommands::Get { id } => {
                let resp = client.get(format!("{}/contracts/{}", base_url, id))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            ContractCommands::Checkin { contract_id } => {
                let req = CheckinReq { contract_id };
                let resp = client.post(format!("{}/contracts/checkin", base_url))
                    .json(&req)
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            ContractCommands::Checkout { contract_id, checkout_date, damage_fee } => {
                let req = CheckoutReq {
                    contract_id,
                    checkout_date: parse_date(&checkout_date)?,
                    damage_fee,
                };
                let resp = client.post(format!("{}/contracts/checkout", base_url))
                    .json(&req)
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            ContractCommands::Renew { contract_id, new_end_date, new_monthly_rent } => {
                let req = RenewContractReq {
                    contract_id,
                    new_end_date: parse_date(&new_end_date)?,
                    new_monthly_rent,
                };
                let resp = client.post(format!("{}/contracts/renew", base_url))
                    .json(&req)
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
            ContractCommands::ProcessExpired { contract_id } => {
                let resp = client.post(format!("{}/contracts/{}/process-expired", base_url, contract_id))
                    .send()?;
                let body = resp.text()?;
                println!("{}", body);
            }
        },
    }

    Ok(())
}
