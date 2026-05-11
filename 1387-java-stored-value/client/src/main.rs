use clap::{Parser, Subcommand};
use colored::*;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Parser, Debug)]
#[command(author, version, about = "会员储值卡系统命令行客户端", long_about = None)]
struct Cli {
    #[arg(short, long, default_value = "http://localhost:3000")]
    server: String,
    
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(short, long)]
        name: String,
        
        #[arg(short, long, default_value = "regular")]
        level: String,
    },
    List,
    Get {
        #[arg(short, long)]
        id: String,
    },
    Recharge {
        #[arg(short, long)]
        id: String,
        
        #[arg(short, long)]
        amount: u64,
    },
    Consume {
        #[arg(short, long)]
        id: String,
        
        #[arg(short, long)]
        price: u64,
        
        #[arg(short, long)]
        desc: String,
    },
    Refund {
        #[arg(short, long)]
        id: String,
        
        #[arg(short, long)]
        txn_id: String,
    },
    Transactions {
        #[arg(short, long)]
        id: String,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateMemberRequest {
    name: String,
    level: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct RechargeRequest {
    amount: u64,
}

#[derive(Debug, Serialize, Deserialize)]
struct ConsumeRequest {
    original_price: u64,
    description: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct RefundRequest {
    consume_transaction_id: String,
}

#[derive(Debug, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

async fn print_member(member: &Value) {
    let id = member.get("id").and_then(|v| v.as_str()).unwrap_or("");
    let name = member.get("name").and_then(|v| v.as_str()).unwrap_or("");
    let level = member.get("level").and_then(|v| v.as_str()).unwrap_or("");
    let principal = member.get("principal_balance").and_then(|v| v.as_u64()).unwrap_or(0);
    let gift = member.get("gift_balance").and_then(|v| v.as_u64()).unwrap_or(0);
    let points = member.get("points").and_then(|v| v.as_u64()).unwrap_or(0);
    
    let level_str = match level {
        "Regular" => "普通会员".cyan(),
        "Silver" => "银卡会员".blue(),
        "Gold" => "金卡会员".yellow(),
        "Diamond" => "钻石会员".purple(),
        _ => level.white(),
    };
    
    println!("{}", "========== 会员信息 ==========".bold().green());
    println!("  ID: {}", id.bright_white());
    println!("  姓名: {}", name.bright_white());
    println!("  等级: {}", level_str);
    println!("  本金余额: {} 元", principal.to_string().green());
    println!("  赠送余额: {} 元", gift.to_string().cyan());
    println!("  总余额: {} 元", (principal + gift).to_string().bold().green());
    println!("  积分: {}", points.to_string().yellow());
    println!("{}", "==============================".bold().green());
}

async fn print_transaction(txn: &Value) {
    let id = txn.get("id").and_then(|v| v.as_str()).unwrap_or("");
    let txn_type = txn.get("transaction_type").and_then(|v| v.as_str()).unwrap_or("");
    let original = txn.get("original_price").and_then(|v| v.as_u64()).unwrap_or(0);
    let discounted = txn.get("discounted_price").and_then(|v| v.as_u64()).unwrap_or(0);
    let principal = txn.get("principal_amount").and_then(|v| v.as_i64()).unwrap_or(0);
    let gift = txn.get("gift_amount").and_then(|v| v.as_i64()).unwrap_or(0);
    let points = txn.get("points").and_then(|v| v.as_i64()).unwrap_or(0);
    let desc = txn.get("description").and_then(|v| v.as_str()).unwrap_or("");
    let created = txn.get("created_at").and_then(|v| v.as_str()).unwrap_or("");
    
    let type_str = match txn_type {
        "Recharge" => "充值".green(),
        "Gift" => "赠送".cyan(),
        "Consume" => "消费".red(),
        "Refund" => "退款".magenta(),
        _ => txn_type.white(),
    };
    
    println!("{}", "--------------------------------".dimmed());
    println!("  交易ID: {}", id.dimmed());
    println!("  类型: {}", type_str);
    println!("  描述: {}", desc);
    if original != discounted {
        println!("  原价: {} 元, 折扣价: {} 元", original, discounted);
    } else {
        println!("  金额: {} 元", original);
    }
    if principal != 0 {
        let sign = if principal > 0 { "+" } else { "" };
        println!("  本金: {}{} 元", sign, principal);
    }
    if gift != 0 {
        let sign = if gift > 0 { "+" } else { "" };
        println!("  赠送: {}{} 元", sign, gift);
    }
    if points != 0 {
        let sign = if points > 0 { "+" } else { "" };
        println!("  积分: {}{}", sign, points);
    }
    println!("  时间: {}", created.dimmed());
}

async fn handle_response<T: serde::de::DeserializeOwned + std::fmt::Debug>(
    response: reqwest::Response,
) -> Result<Option<T>, String> {
    let status = response.status();
    let text = response.text().await.map_err(|e| e.to_string())?;
    
    if !status.is_success() {
        let error: ApiResponse<Value> = serde_json::from_str(&text)
            .map_err(|e| format!("解析响应失败: {}", e))?;
        return Err(error.error.unwrap_or_else(|| format!("HTTP错误: {}", status)));
    }
    
    let api_response: ApiResponse<T> = serde_json::from_str(&text)
        .map_err(|e| format!("解析响应失败: {}", e))?;
    
    if !api_response.success {
        return Err(api_response.error.unwrap_or_else(|| "未知错误".to_string()));
    }
    
    Ok(api_response.data)
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/');
    
    match cli.command {
        Commands::Create { name, level } => {
            println!("{}", "正在创建会员...".cyan());
            let req = CreateMemberRequest { name, level };
            let response = client
                .post(format!("{}/members", base_url))
                .json(&req)
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Value>(response).await {
                Ok(Some(member)) => {
                    println!("{}", "会员创建成功！".green().bold());
                    print_member(&member).await;
                }
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("创建失败: {}", e).red()),
            }
        }
        
        Commands::List => {
            println!("{}", "正在获取会员列表...".cyan());
            let response = client
                .get(format!("{}/members", base_url))
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Vec<Value>>(response).await {
                Ok(Some(members)) => {
                    println!("{}", format!("共 {} 个会员", members.len()).green().bold());
                    for member in members {
                        print_member(&member).await;
                    }
                }
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("获取失败: {}", e).red()),
            }
        }
        
        Commands::Get { id } => {
            println!("{}", format!("正在获取会员 {} ...", id).cyan());
            let response = client
                .get(format!("{}/members/{}", base_url, id))
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Value>(response).await {
                Ok(Some(member)) => print_member(&member).await,
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("获取失败: {}", e).red()),
            }
        }
        
        Commands::Recharge { id, amount } => {
            println!("{}", format!("正在为会员 {} 充值 {} 元...", id, amount).cyan());
            let req = RechargeRequest { amount };
            let response = client
                .post(format!("{}/members/{}/recharge", base_url, id))
                .json(&req)
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Value>(response).await {
                Ok(Some(data)) => {
                    println!("{}", "充值成功！".green().bold());
                    if let Some(member) = data.get("member") {
                        print_member(member).await;
                    }
                    if let Some(txns) = data.get("transactions").and_then(|v| v.as_array()) {
                        println!("{}", "交易记录:".yellow().bold());
                        for txn in txns {
                            print_transaction(txn).await;
                        }
                    }
                }
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("充值失败: {}", e).red()),
            }
        }
        
        Commands::Consume { id, price, desc } => {
            println!("{}", format!("会员 {} 消费 {} 元: {}", id, price, desc).cyan());
            let req = ConsumeRequest {
                original_price: price,
                description: desc,
            };
            let response = client
                .post(format!("{}/members/{}/consume", base_url, id))
                .json(&req)
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Value>(response).await {
                Ok(Some(data)) => {
                    println!("{}", "消费成功！".green().bold());
                    if let Some(member) = data.get("member") {
                        print_member(member).await;
                    }
                    if let Some(breakdown) = data.get("breakdown") {
                        let gift = breakdown.get("gift_deducted").and_then(|v| v.as_u64()).unwrap_or(0);
                        let principal = breakdown.get("principal_deducted").and_then(|v| v.as_u64()).unwrap_or(0);
                        println!("{}", "扣款明细:".yellow().bold());
                        println!("  赠送余额扣除: {} 元", gift.to_string().cyan());
                        println!("  本金余额扣除: {} 元", principal.to_string().green());
                    }
                    if let Some(txn) = data.get("transaction") {
                        println!("{}", "交易记录:".yellow().bold());
                        print_transaction(txn).await;
                    }
                }
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("消费失败: {}", e).red()),
            }
        }
        
        Commands::Refund { id, txn_id } => {
            println!("{}", format!("正在对交易 {} 进行退款...", txn_id).cyan());
            let req = RefundRequest {
                consume_transaction_id: txn_id,
            };
            let response = client
                .post(format!("{}/members/{}/refund", base_url, id))
                .json(&req)
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Value>(response).await {
                Ok(Some(data)) => {
                    println!("{}", "退款成功！".green().bold());
                    if let Some(member) = data.get(0) {
                        print_member(member).await;
                    }
                    if let Some(txn) = data.get(1) {
                        println!("{}", "退款交易记录:".yellow().bold());
                        print_transaction(txn).await;
                    }
                }
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("退款失败: {}", e).red()),
            }
        }
        
        Commands::Transactions { id } => {
            println!("{}", format!("正在获取会员 {} 的交易记录...", id).cyan());
            let response = client
                .get(format!("{}/members/{}/transactions", base_url, id))
                .send()
                .await
                .expect("请求失败");
            
            match handle_response::<Vec<Value>>(response).await {
                Ok(Some(txns)) => {
                    println!("{}", format!("共 {} 条交易记录", txns.len()).green().bold());
                    for txn in txns {
                        print_transaction(&txn).await;
                    }
                    println!("{}", "--------------------------------".dimmed());
                }
                Ok(None) => println!("{}", "无数据返回".yellow()),
                Err(e) => println!("{}", format!("获取失败: {}", e).red()),
            }
        }
    }
}
