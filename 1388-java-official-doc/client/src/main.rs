extern crate core_app as core_module;

use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use std::time::Duration;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(long)]
        title: String,
        #[arg(long)]
        content: String,
        #[arg(long, value_parser = ["通知", "请示", "报告", "函"])]
        doc_type: String,
        #[arg(long, value_parser = ["普通", "加急", "特急"])]
        urgency: String,
        #[arg(long)]
        author: String,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    Submit {
        #[arg(long)]
        id: String,
        #[arg(long)]
        operator: String,
    },
    Review {
        #[arg(long)]
        id: String,
        #[arg(long)]
        operator: String,
        #[arg(long)]
        approve: bool,
        #[arg(long)]
        reason: Option<String>,
    },
    Countersign {
        #[arg(long)]
        id: String,
        #[arg(long)]
        operator: String,
        #[arg(long)]
        approve: bool,
        #[arg(long)]
        comment: Option<String>,
        #[arg(long)]
        reason: Option<String>,
    },
    Issue {
        #[arg(long)]
        id: String,
        #[arg(long)]
        operator: String,
        #[arg(long)]
        approve: bool,
        #[arg(long)]
        reason: Option<String>,
    },
    Resubmit {
        #[arg(long)]
        id: String,
        #[arg(long)]
        operator: String,
        #[arg(long)]
        title: Option<String>,
        #[arg(long)]
        content: Option<String>,
    },
    History {
        #[arg(long)]
        id: String,
    },
    Overdue,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

fn main() {
    let cli = Cli::parse();
    let client = Client::builder()
        .timeout(Duration::from_secs(10))
        .build()
        .unwrap();
    let base_url = cli.server.trim_end_matches('/');

    match cli.command {
        Commands::Create {
            title,
            content,
            doc_type,
            urgency,
            author,
        } => {
            let req = serde_json::json!({
                "title": title,
                "content": content,
                "doc_type": doc_type,
                "urgency": urgency,
                "author": author,
            });
            let url = format!("{}/api/documents", base_url);
            let resp: ApiResponse<serde_json::Value> = client
                .post(&url)
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            print_response(resp);
        }
        Commands::List => {
            let url = format!("{}/api/documents", base_url);
            let resp: ApiResponse<serde_json::Value> = client.get(&url).send().unwrap().json().unwrap();
            print_response(resp);
        }
        Commands::Get { id } => {
            let url = format!("{}/api/documents/{}", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client.get(&url).send().unwrap().json().unwrap();
            print_response(resp);
        }
        Commands::Submit { id, operator } => {
            let req = serde_json::json!({ "operator": operator });
            let url = format!("{}/api/documents/{}/submit", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client
                .post(&url)
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            print_response(resp);
        }
        Commands::Review {
            id,
            operator,
            approve,
            reason,
        } => {
            let req = serde_json::json!({
                "operator": operator,
                "approve": approve,
                "reason": reason,
            });
            let url = format!("{}/api/documents/{}/review", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client
                .post(&url)
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            print_response(resp);
        }
        Commands::Countersign {
            id,
            operator,
            approve,
            comment,
            reason,
        } => {
            let req = serde_json::json!({
                "operator": operator,
                "approve": approve,
                "comment": comment,
                "reason": reason,
            });
            let url = format!("{}/api/documents/{}/countersign", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client
                .post(&url)
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            print_response(resp);
        }
        Commands::Issue {
            id,
            operator,
            approve,
            reason,
        } => {
            let req = serde_json::json!({
                "operator": operator,
                "approve": approve,
                "reason": reason,
            });
            let url = format!("{}/api/documents/{}/issue", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client
                .post(&url)
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            print_response(resp);
        }
        Commands::Resubmit {
            id,
            operator,
            title,
            content,
        } => {
            let req = serde_json::json!({
                "operator": operator,
                "title": title,
                "content": content,
            });
            let url = format!("{}/api/documents/{}/resubmit", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client
                .post(&url)
                .json(&req)
                .send()
                .unwrap()
                .json()
                .unwrap();
            print_response(resp);
        }
        Commands::History { id } => {
            let url = format!("{}/api/documents/{}/history", base_url, id);
            let resp: ApiResponse<serde_json::Value> = client.get(&url).send().unwrap().json().unwrap();
            print_response(resp);
        }
        Commands::Overdue => {
            let url = format!("{}/api/overdue", base_url);
            let resp: ApiResponse<serde_json::Value> = client.get(&url).send().unwrap().json().unwrap();
            print_response(resp);
        }
    }
}

fn print_response<T: Serialize>(resp: ApiResponse<T>) {
    if resp.success {
        if let Some(data) = resp.data {
            println!("{}", serde_json::to_string_pretty(&data).unwrap());
        } else {
            println!("Success");
        }
    } else {
        if let Some(err) = resp.error {
            eprintln!("Error: {}", err);
        } else {
            eprintln!("Unknown error");
        }
    }
}
