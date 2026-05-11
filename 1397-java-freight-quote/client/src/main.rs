use clap::{Parser, Subcommand};
use freight_core::*;
use reqwest::blocking::Client;
use serde::Deserialize;
use std::time::Duration;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about = "货运物流报价系统 CLI", long_about = None)]
struct Cli {
    #[arg(long, env = "FREIGHT_SERVER_URL", default_value = "http://127.0.0.1:8607")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(long)]
        length: f64,
        #[arg(long)]
        width: f64,
        #[arg(long)]
        height: f64,
        #[arg(long)]
        weight: f64,
        #[arg(long, value_enum)]
        destination: DestinationArg,
        #[arg(long, value_enum)]
        cargo_type: CargoTypeArg,
        #[arg(long)]
        declared_value: Option<f64>,
    },
    Batch {
        #[arg(long, help = "JSON 文件路径，包含包裹列表")]
        file: String,
    },
    Get {
        order_id: String,
    },
    List,
    Cancel {
        order_id: String,
    },
}

#[derive(clap::ValueEnum, Clone, Debug)]
enum DestinationArg {
    SameCity,
    SameProvince,
    NeighboringProvince,
    Remote,
}

#[derive(clap::ValueEnum, Clone, Debug)]
enum CargoTypeArg {
    Normal,
    Fragile,
    Liquid,
}

impl From<DestinationArg> for DestinationType {
    fn from(d: DestinationArg) -> Self {
        match d {
            DestinationArg::SameCity => DestinationType::SameCity,
            DestinationArg::SameProvince => DestinationType::SameProvince,
            DestinationArg::NeighboringProvince => DestinationType::NeighboringProvince,
            DestinationArg::Remote => DestinationType::Remote,
        }
    }
}

impl From<CargoTypeArg> for CargoType {
    fn from(c: CargoTypeArg) -> Self {
        match c {
            CargoTypeArg::Normal => CargoType::Normal,
            CargoTypeArg::Fragile => CargoType::Fragile,
            CargoTypeArg::Liquid => CargoType::Liquid,
        }
    }
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn build_client() -> Client {
    Client::builder()
        .timeout(Duration::from_secs(10))
        .build()
        .unwrap()
}

fn print_order(order: &OrderResponse) {
    println!("========================================");
    println!("订单ID: {}", order.id);
    println!("状态: {}", order.status.name());
    println!("创建时间: {}", order.created_at);
    println!("----------------------------------------");
    println!("包裹明细:");
    for (idx, item) in order.items.iter().enumerate() {
        println!("\n包裹 #{}", idx + 1);
        println!("  包裹ID: {}", item.package_id);
        println!("  目的地: {}", item.destination.name());
        println!("  货物类型: {}", item.cargo_type.name());
        println!("  ---");
        println!("  体积重量: {:.2} kg", item.breakdown.volumetric_weight);
        println!("  计费重量: {:.2} kg", item.breakdown.chargeable_weight);
        println!("  基础运费: {:.2} 元", item.breakdown.base_freight);
        println!("  偏远附加费: {:.2} 元", item.breakdown.remote_surcharge);
        println!("  保价费: {:.2} 元", item.breakdown.insurance_fee);
        println!("  小计: {:.2} 元", item.breakdown.total_freight);
    }
    println!("----------------------------------------");
    println!("订单总运费: {:.2} 元", order.total_freight);
    println!("========================================");
}

fn main() {
    let cli = Cli::parse();
    let client = build_client();
    let base_url = cli.server.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Create {
            length,
            width,
            height,
            weight,
            destination,
            cargo_type,
            declared_value,
        } => {
            let req = CreateOrderRequest {
                packages: vec![CreatePackageRequest {
                    length_cm: length,
                    width_cm: width,
                    height_cm: height,
                    actual_weight_kg: weight,
                    destination: destination.into(),
                    cargo_type: cargo_type.into(),
                    declared_value,
                }],
            };

            let resp = client
                .post(&format!("{}/api/orders", base_url))
                .json(&req)
                .send();

            match resp {
                Ok(r) => {
                    if r.status().is_success() {
                        let order: OrderResponse = r.json().unwrap();
                        print_order(&order);
                    } else {
                        let err: ErrorResponse = r.json().unwrap();
                        eprintln!("错误: {}", err.error);
                    }
                }
                Err(e) => eprintln!("请求失败: {}", e),
            }
        }

        Commands::Batch { file } => {
            let content = std::fs::read_to_string(&file);
            match content {
                Ok(json_str) => {
                    let req: CreateOrderRequest = match serde_json::from_str(&json_str) {
                        Ok(r) => r,
                        Err(e) => {
                            eprintln!("JSON 解析错误: {}", e);
                            return;
                        }
                    };

                    let resp = client
                        .post(&format!("{}/api/orders", base_url))
                        .json(&req)
                        .send();

                    match resp {
                        Ok(r) => {
                            if r.status().is_success() {
                                let order: OrderResponse = r.json().unwrap();
                                print_order(&order);
                            } else {
                                let err: ErrorResponse = r.json().unwrap();
                                eprintln!("错误: {}", err.error);
                            }
                        }
                        Err(e) => eprintln!("请求失败: {}", e),
                    }
                }
                Err(e) => eprintln!("读取文件失败: {}", e),
            }
        }

        Commands::Get { order_id } => {
            let id = match Uuid::parse_str(&order_id) {
                Ok(u) => u,
                Err(e) => {
                    eprintln!("订单ID格式错误: {}", e);
                    return;
                }
            };

            let resp = client
                .get(&format!("{}/api/orders/{}", base_url, id))
                .send();

            match resp {
                Ok(r) => {
                    if r.status().is_success() {
                        let order: OrderResponse = r.json().unwrap();
                        print_order(&order);
                    } else {
                        let err: ErrorResponse = r.json().unwrap();
                        eprintln!("错误: {}", err.error);
                    }
                }
                Err(e) => eprintln!("请求失败: {}", e),
            }
        }

        Commands::List => {
            let resp = client.get(&format!("{}/api/orders", base_url)).send();

            match resp {
                Ok(r) => {
                    if r.status().is_success() {
                        let orders: Vec<OrderResponse> = r.json().unwrap();
                        if orders.is_empty() {
                            println!("暂无订单");
                        } else {
                            println!("共 {} 个订单\n", orders.len());
                            for order in orders {
                                print_order(&order);
                                println!();
                            }
                        }
                    } else {
                        let err: ErrorResponse = r.json().unwrap();
                        eprintln!("错误: {}", err.error);
                    }
                }
                Err(e) => eprintln!("请求失败: {}", e),
            }
        }

        Commands::Cancel { order_id } => {
            let id = match Uuid::parse_str(&order_id) {
                Ok(u) => u,
                Err(e) => {
                    eprintln!("订单ID格式错误: {}", e);
                    return;
                }
            };

            let resp = client
                .post(&format!("{}/api/orders/{}/cancel", base_url, id))
                .send();

            match resp {
                Ok(r) => {
                    if r.status().is_success() {
                        let order: OrderResponse = r.json().unwrap();
                        println!("订单已取消:");
                        print_order(&order);
                    } else {
                        let err: ErrorResponse = r.json().unwrap();
                        eprintln!("错误: {}", err.error);
                    }
                }
                Err(e) => eprintln!("请求失败: {}", e),
            }
        }
    }
}
