use checkin_core::{
    ApiResponse, CheckinRequest, CheckinRecord, CheckinReport, CheckoutRequest,
    CreateActivityRequest, CreateEmployeeRequest, GpsCoordinate,
};
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::de::DeserializeOwned;
use std::time::Duration;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "CHECKIN_SERVER", default_value = "http://localhost:8100")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Employee {
        #[command(subcommand)]
        action: EmployeeCommands,
    },
    Activity {
        #[command(subcommand)]
        action: ActivityCommands,
    },
    Checkin {
        #[arg(long)]
        activity_id: Uuid,
        #[arg(long)]
        employee_id: Uuid,
        #[arg(long)]
        device_id: String,
        #[arg(long)]
        latitude: f64,
        #[arg(long)]
        longitude: f64,
    },
    Checkout {
        #[arg(long)]
        activity_id: Uuid,
        #[arg(long)]
        employee_id: Uuid,
        #[arg(long)]
        latitude: f64,
        #[arg(long)]
        longitude: f64,
    },
    Report {
        #[arg(long)]
        activity_id: Uuid,
    },
    Record {
        #[arg(long)]
        activity_id: Uuid,
        #[arg(long)]
        employee_id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        employee_number: String,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ActivityCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        checkin_start: String,
        #[arg(long)]
        checkin_end: String,
        #[arg(long)]
        checkout_deadline: String,
        #[arg(long)]
        latitude: f64,
        #[arg(long)]
        longitude: f64,
        #[arg(long)]
        allowed_distance: f64,
        #[arg(long, value_delimiter = ',')]
        employee_ids: Vec<Uuid>,
    },
    List,
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        Self {
            client: Client::builder()
                .timeout(Duration::from_secs(30))
                .build()
                .unwrap(),
            base_url,
        }
    }

    async fn post<T: DeserializeOwned, B: serde::Serialize>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, Box<dyn std::error::Error>> {
        let url = format!("{}{}", self.base_url, path);
        let resp = self.client.post(&url).json(body).send().await?;
        let api_resp: ApiResponse<T> = resp.json().await?;
        if api_resp.success {
            Ok(api_resp.data.unwrap())
        } else {
            Err(api_resp.message.into())
        }
    }

    async fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T, Box<dyn std::error::Error>> {
        let url = format!("{}{}", self.base_url, path);
        let resp = self.client.get(&url).send().await?;
        let api_resp: ApiResponse<T> = resp.json().await?;
        if api_resp.success {
            Ok(api_resp.data.unwrap())
        } else {
            Err(api_resp.message.into())
        }
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let api_client = ApiClient::new(args.server);

    let result = match args.command {
        Commands::Employee { action } => handle_employee(&api_client, action).await,
        Commands::Activity { action } => handle_activity(&api_client, action).await,
        Commands::Checkin {
            activity_id,
            employee_id,
            device_id,
            latitude,
            longitude,
        } => {
            handle_checkin(
                &api_client,
                activity_id,
                employee_id,
                device_id,
                latitude,
                longitude,
            )
            .await
        }
        Commands::Checkout {
            activity_id,
            employee_id,
            latitude,
            longitude,
        } => handle_checkout(&api_client, activity_id, employee_id, latitude, longitude).await,
        Commands::Report { activity_id } => handle_report(&api_client, activity_id).await,
        Commands::Record {
            activity_id,
            employee_id,
        } => handle_record(&api_client, activity_id, employee_id).await,
    };

    match result {
        Ok(_) => {}
        Err(e) => eprintln!("错误: {}", e),
    }
}

async fn handle_employee(
    client: &ApiClient,
    action: EmployeeCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        EmployeeCommands::Create {
            name,
            employee_number,
        } => {
            let emp: checkin_core::Employee = client
                .post(
                    "/api/employees",
                    &CreateEmployeeRequest {
                        name,
                        employee_number,
                    },
                )
                .await?;
            println!("创建员工成功:");
            println!("  ID: {}", emp.id);
            println!("  姓名: {}", emp.name);
            println!("  工号: {}", emp.employee_number);
        }
        EmployeeCommands::List => {
            let employees: Vec<checkin_core::Employee> = client.get("/api/employees").await?;
            println!("员工列表 (共 {} 人):", employees.len());
            for emp in employees {
                println!("  - {} ({}) [ID: {}]", emp.name, emp.employee_number, emp.id);
            }
        }
        EmployeeCommands::Get { id } => {
            let emp: checkin_core::Employee =
                client.get(&format!("/api/employees/{}", id)).await?;
            println!("员工信息:");
            println!("  ID: {}", emp.id);
            println!("  姓名: {}", emp.name);
            println!("  工号: {}", emp.employee_number);
        }
    }
    Ok(())
}

async fn handle_activity(
    client: &ApiClient,
    action: ActivityCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match action {
        ActivityCommands::Create {
            name,
            checkin_start,
            checkin_end,
            checkout_deadline,
            latitude,
            longitude,
            allowed_distance,
            employee_ids,
        } => {
            let checkin_start = chrono::DateTime::parse_from_rfc3339(&checkin_start)?
                .with_timezone(&chrono::Utc);
            let checkin_end = chrono::DateTime::parse_from_rfc3339(&checkin_end)?
                .with_timezone(&chrono::Utc);
            let checkout_deadline = chrono::DateTime::parse_from_rfc3339(&checkout_deadline)?
                .with_timezone(&chrono::Utc);

            let act: checkin_core::Activity = client
                .post(
                    "/api/activities",
                    &CreateActivityRequest {
                        name,
                        checkin_start,
                        checkin_end,
                        checkout_deadline,
                        location: GpsCoordinate {
                            latitude,
                            longitude,
                        },
                        allowed_distance_meters: allowed_distance,
                        employee_ids,
                    },
                )
                .await?;
            println!("创建活动成功:");
            println!("  ID: {}", act.id);
            println!("  名称: {}", act.name);
            println!("  签到开始: {}", act.checkin_start);
            println!("  签到结束: {}", act.checkin_end);
            println!("  签退截止: {}", act.checkout_deadline);
            println!(
                "  位置: ({}, {})",
                act.location.latitude, act.location.longitude
            );
            println!("  允许偏差: {} 米", act.allowed_distance_meters);
        }
        ActivityCommands::List => {
            let activities: Vec<checkin_core::Activity> = client.get("/api/activities").await?;
            println!("活动列表 (共 {} 个):", activities.len());
            for act in activities {
                println!("  - {} [ID: {}]", act.name, act.id);
                println!("    签到时间: {} - {}", act.checkin_start, act.checkin_end);
            }
        }
        ActivityCommands::Get { id } => {
            let act: checkin_core::Activity =
                client.get(&format!("/api/activities/{}", id)).await?;
            println!("活动信息:");
            println!("  ID: {}", act.id);
            println!("  名称: {}", act.name);
            println!("  签到开始: {}", act.checkin_start);
            println!("  签到结束: {}", act.checkin_end);
            println!("  签退截止: {}", act.checkout_deadline);
            println!(
                "  位置: ({}, {})",
                act.location.latitude, act.location.longitude
            );
            println!("  允许偏差: {} 米", act.allowed_distance_meters);
        }
    }
    Ok(())
}

async fn handle_checkin(
    client: &ApiClient,
    activity_id: Uuid,
    employee_id: Uuid,
    device_id: String,
    latitude: f64,
    longitude: f64,
) -> Result<(), Box<dyn std::error::Error>> {
    let record: CheckinRecord = client
        .post(
            &format!("/api/activities/{}/checkin", activity_id),
            &CheckinRequest {
                employee_id,
                device_id,
                location: GpsCoordinate {
                    latitude,
                    longitude,
                },
            },
        )
        .await?;
    println!("签到成功:");
    println!("  员工 ID: {}", record.employee_id);
    println!("  设备 ID: {}", record.device_id);
    println!("  签到时间: {:?}", record.checkin_time);
    println!("  状态: {:?}", record.status);
    Ok(())
}

async fn handle_checkout(
    client: &ApiClient,
    activity_id: Uuid,
    employee_id: Uuid,
    latitude: f64,
    longitude: f64,
) -> Result<(), Box<dyn std::error::Error>> {
    let record: CheckinRecord = client
        .post(
            &format!("/api/activities/{}/checkout", activity_id),
            &CheckoutRequest {
                employee_id,
                location: GpsCoordinate {
                    latitude,
                    longitude,
                },
            },
        )
        .await?;
    println!("签退成功:");
    println!("  员工 ID: {}", record.employee_id);
    println!("  签退时间: {:?}", record.checkout_time);
    println!("  状态: {:?}", record.status);
    Ok(())
}

async fn handle_report(
    client: &ApiClient,
    activity_id: Uuid,
) -> Result<(), Box<dyn std::error::Error>> {
    let report: CheckinReport = client
        .get(&format!("/api/activities/{}/report", activity_id))
        .await?;
    println!("签到报告:");
    println!("  应到人数: {}", report.total_count);
    println!("  实到人数: {}", report.checked_in_count);
    println!("  签到率: {:.1}%", report.checkin_rate * 100.0);
    println!("  正常: {} 人", report.normal_count);
    println!("  未签退: {} 人", report.not_checked_out_count);
    println!("  早退: {} 人", report.early_checkout_count);
    println!("  未签到: {} 人", report.not_checked_in_count);
    Ok(())
}

async fn handle_record(
    client: &ApiClient,
    activity_id: Uuid,
    employee_id: Uuid,
) -> Result<(), Box<dyn std::error::Error>> {
    let record: CheckinRecord = client
        .get(&format!(
            "/api/activities/{}/records/{}",
            activity_id, employee_id
        ))
        .await?;
    println!("签到记录:");
    println!("  活动 ID: {}", record.activity_id);
    println!("  员工 ID: {}", record.employee_id);
    println!("  设备 ID: {}", record.device_id);
    println!("  签到时间: {:?}", record.checkin_time);
    println!("  签退时间: {:?}", record.checkout_time);
    println!("  状态: {:?}", record.status);
    Ok(())
}
