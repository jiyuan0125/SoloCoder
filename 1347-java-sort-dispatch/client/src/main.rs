extern crate core as logistics_core;

use clap::{Parser, Subcommand};
use logistics_core::models::{OrderRequest, PackageStatus};
use reqwest::blocking::Client;
use serde_json::Value;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateOrder {
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        quantity: u32,
        #[arg(long)]
        destination_address: String,
        #[arg(long)]
        destination_city: String,
        #[arg(long, default_value_t = 1.0)]
        weight_per_item: f64,
        #[arg(long, default_value_t = 0.2)]
        length_per_item: f64,
        #[arg(long, default_value_t = 0.3)]
        width_per_item: f64,
        #[arg(long, default_value_t = 0.1)]
        height_per_item: f64,
    },
    GetOrder {
        #[arg(long)]
        order_id: String,
    },
    GetPackage {
        #[arg(long)]
        package_id: String,
    },
    UpdatePackageStatus {
        #[arg(long)]
        package_id: String,
        #[arg(long)]
        status: String,
    },
    GetExceptions,
    GetWarehouses,
    GetTransferStations,
    ProcessPackages,
}

fn parse_package_status(status: &str) -> Result<PackageStatus, String> {
    match status.to_lowercase().as_str() {
        "created" => Ok(PackageStatus::Created),
        "inwarehouse" | "in_warehouse" => Ok(PackageStatus::InWarehouse),
        "intransit" | "in_transit" => Ok(PackageStatus::InTransit),
        "attransferstation" | "at_transfer_station" => Ok(PackageStatus::AtTransferStation),
        "outfordelivery" | "out_for_delivery" => Ok(PackageStatus::OutForDelivery),
        "delivered" => Ok(PackageStatus::Delivered),
        "lost" => Ok(PackageStatus::Lost),
        "exception" => Ok(PackageStatus::Exception),
        _ => Err(format!("Invalid package status: {}", status)),
    }
}

fn print_json(data: &Value) {
    println!("{}", serde_json::to_string_pretty(data).unwrap());
}

fn main() {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::CreateOrder {
            product_id,
            quantity,
            destination_address,
            destination_city,
            weight_per_item,
            length_per_item,
            width_per_item,
            height_per_item,
        } => {
            let order_request = OrderRequest {
                product_id,
                quantity,
                destination_address,
                destination_city,
                weight_per_item,
                length_per_item,
                width_per_item,
                height_per_item,
            };

            let response = client
                .post(&format!("{}/orders", base_url))
                .json(&order_request)
                .send()
                .expect("Failed to send request");

            let status = response.status();
            let data: Value = response.json().expect("Failed to parse response");

            if status.is_success() {
                println!("Order created successfully:");
                print_json(&data);
            } else {
                eprintln!("Error creating order:");
                print_json(&data);
            }
        }

        Commands::GetOrder { order_id } => {
            let response = client
                .get(&format!("{}/orders/{}", base_url, order_id))
                .send()
                .expect("Failed to send request");

            let status = response.status();
            let data: Value = response.json().expect("Failed to parse response");

            if status.is_success() {
                println!("Order details:");
                print_json(&data);
            } else {
                eprintln!("Error getting order:");
                print_json(&data);
            }
        }

        Commands::GetPackage { package_id } => {
            let response = client
                .get(&format!("{}/packages/{}", base_url, package_id))
                .send()
                .expect("Failed to send request");

            let status = response.status();
            let data: Value = response.json().expect("Failed to parse response");

            if status.is_success() {
                println!("Package details:");
                print_json(&data);
            } else {
                eprintln!("Error getting package:");
                print_json(&data);
            }
        }

        Commands::UpdatePackageStatus { package_id, status } => {
            let package_status = parse_package_status(&status).expect("Invalid package status");

            let response = client
                .post(&format!("{}/packages/{}/status", base_url, package_id))
                .json(&package_status)
                .send()
                .expect("Failed to send request");

            let status = response.status();
            let data: Value = response.json().expect("Failed to parse response");

            if status.is_success() {
                println!("Package status updated:");
                print_json(&data);
            } else {
                eprintln!("Error updating package status:");
                print_json(&data);
            }
        }

        Commands::GetExceptions => {
            let response = client
                .get(&format!("{}/exceptions", base_url))
                .send()
                .expect("Failed to send request");

            let data: Value = response.json().expect("Failed to parse response");
            println!("Current exceptions:");
            print_json(&data);
        }

        Commands::GetWarehouses => {
            let response = client
                .get(&format!("{}/warehouses", base_url))
                .send()
                .expect("Failed to send request");

            let data: Value = response.json().expect("Failed to parse response");
            println!("Warehouses:");
            print_json(&data);
        }

        Commands::GetTransferStations => {
            let response = client
                .get(&format!("{}/transfer-stations", base_url))
                .send()
                .expect("Failed to send request");

            let data: Value = response.json().expect("Failed to parse response");
            println!("Transfer Stations:");
            print_json(&data);
        }

        Commands::ProcessPackages => {
            let response = client
                .post(&format!("{}/process-packages", base_url))
                .send()
                .expect("Failed to send request");

            let data: Value = response.json().expect("Failed to parse response");
            println!("Processing packages:");
            print_json(&data);
        }
    }
}
