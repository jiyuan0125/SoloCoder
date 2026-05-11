use chrono::NaiveDate;
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::Deserialize;
use uuid::Uuid;

use booking_core::{CreateBookingRequest, SubmitFeedbackRequest, TimeSlot};

#[derive(Parser, Debug)]
#[command(name = "booking")]
#[command(about = "预约试驾系统客户端")]
struct Cli {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:8600")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    ListModels,
    ListCars,
    ListCustomers,
    ListAdvisors,
    ListBookings,
    GetBooking { id: String },
    Create {
        #[arg(long)]
        customer_id: String,
        #[arg(long)]
        model_id: String,
        #[arg(long)]
        date: String,
        #[arg(long)]
        time_slot: String,
    },
    Start { id: String },
    Complete { id: String },
    Cancel { id: String },
    Feedback {
        id: String,
        #[arg(long)]
        satisfaction: u8,
        #[arg(long)]
        intent: String,
    },
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

async fn list_models(base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client.get(format!("{}/api/models", base_url)).send().await?;
    let text = resp.text().await?;
    let models: Vec<serde_json::Value> = serde_json::from_str(&text)?;
    println!("{}", serde_json::to_string_pretty(&models)?);
    Ok(())
}

async fn list_cars(base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client.get(format!("{}/api/cars", base_url)).send().await?;
    let text = resp.text().await?;
    let cars: Vec<serde_json::Value> = serde_json::from_str(&text)?;
    println!("{}", serde_json::to_string_pretty(&cars)?);
    Ok(())
}

async fn list_customers(base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client.get(format!("{}/api/customers", base_url)).send().await?;
    let text = resp.text().await?;
    let customers: Vec<serde_json::Value> = serde_json::from_str(&text)?;
    println!("{}", serde_json::to_string_pretty(&customers)?);
    Ok(())
}

async fn list_advisors(base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client.get(format!("{}/api/advisors", base_url)).send().await?;
    let text = resp.text().await?;
    let advisors: Vec<serde_json::Value> = serde_json::from_str(&text)?;
    println!("{}", serde_json::to_string_pretty(&advisors)?);
    Ok(())
}

async fn list_bookings(base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client.get(format!("{}/api/bookings", base_url)).send().await?;
    let text = resp.text().await?;
    let bookings: Vec<serde_json::Value> = serde_json::from_str(&text)?;
    println!("{}", serde_json::to_string_pretty(&bookings)?);
    Ok(())
}

async fn get_booking(base_url: &str, id: &str) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client.get(format!("{}/api/bookings/{}", base_url, id)).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        let booking: serde_json::Value = serde_json::from_str(&text)?;
        println!("{}", serde_json::to_string_pretty(&booking)?);
    } else {
        let err: ErrorResponse = serde_json::from_str(&text)?;
        eprintln!("错误: {}", err.error);
    }
    Ok(())
}

async fn create_booking(
    base_url: &str,
    customer_id: &str,
    model_id: &str,
    date: &str,
    time_slot: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let time_slot = match time_slot.to_lowercase().as_str() {
        "morning" | "上午" => TimeSlot::Morning,
        "afternoon" | "下午" => TimeSlot::Afternoon,
        _ => {
            eprintln!("无效的时段，请使用 morning/上午 或 afternoon/下午");
            return Ok(());
        }
    };

    let date = NaiveDate::parse_from_str(date, "%Y-%m-%d")?;

    let req = CreateBookingRequest {
        customer_id: Uuid::parse_str(customer_id)?,
        model_id: Uuid::parse_str(model_id)?,
        date,
        time_slot,
    };

    let client = Client::new();
    let resp = client
        .post(format!("{}/api/bookings", base_url))
        .json(&req)
        .send()
        .await?;

    let status = resp.status();
    let text = resp.text().await?;

    if status.is_success() {
        let booking: serde_json::Value = serde_json::from_str(&text)?;
        println!("预约成功:");
        println!("{}", serde_json::to_string_pretty(&booking)?);
    } else {
        let err: ErrorResponse = serde_json::from_str(&text)?;
        eprintln!("预约失败: {}", err.error);
    }

    Ok(())
}

async fn update_booking_status(
    base_url: &str,
    id: &str,
    action: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let resp = client
        .post(format!("{}/api/bookings/{}/{}", base_url, id, action))
        .send()
        .await?;

    let status = resp.status();
    let text = resp.text().await?;

    if status.is_success() {
        let booking: serde_json::Value = serde_json::from_str(&text)?;
        println!("操作成功:");
        println!("{}", serde_json::to_string_pretty(&booking)?);
    } else {
        let err: ErrorResponse = serde_json::from_str(&text)?;
        eprintln!("操作失败: {}", err.error);
    }

    Ok(())
}

async fn submit_feedback(
    base_url: &str,
    id: &str,
    satisfaction: u8,
    intent: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let req = SubmitFeedbackRequest {
        satisfaction,
        purchase_intent: intent.to_string(),
    };

    let client = Client::new();
    let resp = client
        .post(format!("{}/api/bookings/{}/feedback", base_url, id))
        .json(&req)
        .send()
        .await?;

    let status = resp.status();
    let text = resp.text().await?;

    if status.is_success() {
        let feedback: serde_json::Value = serde_json::from_str(&text)?;
        println!("反馈提交成功:");
        println!("{}", serde_json::to_string_pretty(&feedback)?);
    } else {
        let err: ErrorResponse = serde_json::from_str(&text)?;
        eprintln!("提交失败: {}", err.error);
    }

    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let base_url = cli.server.trim_end_matches('/');

    match cli.command {
        Commands::ListModels => list_models(base_url).await?,
        Commands::ListCars => list_cars(base_url).await?,
        Commands::ListCustomers => list_customers(base_url).await?,
        Commands::ListAdvisors => list_advisors(base_url).await?,
        Commands::ListBookings => list_bookings(base_url).await?,
        Commands::GetBooking { id } => get_booking(base_url, &id).await?,
        Commands::Create {
            customer_id,
            model_id,
            date,
            time_slot,
        } => {
            create_booking(base_url, &customer_id, &model_id, &date, &time_slot).await?;
        }
        Commands::Start { id } => update_booking_status(base_url, &id, "start").await?,
        Commands::Complete { id } => update_booking_status(base_url, &id, "complete").await?,
        Commands::Cancel { id } => update_booking_status(base_url, &id, "cancel").await?,
        Commands::Feedback {
            id,
            satisfaction,
            intent,
        } => submit_feedback(base_url, &id, satisfaction, &intent).await?,
    }

    Ok(())
}
