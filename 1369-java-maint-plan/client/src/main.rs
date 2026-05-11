use clap::{Parser, Subcommand, ValueEnum};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use maint_core::SparePartUsage;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Device {
        #[command(subcommand)]
        action: DeviceCommands,
    },
    Plan {
        #[command(subcommand)]
        action: PlanCommands,
    },
    MaintenanceOrder {
        #[command(subcommand)]
        action: MaintenanceOrderCommands,
    },
    RepairOrder {
        #[command(subcommand)]
        action: RepairOrderCommands,
    },
    SparePart {
        #[command(subcommand)]
        action: SparePartCommands,
    },
    Notification {
        #[command(subcommand)]
        action: NotificationCommands,
    },
    TriggerDailyTasks,
}

#[derive(Subcommand, Debug)]
enum DeviceCommands {
    List,
    Get {
        id: Uuid,
    },
    Create {
        name: String,
        code: String,
        #[arg(long)]
        description: String,
    },
    SetStatus {
        id: Uuid,
        #[arg(long, value_enum)]
        status: CliDeviceStatus,
    },
    SetHours {
        id: Uuid,
        hours: u64,
    },
    History {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum PlanCommands {
    List,
    GetForDevice {
        device_id: Uuid,
    },
    Create {
        device_id: Uuid,
        name: String,
        #[arg(long, value_enum)]
        cycle_type: CliCycleType,
        cycle_value: u64,
    },
}

#[derive(Subcommand, Debug)]
enum MaintenanceOrderCommands {
    List,
    GetForDevice {
        device_id: Uuid,
    },
    Complete {
        id: Uuid,
        notes: String,
    },
}

#[derive(Subcommand, Debug)]
enum RepairOrderCommands {
    List,
    GetForDevice {
        device_id: Uuid,
    },
    Create {
        device_id: Uuid,
        title: String,
        #[arg(long)]
        description: String,
    },
    Complete {
        id: Uuid,
        fault_category: String,
        #[arg(long)]
        repair_notes: String,
        #[arg(long, value_parser = parse_spare_parts)]
        spare_parts: Vec<SparePartUsage>,
    },
}

#[derive(Subcommand, Debug)]
enum SparePartCommands {
    List,
    Create {
        name: String,
        code: String,
        quantity: u32,
        safety_stock: u32,
        unit: String,
    },
}

#[derive(Subcommand, Debug)]
enum NotificationCommands {
    List,
    MarkRead {
        id: Uuid,
    },
}

#[derive(ValueEnum, Debug, Clone, Copy)]
enum CliDeviceStatus {
    Running,
    Stopped,
    Maintenance,
    Fault,
}

#[derive(ValueEnum, Debug, Clone, Copy)]
enum CliCycleType {
    CalendarTime,
    RunningHours,
}

fn parse_spare_parts(s: &str) -> Result<SparePartUsage, String> {
    let parts: Vec<&str> = s.split(':').collect();
    if parts.len() != 2 {
        return Err("格式错误，应为: spare_part_id:quantity".to_string());
    }
    let id = parts[0].parse::<Uuid>().map_err(|e| e.to_string())?;
    let qty = parts[1].parse::<u32>().map_err(|e| e.to_string())?;
    Ok(SparePartUsage {
        spare_part_id: id,
        quantity: qty,
    })
}

#[derive(Debug, Serialize)]
struct CreateDeviceRequest {
    name: String,
    code: String,
    description: String,
}

#[derive(Debug, Serialize)]
struct UpdateDeviceStatusRequest {
    status: String,
}

#[derive(Debug, Serialize)]
struct UpdateDeviceHoursRequest {
    running_hours: u64,
}

#[derive(Debug, Serialize)]
struct CreatePlanRequest {
    device_id: Uuid,
    name: String,
    cycle_type: String,
    cycle_value: u64,
}

#[derive(Debug, Serialize)]
struct CreateRepairOrderRequest {
    device_id: Uuid,
    title: String,
    description: String,
}

#[derive(Debug, Serialize)]
struct CompleteRepairOrderRequest {
    fault_category: String,
    spare_parts: Vec<SparePartUsage>,
    repair_notes: String,
}

#[derive(Debug, Serialize)]
struct CreateSparePartRequest {
    name: String,
    code: String,
    quantity: u32,
    safety_stock: u32,
    unit: String,
}

#[derive(Debug, Serialize)]
struct CompleteMaintenanceOrderRequest {
    notes: String,
}

#[derive(Debug, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

async fn handle_command(args: Args) -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::Device { action } => match action {
            DeviceCommands::List => {
                let resp = client
                    .get(format!("{}/api/devices", base_url))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            DeviceCommands::Get { id } => {
                let resp = client
                    .get(format!("{}/api/devices/{}", base_url, id))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
            DeviceCommands::Create { name, code, description } => {
                let req = CreateDeviceRequest {
                    name,
                    code,
                    description,
                };
                let resp = client
                    .post(format!("{}/api/devices", base_url))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            DeviceCommands::SetStatus { id, status } => {
                let status_str = match status {
                    CliDeviceStatus::Running => "Running",
                    CliDeviceStatus::Stopped => "Stopped",
                    CliDeviceStatus::Maintenance => "Maintenance",
                    CliDeviceStatus::Fault => "Fault",
                };
                let req = UpdateDeviceStatusRequest {
                    status: status_str.to_string(),
                };
                let resp = client
                    .put(format!("{}/api/devices/{}/status", base_url, id))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
            DeviceCommands::SetHours { id, hours } => {
                let req = UpdateDeviceHoursRequest { running_hours: hours };
                let resp = client
                    .put(format!("{}/api/devices/{}/hours", base_url, id))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
            DeviceCommands::History { id } => {
                let resp = client
                    .get(format!("{}/api/devices/{}/history", base_url, id))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
        },
        Commands::Plan { action } => match action {
            PlanCommands::List => {
                let resp = client
                    .get(format!("{}/api/maintenance-plans", base_url))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            PlanCommands::GetForDevice { device_id } => {
                let resp = client
                    .get(format!("{}/api/devices/{}/maintenance-plans", base_url, device_id))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            PlanCommands::Create {
                device_id,
                name,
                cycle_type,
                cycle_value,
            } => {
                let cycle_type_str = match cycle_type {
                    CliCycleType::CalendarTime => "CalendarTime",
                    CliCycleType::RunningHours => "RunningHours",
                };
                let req = CreatePlanRequest {
                    device_id,
                    name,
                    cycle_type: cycle_type_str.to_string(),
                    cycle_value,
                };
                let resp = client
                    .post(format!("{}/api/maintenance-plans", base_url))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
        },
        Commands::MaintenanceOrder { action } => match action {
            MaintenanceOrderCommands::List => {
                let resp = client
                    .get(format!("{}/api/maintenance-orders", base_url))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            MaintenanceOrderCommands::GetForDevice { device_id } => {
                let resp = client
                    .get(format!("{}/api/devices/{}/maintenance-orders", base_url, device_id))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            MaintenanceOrderCommands::Complete { id, notes } => {
                let req = CompleteMaintenanceOrderRequest { notes };
                let resp = client
                    .post(format!("{}/api/maintenance-orders/{}/complete", base_url, id))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
        },
        Commands::RepairOrder { action } => match action {
            RepairOrderCommands::List => {
                let resp = client
                    .get(format!("{}/api/repair-orders", base_url))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            RepairOrderCommands::GetForDevice { device_id } => {
                let resp = client
                    .get(format!("{}/api/devices/{}/repair-orders", base_url, device_id))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            RepairOrderCommands::Create {
                device_id,
                title,
                description,
            } => {
                let req = CreateRepairOrderRequest {
                    device_id,
                    title,
                    description,
                };
                let resp = client
                    .post(format!("{}/api/repair-orders", base_url))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
            RepairOrderCommands::Complete {
                id,
                fault_category,
                repair_notes,
                spare_parts,
            } => {
                let req = CompleteRepairOrderRequest {
                    fault_category,
                    spare_parts,
                    repair_notes,
                };
                let resp = client
                    .post(format!("{}/api/repair-orders/{}/complete", base_url, id))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
        },
        Commands::SparePart { action } => match action {
            SparePartCommands::List => {
                let resp = client
                    .get(format!("{}/api/spare-parts", base_url))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            SparePartCommands::Create {
                name,
                code,
                quantity,
                safety_stock,
                unit,
            } => {
                let req = CreateSparePartRequest {
                    name,
                    code,
                    quantity,
                    safety_stock,
                    unit,
                };
                let resp = client
                    .post(format!("{}/api/spare-parts", base_url))
                    .json(&req)
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
        },
        Commands::Notification { action } => match action {
            NotificationCommands::List => {
                let resp = client
                    .get(format!("{}/api/notifications", base_url))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                println!("{}", serde_json::to_string_pretty(&result.data.unwrap())?);
            }
            NotificationCommands::MarkRead { id } => {
                let resp = client
                    .post(format!("{}/api/notifications/{}/read", base_url, id))
                    .send()
                    .await?;
                let result: ApiResponse<serde_json::Value> = resp.json().await?;
                if result.success {
                    println!("通知已标记为已读");
                } else {
                    println!("错误: {}", result.message.unwrap());
                }
            }
        },
        Commands::TriggerDailyTasks => {
            let resp = client
                .post(format!("{}/api/daily-tasks", base_url))
                .send()
                .await?;
            let result: ApiResponse<serde_json::Value> = resp.json().await?;
            if result.success {
                println!("每日任务已触发");
            } else {
                println!("错误: {}", result.message.unwrap());
            }
        }
    }

    Ok(())
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    if let Err(e) = handle_command(args).await {
        eprintln!("错误: {}", e);
        std::process::exit(1);
    }
}
