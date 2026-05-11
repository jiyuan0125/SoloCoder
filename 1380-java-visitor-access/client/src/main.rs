use chrono::{DateTime, Utc};
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use visitor_core::models::{
    ApprovalDecision, PasscodeVerification, VisitorRegistration, VisitorUpdate,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "VISITOR_SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Register {
        #[arg(long)]
        name: String,
        #[arg(long)]
        phone: String,
        #[arg(long)]
        id_card: String,
        #[arg(long)]
        purpose: String,
        #[arg(long)]
        host: String,
        #[arg(long, default_value_t = Utc::now())]
        expected_arrival_time: DateTime<Utc>,
    },
    Update {
        #[arg(long)]
        id: Uuid,
        #[arg(long)]
        name: String,
        #[arg(long)]
        phone: String,
        #[arg(long)]
        id_card: String,
        #[arg(long)]
        purpose: String,
        #[arg(long)]
        host: String,
        #[arg(long)]
        expected_arrival_time: Option<DateTime<Utc>>,
    },
    Approve {
        #[arg(long)]
        id: Uuid,
        #[arg(long, action = clap::ArgAction::SetTrue)]
        approve: bool,
        #[arg(long, action = clap::ArgAction::SetTrue)]
        reject: bool,
    },
    Verify {
        #[arg(long)]
        passcode: String,
    },
    Signout {
        #[arg(long)]
        id: Uuid,
    },
    Get {
        #[arg(long)]
        id: Uuid,
    },
    List,
    ListByHost {
        #[arg(long)]
        host: String,
    },
}

#[derive(Debug, Deserialize, Serialize)]
struct VisitorRecordResponse {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
    pub id_card: String,
    pub purpose: String,
    pub host: String,
    pub expected_arrival_time: DateTime<Utc>,
    pub status: String,
    pub passcode: Option<String>,
    pub passcode_expires_at: Option<DateTime<Utc>>,
    pub passcode_used: bool,
    pub arrival_time: Option<DateTime<Utc>>,
    pub departure_time: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn print_visitor(record: &VisitorRecordResponse) {
    println!("Visitor ID: {}", record.id);
    println!("  姓名: {}", record.name);
    println!("  手机号: {}", record.phone);
    println!("  身份证号: {}", record.id_card);
    println!("  来访事由: {}", record.purpose);
    println!("  被访人: {}", record.host);
    println!("  预计到访时间: {}", record.expected_arrival_time);
    println!("  状态: {}", record.status);
    if let Some(passcode) = &record.passcode {
        println!("  通行码: {}", passcode);
    }
    if let Some(expires_at) = record.passcode_expires_at {
        println!("  通行码过期时间: {}", expires_at);
    }
    println!("  通行码已使用: {}", record.passcode_used);
    if let Some(arrival) = record.arrival_time {
        println!("  到达时间: {}", arrival);
    }
    if let Some(departure) = record.departure_time {
        println!("  离开时间: {}", departure);
    }
    println!("  创建时间: {}", record.created_at);
    println!("  更新时间: {}", record.updated_at);
    println!();
}

async fn handle_error(response: reqwest::Response) -> String {
    let status = response.status();
    if let Ok(error_response) = response.json::<ErrorResponse>().await {
        format!("Error: {}", error_response.error)
    } else {
        format!("Request failed with status: {}", status)
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::Register {
            name,
            phone,
            id_card,
            purpose,
            host,
            expected_arrival_time,
        } => {
            let registration = VisitorRegistration {
                name,
                phone,
                id_card,
                purpose,
                host,
                expected_arrival_time,
            };

            let response = client
                .post(format!("{}/visitors", base_url))
                .json(&registration)
                .send()
                .await?;

            if response.status().is_success() {
                let visitor: VisitorRecordResponse = response.json().await?;
                println!("✅ 访客登记成功！");
                print_visitor(&visitor);
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::Update {
            id,
            name,
            phone,
            id_card,
            purpose,
            host,
            expected_arrival_time,
        } => {
            let registration = VisitorRegistration {
                name,
                phone,
                id_card,
                purpose,
                host,
                expected_arrival_time: expected_arrival_time.unwrap_or_else(Utc::now),
            };

            let update = VisitorUpdate { id, registration };

            let response = client
                .put(format!("{}/visitors/update", base_url))
                .json(&update)
                .send()
                .await?;

            if response.status().is_success() {
                let visitor: VisitorRecordResponse = response.json().await?;
                println!("✅ 访客信息更新成功！");
                print_visitor(&visitor);
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::Approve {
            id,
            approve,
            reject,
        } => {
            if approve && reject {
                eprintln!("Error: 不能同时使用 --approve 和 --reject");
                std::process::exit(1);
            }
            if !approve && !reject {
                eprintln!("Error: 必须指定 --approve 或 --reject");
                std::process::exit(1);
            }

            let is_approved = approve;
            let decision = ApprovalDecision {
                approved: is_approved,
            };

            let response = client
                .post(format!("{}/visitors/{}/approve", base_url, id))
                .json(&decision)
                .send()
                .await?;

            if response.status().is_success() {
                let visitor: VisitorRecordResponse = response.json().await?;
                if is_approved {
                    println!("✅ 访客已批准！");
                    if let Some(passcode) = &visitor.passcode {
                        println!("通行码: {}", passcode);
                    }
                } else {
                    println!("❌ 访客已拒绝");
                }
                print_visitor(&visitor);
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::Verify { passcode } => {
            let verification = PasscodeVerification { passcode };

            let response = client
                .post(format!("{}/passcode/verify", base_url))
                .json(&verification)
                .send()
                .await?;

            if response.status().is_success() {
                let visitor: VisitorRecordResponse = response.json().await?;
                println!("✅ 通行码验证成功！访客已到达。");
                print_visitor(&visitor);
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::Signout { id } => {
            let response = client
                .post(format!("{}/visitors/{}/signout", base_url, id))
                .send()
                .await?;

            if response.status().is_success() {
                let visitor: VisitorRecordResponse = response.json().await?;
                println!("✅ 访客已签退！");
                print_visitor(&visitor);
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::Get { id } => {
            let response = client
                .get(format!("{}/visitors/{}", base_url, id))
                .send()
                .await?;

            if response.status().is_success() {
                let visitor: VisitorRecordResponse = response.json().await?;
                print_visitor(&visitor);
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::List => {
            let response = client.get(format!("{}/visitors", base_url)).send().await?;

            if response.status().is_success() {
                let visitors: Vec<VisitorRecordResponse> = response.json().await?;
                if visitors.is_empty() {
                    println!("暂无访客记录");
                } else {
                    println!("=== 访客列表 (共 {} 条) ===", visitors.len());
                    for visitor in &visitors {
                        print_visitor(visitor);
                    }
                }
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }

        Commands::ListByHost { host } => {
            let response = client
                .get(format!("{}/visitors/host/{}", base_url, host))
                .send()
                .await?;

            if response.status().is_success() {
                let visitors: Vec<VisitorRecordResponse> = response.json().await?;
                if visitors.is_empty() {
                    println!("被访人 {} 暂无访客记录", host);
                } else {
                    println!("=== 被访人 {} 的访客列表 (共 {} 条) ===", host, visitors.len());
                    for visitor in &visitors {
                        print_visitor(visitor);
                    }
                }
            } else {
                eprintln!("{}", handle_error(response).await);
                std::process::exit(1);
            }
        }
    }

    Ok(())
}
