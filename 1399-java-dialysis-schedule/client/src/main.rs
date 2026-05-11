

use chrono::{NaiveDate, Weekday};
use clap::{Parser, Subcommand};
use dialysis_core::*;
use reqwest::blocking::Client;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, env = "DIALYSIS_SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Patient {
        #[command(subcommand)]
        action: PatientCommand,
    },
    Machine {
        #[command(subcommand)]
        action: MachineCommand,
    },
    Schedule {
        #[command(subcommand)]
        action: ScheduleCommand,
    },
    Maintenance {
        #[command(subcommand)]
        action: MaintenanceCommand,
    },
}

#[derive(Subcommand, Debug)]
enum PatientCommand {
    List,
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long, value_parser = parse_disease)]
        disease: DiseaseType,
        #[arg(short, long, default_value_t = 3)]
        frequency: u8,
        #[arg(short, long, value_delimiter = ',', value_parser = parse_weekday)]
        days: Vec<Weekday>,
        #[arg(short, long, value_delimiter = ',', value_parser = parse_timeslot)]
        slots: Vec<TimeSlot>,
    },
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum MachineCommand {
    List,
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(long)]
        hepatitis_b_only: bool,
    },
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ScheduleCommand {
    List,
    ListByDate {
        #[arg(short, long, value_parser = parse_date)]
        date: NaiveDate,
    },
    Create {
        #[arg(short, long)]
        patient: Uuid,
        #[arg(short, long)]
        machine: Uuid,
        #[arg(short, long, value_parser = parse_date)]
        date: NaiveDate,
        #[arg(short, long, value_parser = parse_timeslot)]
        slot: TimeSlot,
    },
    BulkCreate {
        #[arg(short, long, value_parser = parse_bulk_entry)]
        entries: Vec<CreateScheduleRequest>,
    },
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Cancel {
        #[arg(short, long)]
        id: Uuid,
    },
    Reschedule {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, value_parser = parse_date)]
        new_date: NaiveDate,
        #[arg(short, long, value_parser = parse_timeslot)]
        new_slot: TimeSlot,
        #[arg(long)]
        new_machine: Option<Uuid>,
    },
}

#[derive(Subcommand, Debug)]
enum MaintenanceCommand {
    List,
    Create {
        #[arg(short, long)]
        machine: Uuid,
        #[arg(short, long, value_parser = parse_date)]
        date: NaiveDate,
        #[arg(short, long, value_parser = parse_timeslot)]
        slot: TimeSlot,
        #[arg(long)]
        emergency: bool,
        #[arg(short, long)]
        description: String,
    },
    Cancel {
        #[arg(short, long)]
        id: Uuid,
    },
}

fn parse_disease(s: &str) -> Result<DiseaseType, String> {
    match s.to_lowercase().as_str() {
        "none" => Ok(DiseaseType::None),
        "hepatitisb" | "hbv" | "乙肝" => Ok(DiseaseType::HepatitisB),
        "hepatitisc" | "hcv" | "丙肝" => Ok(DiseaseType::HepatitisC),
        other => Ok(DiseaseType::Other(other.to_string())),
    }
}

fn parse_weekday(s: &str) -> Result<Weekday, String> {
    match s.to_lowercase().as_str() {
        "mon" | "monday" | "周一" | "星期一" => Ok(Weekday::Mon),
        "tue" | "tuesday" | "周二" | "星期二" => Ok(Weekday::Tue),
        "wed" | "wednesday" | "周三" | "星期三" => Ok(Weekday::Wed),
        "thu" | "thursday" | "周四" | "星期四" => Ok(Weekday::Thu),
        "fri" | "friday" | "周五" | "星期五" => Ok(Weekday::Fri),
        "sat" | "saturday" | "周六" | "星期六" => Ok(Weekday::Sat),
        "sun" | "sunday" | "周日" | "星期日" => Ok(Weekday::Sun),
        _ => Err(format!("无效的星期: {}", s)),
    }
}

fn parse_timeslot(s: &str) -> Result<TimeSlot, String> {
    match s.to_lowercase().as_str() {
        "morning" | "am" | "上午" => Ok(TimeSlot::Morning),
        "afternoon" | "pm" | "下午" => Ok(TimeSlot::Afternoon),
        "evening" | "night" | "晚上" => Ok(TimeSlot::Evening),
        _ => Err(format!("无效的时段: {}", s)),
    }
}

fn parse_date(s: &str) -> Result<NaiveDate, String> {
    NaiveDate::parse_from_str(s, "%Y-%m-%d")
        .map_err(|e| format!("无效的日期格式 (YYYY-MM-DD): {}", e))
}

fn parse_bulk_entry(s: &str) -> Result<CreateScheduleRequest, String> {
    let parts: Vec<&str> = s.split(',').collect();
    if parts.len() != 4 {
        return Err(format!("格式应为: patient_id,machine_id,date,slot"));
    }
    Ok(CreateScheduleRequest {
        patient_id: Uuid::parse_str(parts[0]).map_err(|e| e.to_string())?,
        machine_id: Uuid::parse_str(parts[1]).map_err(|e| e.to_string())?,
        date: parse_date(parts[2])?,
        time_slot: parse_timeslot(parts[3])?,
    })
}

fn print_patient(p: &Patient) {
    let disease_str = match &p.disease {
        DiseaseType::None => "无",
        DiseaseType::HepatitisB => "乙肝",
        DiseaseType::HepatitisC => "丙肝",
        DiseaseType::Other(o) => o,
    };
    let days: Vec<_> = p.preferred_days.iter().map(|d| format!("{:?}", d)).collect();
    let slots: Vec<_> = p.preferred_slots.iter().map(|s| s.to_string()).collect();
    println!(
        "  ID: {}\n    姓名: {}\n    传染病: {}\n    每周透析次数: {}\n    偏好日期: {}\n    偏好时段: {}",
        p.id,
        p.name,
        disease_str,
        p.dialysis_frequency_per_week,
        days.join(", "),
        slots.join(", ")
    );
}

fn print_machine(m: &Machine) {
    println!(
        "  ID: {}\n    名称: {}\n    类型: {}",
        m.id,
        m.name,
        if m.is_hepatitis_b_only {
            "乙肝专用"
        } else {
            "普通"
        }
    );
}

fn print_schedule(s: &ScheduleEntry) {
    println!(
        "  ID: {}\n    患者: {}\n    机器: {}\n    日期: {}\n    时段: {}",
        s.id,
        s.patient_id,
        s.machine_id,
        s.date,
        s.time_slot
    );
}

fn print_maintenance(m: &Maintenance) {
    println!(
        "  ID: {}\n    机器: {}\n    日期: {}\n    时段: {}\n    紧急: {}\n    描述: {}",
        m.id,
        m.machine_id,
        m.date,
        m.time_slot,
        m.is_emergency,
        m.description
    );
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Patient { action } => match action {
            PatientCommand::List => {
                let patients: Vec<Patient> = client
                    .get(&format!("{}/patients", base_url))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("患者列表 ({} 人):", patients.len());
                for p in patients {
                    print_patient(&p);
                }
            }
            PatientCommand::Create {
                name,
                disease,
                frequency,
                days,
                slots,
            } => {
                let req = CreatePatientRequest {
                    name,
                    disease,
                    dialysis_frequency_per_week: frequency,
                    preferred_days: days,
                    preferred_slots: slots,
                };
                let patient: Patient = client
                    .post(&format!("{}/patients", base_url))
                    .json(&req)
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("已创建患者:");
                print_patient(&patient);
            }
            PatientCommand::Get { id } => {
                let patient: Patient = client
                    .get(&format!("{}/patients/{}", base_url, id))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("患者详情:");
                print_patient(&patient);
            }
        },
        Commands::Machine { action } => match action {
            MachineCommand::List => {
                let machines: Vec<Machine> = client
                    .get(&format!("{}/machines", base_url))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("机器列表 ({} 台):", machines.len());
                for m in machines {
                    print_machine(&m);
                }
            }
            MachineCommand::Create {
                name,
                hepatitis_b_only,
            } => {
                let req = CreateMachineRequest {
                    name,
                    is_hepatitis_b_only: hepatitis_b_only,
                };
                let machine: Machine = client
                    .post(&format!("{}/machines", base_url))
                    .json(&req)
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("已创建机器:");
                print_machine(&machine);
            }
            MachineCommand::Get { id } => {
                let machine: Machine = client
                    .get(&format!("{}/machines/{}", base_url, id))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("机器详情:");
                print_machine(&machine);
            }
        },
        Commands::Schedule { action } => match action {
            ScheduleCommand::List => {
                let schedules: Vec<ScheduleEntry> = client
                    .get(&format!("{}/schedules", base_url))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("排班列表 ({} 条):", schedules.len());
                for s in schedules {
                    print_schedule(&s);
                }
            }
            ScheduleCommand::ListByDate { date } => {
                let schedules: Vec<ScheduleEntry> = client
                    .get(&format!("{}/schedules/date/{}", base_url, date))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("{} 的排班 ({} 条):", date, schedules.len());
                for s in schedules {
                    print_schedule(&s);
                }
            }
            ScheduleCommand::Create {
                patient,
                machine,
                date,
                slot,
            } => {
                let req = CreateScheduleRequest {
                    patient_id: patient,
                    machine_id: machine,
                    date,
                    time_slot: slot,
                };
                let schedule: ScheduleEntry = client
                    .post(&format!("{}/schedules", base_url))
                    .json(&req)
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("已创建排班:");
                print_schedule(&schedule);
            }
            ScheduleCommand::BulkCreate { entries } => {
                let req = BulkCreateScheduleRequest { schedules: entries };
                let schedules: Vec<ScheduleEntry> = client
                    .post(&format!("{}/schedules/bulk", base_url))
                    .json(&req)
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("批量创建排班 ({} 条):", schedules.len());
                for s in schedules {
                    print_schedule(&s);
                }
            }
            ScheduleCommand::Get { id } => {
                let schedule: ScheduleEntry = client
                    .get(&format!("{}/schedules/{}", base_url, id))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("排班详情:");
                print_schedule(&schedule);
            }
            ScheduleCommand::Cancel { id } => {
                let schedule: ScheduleEntry = client
                    .delete(&format!("{}/schedules/{}", base_url, id))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("已取消排班:");
                print_schedule(&schedule);
            }
            ScheduleCommand::Reschedule {
                id,
                new_date,
                new_slot,
                new_machine,
            } => {
                let req = RescheduleRequest {
                    schedule_id: id,
                    new_date,
                    new_time_slot: new_slot,
                    new_machine_id: new_machine,
                };
                let schedule: ScheduleEntry = client
                    .post(&format!("{}/schedules/reschedule", base_url))
                    .json(&req)
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("已调班:");
                print_schedule(&schedule);
            }
        },
        Commands::Maintenance { action } => match action {
            MaintenanceCommand::List => {
                let maintenances: Vec<Maintenance> = client
                    .get(&format!("{}/maintenances", base_url))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("维护列表 ({} 条):", maintenances.len());
                for m in maintenances {
                    print_maintenance(&m);
                }
            }
            MaintenanceCommand::Create {
                machine,
                date,
                slot,
                emergency,
                description,
            } => {
                let req = CreateMaintenanceRequest {
                    machine_id: machine,
                    date,
                    time_slot: slot,
                    is_emergency: emergency,
                    description,
                };
                let resp = client
                    .post(&format!("{}/maintenances", base_url))
                    .json(&req)
                    .send()?
                    .error_for_status()?
                    .json::<serde_json::Value>()?;
                println!("已创建维护:");
                let maint: Maintenance = serde_json::from_value(resp["maintenance"].clone())?;
                print_maintenance(&maint);
                let affected: Vec<serde_json::Value> =
                    serde_json::from_value(resp["affected"].clone())?;
                if !affected.is_empty() {
                    println!("  受影响的排班 ({} 条):", affected.len());
                    for a in affected {
                        println!(
                            "    原始排班: {}, 取消: {}, 消息: {}",
                            a["original_schedule_id"], a["canceled"], a["message"]
                        );
                    }
                }
            }
            MaintenanceCommand::Cancel { id } => {
                let maint: Maintenance = client
                    .delete(&format!("{}/maintenances/{}", base_url, id))
                    .send()?
                    .error_for_status()?
                    .json()?;
                println!("已取消维护:");
                print_maintenance(&maint);
            }
        },
    }

    Ok(())
}
