use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use uuid::Uuid;
use chrono::Utc;
use contract_core::{
    Contract, ContractType, CreateContractRequest, ApprovalRequest,
    ReminderInfo, SystemConfig, PriceAdjustmentRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        r#type: String,
        #[arg(long)]
        amount: f64,
        #[arg(long, default_value_t = false)]
        framework: bool,
        #[arg(long)]
        parent: Option<Uuid>,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
    Activate {
        #[arg(long)]
        id: Uuid,
    },
    Renew {
        #[arg(long)]
        id: Uuid,
        #[arg(long)]
        amount: f64,
        #[arg(long)]
        days: i64,
    },
    Price {
        #[arg(long)]
        id: Uuid,
        #[arg(long)]
        amount: f64,
    },
    ApproveRenewal {
        #[arg(long)]
        approval: Uuid,
        #[arg(long)]
        days: i64,
    },
    RejectRenewal {
        #[arg(long)]
        approval: Uuid,
        #[arg(long)]
        reason: String,
    },
    ApprovePrice {
        #[arg(long)]
        approval: Uuid,
    },
    RejectPrice {
        #[arg(long)]
        approval: Uuid,
        #[arg(long)]
        reason: String,
    },
    Terminate {
        #[arg(long)]
        id: Uuid,
    },
    Close {
        #[arg(long)]
        id: Uuid,
    },
    Reminders,
    Config,
    GetApproval {
        #[arg(long)]
        id: Uuid,
    },
}

fn parse_contract_type(s: &str) -> Result<ContractType, String> {
    match s.to_lowercase().as_str() {
        "purchase" | "采购" => Ok(ContractType::Purchase),
        "sales" | "销售" => Ok(ContractType::Sales),
        "service" | "服务" => Ok(ContractType::Service),
        _ => Err(format!("未知合同类型: {}", s)),
    }
}

fn main() {
    let cli = Cli::parse();
    let client = Client::new();

    let result: Result<serde_json::Value, String> = match cli.command {
        Commands::Create { name, r#type, amount, framework, parent } => {
            let contract_type = parse_contract_type(&r#type).unwrap_or_else(|e| {
                eprintln!("{}", e);
                std::process::exit(1);
            });
            let req = CreateContractRequest {
                name,
                contract_type,
                amount,
                start_date: Utc::now(),
                end_date: None,
                is_framework: framework,
                parent_id: parent,
            };
            post::<_, Contract>(&client, &format!("{}/api/contracts", cli.server), &req)
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::List => {
            get::<Vec<Contract>>(&client, &format!("{}/api/contracts", cli.server))
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Get { id } => {
            get::<Contract>(&client, &format!("{}/api/contracts/{}", cli.server, id))
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Activate { id } => {
            put::<_, Contract>(&client, &format!("{}/api/contracts/{}/activate", cli.server, id), &())
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Renew { id, amount, days } => {
            let body = serde_json::json!({
                "new_amount": amount,
                "end_date": (Utc::now() + chrono::Duration::days(days)).to_rfc3339()
            });
            put::<_, ApprovalRequest>(&client, &format!("{}/api/contracts/{}/renew", cli.server, id), &body)
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Price { id, amount } => {
            let req = PriceAdjustmentRequest { new_amount: amount };
            put::<_, Option<ApprovalRequest>>(&client, &format!("{}/api/contracts/{}/price", cli.server, id), &req)
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::ApproveRenewal { approval, days } => {
            let body = serde_json::json!({
                "end_date": (Utc::now() + chrono::Duration::days(days)).to_rfc3339()
            });
            put::<_, Contract>(&client, &format!("{}/api/approvals/{}/approve-renewal", cli.server, approval), &body)
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::RejectRenewal { approval, reason } => {
            let body = serde_json::json!({ "reason": reason });
            put::<_, ()>(&client, &format!("{}/api/approvals/{}/reject-renewal", cli.server, approval), &body)
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::ApprovePrice { approval } => {
            put::<_, Contract>(&client, &format!("{}/api/approvals/{}/approve-price", cli.server, approval), &())
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::RejectPrice { approval, reason } => {
            let body = serde_json::json!({ "reason": reason });
            put::<_, ()>(&client, &format!("{}/api/approvals/{}/reject-price", cli.server, approval), &body)
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Terminate { id } => {
            put::<_, Contract>(&client, &format!("{}/api/contracts/{}/terminate", cli.server, id), &())
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Close { id } => {
            put::<_, Contract>(&client, &format!("{}/api/contracts/{}/close", cli.server, id), &())
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Reminders => {
            get::<Vec<ReminderInfo>>(&client, &format!("{}/api/reminders", cli.server))
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::Config => {
            get::<SystemConfig>(&client, &format!("{}/api/config", cli.server))
                .map(|v| serde_json::to_value(v).unwrap())
        }
        Commands::GetApproval { id } => {
            get::<ApprovalRequest>(&client, &format!("{}/api/approvals/{}", cli.server, id))
                .map(|v| serde_json::to_value(v).unwrap())
        }
    };

    match result {
        Ok(json) => {
            println!("{}", serde_json::to_string_pretty(&json).unwrap());
        }
        Err(e) => {
            eprintln!("Error: {}", e);
            std::process::exit(1);
        }
    }
}

fn post<T: serde::Serialize, R: serde::de::DeserializeOwned>(
    client: &Client,
    url: &str,
    body: &T,
) -> Result<R, String> {
    let response = client
        .post(url)
        .json(body)
        .send()
        .map_err(|e| e.to_string())?;
    
    handle_response(response)
}

fn get<R: serde::de::DeserializeOwned>(
    client: &Client,
    url: &str,
) -> Result<R, String> {
    let response = client
        .get(url)
        .send()
        .map_err(|e| e.to_string())?;
    
    handle_response(response)
}

fn put<T: serde::Serialize, R: serde::de::DeserializeOwned>(
    client: &Client,
    url: &str,
    body: &T,
) -> Result<R, String> {
    let response = client
        .put(url)
        .json(body)
        .send()
        .map_err(|e| e.to_string())?;
    
    handle_response(response)
}

fn handle_response<R: serde::de::DeserializeOwned>(
    response: reqwest::blocking::Response,
) -> Result<R, String> {
    if response.status().is_success() {
        if response.content_length() == Some(0) {
            return serde_json::from_str("null").map_err(|e| e.to_string());
        }
        response.json().map_err(|e| e.to_string())
    } else {
        let status = response.status();
        let text = response.text().unwrap_or_default();
        Err(format!("HTTP {}: {}", status, text))
    }
}
