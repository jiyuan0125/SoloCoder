use chrono::{DateTime, Utc};
use clap::{Parser, Subcommand};
use maintenance_core::*;
use reqwest::Client;
use serde::Deserialize;

#[derive(Debug, Parser)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, env = "MAINTENANCE_SERVER", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Debug, Subcommand)]
enum Commands {
    Equipment {
        #[command(subcommand)]
        command: EquipmentCommands,
    },
    Maintenance {
        #[command(subcommand)]
        command: MaintenanceCommands,
    },
    Stats {
        #[command(subcommand)]
        command: StatsCommands,
    },
    Reminders,
    Records,
    Personnel,
}

#[derive(Debug, Subcommand)]
enum EquipmentCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        install_date: String,
        #[arg(long)]
        hours_interval: Option<u64>,
        #[arg(long)]
        days_interval: Option<u32>,
    },
    List,
    Get {
        id: String,
    },
    AddRuntime {
        id: String,
        #[arg(long)]
        hours: u64,
    },
}

#[derive(Debug, Subcommand)]
enum MaintenanceCommands {
    Perform {
        id: String,
        #[arg(long, value_parser = parse_type)]
        maintenance_type: MaintenanceType,
        #[arg(long)]
        date: Option<String>,
        #[arg(long)]
        labor_hours: f64,
        #[arg(long, value_delimiter = ',')]
        materials: Vec<String>,
        #[arg(long)]
        personnel: String,
        #[arg(long)]
        notes: Option<String>,
    },
    List {
        equipment_id: Option<String>,
    },
}

#[derive(Debug, Subcommand)]
enum StatsCommands {
    ByEquipment {
        equipment_id: String,
    },
    ByPersonnel {
        name: String,
    },
    ByTime {
        #[arg(long)]
        start: String,
        #[arg(long)]
        end: String,
    },
}

fn parse_type(s: &str) -> Result<MaintenanceType, String> {
    match s.to_lowercase().as_str() {
        "hourly" => Ok(MaintenanceType::Hourly),
        "calendar" => Ok(MaintenanceType::Calendar),
        "both" => Ok(MaintenanceType::Both),
        _ => Err(format!("Invalid type: {}", s)),
    }
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Equipment { command } => handle_equipment(&client, &base_url, command).await?,
        Commands::Maintenance { command } => handle_maintenance(&client, &base_url, command).await?,
        Commands::Stats { command } => handle_stats(&client, &base_url, command).await?,
        Commands::Reminders => handle_reminders(&client, &base_url).await?,
        Commands::Records => handle_all_records(&client, &base_url).await?,
        Commands::Personnel => handle_personnel(&client, &base_url).await?,
    }

    Ok(())
}

async fn handle_equipment(
    client: &Client,
    base_url: &str,
    command: EquipmentCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match command {
        EquipmentCommands::Create {
            name,
            install_date,
            hours_interval,
            days_interval,
        } => {
            let install_dt = DateTime::parse_from_rfc3339(&install_date)?.with_timezone(&Utc);
            let req = CreateEquipmentRequest {
                name,
                install_date: install_dt,
                hours_interval,
                days_interval,
            };
            let resp = client
                .post(format!("{}/equipment", base_url))
                .json(&req)
                .send()
                .await?;
            if resp.status().is_success() {
                let eq: Equipment = resp.json().await?;
                println!("设备创建成功:");
                print_equipment(&eq);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        EquipmentCommands::List => {
            let resp = client.get(format!("{}/equipment", base_url)).send().await?;
            let eqs: Vec<Equipment> = resp.json().await?;
            println!("设备列表 (共 {} 台):", eqs.len());
            for eq in eqs {
                println!("\n---");
                print_equipment(&eq);
            }
        }
        EquipmentCommands::Get { id } => {
            let resp = client.get(format!("{}/equipment/{}", base_url, id)).send().await?;
            if resp.status().is_success() {
                let eq: Equipment = resp.json().await?;
                print_equipment(&eq);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        EquipmentCommands::AddRuntime { id, hours } => {
            let req = UpdateRuntimeRequest { additional_hours: hours };
            let resp = client
                .put(format!("{}/equipment/{}/runtime", base_url, id))
                .json(&req)
                .send()
                .await?;
            if resp.status().is_success() {
                let eq: Equipment = resp.json().await?;
                println!("运行时长已更新:");
                print_equipment(&eq);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
    }
    Ok(())
}

fn print_equipment(eq: &Equipment) {
    println!("ID: {}", eq.id);
    println!("名称: {}", eq.name);
    println!("安装日期: {}", eq.install_date.format("%Y-%m-%d"));
    println!("累计运行时长: {} 小时", eq.total_runtime_hours);
    println!("保养规则:");
    if let Some(h) = eq.maintenance_rule.hours_interval {
        println!("  - 每 {} 小时", h);
    }
    if let Some(d) = eq.maintenance_rule.days_interval {
        println!("  - 每 {} 天", d);
    }
    if let Some(dt) = eq.last_maintenance_date {
        println!("上次保养日期: {}", dt.format("%Y-%m-%d"));
    }
    if let Some(rt) = eq.last_maintenance_runtime {
        println!("上次保养时长: {} 小时", rt);
    }
}

async fn handle_maintenance(
    client: &Client,
    base_url: &str,
    command: MaintenanceCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    match command {
        MaintenanceCommands::Perform {
            id,
            maintenance_type,
            date,
            labor_hours,
            materials,
            personnel,
            notes,
        } => {
            let maint_date = if let Some(d) = date {
                DateTime::parse_from_rfc3339(&d)?.with_timezone(&Utc)
            } else {
                Utc::now()
            };

            let material_usages: Vec<MaterialUsage> = materials
                .into_iter()
                .filter_map(|m| parse_material(&m))
                .collect();

            let req = PerformMaintenanceRequest {
                maintenance_type,
                date: maint_date,
                labor_hours,
                materials: material_usages,
                personnel,
                notes,
            };

            let resp = client
                .post(format!("{}/equipment/{}/maintenance", base_url, id))
                .json(&req)
                .send()
                .await?;

            if resp.status().is_success() {
                let record: MaintenanceRecord = resp.json().await?;
                println!("保养记录已创建:");
                print_record(&record);
            } else {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
            }
        }
        MaintenanceCommands::List { equipment_id } => {
            let records = if let Some(id) = equipment_id {
                let resp = client
                    .get(format!("{}/equipment/{}/records", base_url, id))
                    .send()
                    .await?;
                if !resp.status().is_success() {
                    let err: ErrorResponse = resp.json().await?;
                    eprintln!("错误: {}", err.error);
                    return Ok(());
                }
                resp.json().await?
            } else {
                let resp = client.get(format!("{}/records", base_url)).send().await?;
                resp.json().await?
            };

            let records: Vec<MaintenanceRecord> = records;
            println!("保养记录 (共 {} 条):", records.len());
            for rec in records {
                println!("\n---");
                print_record(&rec);
            }
        }
    }
    Ok(())
}

fn parse_material(s: &str) -> Option<MaterialUsage> {
    let parts: Vec<&str> = s.splitn(3, ':').collect();
    if parts.len() != 3 {
        return None;
    }
    let quantity: f64 = parts[1].parse().ok()?;
    let unit_cost: f64 = parts[2].parse().ok()?;
    Some(MaterialUsage {
        name: parts[0].to_string(),
        quantity,
        unit_cost,
    })
}

fn print_record(rec: &MaintenanceRecord) {
    println!("记录ID: {}", rec.id);
    println!("设备ID: {}", rec.equipment_id);
    println!(
        "类型: {}",
        match rec.maintenance_type {
            MaintenanceType::Hourly => "按时长",
            MaintenanceType::Calendar => "按日历",
            MaintenanceType::Both => "两者",
        }
    );
    println!("日期: {}", rec.date.format("%Y-%m-%d %H:%M:%S"));
    println!("工时: {} 小时", rec.labor_hours);
    println!("人员: {}", rec.personnel);
    if !rec.materials.is_empty() {
        println!("耗材:");
        for m in &rec.materials {
            println!(
                "  - {}: {} x {} = {}",
                m.name,
                m.quantity,
                m.unit_cost,
                m.total_cost()
            );
        }
    }
    println!("耗材费用: {}", rec.material_cost());
    println!("总费用: {}", rec.total_cost());
    println!("保养时累计时长: {} 小时", rec.runtime_at_maintenance);
    if let Some(n) = &rec.notes {
        println!("备注: {}", n);
    }
}

async fn handle_stats(
    client: &Client,
    base_url: &str,
    command: StatsCommands,
) -> Result<(), Box<dyn std::error::Error>> {
    let stats: StatisticsResult = match command {
        StatsCommands::ByEquipment { equipment_id } => {
            let resp = client
                .get(format!("{}/stats?equipment={}", base_url, equipment_id))
                .send()
                .await?;
            if !resp.status().is_success() {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
                return Ok(());
            }
            resp.json().await?
        }
        StatsCommands::ByPersonnel { name } => {
            let resp = client
                .get(format!("{}/stats?personnel={}", base_url, name))
                .send()
                .await?;
            resp.json().await?
        }
        StatsCommands::ByTime { start, end } => {
            let resp = client
                .get(format!("{}/stats?start={}&end={}", base_url, start, end))
                .send()
                .await?;
            if !resp.status().is_success() {
                let err: ErrorResponse = resp.json().await?;
                eprintln!("错误: {}", err.error);
                return Ok(());
            }
            resp.json().await?
        }
    };

    print_stats(&stats);
    Ok(())
}

fn print_stats(stats: &StatisticsResult) {
    println!("统计结果:");
    println!("  记录数: {}", stats.record_count);
    println!("  总工时: {} 小时", stats.total_labor_hours);
    println!("  耗材总费用: {}", stats.total_material_cost);
    println!("  总费用: {}", stats.total_cost);
}

async fn handle_reminders(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp = client.get(format!("{}/reminders", base_url)).send().await?;
    let reminders: Vec<MaintenanceReminder> = resp.json().await?;

    println!("保养提醒 (共 {} 台设备需要保养):", reminders.len());
    for rem in reminders {
        println!("\n---");
        println!("设备ID: {}", rem.equipment_id);
        println!("设备名称: {}", rem.equipment_name);
        println!(
            "触发类型: {}",
            match rem.trigger_type {
                MaintenanceType::Hourly => "按时长",
                MaintenanceType::Calendar => "按日历",
                MaintenanceType::Both => "两者",
            }
        );
        if let Some(h) = rem.hours_since_last {
            println!("距上次保养时长: {} 小时", h);
        }
        if let Some(d) = rem.days_since_last {
            println!("距上次保养天数: {} 天", d);
        }
    }
    Ok(())
}

async fn handle_all_records(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp = client.get(format!("{}/records", base_url)).send().await?;
    let records: Vec<MaintenanceRecord> = resp.json().await?;

    println!("所有保养记录 (共 {} 条):", records.len());
    for rec in records {
        println!("\n---");
        print_record(&rec);
    }
    Ok(())
}

async fn handle_personnel(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let resp = client.get(format!("{}/personnel", base_url)).send().await?;
    let personnel: Vec<String> = resp.json().await?;

    println!("维保人员名单 (共 {} 人):", personnel.len());
    for p in personnel {
        println!("  - {}", p);
    }
    Ok(())
}
