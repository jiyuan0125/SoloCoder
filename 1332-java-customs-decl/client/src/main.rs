use clap::{Parser, Subcommand};
use decl_core::*;
use reqwest::Client;
use rust_decimal::Decimal;
use serde_json::Value;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand, Debug)]
enum Command {
    RateList,
    RateAdd {
        #[arg(long)]
        date: String,
        #[arg(long)]
        currency: String,
        #[arg(long)]
        rate: Decimal,
    },
    DeclList,
    DeclGet {
        #[arg(long)]
        id: Uuid,
    },
    DeclCreate {
        #[arg(long)]
        no: String,
        #[arg(long)]
        items: String,
    },
    DeclDelete {
        #[arg(long)]
        id: Uuid,
    },
    DeclMerge {
        #[arg(long)]
        ids: String,
        #[arg(long)]
        new_no: String,
    },
    DeclDeclare {
        #[arg(long)]
        id: Uuid,
        #[arg(long)]
        date: String,
    },
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base = cli.server.trim_end_matches('/').to_string();

    match cli.command {
        Command::RateList => {
            let resp = client.get(format!("{}/api/rates", base)).send().await?;
            let rates: Vec<ExchangeRate> = resp.json().await?;
            print_json(&rates);
        }
        Command::RateAdd { date, currency, rate } => {
            let currency = parse_currency(&currency)?;
            let date = chrono::NaiveDate::parse_from_str(&date, "%Y-%m-%d")?;
            let req = CreateExchangeRateRequest {
                date,
                currency,
                rate_to_cny: rate,
            };
            let resp = client
                .post(format!("{}/api/rates", base))
                .json(&req)
                .send()
                .await?;
            let rate: ExchangeRate = resp.json().await?;
            print_json(&rate);
        }
        Command::DeclList => {
            let resp = client.get(format!("{}/api/declarations", base)).send().await?;
            let decls: Vec<DeclarationSummary> = resp.json().await?;
            print_json(&decls);
        }
        Command::DeclGet { id } => {
            let resp = client
                .get(format!("{}/api/declarations/{}", base, id))
                .send()
                .await?;
            let decl: Declaration = resp.json().await?;
            print_json(&decl);
        }
        Command::DeclCreate { no, items } => {
            let items: Vec<DeclarationItem> = serde_json::from_str(&items)?;
            let req = CreateDeclarationRequest {
                declaration_no: no,
                items,
            };
            let resp = client
                .post(format!("{}/api/declarations", base))
                .json(&req)
                .send()
                .await?;
            let decl: Declaration = resp.json().await?;
            print_json(&decl);
        }
        Command::DeclDelete { id } => {
            let resp = client
                .delete(format!("{}/api/declarations/{}", base, id))
                .send()
                .await?;
            if resp.status().is_success() {
                println!("Deleted");
            } else {
                let body: Value = resp.json().await?;
                print_json(&body);
            }
        }
        Command::DeclMerge { ids, new_no } => {
            let ids: Vec<Uuid> = serde_json::from_str(&ids)?;
            let req = MergeDeclarationRequest {
                declaration_ids: ids,
                new_declaration_no: new_no,
            };
            let resp = client
                .post(format!("{}/api/declarations/merge", base))
                .json(&req)
                .send()
                .await?;
            let decl: Declaration = resp.json().await?;
            print_json(&decl);
        }
        Command::DeclDeclare { id, date } => {
            let date = chrono::NaiveDate::parse_from_str(&date, "%Y-%m-%d")?;
            let req = DeclareRequest {
                declaration_id: id,
                declare_date: date,
            };
            let resp = client
                .post(format!("{}/api/declarations/declare", base))
                .json(&req)
                .send()
                .await?;
            let decl: Declaration = resp.json().await?;
            print_json(&decl);
        }
    }

    Ok(())
}

fn parse_currency(s: &str) -> Result<Currency, String> {
    match s.to_uppercase().as_str() {
        "USD" => Ok(Currency::USD),
        "EUR" => Ok(Currency::EUR),
        "JPY" => Ok(Currency::JPY),
        "GBP" => Ok(Currency::GBP),
        "CNY" => Ok(Currency::CNY),
        _ => Err(format!("Unknown currency: {}", s)),
    }
}

fn print_json<T: serde::Serialize>(value: &T) {
    let json = serde_json::to_string_pretty(value).unwrap();
    println!("{}", json);
}
