use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde_json::json;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateShipment {
        #[arg(long)]
        cargo_type: String,
        #[arg(long)]
        vehicle_id: String,
        #[arg(long)]
        min_temp: f64,
        #[arg(long)]
        max_temp: f64,
    },
    ListShipments,
    GetShipment {
        #[arg(long)]
        id: String,
    },
    StartShipment {
        #[arg(long)]
        id: String,
    },
    ReportTemperature {
        #[arg(long)]
        shipment_id: String,
        #[arg(long)]
        temperature: f64,
    },
    CompleteShipment {
        #[arg(long)]
        id: String,
    },
    AcceptShipment {
        #[arg(long)]
        id: String,
    },
    RejectShipment {
        #[arg(long)]
        id: String,
    },
    ListAlerts {
        #[arg(long)]
        shipment_id: Option<String>,
    },
    VehicleStats {
        #[arg(long)]
        vehicle_id: String,
        #[arg(long)]
        year: i32,
        #[arg(long)]
        month: u32,
    },
    CargoStats {
        #[arg(long)]
        cargo_type: String,
    },
}

fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();

    let result = match cli.command {
        Commands::CreateShipment {
            cargo_type,
            vehicle_id,
            min_temp,
            max_temp,
        } => create_shipment(&client, &base_url, &cargo_type, &vehicle_id, min_temp, max_temp),
        Commands::ListShipments => list_shipments(&client, &base_url),
        Commands::GetShipment { id } => get_shipment(&client, &base_url, &id),
        Commands::StartShipment { id } => start_shipment(&client, &base_url, &id),
        Commands::ReportTemperature { shipment_id, temperature } => {
            report_temperature(&client, &base_url, &shipment_id, temperature)
        }
        Commands::CompleteShipment { id } => complete_shipment(&client, &base_url, &id),
        Commands::AcceptShipment { id } => accept_shipment(&client, &base_url, &id),
        Commands::RejectShipment { id } => reject_shipment(&client, &base_url, &id),
        Commands::ListAlerts { shipment_id } => list_alerts(&client, &base_url, shipment_id.as_deref()),
        Commands::VehicleStats { vehicle_id, year, month } => {
            vehicle_stats(&client, &base_url, &vehicle_id, year, month)
        }
        Commands::CargoStats { cargo_type } => cargo_stats(&client, &base_url, &cargo_type),
    };

    match result {
        Ok(response) => println!("{}", response),
        Err(e) => eprintln!("Error: {}", e),
    }
}

fn create_shipment(
    client: &Client,
    base_url: &str,
    cargo_type: &str,
    vehicle_id: &str,
    min_temp: f64,
    max_temp: f64,
) -> Result<String, String> {
    let body = json!({
        "cargo_type": cargo_type,
        "vehicle_id": vehicle_id,
        "min_temp": min_temp,
        "max_temp": max_temp,
    });

    let response = client
        .post(&format!("{}/api/shipments", base_url))
        .json(&body)
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipment created:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn list_shipments(client: &Client, base_url: &str) -> Result<String, String> {
    let response = client
        .get(&format!("{}/api/shipments", base_url))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipments:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn get_shipment(client: &Client, base_url: &str, id: &str) -> Result<String, String> {
    let response = client
        .get(&format!("{}/api/shipments/{}", base_url, id))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipment details:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn start_shipment(client: &Client, base_url: &str, id: &str) -> Result<String, String> {
    let response = client
        .post(&format!("{}/api/shipments/{}/start", base_url, id))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipment started:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn report_temperature(
    client: &Client,
    base_url: &str,
    shipment_id: &str,
    temperature: f64,
) -> Result<String, String> {
    let body = json!({
        "temperature": temperature,
    });

    let response = client
        .post(&format!("{}/api/shipments/{}/temperature", base_url, shipment_id))
        .json(&body)
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Temperature reported:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn complete_shipment(client: &Client, base_url: &str, id: &str) -> Result<String, String> {
    let response = client
        .post(&format!("{}/api/shipments/{}/complete", base_url, id))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipment completed:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn accept_shipment(client: &Client, base_url: &str, id: &str) -> Result<String, String> {
    let response = client
        .post(&format!("{}/api/shipments/{}/accept", base_url, id))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipment accepted:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn reject_shipment(client: &Client, base_url: &str, id: &str) -> Result<String, String> {
    let response = client
        .post(&format!("{}/api/shipments/{}/reject", base_url, id))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Shipment rejected:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn list_alerts(client: &Client, base_url: &str, shipment_id: Option<&str>) -> Result<String, String> {
    let url = match shipment_id {
        Some(id) => format!("{}/api/alerts?shipment_id={}", base_url, id),
        None => format!("{}/api/alerts", base_url),
    };

    let response = client
        .get(&url)
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Alerts:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn vehicle_stats(
    client: &Client,
    base_url: &str,
    vehicle_id: &str,
    year: i32,
    month: u32,
) -> Result<String, String> {
    let response = client
        .get(&format!(
            "{}/api/stats/vehicle/{}/{}/{}",
            base_url, vehicle_id, year, month
        ))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Vehicle stats:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn cargo_stats(client: &Client, base_url: &str, cargo_type: &str) -> Result<String, String> {
    let response = client
        .get(&format!("{}/api/stats/cargo/{}", base_url, cargo_type))
        .send()
        .map_err(|e| format!("Request failed: {}", e))?;

    let status = response.status();
    let text = response.text().map_err(|e| format!("Failed to read response: {}", e))?;

    if status.is_success() {
        Ok(format!("Cargo stats:\n{}", pretty_json(&text)))
    } else {
        Err(format!("Request failed ({}): {}", status, text))
    }
}

fn pretty_json(text: &str) -> String {
    match serde_json::from_str::<serde_json::Value>(text) {
        Ok(v) => serde_json::to_string_pretty(&v).unwrap_or_else(|_| text.to_string()),
        Err(_) => text.to_string(),
    }
}
