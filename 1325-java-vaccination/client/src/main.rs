use clap::{Parser, Subcommand};
use chrono::{DateTime, Utc};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(long, env = "VACCINATION_SERVER", default_value = "http://127.0.0.1:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Vaccines {
        #[command(subcommand)]
        action: VaccineCommands,
    },
    Persons {
        #[command(subcommand)]
        action: PersonCommands,
    },
    Vaccinations {
        #[command(subcommand)]
        action: VaccinationCommands,
    },
}

#[derive(Subcommand, Debug)]
enum VaccineCommands {
    List,
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        total_doses: u32,
        #[arg(long, value_delimiter = ',')]
        intervals: Vec<u32>,
        #[arg(long, value_delimiter = ',', default_value = "")]
        contraindications: Vec<String>,
    },
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum PersonCommands {
    List,
    Create {
        #[arg(long)]
        name: String,
        #[arg(long, value_delimiter = ',', default_value = "")]
        contraindications: Vec<String>,
    },
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum VaccinationCommands {
    List,
    Validate {
        #[arg(long)]
        person_id: Uuid,
        #[arg(long)]
        vaccine_id: Uuid,
        #[arg(long)]
        date: String,
    },
    Record {
        #[arg(long)]
        person_id: Uuid,
        #[arg(long)]
        vaccine_id: Uuid,
        #[arg(long)]
        date: String,
    },
    Void {
        #[arg(long)]
        id: Uuid,
    },
    Get {
        #[arg(long)]
        id: Uuid,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateVaccineRequest {
    name: String,
    total_doses: u32,
    dose_schedules: Vec<VaccineDoseSchedule>,
    contraindications: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
struct VaccineDoseSchedule {
    dose_number: u32,
    min_interval_days: u32,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreatePersonRequest {
    name: String,
    contraindications: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ValidateVaccinationRequest {
    person_id: Uuid,
    vaccine_id: Uuid,
    vaccination_date: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize)]
struct RecordVaccinationRequest {
    person_id: Uuid,
    vaccine_id: Uuid,
    vaccination_date: DateTime<Utc>,
}

fn parse_date(date_str: &str) -> Result<DateTime<Utc>, String> {
    if let Ok(dt) = DateTime::parse_from_rfc3339(date_str) {
        return Ok(dt.with_timezone(&Utc));
    }
    if let Ok(naive) = chrono::NaiveDate::parse_from_str(date_str, "%Y-%m-%d") {
        let naive_dt = naive.and_hms_opt(0, 0, 0).ok_or("Invalid time")?;
        return Ok(DateTime::<Utc>::from_naive_utc_and_offset(naive_dt, Utc));
    }
    Err(format!("无法解析日期: {}. 请使用 RFC3339 格式 或 YYYY-MM-DD 格式", date_str))
}

async fn list_vaccines(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/vaccines", base_url);
    let resp = client.get(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        let vaccines: Vec<serde_json::Value> = serde_json::from_str(&text)?;
        println!("疫苗列表:");
        for v in vaccines {
            println!("  ID: {}", v["id"]);
            println!("    名称: {}", v["name"]);
            println!("    总针数: {}", v["total_doses"]);
            println!("    针次间隔: {:?}", v["dose_schedules"]);
            println!("    禁忌症: {:?}", v["contraindications"]);
            println!();
        }
    } else {
        eprintln!("错误: {}", text);
    }
    Ok(())
}

async fn create_vaccine(
    client: &Client,
    base_url: &str,
    name: String,
    total_doses: u32,
    intervals: Vec<u32>,
    contraindications: Vec<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut dose_schedules = Vec::new();
    for (i, &interval) in intervals.iter().enumerate() {
        dose_schedules.push(VaccineDoseSchedule {
            dose_number: (i + 1) as u32,
            min_interval_days: interval,
        });
    }

    let req = CreateVaccineRequest {
        name,
        total_doses,
        dose_schedules,
        contraindications: contraindications.into_iter().filter(|s| !s.is_empty()).collect(),
    };

    let url = format!("{}/api/vaccines", base_url);
    let resp = client.post(&url).json(&req).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("创建疫苗成功:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn get_vaccine(
    client: &Client,
    base_url: &str,
    id: Uuid,
) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/vaccines/{}", base_url, id);
    let resp = client.get(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("疫苗详情:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn list_persons(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/persons", base_url);
    let resp = client.get(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        let persons: Vec<serde_json::Value> = serde_json::from_str(&text)?;
        println!("人员列表:");
        for p in persons {
            println!("  ID: {}", p["id"]);
            println!("    姓名: {}", p["name"]);
            println!("    禁忌症: {:?}", p["contraindications"]);
            println!();
        }
    } else {
        eprintln!("错误: {}", text);
    }
    Ok(())
}

async fn create_person(
    client: &Client,
    base_url: &str,
    name: String,
    contraindications: Vec<String>,
) -> Result<(), Box<dyn std::error::Error>> {
    let req = CreatePersonRequest {
        name,
        contraindications: contraindications.into_iter().filter(|s| !s.is_empty()).collect(),
    };

    let url = format!("{}/api/persons", base_url);
    let resp = client.post(&url).json(&req).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("创建人员成功:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn get_person(
    client: &Client,
    base_url: &str,
    id: Uuid,
) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/persons/{}", base_url, id);
    let resp = client.get(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("人员详情:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn list_vaccinations(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/vaccinations", base_url);
    let resp = client.get(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        let records: Vec<serde_json::Value> = serde_json::from_str(&text)?;
        println!("接种记录列表:");
        for r in records {
            println!("  ID: {}", r["id"]);
            println!("    人员ID: {}", r["person_id"]);
            println!("    疫苗ID: {}", r["vaccine_id"]);
            println!("    针次: {}", r["dose_number"]);
            println!("    接种日期: {}", r["vaccination_date"]);
            println!("    已作废: {}", r["voided"]);
            println!();
        }
    } else {
        eprintln!("错误: {}", text);
    }
    Ok(())
}

async fn validate_vaccination(
    client: &Client,
    base_url: &str,
    person_id: Uuid,
    vaccine_id: Uuid,
    date: String,
) -> Result<(), Box<dyn std::error::Error>> {
    let vaccination_date = parse_date(&date)?;
    
    let req = ValidateVaccinationRequest {
        person_id,
        vaccine_id,
        vaccination_date,
    };

    let url = format!("{}/api/vaccinations/validate", base_url);
    let resp = client.post(&url).json(&req).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("校验结果:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn record_vaccination(
    client: &Client,
    base_url: &str,
    person_id: Uuid,
    vaccine_id: Uuid,
    date: String,
) -> Result<(), Box<dyn std::error::Error>> {
    let vaccination_date = parse_date(&date)?;
    
    let req = RecordVaccinationRequest {
        person_id,
        vaccine_id,
        vaccination_date,
    };

    let url = format!("{}/api/vaccinations", base_url);
    let resp = client.post(&url).json(&req).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("创建接种记录成功:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn void_vaccination(
    client: &Client,
    base_url: &str,
    id: Uuid,
) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/vaccinations/{}/void", base_url, id);
    let resp = client.post(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("作废记录成功:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

async fn get_vaccination(
    client: &Client,
    base_url: &str,
    id: Uuid,
) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/vaccinations/{}", base_url, id);
    let resp = client.get(&url).send().await?;
    let status = resp.status();
    let text = resp.text().await?;
    
    if status.is_success() {
        println!("接种记录详情:");
        println!("{}", serde_json::to_string_pretty(&serde_json::from_str::<serde_json::Value>(&text)?)?);
    } else {
        eprintln!("错误 ({}): {}", status, text);
    }
    Ok(())
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();

    match args.command {
        Commands::Vaccines { action } => match action {
            VaccineCommands::List => list_vaccines(&client, &args.server).await?,
            VaccineCommands::Create { name, total_doses, intervals, contraindications } => {
                create_vaccine(&client, &args.server, name, total_doses, intervals, contraindications).await?
            }
            VaccineCommands::Get { id } => get_vaccine(&client, &args.server, id).await?,
        },
        Commands::Persons { action } => match action {
            PersonCommands::List => list_persons(&client, &args.server).await?,
            PersonCommands::Create { name, contraindications } => {
                create_person(&client, &args.server, name, contraindications).await?
            }
            PersonCommands::Get { id } => get_person(&client, &args.server, id).await?,
        },
        Commands::Vaccinations { action } => match action {
            VaccinationCommands::List => list_vaccinations(&client, &args.server).await?,
            VaccinationCommands::Validate { person_id, vaccine_id, date } => {
                validate_vaccination(&client, &args.server, person_id, vaccine_id, date).await?
            }
            VaccinationCommands::Record { person_id, vaccine_id, date } => {
                record_vaccination(&client, &args.server, person_id, vaccine_id, date).await?
            }
            VaccinationCommands::Void { id } => void_vaccination(&client, &args.server, id).await?,
            VaccinationCommands::Get { id } => get_vaccination(&client, &args.server, id).await?,
        },
    }

    Ok(())
}
