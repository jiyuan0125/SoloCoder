use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Serialize, Deserialize};
use serde_json::Value;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about = "快递驿站管理系统客户端", long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Inbound {
        #[arg(long)]
        tracking_number: String,
        #[arg(long)]
        recipient_phone: String,
        #[arg(long)]
        recipient_name: Option<String>,
        #[arg(long)]
        courier_name: String,
        #[arg(long)]
        courier_phone: String,
        #[arg(long)]
        express_company: String,
    },
    Pickup {
        #[arg(long)]
        code: String,
    },
    ManualVerify {
        #[arg(long)]
        phone: String,
        #[arg(long)]
        id_last_four: String,
        #[arg(long)]
        package_id: Uuid,
    },
    RegisterProxy {
        #[arg(long)]
        package_id: Uuid,
        #[arg(long)]
        proxy_phone: String,
        #[arg(long)]
        proxy_name: Option<String>,
    },
    ProxyFeedback {
        #[arg(long)]
        package_id: Uuid,
        #[arg(long, value_parser = ["confirm", "reject"])]
        feedback: String,
    },
    ListPackages,
    GetPackage {
        #[arg(long)]
        id: Uuid,
    },
    GetByPhone {
        #[arg(long)]
        phone: String,
    },
    GetByCode {
        #[arg(long)]
        code: String,
    },
    ListProxyRecords,
    ListNotifications,
    CheckExpiredProxies,
    CheckStranded,
    Health,
}

#[derive(Debug, Serialize)]
struct InboundReq {
    tracking_number: String,
    recipient_phone: String,
    recipient_name: Option<String>,
    courier_name: String,
    courier_phone: String,
    express_company: String,
}

#[derive(Debug, Serialize)]
struct PickupReq {
    pickup_code: String,
}

#[derive(Debug, Serialize)]
struct ManualVerifyReq {
    phone: String,
    id_last_four: String,
    package_id: Uuid,
}

#[derive(Debug, Serialize)]
struct ProxyReq {
    package_id: Uuid,
    proxy_phone: String,
    proxy_name: Option<String>,
}

#[derive(Debug, Serialize)]
struct ProxyFeedbackReq {
    package_id: Uuid,
    feedback: String,
}

#[derive(Debug, Deserialize)]
struct ApiResponse {
    success: bool,
    error: Option<String>,
    data: Option<Value>,
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();

    match cli.command {
        Commands::Inbound {
            tracking_number,
            recipient_phone,
            recipient_name,
            courier_name,
            courier_phone,
            express_company,
        } => {
            let req = InboundReq {
                tracking_number,
                recipient_phone,
                recipient_name,
                courier_name,
                courier_phone,
                express_company,
            };
            post_json(&client, &cli.server, "/api/packages", &req).await;
        }
        Commands::Pickup { code } => {
            let req = PickupReq { pickup_code: code };
            post_json(&client, &cli.server, "/api/pickup", &req).await;
        }
        Commands::ManualVerify { phone, id_last_four, package_id } => {
            let req = ManualVerifyReq { phone, id_last_four, package_id };
            post_json(&client, &cli.server, "/api/manual-verify", &req).await;
        }
        Commands::RegisterProxy { package_id, proxy_phone, proxy_name } => {
            let req = ProxyReq { package_id, proxy_phone, proxy_name };
            post_json(&client, &cli.server, "/api/proxy", &req).await;
        }
        Commands::ProxyFeedback { package_id, feedback } => {
            let fb = if feedback == "confirm" { "Confirmed" } else { "Rejected" };
            let req = ProxyFeedbackReq { package_id, feedback: fb.to_string() };
            post_json(&client, &cli.server, "/api/proxy/feedback", &req).await;
        }
        Commands::ListPackages => {
            get_json(&client, &cli.server, "/api/packages").await;
        }
        Commands::GetPackage { id } => {
            get_json(&client, &cli.server, &format!("/api/packages/{}", id)).await;
        }
        Commands::GetByPhone { phone } => {
            get_json(&client, &cli.server, &format!("/api/packages/by-phone?phone={}", phone)).await;
        }
        Commands::GetByCode { code } => {
            get_json(&client, &cli.server, &format!("/api/packages/by-code?code={}", code)).await;
        }
        Commands::ListProxyRecords => {
            get_json(&client, &cli.server, "/api/proxy-records").await;
        }
        Commands::ListNotifications => {
            get_json(&client, &cli.server, "/api/notifications").await;
        }
        Commands::CheckExpiredProxies => {
            post_empty(&client, &cli.server, "/api/check-expired-proxies").await;
        }
        Commands::CheckStranded => {
            post_empty(&client, &cli.server, "/api/check-stranded").await;
        }
        Commands::Health => {
            let url = format!("{}/api/health", cli.server);
            match client.get(&url).send().await {
                Ok(resp) => {
                    if resp.status().is_success() {
                        println!("服务运行正常");
                    } else {
                        println!("服务异常: {}", resp.status());
                    }
                }
                Err(e) => println!("连接失败: {}", e),
            }
        }
    }
}

async fn post_json<T: Serialize>(client: &Client, server: &str, path: &str, body: &T) {
    let url = format!("{}{}", server, path);
    match client.post(&url).json(body).send().await {
        Ok(resp) => {
            let result: Result<ApiResponse, _> = resp.json().await;
            match result {
                Ok(data) => print_response(&data),
                Err(e) => println!("解析响应失败: {}", e),
            }
        }
        Err(e) => println!("请求失败: {}", e),
    }
}

async fn post_empty(client: &Client, server: &str, path: &str) {
    let url = format!("{}{}", server, path);
    match client.post(&url).send().await {
        Ok(resp) => {
            let result: Result<ApiResponse, _> = resp.json().await;
            match result {
                Ok(data) => print_response(&data),
                Err(e) => println!("解析响应失败: {}", e),
            }
        }
        Err(e) => println!("请求失败: {}", e),
    }
}

async fn get_json(client: &Client, server: &str, path: &str) {
    let url = format!("{}{}", server, path);
    match client.get(&url).send().await {
        Ok(resp) => {
            let result: Result<ApiResponse, _> = resp.json().await;
            match result {
                Ok(data) => print_response(&data),
                Err(e) => println!("解析响应失败: {}", e),
            }
        }
        Err(e) => println!("请求失败: {}", e),
    }
}

fn print_response(resp: &ApiResponse) {
    if resp.success {
        if let Some(data) = &resp.data {
            println!("{}", serde_json::to_string_pretty(data).unwrap());
        } else {
            println!("操作成功");
        }
    } else {
        if let Some(err) = &resp.error {
            println!("错误: {}", err);
        } else {
            println!("操作失败");
        }
    }
}
