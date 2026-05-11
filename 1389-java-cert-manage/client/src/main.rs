use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Serialize, Deserialize};
use serde_json::Value;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Department {
        #[command(subcommand)]
        action: DepartmentCommands,
    },
    Employee {
        #[command(subcommand)]
        action: EmployeeCommands,
    },
    Certificate {
        #[command(subcommand)]
        action: CertificateCommands,
    },
    Training {
        #[command(subcommand)]
        action: TrainingCommands,
    },
    Stats {
        #[command(subcommand)]
        action: StatsCommands,
    },
    Alerts,
}

#[derive(Subcommand, Debug)]
enum DepartmentCommands {
    List,
    Create {
        #[arg(long)]
        name: String,
    },
    Get {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    List,
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        department_id: String,
    },
    Get {
        #[arg(long)]
        id: String,
    },
    Certificates {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum CertificateCommands {
    List,
    Create {
        #[arg(long)]
        employee_id: String,
        #[arg(long)]
        cert_number: String,
        #[arg(long)]
        issuer: String,
        #[arg(long)]
        issue_date: String,
        #[arg(long)]
        expiry_date: String,
        #[arg(long)]
        cert_type: String,
        #[arg(long)]
        required_hours: u32,
    },
    Get {
        #[arg(long)]
        id: String,
    },
    Renew {
        #[arg(long)]
        id: String,
        #[arg(long)]
        new_expiry: String,
    },
    Revoke {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum TrainingCommands {
    List,
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        date: String,
        #[arg(long)]
        hours: u32,
        #[arg(long, value_delimiter = ',')]
        attendees: Vec<String>,
    },
    Get {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum StatsCommands {
    DepartmentRates,
    ExpiryRates,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "client=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Department { action } => handle_department(&client, &base_url, action).await,
        Commands::Employee { action } => handle_employee(&client, &base_url, action).await,
        Commands::Certificate { action } => handle_certificate(&client, &base_url, action).await,
        Commands::Training { action } => handle_training(&client, &base_url, action).await,
        Commands::Stats { action } => handle_stats(&client, &base_url, action).await,
        Commands::Alerts => handle_alerts(&client, &base_url).await,
    }
}

async fn handle_department(client: &Client, base_url: &str, action: DepartmentCommands) -> anyhow::Result<()> {
    match action {
        DepartmentCommands::List => {
            let resp = client.get(format!("{}/departments", base_url)).send().await?;
            print_response(resp).await;
        }
        DepartmentCommands::Create { name } => {
            let resp = client.post(format!("{}/departments", base_url))
                .json(&serde_json::json!({"name": name}))
                .send().await?;
            print_response(resp).await;
        }
        DepartmentCommands::Get { id } => {
            let resp = client.get(format!("{}/departments/{}", base_url, id)).send().await?;
            print_response(resp).await;
        }
    }
    Ok(())
}

async fn handle_employee(client: &Client, base_url: &str, action: EmployeeCommands) -> anyhow::Result<()> {
    match action {
        EmployeeCommands::List => {
            let resp = client.get(format!("{}/employees", base_url)).send().await?;
            print_response(resp).await;
        }
        EmployeeCommands::Create { name, department_id } => {
            let resp = client.post(format!("{}/employees", base_url))
                .json(&serde_json::json!({
                    "name": name,
                    "department_id": department_id
                }))
                .send().await?;
            print_response(resp).await;
        }
        EmployeeCommands::Get { id } => {
            let resp = client.get(format!("{}/employees/{}", base_url, id)).send().await?;
            print_response(resp).await;
        }
        EmployeeCommands::Certificates { id } => {
            let resp = client.get(format!("{}/employees/{}/certificates", base_url, id)).send().await?;
            print_response(resp).await;
        }
    }
    Ok(())
}

async fn handle_certificate(client: &Client, base_url: &str, action: CertificateCommands) -> anyhow::Result<()> {
    match action {
        CertificateCommands::List => {
            let resp = client.get(format!("{}/certificates", base_url)).send().await?;
            print_response(resp).await;
        }
        CertificateCommands::Create { 
            employee_id, cert_number, issuer, issue_date, 
            expiry_date, cert_type, required_hours 
        } => {
            let resp = client.post(format!("{}/certificates", base_url))
                .json(&serde_json::json!({
                    "employee_id": employee_id,
                    "certificate_number": cert_number,
                    "issuer": issuer,
                    "issue_date": issue_date,
                    "expiry_date": expiry_date,
                    "certificate_type": cert_type,
                    "required_hours_for_renewal": required_hours
                }))
                .send().await?;
            print_response(resp).await;
        }
        CertificateCommands::Get { id } => {
            let resp = client.get(format!("{}/certificates/{}", base_url, id)).send().await?;
            print_response(resp).await;
        }
        CertificateCommands::Renew { id, new_expiry } => {
            let resp = client.put(format!("{}/certificates/{}/renew", base_url, id))
                .json(&serde_json::json!({
                    "new_expiry_date": new_expiry
                }))
                .send().await?;
            print_response(resp).await;
        }
        CertificateCommands::Revoke { id } => {
            let resp = client.put(format!("{}/certificates/{}/revoke", base_url, id))
                .send().await?;
            print_response(resp).await;
        }
    }
    Ok(())
}

async fn handle_training(client: &Client, base_url: &str, action: TrainingCommands) -> anyhow::Result<()> {
    match action {
        TrainingCommands::List => {
            let resp = client.get(format!("{}/training", base_url)).send().await?;
            print_response(resp).await;
        }
        TrainingCommands::Create { name, date, hours, attendees } => {
            let resp = client.post(format!("{}/training", base_url))
                .json(&serde_json::json!({
                    "name": name,
                    "date": date,
                    "hours": hours,
                    "attendee_ids": attendees
                }))
                .send().await?;
            print_response(resp).await;
        }
        TrainingCommands::Get { id } => {
            let resp = client.get(format!("{}/training/{}", base_url, id)).send().await?;
            print_response(resp).await;
        }
    }
    Ok(())
}

async fn handle_stats(client: &Client, base_url: &str, action: StatsCommands) -> anyhow::Result<()> {
    match action {
        StatsCommands::DepartmentRates => {
            let resp = client.get(format!("{}/stats/department-rates", base_url)).send().await?;
            print_response(resp).await;
        }
        StatsCommands::ExpiryRates => {
            let resp = client.get(format!("{}/stats/expiry-rates", base_url)).send().await?;
            print_response(resp).await;
        }
    }
    Ok(())
}

async fn handle_alerts(client: &Client, base_url: &str) -> anyhow::Result<()> {
    let resp = client.get(format!("{}/alerts/expiring", base_url)).send().await?;
    print_response(resp).await;
    Ok(())
}

async fn print_response(resp: reqwest::Response) {
    let status = resp.status();
    let text = resp.text().await.unwrap_or_default();
    
    if status.is_success() {
        if let Ok(json) = serde_json::from_str::<Value>(&text) {
            println!("{}", serde_json::to_string_pretty(&json).unwrap_or_else(|_| json.to_string()));
        } else {
            println!("{}", text);
        }
    } else {
        eprintln!("Error {}: {}", status, text);
    }
}
