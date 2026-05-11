use clap::{Parser, Subcommand};
use reqwest::Client;
use serde_json::json;
use uuid::Uuid;
use chrono::{DateTime, Utc, NaiveDateTime, TimeZone};

use otl_core::models::{WeekendCompensation};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Employee {
        #[command(subcommand)]
        action: EmployeeCommands,
    },
    Overtime {
        #[command(subcommand)]
        action: OvertimeCommands,
    },
    Leave {
        #[command(subcommand)]
        action: LeaveCommands,
    },
    Holiday {
        #[command(subcommand)]
        action: HolidayCommands,
    },
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        monthly_salary: f64,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
    Balance {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum OvertimeCommands {
    Create {
        #[arg(long)]
        employee_id: Uuid,
        #[arg(long)]
        start_time: String,
        #[arg(long)]
        end_time: String,
        #[arg(long)]
        compensation: Option<String>,
    },
    Approve {
        #[arg(long)]
        id: Uuid,
    },
    Reject {
        #[arg(long)]
        id: Uuid,
    },
    List {
        #[arg(long)]
        employee_id: Option<Uuid>,
    },
}

#[derive(Subcommand, Debug)]
enum LeaveCommands {
    Create {
        #[arg(long)]
        employee_id: Uuid,
        #[arg(long)]
        start_time: String,
        #[arg(long)]
        end_time: String,
    },
    Approve {
        #[arg(long)]
        id: Uuid,
    },
    Reject {
        #[arg(long)]
        id: Uuid,
    },
    Cancel {
        #[arg(long)]
        id: Uuid,
    },
    List {
        #[arg(long)]
        employee_id: Option<Uuid>,
    },
}

#[derive(Subcommand, Debug)]
enum HolidayCommands {
    Create {
        #[arg(long)]
        date: String,
        #[arg(long)]
        name: String,
    },
    List,
}

fn parse_datetime(s: &str) -> DateTime<Utc> {
    if let Ok(dt) = DateTime::parse_from_rfc3339(s) {
        dt.with_timezone(&Utc)
    } else {
        let naive = NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S").expect("Invalid datetime format");
        Utc.from_utc_datetime(&naive)
    }
}

fn parse_date(s: &str) -> DateTime<Utc> {
    if let Ok(dt) = DateTime::parse_from_rfc3339(s) {
        dt.with_timezone(&Utc)
    } else {
        let naive = chrono::NaiveDate::parse_from_str(s, "%Y-%m-%d")
            .expect("Invalid date format")
            .and_hms_opt(0, 0, 0)
            .expect("Invalid time");
        Utc.from_utc_datetime(&naive)
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::Employee { action } => handle_employee(&client, &base_url, action).await?,
        Commands::Overtime { action } => handle_overtime(&client, &base_url, action).await?,
        Commands::Leave { action } => handle_leave(&client, &base_url, action).await?,
        Commands::Holiday { action } => handle_holiday(&client, &base_url, action).await?,
    }

    Ok(())
}

async fn handle_employee(
    client: &Client,
    base_url: &str,
    action: EmployeeCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        EmployeeCommands::Create { name, monthly_salary } => {
            let resp = client
                .post(&format!("{}/api/employees", base_url))
                .json(&json!({
                    "name": name,
                    "monthly_salary": monthly_salary
                }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        EmployeeCommands::List => {
            let resp = client.get(&format!("{}/api/employees", base_url)).send().await?;
            print_response(resp).await?;
        }
        EmployeeCommands::Get { id } => {
            let resp = client.get(&format!("{}/api/employees/{}", base_url, id)).send().await?;
            print_response(resp).await?;
        }
        EmployeeCommands::Balance { id } => {
            let resp = client.get(&format!("{}/api/employees/{}/balance", base_url, id)).send().await?;
            print_response(resp).await?;
        }
    }
    Ok(())
}

async fn handle_overtime(
    client: &Client,
    base_url: &str,
    action: OvertimeCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        OvertimeCommands::Create { employee_id, start_time, end_time, compensation } => {
            let start = parse_datetime(&start_time);
            let end = parse_datetime(&end_time);
            let compensation_type = match compensation.as_deref() {
                Some("leave") => Some(WeekendCompensation::Leave),
                Some("pay") => Some(WeekendCompensation::OvertimePay),
                Some(_) => panic!("Invalid compensation type. Use 'leave' or 'pay'"),
                None => None,
            };
            
            let mut body = json!({
                "employee_id": employee_id,
                "start_time": start,
                "end_time": end,
            });
            
            if let Some(ct) = compensation_type {
                body["weekend_compensation"] = json!(ct);
            }
            
            let resp = client
                .post(&format!("{}/api/overtimes", base_url))
                .json(&body)
                .send()
                .await?;
            print_response(resp).await?;
        }
        OvertimeCommands::Approve { id } => {
            let resp = client
                .post(&format!("{}/api/overtimes/approve", base_url))
                .json(&json!({ "overtime_id": id }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        OvertimeCommands::Reject { id } => {
            let resp = client
                .post(&format!("{}/api/overtimes/reject", base_url))
                .json(&json!({ "overtime_id": id }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        OvertimeCommands::List { employee_id } => {
            let url = match employee_id {
                Some(id) => format!("{}/api/overtimes?employee_id={}", base_url, id),
                None => format!("{}/api/overtimes", base_url),
            };
            let resp = client.get(&url).send().await?;
            print_response(resp).await?;
        }
    }
    Ok(())
}

async fn handle_leave(
    client: &Client,
    base_url: &str,
    action: LeaveCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        LeaveCommands::Create { employee_id, start_time, end_time } => {
            let start = parse_datetime(&start_time);
            let end = parse_datetime(&end_time);
            
            let resp = client
                .post(&format!("{}/api/leaves", base_url))
                .json(&json!({
                    "employee_id": employee_id,
                    "start_time": start,
                    "end_time": end
                }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        LeaveCommands::Approve { id } => {
            let resp = client
                .post(&format!("{}/api/leaves/approve", base_url))
                .json(&json!({ "leave_id": id }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        LeaveCommands::Reject { id } => {
            let resp = client
                .post(&format!("{}/api/leaves/reject", base_url))
                .json(&json!({ "leave_id": id }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        LeaveCommands::Cancel { id } => {
            let resp = client
                .post(&format!("{}/api/leaves/cancel", base_url))
                .json(&json!({ "leave_id": id }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        LeaveCommands::List { employee_id } => {
            let url = match employee_id {
                Some(id) => format!("{}/api/leaves?employee_id={}", base_url, id),
                None => format!("{}/api/leaves", base_url),
            };
            let resp = client.get(&url).send().await?;
            print_response(resp).await?;
        }
    }
    Ok(())
}

async fn handle_holiday(
    client: &Client,
    base_url: &str,
    action: HolidayCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        HolidayCommands::Create { date, name } => {
            let dt = parse_date(&date);
            let resp = client
                .post(&format!("{}/api/holidays", base_url))
                .json(&json!({
                    "date": dt,
                    "name": name
                }))
                .send()
                .await?;
            print_response(resp).await?;
        }
        HolidayCommands::List => {
            let resp = client.get(&format!("{}/api/holidays", base_url)).send().await?;
            print_response(resp).await?;
        }
    }
    Ok(())
}

async fn print_response(resp: reqwest::Response) -> Result<(), Box<dyn std::error::Error>> {
    let status = resp.status();
    let body = resp.text().await?;
    
    if status.is_success() {
        if let Ok(json) = serde_json::from_str::<serde_json::Value>(&body) {
            println!("{}", serde_json::to_string_pretty(&json)?);
        } else {
            println!("{}", body);
        }
    } else {
        eprintln!("Error {}: {}", status.as_u16(), body);
    }
    
    Ok(())
}
