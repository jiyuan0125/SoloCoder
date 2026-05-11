use clap::{Parser, Subcommand};
use futures_sim_core::{Contract, User};
use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateContract {
        #[arg(long)]
        code: String,
        #[arg(long)]
        price: f64,
        #[arg(long)]
        multiplier: f64,
        #[arg(long)]
        margin_ratio: f64,
    },
    ListContracts,
    GetContract {
        #[arg(long)]
        code: String,
    },
    UpdatePrice {
        #[arg(long)]
        code: String,
        #[arg(long)]
        price: f64,
    },
    RegisterUser {
        #[arg(long)]
        user_id: String,
    },
    ListUsers,
    GetUser {
        #[arg(long)]
        user_id: String,
    },
    OpenPosition {
        #[arg(long)]
        user_id: String,
        #[arg(long)]
        contract_code: String,
        #[arg(long)]
        position_type: String,
        #[arg(long)]
        lots: i32,
    },
    ClosePosition {
        #[arg(long)]
        user_id: String,
        #[arg(long)]
        contract_code: String,
        #[arg(long)]
        position_type: String,
        #[arg(long)]
        lots: i32,
    },
    Settle,
}

#[derive(Serialize)]
struct CreateContractReq {
    code: String,
    price: f64,
    multiplier: f64,
    margin_ratio: f64,
}

#[derive(Serialize)]
struct OpenPositionReq {
    user_id: String,
    contract_code: String,
    position_type: String,
    lots: i32,
}

#[derive(Serialize)]
struct ClosePositionReq {
    user_id: String,
    contract_code: String,
    position_type: String,
    lots: i32,
}

#[derive(Serialize)]
struct UpdatePriceReq {
    price: f64,
}

#[derive(Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> fmt::Display for ApiResponse<T>
where
    T: fmt::Debug,
{
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.success {
            if let Some(data) = &self.data {
                write!(f, "{:#?}", data)
            } else {
                write!(f, "成功")
            }
        } else {
            write!(f, "错误: {}", self.error.as_deref().unwrap_or("未知错误"))
        }
    }
}

fn main() {
    let args = Args::parse();
    let client = reqwest::blocking::Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();

    match args.command {
        Commands::CreateContract {
            code,
            price,
            multiplier,
            margin_ratio,
        } => {
            let req = CreateContractReq {
                code,
                price,
                multiplier,
                margin_ratio,
            };
            let resp: ApiResponse<Contract> = client
                .post(format!("{}/api/contracts", base_url))
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::ListContracts => {
            let resp: ApiResponse<Vec<Contract>> = client
                .get(format!("{}/api/contracts", base_url))
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::GetContract { code } => {
            let resp: ApiResponse<Contract> = client
                .get(format!("{}/api/contracts/{}", base_url, code))
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::UpdatePrice { code, price } => {
            let req = UpdatePriceReq { price };
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/contracts/{}/price", base_url, code))
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::RegisterUser { user_id } => {
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/users/{}", base_url, user_id))
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::ListUsers => {
            let resp: ApiResponse<Vec<User>> = client
                .get(format!("{}/api/users", base_url))
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::GetUser { user_id } => {
            let resp: ApiResponse<User> = client
                .get(format!("{}/api/users/{}", base_url, user_id))
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::OpenPosition {
            user_id,
            contract_code,
            position_type,
            lots,
        } => {
            let req = OpenPositionReq {
                user_id,
                contract_code,
                position_type,
                lots,
            };
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/positions/open", base_url))
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::ClosePosition {
            user_id,
            contract_code,
            position_type,
            lots,
        } => {
            let req = ClosePositionReq {
                user_id,
                contract_code,
                position_type,
                lots,
            };
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/positions/close", base_url))
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
        Commands::Settle => {
            let resp: ApiResponse<serde_json::Value> = client
                .post(format!("{}/api/settle", base_url))
                .send()
                .unwrap()
                .json()
                .unwrap();
            println!("{}", resp);
        }
    }
}
