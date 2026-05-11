use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use forex_core::{Account, Balance, Currency, ExchangeRate, Order};
use std::collections::HashMap;

#[derive(Parser, Debug)]
#[command(author, version, about = "外汇交易模拟系统客户端", long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateAccount {
        name: String,
        #[arg(long)]
        reviewer: bool,
    },
    GetAccount {
        account_id: String,
    },
    GetBalances {
        account_id: String,
    },
    ListRates,
    PlaceOrder {
        #[arg(short, long)]
        account: String,
        #[arg(short, long)]
        pair: String,
        #[arg(short, long)]
        side: String,
        #[arg(short, long)]
        amount: f64,
    },
    GetOrder {
        order_id: String,
    },
    ListOrders {
        account_id: String,
    },
    PendingApproval,
    ApproveOrder {
        #[arg(short, long)]
        reviewer: String,
        #[arg(short, long)]
        order: String,
        #[arg(short, long)]
        comment: Option<String>,
    },
    RejectOrder {
        #[arg(short, long)]
        reviewer: String,
        #[arg(short, long)]
        order: String,
        #[arg(short, long)]
        comment: Option<String>,
    },
    ProcessSettlement,
}

#[derive(Debug, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

#[derive(Debug, Serialize)]
struct CreateAccountReq {
    name: String,
    is_reviewer: bool,
}

#[derive(Debug, Serialize)]
struct PlaceOrderReq {
    account_id: String,
    pair: String,
    side: String,
    amount: f64,
}

#[derive(Debug, Serialize)]
struct ApproveOrderReq {
    reviewer_id: String,
    order_id: String,
    comment: Option<String>,
}

#[derive(Debug, Serialize)]
struct RejectOrderReq {
    reviewer_id: String,
    order_id: String,
    comment: Option<String>,
}

fn format_balance(balance: &Balance, currency: Currency) -> String {
    let precision = currency.precision();
    format!(
        "可用: {:.precision$} | 冻结: {:.precision$} | 总计: {:.precision$}",
        balance.available,
        balance.frozen,
        balance.total(),
        precision = precision as usize
    )
}

fn print_order(order: &Order) {
    println!("订单ID: {}", order.id);
    println!("账户ID: {}", order.account_id);
    println!("货币对: {}", order.pair);
    println!("方向: {}", order.side);
    println!("金额: {} {}", order.amount, order.pair.base);
    println!("汇率: {}", order.rate);
    println!("状态: {}", order.status);
    println!("创建时间: {}", order.created_at);
    println!("交割日期: {}", order.settlement_date);
    println!("等值美元: ${:.2}", order.usd_equivalent);
    if let Some(comment) = &order.approval_comment {
        println!("审批意见: {}", comment);
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();

    match cli.command {
        Commands::CreateAccount { name, reviewer } => {
            let resp = client
                .post(format!("{}/api/accounts", cli.server))
                .json(&CreateAccountReq { name, is_reviewer: reviewer })
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Account> = resp.json().await.unwrap();
            if result.success {
                let acc = result.data.unwrap();
                println!("账户创建成功!");
                println!("账户ID: {}", acc.id);
                println!("名称: {}", acc.name);
                println!("类型: {}", if acc.is_reviewer { "复核员" } else { "普通用户" });
                if !acc.is_reviewer {
                    println!("初始余额: 100,000.00 CNY");
                }
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::GetAccount { account_id } => {
            let resp = client
                .get(format!("{}/api/accounts/{}", cli.server, account_id))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Account> = resp.json().await.unwrap();
            if result.success {
                let acc = result.data.unwrap();
                println!("账户ID: {}", acc.id);
                println!("名称: {}", acc.name);
                println!("类型: {}", if acc.is_reviewer { "复核员" } else { "普通用户" });
                println!("创建时间: {}", acc.created_at);
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::GetBalances { account_id } => {
            let resp = client
                .get(format!("{}/api/accounts/{}/balances", cli.server, account_id))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<HashMap<Currency, Balance>> = resp.json().await.unwrap();
            if result.success {
                let balances = result.data.unwrap();
                println!("账户余额:");
                let mut currencies: Vec<_> = balances.iter().collect();
                currencies.sort_by_key(|(c, _)| c.symbol());
                for (currency, balance) in currencies {
                    println!("  {}: {}", currency, format_balance(balance, *currency));
                }
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::ListRates => {
            let resp = client
                .get(format!("{}/api/rates", cli.server))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Vec<ExchangeRate>> = resp.json().await.unwrap();
            if result.success {
                let rates = result.data.unwrap();
                println!("当前汇率:");
                println!("{:<12} {:<12} {:<12}", "货币对", "买价(Bid)", "卖价(Ask)");
                println!("{}", "-".repeat(40));
                for rate in rates {
                    println!("{:<12} {:<12.4} {:<12.4}", rate.pair, rate.bid, rate.ask);
                }
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::PlaceOrder { account, pair, side, amount } => {
            let resp = client
                .post(format!("{}/api/orders", cli.server))
                .json(&PlaceOrderReq {
                    account_id: account,
                    pair,
                    side,
                    amount,
                })
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Order> = resp.json().await.unwrap();
            if result.success {
                let order = result.data.unwrap();
                println!("订单提交成功!");
                print_order(&order);
                if order.status == forex_core::OrderStatus::PendingApproval {
                    println!("\n注意: 此订单金额超过等值5万美元，需要复核员审批");
                }
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::GetOrder { order_id } => {
            let resp = client
                .get(format!("{}/api/orders/{}", cli.server, order_id))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Order> = resp.json().await.unwrap();
            if result.success {
                print_order(&result.data.unwrap());
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::ListOrders { account_id } => {
            let resp = client
                .get(format!("{}/api/accounts/{}/orders", cli.server, account_id))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Vec<Order>> = resp.json().await.unwrap();
            if result.success {
                let orders = result.data.unwrap();
                if orders.is_empty() {
                    println!("暂无订单");
                } else {
                    println!("{:<10} {:<12} {:<6} {:<10} {:<10} {:<12}", 
                        "状态", "货币对", "方向", "金额", "汇率", "交割日期");
                    println!("{}", "-".repeat(65));
                    for order in orders {
                        println!("{:<10} {:<12} {:<6} {:<10.2} {:<10.4} {:<12}", 
                            order.status.to_string(),
                            order.pair.to_string(),
                            order.side.to_string(),
                            order.amount,
                            order.rate,
                            order.settlement_date.to_string());
                    }
                }
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::PendingApproval => {
            let resp = client
                .get(format!("{}/api/orders/pending-approval", cli.server))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Vec<Order>> = resp.json().await.unwrap();
            if result.success {
                let orders = result.data.unwrap();
                if orders.is_empty() {
                    println!("暂无待审批订单");
                } else {
                    println!("待审批订单列表:");
                    for order in orders {
                        println!("\n---");
                        print_order(&order);
                    }
                }
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::ApproveOrder { reviewer, order, comment } => {
            let resp = client
                .post(format!("{}/api/orders/approve", cli.server))
                .json(&ApproveOrderReq {
                    reviewer_id: reviewer,
                    order_id: order,
                    comment,
                })
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Order> = resp.json().await.unwrap();
            if result.success {
                println!("订单已批准!");
                print_order(&result.data.unwrap());
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::RejectOrder { reviewer, order, comment } => {
            let resp = client
                .post(format!("{}/api/orders/reject", cli.server))
                .json(&RejectOrderReq {
                    reviewer_id: reviewer,
                    order_id: order,
                    comment,
                })
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<Order> = resp.json().await.unwrap();
            if result.success {
                println!("订单已拒绝!");
                print_order(&result.data.unwrap());
                println!("\n冻结金额已退回可用余额");
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }

        Commands::ProcessSettlement => {
            let resp = client
                .post(format!("{}/api/settlement/process", cli.server))
                .send()
                .await
                .unwrap();
            
            let result: ApiResponse<serde_json::Value> = resp.json().await.unwrap();
            if result.success {
                let data = result.data.unwrap();
                let settled = data["settled"].as_u64().unwrap_or(0);
                println!("交割处理完成，共交割 {} 笔订单", settled);
            } else {
                println!("错误: {}", result.error.unwrap());
            }
        }
    }
}
