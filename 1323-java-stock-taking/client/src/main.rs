use std::collections::HashMap;

use clap::{Parser, Subcommand};
use rust_decimal::Decimal;
use stock_core::{
    ApproveDifferenceRequest, CreateCountRecordRequest, CreateBatchRequest, Material,
    SubmitDifferenceRequest,
};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, env = "STOCK_SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,

    Material {
        #[command(subcommand)]
        cmd: MaterialCommands,
    },

    Batch {
        #[command(subcommand)]
        cmd: BatchCommands,
    },

    Count {
        #[command(subcommand)]
        cmd: CountCommands,
    },

    Difference {
        #[command(subcommand)]
        cmd: DifferenceCommands,
    },
}

#[derive(Subcommand, Debug)]
enum MaterialCommands {
    List,
    Get {
        code: String,
    },
    Add {
        code: String,
        name: String,
        price: Decimal,
        #[arg(long)]
        quantity: i64,
    },
}

#[derive(Subcommand, Debug)]
enum BatchCommands {
    List,
    Get {
        batch_no: String,
    },
    Create {
        operator: String,
    },
    Cancel {
        batch_no: String,
    },
    Complete {
        batch_no: String,
    },
    Records {
        batch_no: String,
    },
    Differences {
        batch_no: String,
    },
}

#[derive(Subcommand, Debug)]
enum CountCommands {
    Submit {
        batch_no: String,
        material_code: String,
        actual_quantity: i64,
        operator: String,
    },
}

#[derive(Subcommand, Debug)]
enum DifferenceCommands {
    Get {
        id: Uuid,
    },
    Submit {
        id: Uuid,
        reason: String,
        operator: String,
    },
    Approve {
        id: Uuid,
        #[arg(long)]
        approved: bool,
        approver: String,
    },
}

fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    let client = reqwest::blocking::Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Health => {
            let resp = client.get(format!("{}/health", base_url)).send()?;
            print_response(resp);
        }

        Commands::Material { cmd } => match cmd {
            MaterialCommands::List => {
                let resp = client.get(format!("{}/materials", base_url)).send()?;
                print_response(resp);
            }
            MaterialCommands::Get { code } => {
                let resp = client.get(format!("{}/materials/{}", base_url, code)).send()?;
                print_response(resp);
            }
            MaterialCommands::Add {
                code,
                name,
                price,
                quantity,
            } => {
                let material = Material {
                    code,
                    name,
                    price,
                    book_quantity: quantity,
                };
                let resp = client
                    .post(format!("{}/materials", base_url))
                    .json(&material)
                    .send()?;
                print_response(resp);
            }
        },

        Commands::Batch { cmd } => match cmd {
            BatchCommands::List => {
                let resp = client.get(format!("{}/batches", base_url)).send()?;
                print_response(resp);
            }
            BatchCommands::Get { batch_no } => {
                let resp = client.get(format!("{}/batches/{}", base_url, batch_no)).send()?;
                print_response(resp);
            }
            BatchCommands::Create { operator } => {
                let req = CreateBatchRequest { operator };
                let resp = client
                    .post(format!("{}/batches", base_url))
                    .json(&req)
                    .send()?;
                print_response(resp);
            }
            BatchCommands::Cancel { batch_no } => {
                let resp = client
                    .put(format!("{}/batches/{}/cancel", base_url, batch_no))
                    .json(&HashMap::<String, String>::new())
                    .send()?;
                print_response(resp);
            }
            BatchCommands::Complete { batch_no } => {
                let resp = client
                    .put(format!("{}/batches/{}/complete", base_url, batch_no))
                    .json(&HashMap::<String, String>::new())
                    .send()?;
                print_response(resp);
            }
            BatchCommands::Records { batch_no } => {
                let resp = client
                    .get(format!("{}/batches/{}/records", base_url, batch_no))
                    .send()?;
                print_response(resp);
            }
            BatchCommands::Differences { batch_no } => {
                let resp = client
                    .get(format!("{}/batches/{}/differences", base_url, batch_no))
                    .send()?;
                print_response(resp);
            }
        },

        Commands::Count { cmd } => match cmd {
            CountCommands::Submit {
                batch_no,
                material_code,
                actual_quantity,
                operator,
            } => {
                let req = CreateCountRecordRequest {
                    batch_no,
                    material_code,
                    actual_quantity,
                    operator,
                };
                let resp = client
                    .post(format!("{}/counts", base_url))
                    .json(&req)
                    .send()?;
                print_response(resp);
            }
        },

        Commands::Difference { cmd } => match cmd {
            DifferenceCommands::Get { id } => {
                let resp = client
                    .get(format!("{}/differences/{}", base_url, id))
                    .send()?;
                print_response(resp);
            }
            DifferenceCommands::Submit {
                id,
                reason,
                operator,
            } => {
                let req = SubmitDifferenceRequest { reason, operator };
                let resp = client
                    .put(format!("{}/differences/{}/submit", base_url, id))
                    .json(&req)
                    .send()?;
                print_response(resp);
            }
            DifferenceCommands::Approve {
                id,
                approved,
                approver,
            } => {
                let req = ApproveDifferenceRequest { approved, approver };
                let resp = client
                    .put(format!("{}/differences/{}/approve", base_url, id))
                    .json(&req)
                    .send()?;
                print_response(resp);
            }
        },
    }

    Ok(())
}

fn print_response(resp: reqwest::blocking::Response) {
    let status = resp.status();
    let body: serde_json::Value = resp.json().unwrap_or(serde_json::json!({"raw": "无法解析JSON"}));
    println!("Status: {}", status);
    println!("{}", serde_json::to_string_pretty(&body).unwrap());
}
