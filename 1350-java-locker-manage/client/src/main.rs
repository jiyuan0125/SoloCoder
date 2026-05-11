use clap::{Parser, Subcommand};
use locker_core::*;
use reqwest::Client;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "LOCKER_SERVER", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    ListLockers,
    GetLocker {
        #[arg(short, long)]
        id: String,
    },
    GetLockerInfo {
        #[arg(short, long)]
        id: String,
    },
    Deposit {
        #[arg(short, long)]
        locker: String,

        #[arg(short, long, value_parser = parse_size)]
        size: PackageSize,

        #[arg(short, long)]
        phone: String,

        #[arg(short, long)]
        courier: String,
    },
    Pickup {
        #[arg(short, long)]
        code: String,
    },
    ListPackages {
        #[arg(short, long)]
        phone: String,
    },
}

fn parse_size(s: &str) -> Result<PackageSize, String> {
    match s.to_lowercase().as_str() {
        "small" | "s" => Ok(PackageSize::Small),
        "medium" | "m" => Ok(PackageSize::Medium),
        "large" | "l" => Ok(PackageSize::Large),
        _ => Err(format!("Invalid size: {}. Use small/medium/large", s)),
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
    message: String,
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();

    let result = match &cli.command {
        Commands::Health => health_check(&client, &cli.server).await,
        Commands::ListLockers => list_lockers(&client, &cli.server).await,
        Commands::GetLocker { id } => get_locker(&client, &cli.server, id).await,
        Commands::GetLockerInfo { id } => get_locker_info(&client, &cli.server, id).await,
        Commands::Deposit {
            locker,
            size,
            phone,
            courier,
        } => {
            deposit(&client, &cli.server, locker, *size, phone, courier).await
        }
        Commands::Pickup { code } => pickup(&client, &cli.server, code).await,
        Commands::ListPackages { phone } => list_packages(&client, &cli.server, phone).await,
    };

    if let Err(e) = result {
        eprintln!("Error: {}", e);
        std::process::exit(1);
    }
}

async fn health_check(client: &Client, server: &str) -> Result<(), reqwest::Error> {
    let url = format!("{}/health", server);
    let response = client.get(&url).send().await?;

    if response.status().is_success() {
        let body: serde_json::Value = response.json().await?;
        println!("{}", serde_json::to_string_pretty(&body).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}

async fn list_lockers(client: &Client, server: &str) -> Result<(), reqwest::Error> {
    let url = format!("{}/lockers", server);
    let response = client.get(&url).send().await?;

    if response.status().is_success() {
        let lockers: Vec<LockerInfo> = response.json().await?;
        println!("{}", serde_json::to_string_pretty(&lockers).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}

async fn get_locker(client: &Client, server: &str, id: &str) -> Result<(), reqwest::Error> {
    let url = format!("{}/lockers/{}", server, id);
    let response = client.get(&url).send().await?;

    if response.status().is_success() {
        let locker: LockerDetails = response.json().await?;
        println!("{}", serde_json::to_string_pretty(&locker).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}

async fn get_locker_info(client: &Client, server: &str, id: &str) -> Result<(), reqwest::Error> {
    let url = format!("{}/lockers/{}/info", server, id);
    let response = client.get(&url).send().await?;

    if response.status().is_success() {
        let info: LockerInfo = response.json().await?;
        println!("{}", serde_json::to_string_pretty(&info).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}

async fn deposit(
    client: &Client,
    server: &str,
    locker: &str,
    size: PackageSize,
    phone: &str,
    courier: &str,
) -> Result<(), reqwest::Error> {
    let url = format!("{}/deposit", server);
    let request = DepositRequest {
        locker_id: locker.to_string(),
        package_size: size,
        phone: phone.to_string(),
        courier_name: courier.to_string(),
    };

    let response = client.post(&url).json(&request).send().await?;

    if response.status().is_success() {
        let result: DepositResponse = response.json().await?;
        println!("Deposit successful!");
        println!("{}", serde_json::to_string_pretty(&result).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}

async fn pickup(client: &Client, server: &str, code: &str) -> Result<(), reqwest::Error> {
    let url = format!("{}/pickup", server);
    let request = PickupRequest {
        pickup_code: code.to_string(),
    };

    let response = client.post(&url).json(&request).send().await?;

    if response.status().is_success() {
        let result: PickupResponse = response.json().await?;
        println!("Pickup successful!");
        println!("{}", serde_json::to_string_pretty(&result).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}

async fn list_packages(client: &Client, server: &str, phone: &str) -> Result<(), reqwest::Error> {
    let url = format!("{}/packages/phone/{}", server, phone);
    let response = client.get(&url).send().await?;

    if response.status().is_success() {
        let packages: Vec<PackageInfo> = response.json().await?;
        println!("{}", serde_json::to_string_pretty(&packages).unwrap());
        Ok(())
    } else {
        let status = response.status();
        let error: ErrorResponse = response.json().await?;
        eprintln!("HTTP {}: {}", status, error.message);
        std::process::exit(1);
    }
}
