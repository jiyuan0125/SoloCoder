use clap::{Parser, Subcommand};
use reqwest::Client;
use rust_decimal::Decimal;
use rust_decimal::prelude::FromPrimitive;
use serde::Deserialize;
use insurance_claim_core::*;

#[derive(Parser, Debug)]
#[command(name = "insurance-claim-client")]
struct Cli {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Policy {
        #[command(subcommand)]
        cmd: PolicyCommands,
    },
    Claim {
        #[command(subcommand)]
        cmd: ClaimCommands,
    },
}

#[derive(Subcommand, Debug)]
enum PolicyCommands {
    Create {
        #[arg(long)]
        policy_number: String,
        #[arg(long)]
        holder_name: String,
        #[arg(long)]
        deductible: f64,
        #[arg(long)]
        policy_year: i32,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum ClaimCommands {
    Create {
        #[arg(long)]
        case_number: String,
        #[arg(long)]
        total_loss: f64,
        #[arg(long, value_delimiter = ',')]
        parties: Vec<String>,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    Close {
        #[arg(long)]
        id: String,
    },
}

#[derive(Debug, Deserialize)]
struct ApiError {
    error: String,
}

async fn handle_response(response: reqwest::Response) -> Result<String, String> {
    if response.status().is_success() {
        let text = response.text().await.map_err(|e| e.to_string())?;
        let json: serde_json::Value = serde_json::from_str(&text).map_err(|e| e.to_string())?;
        Ok(serde_json::to_string_pretty(&json).map_err(|e| e.to_string())?)
    } else {
        let status = response.status();
        let text = response.text().await.map_err(|e| e.to_string())?;
        if let Ok(api_error) = serde_json::from_str::<ApiError>(&text) {
            Err(format!("错误 {}: {}", status, api_error.error))
        } else {
            Err(format!("错误 {}: {}", status, text))
        }
    }
}

fn parse_party(s: &str) -> Result<CreatePartyRequest, String> {
    let parts: Vec<&str> = s.split('|').collect();
    if parts.len() != 4 {
        return Err(format!("无效的参与方格式: {}. 格式应为: 姓名|保单ID|责任比例|自身损失", s));
    }
    
    let ratio = parts[2].parse::<f64>().map_err(|_| format!("无效的责任比例: {}", parts[2]))?;
    let own_loss = parts[3].parse::<f64>().map_err(|_| format!("无效的自身损失: {}", parts[3]))?;
    
    Ok(CreatePartyRequest {
        name: parts[0].to_string(),
        policy_id: parts[1].to_string(),
        liability_ratio: Decimal::from_f64(ratio).unwrap(),
        own_loss: Decimal::from_f64(own_loss).unwrap(),
    })
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    let result = match cli.command {
        Commands::Policy { cmd } => handle_policy_cmd(&client, &base_url, cmd).await,
        Commands::Claim { cmd } => handle_claim_cmd(&client, &base_url, cmd).await,
    };

    match result {
        Ok(output) => println!("{}", output),
        Err(e) => eprintln!("{}", e),
    }
}

async fn handle_policy_cmd(
    client: &Client,
    base_url: &str,
    cmd: PolicyCommands,
) -> Result<String, String> {
    match cmd {
        PolicyCommands::Create { policy_number, holder_name, deductible, policy_year } => {
            let req = CreatePolicyRequest {
                policy_number,
                holder_name,
                deductible: Decimal::from_f64(deductible).unwrap(),
                policy_year,
            };
            
            let response = client
                .post(format!("{}/policies", base_url))
                .json(&req)
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
        PolicyCommands::List => {
            let response = client
                .get(format!("{}/policies", base_url))
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
        PolicyCommands::Get { id } => {
            let response = client
                .get(format!("{}/policies/{}", base_url, id))
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
    }
}

async fn handle_claim_cmd(
    client: &Client,
    base_url: &str,
    cmd: ClaimCommands,
) -> Result<String, String> {
    match cmd {
        ClaimCommands::Create { case_number, total_loss, parties } => {
            let party_requests: Vec<CreatePartyRequest> = parties
                .iter()
                .map(|s| parse_party(s))
                .collect::<Result<_, _>>()?;
            
            let req = CreateClaimRequest {
                case_number,
                total_loss: Decimal::from_f64(total_loss).unwrap(),
                parties: party_requests,
            };
            
            let response = client
                .post(format!("{}/claims", base_url))
                .json(&req)
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
        ClaimCommands::List => {
            let response = client
                .get(format!("{}/claims", base_url))
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
        ClaimCommands::Get { id } => {
            let response = client
                .get(format!("{}/claims/{}", base_url, id))
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
        ClaimCommands::Close { id } => {
            let response = client
                .post(format!("{}/claims/{}", base_url, id))
                .send()
                .await
                .map_err(|e| e.to_string())?;
            
            handle_response(response).await
        }
    }
}
