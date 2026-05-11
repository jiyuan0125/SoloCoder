use chrono::{DateTime, Utc, TimeZone};
use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};

use drug_trace_core::{InboundRequest, OutboundRequest, RecallRequest};

#[derive(Parser, Debug)]
#[command(author, version, about = "药品批号追踪和召回系统命令行工具", long_about = None)]
struct Cli {
    #[arg(long, env = "DRUG_TRACE_SERVER", default_value = "http://localhost:8100")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Inbound {
        #[arg(long)]
        name: String,
        #[arg(long)]
        batch: String,
        #[arg(long)]
        production: String,
        #[arg(long)]
        expiry: String,
        #[arg(long)]
        supplier: String,
        #[arg(long)]
        quantity: u32,
    },
    Outbound {
        #[arg(long)]
        name: String,
        #[arg(long)]
        customer: String,
        #[arg(long)]
        quantity: u32,
    },
    Recall {
        #[arg(long)]
        batch: String,
    },
    ListBatches,
    ListAlerts,
    ListInbound,
    ListOutbound,
    Health,
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn parse_datetime(s: &str) -> Result<DateTime<Utc>, String> {
    if let Ok(dt) = DateTime::parse_from_rfc3339(s) {
        return Ok(dt.with_timezone(&Utc));
    }

    if let Ok(naive) = chrono::NaiveDate::parse_from_str(s, "%Y-%m-%d") {
        let dt = naive
            .and_hms_opt(0, 0, 0)
            .ok_or("无效日期")?;
        return Ok(Utc.from_utc_datetime(&dt));
    }

    if let Ok(naive) = chrono::NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S") {
        return Ok(Utc.from_utc_datetime(&naive));
    }

    Err(format!("无法解析日期: {}", s))
}

fn print_json<T: Serialize>(data: &T) {
    println!(
        "{}",
        serde_json::to_string_pretty(data).unwrap_or_else(|_| "{}".to_string())
    );
}

fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();

    let result = match cli.command {
        Commands::Inbound {
            name,
            batch,
            production,
            expiry,
            supplier,
            quantity,
        } => handle_inbound(&client, &base_url, name, batch, production, expiry, supplier, quantity),
        Commands::Outbound {
            name,
            customer,
            quantity,
        } => handle_outbound(&client, &base_url, name, customer, quantity),
        Commands::Recall { batch } => handle_recall(&client, &base_url, batch),
        Commands::ListBatches => handle_list_batches(&client, &base_url),
        Commands::ListAlerts => handle_list_alerts(&client, &base_url),
        Commands::ListInbound => handle_list_inbound(&client, &base_url),
        Commands::ListOutbound => handle_list_outbound(&client, &base_url),
        Commands::Health => handle_health(&client, &base_url),
    };

    if let Err(e) = result {
        eprintln!("错误: {}", e);
        std::process::exit(1);
    }
}

fn handle_inbound(
    client: &Client,
    base_url: &str,
    name: String,
    batch: String,
    production: String,
    expiry: String,
    supplier: String,
    quantity: u32,
) -> Result<(), String> {
    let production_dt = parse_datetime(&production)?;
    let expiry_dt = parse_datetime(&expiry)?;

    let req = InboundRequest {
        drug_name: name,
        batch_no: batch,
        production_date: production_dt,
        expiry_date: expiry_dt,
        supplier,
        quantity,
    };

    let resp = client
        .post(format!("{}/api/inbound", base_url))
        .json(&req)
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    if !resp.status().is_success() {
        let err: ErrorResponse = resp
            .json()
            .unwrap_or(ErrorResponse {
                error: "未知错误".to_string(),
            });
        return Err(err.error);
    }

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("入库成功:");
    print_json(&data);
    Ok(())
}

fn handle_outbound(
    client: &Client,
    base_url: &str,
    name: String,
    customer: String,
    quantity: u32,
) -> Result<(), String> {
    let req = OutboundRequest {
        drug_name: name,
        customer,
        quantity,
    };

    let resp = client
        .post(format!("{}/api/outbound", base_url))
        .json(&req)
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    if !resp.status().is_success() {
        let err: ErrorResponse = resp
            .json()
            .unwrap_or(ErrorResponse {
                error: "未知错误".to_string(),
            });
        return Err(err.error);
    }

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("出库成功:");
    print_json(&data);
    Ok(())
}

fn handle_recall(client: &Client, base_url: &str, batch: String) -> Result<(), String> {
    let req = RecallRequest { batch_no: batch };

    let resp = client
        .post(format!("{}/api/recall", base_url))
        .json(&req)
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    if !resp.status().is_success() {
        let err: ErrorResponse = resp
            .json()
            .unwrap_or(ErrorResponse {
                error: "未知错误".to_string(),
            });
        return Err(err.error);
    }

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("召回记录:");
    print_json(&data);
    Ok(())
}

fn handle_list_batches(client: &Client, base_url: &str) -> Result<(), String> {
    let resp = client
        .get(format!("{}/api/batches", base_url))
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("批次列表:");
    print_json(&data);
    Ok(())
}

fn handle_list_alerts(client: &Client, base_url: &str) -> Result<(), String> {
    let resp = client
        .get(format!("{}/api/alerts", base_url))
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("过期预警:");
    print_json(&data);
    Ok(())
}

fn handle_list_inbound(client: &Client, base_url: &str) -> Result<(), String> {
    let resp = client
        .get(format!("{}/api/records/inbound", base_url))
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("入库记录:");
    print_json(&data);
    Ok(())
}

fn handle_list_outbound(client: &Client, base_url: &str) -> Result<(), String> {
    let resp = client
        .get(format!("{}/api/records/outbound", base_url))
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    let data: serde_json::Value = resp
        .json()
        .map_err(|e| format!("解析响应失败: {}", e))?;

    println!("出库记录:");
    print_json(&data);
    Ok(())
}

fn handle_health(client: &Client, base_url: &str) -> Result<(), String> {
    let resp = client
        .get(format!("{}/health", base_url))
        .send()
        .map_err(|e| format!("请求失败: {}", e))?;

    let status = resp.status();
    let body = resp
        .text()
        .map_err(|e| format!("读取响应失败: {}", e))?;

    if status.is_success() {
        println!("服务状态: 正常");
        println!("响应: {}", body);
        Ok(())
    } else {
        Err(format!("服务异常: {}", status))
    }
}
