use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::NaiveDate;
use ticket_core::models::{CreateOrderRequest, TicketRequest, TicketType};

#[derive(Parser, Debug)]
#[command(author, version, about = "景区门票销售系统客户端", long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    
    ListScenics,
    
    GetScenic {
        id: Uuid,
    },
    
    CreateScenic {
        #[arg(long)]
        name: String,
        #[arg(long)]
        base_price: u32,
        #[arg(long)]
        daily_capacity: u32,
    },
    
    BuyTicket {
        scenic_id: Uuid,
        #[arg(long)]
        date: String,
        #[arg(long, value_parser = parse_ticket_request)]
        tickets: Vec<TicketRequestArg>,
    },
    
    GetOrder {
        id: Uuid,
    },
    
    RefundOrder {
        id: Uuid,
    },
    
    GetStats {
        scenic_id: Uuid,
        date: String,
    },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct TicketRequestArg {
    ticket_type: TicketType,
    id_card: Option<String>,
    name: Option<String>,
    child_height: Option<f32>,
    student_id: Option<String>,
}

fn parse_ticket_request(s: &str) -> Result<TicketRequestArg, String> {
    let parts: Vec<&str> = s.split(',').collect();
    if parts.is_empty() {
        return Err("无效的票型格式".into());
    }
    
    let ticket_type = match parts[0].to_lowercase().as_str() {
        "adult" | "成人" => TicketType::Adult,
        "child" | "儿童" => TicketType::Child,
        "elder" | "老人" => TicketType::Elder,
        "student" | "学生" => TicketType::Student,
        _ => return Err("无效的票类型".into()),
    };
    
    let mut arg = TicketRequestArg {
        ticket_type,
        id_card: None,
        name: None,
        child_height: None,
        student_id: None,
    };
    
    for part in &parts[1..] {
        let kv: Vec<&str> = part.splitn(2, '=').collect();
        if kv.len() != 2 {
            continue;
        }
        
        match kv[0].to_lowercase().as_str() {
            "id" | "id_card" | "身份证" => arg.id_card = Some(kv[1].to_string()),
            "name" | "姓名" => arg.name = Some(kv[1].to_string()),
            "height" | "身高" => {
                arg.child_height = Some(kv[1].parse().map_err(|_| "无效的身高值")?);
            }
            "student_id" | "学号" => arg.student_id = Some(kv[1].to_string()),
            _ => {}
        }
    }
    
    Ok(arg)
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

struct ApiClient {
    client: reqwest::Client,
    base_url: String,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        Self {
            client: reqwest::Client::new(),
            base_url,
        }
    }

    async fn health_check(&self) -> Result<(), String> {
        let url = format!("{}/health", self.base_url);
        let resp = self.client.get(&url).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        if resp.status().is_success() {
            println!("服务状态: 正常");
            Ok(())
        } else {
            Err("服务异常".into())
        }
    }

    async fn list_scenics(&self) -> Result<(), String> {
        let url = format!("{}/scenics", self.base_url);
        let resp = self.client.get(&url).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("景区列表:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }

    async fn get_scenic(&self, id: Uuid) -> Result<(), String> {
        let url = format!("{}/scenics/{}", self.base_url, id);
        let resp = self.client.get(&url).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("景区信息:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }

    async fn create_scenic(&self, name: String, base_price: u32, daily_capacity: u32) -> Result<(), String> {
        let url = format!("{}/scenics", self.base_url);
        let body = serde_json::json!({
            "name": name,
            "base_price": base_price,
            "daily_capacity": daily_capacity,
        });
        
        let resp = self.client.post(&url).json(&body).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("创建成功:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }

    async fn buy_ticket(&self, scenic_id: Uuid, date_str: String, tickets: Vec<TicketRequestArg>) -> Result<(), String> {
        let date = NaiveDate::parse_from_str(&date_str, "%Y-%m-%d")
            .map_err(|e| format!("无效的日期格式: {}", e))?;
        
        let ticket_requests: Vec<TicketRequest> = tickets.into_iter().map(|t| TicketRequest {
            ticket_type: t.ticket_type,
            id_card: t.id_card,
            name: t.name,
            child_height: t.child_height,
            student_id: t.student_id,
        }).collect();
        
        let request = CreateOrderRequest {
            scenic_id,
            use_date: date,
            tickets: ticket_requests,
        };
        
        let url = format!("{}/orders", self.base_url);
        let resp = self.client.post(&url).json(&request).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("购票成功:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }

    async fn get_order(&self, id: Uuid) -> Result<(), String> {
        let url = format!("{}/orders/{}", self.base_url, id);
        let resp = self.client.get(&url).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("订单信息:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }

    async fn refund_order(&self, id: Uuid) -> Result<(), String> {
        let url = format!("{}/orders/{}/refund", self.base_url, id);
        let resp = self.client.post(&url).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("退票成功:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }

    async fn get_stats(&self, scenic_id: Uuid, date_str: String) -> Result<(), String> {
        let date = NaiveDate::parse_from_str(&date_str, "%Y-%m-%d")
            .map_err(|e| format!("无效的日期格式: {}", e))?;
        
        let url = format!("{}/stats/{}/{}", self.base_url, scenic_id, date);
        let resp = self.client.get(&url).send().await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        let api_resp: ApiResponse<serde_json::Value> = resp.json().await
            .map_err(|e| format!("解析响应失败: {}", e))?;
        
        if api_resp.success {
            println!("统计信息:");
            println!("{}", serde_json::to_string_pretty(&api_resp.data.unwrap()).unwrap());
            Ok(())
        } else {
            Err(api_resp.error.unwrap_or("未知错误".into()))
        }
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let cli = Cli::parse();
    let client = ApiClient::new(cli.server_url);

    let result = match cli.command {
        Commands::Health => client.health_check().await,
        Commands::ListScenics => client.list_scenics().await,
        Commands::GetScenic { id } => client.get_scenic(id).await,
        Commands::CreateScenic { name, base_price, daily_capacity } => {
            client.create_scenic(name, base_price, daily_capacity).await
        }
        Commands::BuyTicket { scenic_id, date, tickets } => {
            client.buy_ticket(scenic_id, date, tickets).await
        }
        Commands::GetOrder { id } => client.get_order(id).await,
        Commands::RefundOrder { id } => client.refund_order(id).await,
        Commands::GetStats { scenic_id, date } => client.get_stats(scenic_id, date).await,
    };

    if let Err(e) = result {
        eprintln!("错误: {}", e);
        std::process::exit(1);
    }
}
