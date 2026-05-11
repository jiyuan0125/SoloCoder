use std::env;

use bid_procure_core::{
    AcceptInvitationRequest, CreateProcurementRequest, InviteSuppliersRequest, Procurement,
    QualificationLevel, SubmitBidRequest, Supplier, WithdrawBidRequest,
};
use chrono::{DateTime, Utc};
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::Deserialize;
use uuid::Uuid;

#[derive(Parser)]
#[command(name = "bid-procure-cli")]
#[command(about = "竞价采购管理系统命令行客户端", long_about = None)]
struct Cli {
    #[arg(long, default_value = "http://localhost:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Supplier {
        #[command(subcommand)]
        action: SupplierCommands,
    },
    Procurement {
        #[command(subcommand)]
        action: ProcurementCommands,
    },
}

#[derive(Subcommand)]
enum SupplierCommands {
    Create {
        name: String,
        qualification: String,
    },
    List,
    Get {
        id: String,
    },
}

#[derive(Subcommand)]
enum ProcurementCommands {
    Create {
        title: String,
        description: String,
        #[arg(long)]
        required_qualification: Option<String>,
    },
    List,
    Get {
        id: String,
    },
    Invite {
        procurement_id: String,
        supplier_ids: Vec<String>,
    },
    Accept {
        procurement_id: String,
        supplier_id: String,
    },
    Start {
        id: String,
    },
    Bid {
        procurement_id: String,
        supplier_id: String,
        price: f64,
        delivery_date: String,
    },
    Withdraw {
        procurement_id: String,
        supplier_id: String,
    },
    Round {
        procurement_id: String,
        round: u32,
    },
    Advance {
        id: String,
    },
    Evaluate {
        id: String,
    },
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn get_server_url() -> String {
    env::var("SERVER_URL").unwrap_or_else(|_| "http://localhost:8080".to_string())
}

async fn create_supplier(server: &str, name: &str, qualification: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/suppliers", server);
    let body = serde_json::json!({
        "name": name,
        "qualification": qualification,
    });

    let response = client
        .post(&url)
        .json(&body)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let supplier: Supplier = response.json().await.map_err(|e| format!("解析失败: {}", e))?;
        println!("供应商创建成功:");
        println!("  ID: {}", supplier.id);
        println!("  名称: {}", supplier.name);
        println!("  资质: {}", supplier.qualification);
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn list_suppliers(server: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/suppliers", server);

    let response = client
        .get(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let suppliers: Vec<Supplier> = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!("供应商列表 (共{}个):", suppliers.len());
        for s in suppliers {
            println!("  ID: {}, 名称: {}, 资质: {}", s.id, s.name, s.qualification);
        }
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn get_supplier(server: &str, id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/suppliers/{}", server, id);

    let response = client
        .get(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let supplier: Supplier = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!("供应商详情:");
        println!("  ID: {}", supplier.id);
        println!("  名称: {}", supplier.name);
        println!("  资质: {}", supplier.qualification);
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn create_procurement(
    server: &str,
    title: &str,
    description: &str,
    required_qualification: Option<&str>,
) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements", server);

    let req = CreateProcurementRequest {
        title: title.to_string(),
        description: description.to_string(),
        required_qualification: required_qualification
            .map(|q| q.parse::<QualificationLevel>().ok())
            .flatten(),
    };

    let response = client
        .post(&url)
        .json(&req)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurement: Procurement = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        print_procurement(&procurement);
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn list_procurements(server: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements", server);

    let response = client
        .get(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurements: Vec<Procurement> = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!("采购需求列表 (共{}个):", procurements.len());
        for p in procurements {
            println!(
                "  ID: {}, 标题: {}, 状态: {}",
                p.id, p.title, p.status
            );
        }
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn get_procurement(server: &str, id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/{}", server, id);

    let response = client
        .get(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurement: Procurement = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        print_procurement(&procurement);
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

fn print_procurement(p: &Procurement) {
    println!("采购需求详情:");
    println!("  ID: {}", p.id);
    println!("  标题: {}", p.title);
    println!("  描述: {}", p.description);
    println!("  发布日期: {}", p.publish_date);
    println!("  状态: {}", p.status);
    println!(
        "  要求资质: {}",
        p.required_qualification
            .map(|q| q.to_string())
            .unwrap_or_else(|| "无".to_string())
    );
    println!("  当前轮次: {}/{}", p.current_round, p.max_rounds);
    println!("  已邀请供应商: {}", p.invited_suppliers.len());
    println!("  已接受邀请: {}", p.accepted_suppliers.len());
    println!("  已弃权: {}", p.withdrawn_suppliers.len());

    if let Some(winner_id) = p.winner_id {
        println!("  中标供应商: {}", winner_id);
    }
    if let Some(bid) = &p.winner_bid {
        println!("  中标价格: {}", bid.price);
        println!("  交货期: {}", bid.delivery_date);
    }
}

async fn invite_suppliers(server: &str, procurement_id: &str, supplier_ids: &[String]) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/invite", server);

    let procurement_uuid = Uuid::parse_str(procurement_id).map_err(|_| "无效的采购需求ID".to_string())?;
    let supplier_uuids: Result<Vec<Uuid>, _> = supplier_ids
        .iter()
        .map(|s| Uuid::parse_str(s))
        .collect();
    let supplier_uuids = supplier_uuids.map_err(|_| "无效的供应商ID".to_string())?;

    let req = InviteSuppliersRequest {
        procurement_id: procurement_uuid,
        supplier_ids: supplier_uuids,
    };

    let response = client
        .post(&url)
        .json(&req)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurement: Procurement = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!("邀请成功，当前已邀请 {} 个供应商", procurement.invited_suppliers.len());
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn accept_invitation(server: &str, procurement_id: &str, supplier_id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/accept", server);

    let procurement_uuid = Uuid::parse_str(procurement_id).map_err(|_| "无效的采购需求ID".to_string())?;
    let supplier_uuid = Uuid::parse_str(supplier_id).map_err(|_| "无效的供应商ID".to_string())?;

    let req = AcceptInvitationRequest {
        procurement_id: procurement_uuid,
        supplier_id: supplier_uuid,
    };

    let response = client
        .post(&url)
        .json(&req)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        println!("接受邀请成功");
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn start_bidding(server: &str, id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/{}/start", server, id);

    let response = client
        .post(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurement: Procurement = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!("报价已开始，当前状态: {}", procurement.status);
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn submit_bid(
    server: &str,
    procurement_id: &str,
    supplier_id: &str,
    price: f64,
    delivery_date: &str,
) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/bid", server);

    let procurement_uuid = Uuid::parse_str(procurement_id).map_err(|_| "无效的采购需求ID".to_string())?;
    let supplier_uuid = Uuid::parse_str(supplier_id).map_err(|_| "无效的供应商ID".to_string())?;
    let delivery: DateTime<Utc> = delivery_date
        .parse()
        .map_err(|_| "无效的日期格式，请使用 RFC3339 格式，如 2026-06-01T00:00:00Z".to_string())?;

    let req = SubmitBidRequest {
        procurement_id: procurement_uuid,
        supplier_id: supplier_uuid,
        price,
        delivery_date: delivery,
    };

    let response = client
        .post(&url)
        .json(&req)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        println!("报价提交成功");
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn withdraw_bid(server: &str, procurement_id: &str, supplier_id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/withdraw", server);

    let procurement_uuid = Uuid::parse_str(procurement_id).map_err(|_| "无效的采购需求ID".to_string())?;
    let supplier_uuid = Uuid::parse_str(supplier_id).map_err(|_| "无效的供应商ID".to_string())?;

    let req = WithdrawBidRequest {
        procurement_id: procurement_uuid,
        supplier_id: supplier_uuid,
    };

    let response = client
        .post(&url)
        .json(&req)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        println!("弃权成功");
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn get_round_result(server: &str, procurement_id: &str, round: u32) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/{}/round/{}", server, procurement_id, round);

    let response = client
        .get(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let result: serde_json::Value = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!("第 {} 轮报价结果:", round);
        println!("{}", serde_json::to_string_pretty(&result).unwrap());
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn advance_round(server: &str, id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/{}/advance", server, id);

    let response = client
        .post(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurement: Procurement = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        println!(
            "进入下一轮，当前轮次: {}/{}，状态: {}",
            procurement.current_round, procurement.max_rounds, procurement.status
        );
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

async fn evaluate_winner(server: &str, id: &str) -> Result<(), String> {
    let client = Client::new();
    let url = format!("{}/procurements/{}/evaluate", server, id);

    let response = client
        .post(&url)
        .send()
        .await
        .map_err(|e| format!("请求失败: {}", e))?;

    if response.status().is_success() {
        let procurement: Procurement = response
            .json()
            .await
            .map_err(|e| format!("解析失败: {}", e))?;
        print_procurement(&procurement);
        Ok(())
    } else {
        let err: ErrorResponse = response.json().await.unwrap_or(ErrorResponse {
            error: "未知错误".to_string(),
        });
        Err(err.error)
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let server = if cli.server == "http://localhost:8080" {
        get_server_url()
    } else {
        cli.server
    };

    let result = match &cli.command {
        Commands::Supplier { action } => match action {
            SupplierCommands::Create { name, qualification } => {
                create_supplier(&server, name, qualification).await
            }
            SupplierCommands::List => list_suppliers(&server).await,
            SupplierCommands::Get { id } => get_supplier(&server, id).await,
        },
        Commands::Procurement { action } => match action {
            ProcurementCommands::Create {
                title,
                description,
                required_qualification,
            } => {
                create_procurement(
                    &server,
                    title,
                    description,
                    required_qualification.as_deref(),
                )
                .await
            }
            ProcurementCommands::List => list_procurements(&server).await,
            ProcurementCommands::Get { id } => get_procurement(&server, id).await,
            ProcurementCommands::Invite {
                procurement_id,
                supplier_ids,
            } => invite_suppliers(&server, procurement_id, supplier_ids).await,
            ProcurementCommands::Accept {
                procurement_id,
                supplier_id,
            } => accept_invitation(&server, procurement_id, supplier_id).await,
            ProcurementCommands::Start { id } => start_bidding(&server, id).await,
            ProcurementCommands::Bid {
                procurement_id,
                supplier_id,
                price,
                delivery_date,
            } => submit_bid(&server, procurement_id, supplier_id, *price, delivery_date).await,
            ProcurementCommands::Withdraw {
                procurement_id,
                supplier_id,
            } => withdraw_bid(&server, procurement_id, supplier_id).await,
            ProcurementCommands::Round {
                procurement_id,
                round,
            } => get_round_result(&server, procurement_id, *round).await,
            ProcurementCommands::Advance { id } => advance_round(&server, id).await,
            ProcurementCommands::Evaluate { id } => evaluate_winner(&server, id).await,
        },
    };

    if let Err(e) = result {
        eprintln!("错误: {}", e);
        std::process::exit(1);
    }
}
