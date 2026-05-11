use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::json;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = "家政服务平台管理系统命令行客户端")]
struct Cli {
    #[arg(short, long, default_value = "http://localhost:8605")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Skills,
    CreateAunt {
        #[arg(long)]
        name: String,
        #[arg(long)]
        phone: String,
        #[arg(long, value_delimiter = ',')]
        skills: Vec<String>,
        #[arg(long)]
        hourly_rate: u32,
    },
    ListAunts,
    GetAunt {
        id: String,
    },
    CreateCustomer {
        #[arg(long)]
        name: String,
        #[arg(long)]
        phone: String,
        #[arg(long)]
        address: String,
    },
    ListCustomers,
    RecommendAunts {
        #[arg(long)]
        skill_type: String,
        #[arg(long)]
        date: String,
        #[arg(long)]
        start_time: String,
        #[arg(long)]
        end_time: String,
    },
    CreateOrder {
        #[arg(long)]
        customer_id: String,
        #[arg(long)]
        skill_type: String,
        #[arg(long)]
        date: String,
        #[arg(long)]
        start_time: String,
        #[arg(long)]
        end_time: String,
        #[arg(long)]
        aunt_id: Option<String>,
    },
    ListOrders,
    GetOrder {
        id: String,
    },
    ConfirmOrder {
        id: String,
    },
    StartService {
        id: String,
    },
    CompleteOrder {
        id: String,
    },
    CancelOrder {
        id: String,
    },
    RequestRefund {
        #[arg(long)]
        order_id: String,
        #[arg(long)]
        customer_id: String,
        #[arg(long)]
        reason: String,
        #[arg(long)]
        percentage: u32,
    },
    ApproveRefund {
        #[arg(long)]
        refund_request_id: String,
        #[arg(long)]
        percentage: u32,
    },
    RejectRefund {
        #[arg(long)]
        refund_request_id: String,
        #[arg(long)]
        reason: String,
    },
    GetEarnings {
        #[arg(long)]
        aunt_id: String,
        #[arg(long)]
        year: i32,
        #[arg(long)]
        month: u32,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<T>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
}

fn print_pretty(value: &serde_json::Value) {
    println!("{}", serde_json::to_string_pretty(value).unwrap());
}

fn parse_skill_type(input: &str) -> String {
    match input.to_lowercase().as_str() {
        "cleaning" | "保洁" => "Cleaning".to_string(),
        "cooking" | "做饭" => "Cooking".to_string(),
        "nanny" | "月嫂" => "Nanny".to_string(),
        "eldercare" | "elder_care" | "老人护理" => "ElderCare".to_string(),
        "petcare" | "pet_care" | "宠物照料" => "PetCare".to_string(),
        "tutoring" | "家教" => "Tutoring".to_string(),
        _ => input.to_string(),
    }
}

fn default_available_times() -> Vec<serde_json::Value> {
    vec![
        json!({
            "weekday": "Mon",
            "start_time": "08:00:00",
            "end_time": "18:00:00"
        }),
        json!({
            "weekday": "Tue",
            "start_time": "08:00:00",
            "end_time": "18:00:00"
        }),
        json!({
            "weekday": "Wed",
            "start_time": "08:00:00",
            "end_time": "18:00:00"
        }),
        json!({
            "weekday": "Thu",
            "start_time": "08:00:00",
            "end_time": "18:00:00"
        }),
        json!({
            "weekday": "Fri",
            "start_time": "08:00:00",
            "end_time": "18:00:00"
        }),
        json!({
            "weekday": "Sat",
            "start_time": "09:00:00",
            "end_time": "17:00:00"
        }),
        json!({
            "weekday": "Sun",
            "start_time": "09:00:00",
            "end_time": "17:00:00"
        }),
    ]
}

async fn handle_commands(cli: Cli) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let base_url = cli.server;

    match cli.command {
        Commands::Skills => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/skills", base_url))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::CreateAunt { name, phone, skills, hourly_rate } => {
            let parsed_skills: Vec<String> = skills.into_iter().map(|s| parse_skill_type(&s)).collect();
            let body = json!({
                "name": name,
                "phone": phone,
                "skills": parsed_skills,
                "hourly_rate": hourly_rate,
                "available_times": default_available_times()
            });

            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/aunts", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;

            if resp.success {
                println!("阿姨创建成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::ListAunts => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/aunts", base_url))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::GetAunt { id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/aunts/{}", base_url, id))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::CreateCustomer { name, phone, address } => {
            let body = json!({
                "name": name,
                "phone": phone,
                "address": address
            });

            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/customers", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;

            if resp.success {
                println!("客户创建成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::ListCustomers => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/customers", base_url))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::RecommendAunts { skill_type, date, start_time, end_time } => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/aunts/recommend", base_url))
                .query(&[
                    ("skill_type", parse_skill_type(&skill_type)),
                    ("date", date),
                    ("start_time", start_time),
                    ("end_time", end_time),
                ])
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::CreateOrder { customer_id, skill_type, date, start_time, end_time, aunt_id } => {
            let body = json!({
                "customer_id": customer_id,
                "skill_type": parse_skill_type(&skill_type),
                "time_slot": {
                    "date": date,
                    "start_time": start_time,
                    "end_time": end_time
                },
                "aunt_id": aunt_id
            });

            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/orders", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;

            if resp.success {
                println!("订单创建成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::ListOrders => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/orders", base_url))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::GetOrder { id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/orders/{}", base_url, id))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::ConfirmOrder { id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/orders/{}/confirm", base_url, id))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                println!("订单确认成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::StartService { id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/orders/{}/start", base_url, id))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                println!("服务开始成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::CompleteOrder { id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/orders/{}/complete", base_url, id))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                println!("订单完成成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::CancelOrder { id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/orders/{}/cancel", base_url, id))
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                println!("订单取消成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::RequestRefund { order_id, customer_id, reason, percentage } => {
            let body = json!({
                "order_id": order_id,
                "customer_id": customer_id,
                "reason": reason,
                "requested_percentage": percentage
            });

            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/refunds/request", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;

            if resp.success {
                println!("退款申请成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::ApproveRefund { refund_request_id, percentage } => {
            let body = json!({
                "refund_request_id": refund_request_id,
                "approved_percentage": percentage
            });

            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/refunds/approve", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;

            if resp.success {
                println!("退款批准成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::RejectRefund { refund_request_id, reason } => {
            let body = json!({
                "refund_request_id": refund_request_id,
                "reason": reason
            });

            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/refunds/reject", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;

            if resp.success {
                println!("退款拒绝成功！");
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }

        Commands::GetEarnings { aunt_id, year, month } => {
            let resp: ApiResponse<serde_json::Value> = client
                .get(format!("{}/api/aunts/{}/earnings", base_url, aunt_id))
                .query(&[
                    ("year", year.to_string()),
                    ("month", month.to_string()),
                ])
                .send()
                .await?
                .json()
                .await?;
            if resp.success {
                print_pretty(&resp.data.unwrap());
            } else {
                eprintln!("错误: {}", resp.error.unwrap_or_default());
            }
        }
    }

    Ok(())
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    if let Err(e) = handle_commands(cli).await {
        eprintln!("执行错误: {}", e);
    }
}
