use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about = "Factory Energy Monitor CLI", long_about = None)]
struct Cli {
    #[arg(long, default_value = "http://127.0.0.1:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    AddRecord {
        #[arg(short, long)]
        point_id: String,
        #[arg(short, long)]
        date: String,
        #[arg(long)]
        peak: f64,
        #[arg(long)]
        flat: f64,
        #[arg(long)]
        valley: f64,
    },
    GetRecord {
        #[arg(short, long)]
        point_id: String,
        #[arg(short, long)]
        date: String,
    },
    ListRecords {
        #[arg(short, long)]
        point_id: String,
    },
    GetTariff,
    UpdateTariff {
        #[arg(long)]
        peak_price: Option<f64>,
        #[arg(long)]
        flat_price: Option<f64>,
        #[arg(long)]
        valley_price: Option<f64>,
    },
    MarkShutdown {
        #[arg(short, long)]
        date: String,
    },
    UnmarkShutdown {
        #[arg(short, long)]
        date: String,
    },
    CheckAlerts {
        #[arg(short, long)]
        point_id: Option<String>,
    },
}

#[derive(Debug, Serialize)]
struct RecordRequest {
    point_id: String,
    date: String,
    peak: f64,
    flat: f64,
    valley: f64,
}

#[derive(Debug, Deserialize)]
struct RecordResponse {
    point_id: String,
    date: String,
    peak: f64,
    flat: f64,
    valley: f64,
    total_usage: f64,
    cost: CostBreakdown,
}

#[derive(Debug, Deserialize)]
struct CostBreakdown {
    peak_cost: f64,
    flat_cost: f64,
    valley_cost: f64,
    total_cost: f64,
}

#[derive(Debug, Deserialize)]
struct TariffConfigResponse {
    peak_price: f64,
    flat_price: f64,
    valley_price: f64,
}

#[derive(Debug, Serialize)]
struct TariffUpdateRequest {
    peak_price: Option<f64>,
    flat_price: Option<f64>,
    valley_price: Option<f64>,
}

#[derive(Debug, Serialize)]
struct ShutdownRequest {
    date: String,
}

#[derive(Debug, Deserialize)]
struct AlertResponse {
    point_id: String,
    date: String,
    current_usage: f64,
    previous_usage: f64,
    increase_percent: f64,
}

fn main() {
    let cli = Cli::parse();
    let client = Client::new();

    match cli.command {
        Commands::AddRecord { point_id, date, peak, flat, valley } => {
            let url = format!("{}/records", cli.server);
            let req = RecordRequest { point_id, date, peak, flat, valley };
            let resp: RecordResponse = client.post(&url).json(&req).send().unwrap().json().unwrap();
            print_record(&resp);
        }
        Commands::GetRecord { point_id, date } => {
            let url = format!("{}/records/{}/{}", cli.server, point_id, date);
            let resp: RecordResponse = client.get(&url).send().unwrap().json().unwrap();
            print_record(&resp);
        }
        Commands::ListRecords { point_id } => {
            let url = format!("{}/records/{}", cli.server, point_id);
            let resp: Vec<RecordResponse> = client.get(&url).send().unwrap().json().unwrap();
            for r in resp {
                print_record(&r);
                println!("---");
            }
        }
        Commands::GetTariff => {
            let url = format!("{}/tariff", cli.server);
            let resp: TariffConfigResponse = client.get(&url).send().unwrap().json().unwrap();
            println!("Current Tariff Configuration:");
            println!("  Peak:   {:.2} 元/度", resp.peak_price);
            println!("  Flat:   {:.2} 元/度", resp.flat_price);
            println!("  Valley: {:.2} 元/度", resp.valley_price);
        }
        Commands::UpdateTariff { peak_price, flat_price, valley_price } => {
            let url = format!("{}/tariff", cli.server);
            let req = TariffUpdateRequest { peak_price, flat_price, valley_price };
            let resp: TariffConfigResponse = client.put(&url).json(&req).send().unwrap().json().unwrap();
            println!("Updated Tariff Configuration:");
            println!("  Peak:   {:.2} 元/度", resp.peak_price);
            println!("  Flat:   {:.2} 元/度", resp.flat_price);
            println!("  Valley: {:.2} 元/度", resp.valley_price);
        }
        Commands::MarkShutdown { date } => {
            let url = format!("{}/shutdown", cli.server);
            let req = ShutdownRequest { date };
            let resp: serde_json::Value = client.post(&url).json(&req).send().unwrap().json().unwrap();
            println!("{}", resp);
        }
        Commands::UnmarkShutdown { date } => {
            let url = format!("{}/shutdown/{}", cli.server, date);
            let resp: serde_json::Value = client.delete(&url).send().unwrap().json().unwrap();
            println!("{}", resp);
        }
        Commands::CheckAlerts { point_id } => {
            match point_id {
                Some(pid) => {
                    let url = format!("{}/alerts/{}", cli.server, pid);
                    let resp: Option<AlertResponse> = client.get(&url).send().unwrap().json().unwrap();
                    match resp {
                        Some(alert) => print_alert(&alert),
                        None => println!("No alerts for point {}", pid),
                    }
                }
                None => {
                    let url = format!("{}/alerts", cli.server);
                    let resp: Vec<AlertResponse> = client.get(&url).send().unwrap().json().unwrap();
                    if resp.is_empty() {
                        println!("No alerts for any points");
                    } else {
                        for alert in resp {
                            print_alert(&alert);
                            println!("---");
                        }
                    }
                }
            }
        }
    }
}

fn print_record(r: &RecordResponse) {
    println!("Point: {}, Date: {}", r.point_id, r.date);
    println!("  Usage - Peak: {:.2} kWh, Flat: {:.2} kWh, Valley: {:.2} kWh, Total: {:.2} kWh",
             r.peak, r.flat, r.valley, r.total_usage);
    println!("  Cost  - Peak: {:.2} 元, Flat: {:.2} 元, Valley: {:.2} 元, Total: {:.2} 元",
             r.cost.peak_cost, r.cost.flat_cost, r.cost.valley_cost, r.cost.total_cost);
}

fn print_alert(a: &AlertResponse) {
    println!("ALERT! Point: {}, Date: {}", a.point_id, a.date);
    println!("  Previous usage: {:.2} kWh", a.previous_usage);
    println!("  Current usage:  {:.2} kWh", a.current_usage);
    println!("  Increase:       {:.2}% (> 50% threshold)", a.increase_percent);
}
