use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use std::env;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "DORM_SERVER_URL", default_value = "http://127.0.0.1:8608")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Building {
        #[command(subcommand)]
        cmd: BuildingCommands,
    },
    Room {
        #[command(subcommand)]
        cmd: RoomCommands,
    },
    Bed {
        #[command(subcommand)]
        cmd: BedCommands,
    },
    Student {
        #[command(subcommand)]
        cmd: StudentCommands,
    },
    Allocate {
        department: String,
    },
    Swap {
        #[command(subcommand)]
        cmd: SwapCommands,
    },
    Stats,
}

#[derive(Subcommand, Debug)]
enum BuildingCommands {
    List,
    Create {
        name: String,
        floor_count: u32,
        rooms_per_floor: u32,
    },
    Get {
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum RoomCommands {
    List {
        #[arg(long)]
        building_id: Option<String>,
    },
    Create {
        building_id: String,
        floor_number: u32,
        room_number: u32,
        #[arg(short, long, default_value = "four")]
        room_type: String,
    },
    Get {
        id: String,
    },
    Maintenance {
        id: String,
        #[arg(short, long)]
        status: String,
    },
    Hygiene {
        id: String,
        score: String,
    },
}

#[derive(Subcommand, Debug)]
enum BedCommands {
    List {
        #[arg(long)]
        room_id: Option<String>,
    },
    Get {
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum StudentCommands {
    List {
        #[arg(long)]
        department: Option<String>,
        #[arg(long)]
        status: Option<String>,
    },
    Create {
        student_id: String,
        name: String,
        department: String,
    },
    Get {
        id: String,
    },
    Checkout {
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum SwapCommands {
    List {
        #[arg(long)]
        status: Option<String>,
    },
    Create {
        requester_id: String,
        target_id: String,
    },
    Get {
        id: String,
    },
    Respond {
        id: String,
        #[arg(short, long)]
        accept: bool,
    },
}

#[derive(Debug, Deserialize, Serialize)]
struct ApiResponse<T> {
    data: Option<T>,
    error: Option<String>,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Commands::Building { cmd } => handle_building(&client, &base_url, cmd),
        Commands::Room { cmd } => handle_room(&client, &base_url, cmd),
        Commands::Bed { cmd } => handle_bed(&client, &base_url, cmd),
        Commands::Student { cmd } => handle_student(&client, &base_url, cmd),
        Commands::Allocate { department } => handle_allocate(&client, &base_url, department),
        Commands::Swap { cmd } => handle_swap(&client, &base_url, cmd),
        Commands::Stats => handle_stats(&client, &base_url),
    }
}

fn handle_building(client: &Client, base_url: &str, cmd: BuildingCommands) -> Result<(), Box<dyn std::error::Error>> {
    match cmd {
        BuildingCommands::List => {
            let url = format!("{}/api/buildings", base_url);
            let response = client.get(&url).send()?;
            let buildings: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&buildings)?);
        }
        BuildingCommands::Create { name, floor_count, rooms_per_floor } => {
            let url = format!("{}/api/buildings", base_url);
            let body = serde_json::json!({
                "name": name,
                "floor_count": floor_count,
                "rooms_per_floor": rooms_per_floor
            });
            let response = client.post(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        BuildingCommands::Get { id } => {
            let url = format!("{}/api/buildings/{}", base_url, id);
            let response = client.get(&url).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
    }
    Ok(())
}

fn handle_room(client: &Client, base_url: &str, cmd: RoomCommands) -> Result<(), Box<dyn std::error::Error>> {
    match cmd {
        RoomCommands::List { building_id } => {
            let mut url = format!("{}/api/rooms", base_url);
            if let Some(bid) = building_id {
                url = format!("{}?building_id={}", url, bid);
            }
            let response = client.get(&url).send()?;
            let rooms: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&rooms)?);
        }
        RoomCommands::Create { building_id, floor_number, room_number, room_type } => {
            let url = format!("{}/api/rooms", base_url);
            let room_type_value = match room_type.to_lowercase().as_str() {
                "six" => "SixPerson",
                _ => "FourPerson",
            };
            let body = serde_json::json!({
                "building_id": building_id,
                "floor_number": floor_number,
                "room_number": room_number,
                "room_type": room_type_value
            });
            let response = client.post(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        RoomCommands::Get { id } => {
            let url = format!("{}/api/rooms/{}", base_url, id);
            let response = client.get(&url).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        RoomCommands::Maintenance { id, status } => {
            let url = format!("{}/api/rooms/{}/maintenance", base_url, id);
            let status_value = match status.to_lowercase().as_str() {
                "maintenance" | "under_maintenance" => "UnderMaintenance",
                _ => "Normal",
            };
            let body = serde_json::json!({ "status": status_value });
            let response = client.put(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        RoomCommands::Hygiene { id, score } => {
            let url = format!("{}/api/rooms/{}/hygiene", base_url, id);
            let body = serde_json::json!({ "score": score });
            let response = client.put(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
    }
    Ok(())
}

fn handle_bed(client: &Client, base_url: &str, cmd: BedCommands) -> Result<(), Box<dyn std::error::Error>> {
    match cmd {
        BedCommands::List { room_id } => {
            let mut url = format!("{}/api/beds", base_url);
            if let Some(rid) = room_id {
                url = format!("{}?building_id={}", url, rid);
            }
            let response = client.get(&url).send()?;
            let beds: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&beds)?);
        }
        BedCommands::Get { id } => {
            let url = format!("{}/api/beds/{}", base_url, id);
            let response = client.get(&url).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
    }
    Ok(())
}

fn handle_student(client: &Client, base_url: &str, cmd: StudentCommands) -> Result<(), Box<dyn std::error::Error>> {
    match cmd {
        StudentCommands::List { department, status } => {
            let mut url = format!("{}/api/students", base_url);
            let mut params = Vec::new();
            if let Some(dept) = department {
                params.push(format!("department={}", dept));
            }
            if let Some(st) = status {
                params.push(format!("status={}", st));
            }
            if !params.is_empty() {
                url = format!("{}?{}", url, params.join("&"));
            }
            let response = client.get(&url).send()?;
            let students: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&students)?);
        }
        StudentCommands::Create { student_id, name, department } => {
            let url = format!("{}/api/students", base_url);
            let body = serde_json::json!({
                "student_id": student_id,
                "name": name,
                "department": department
            });
            let response = client.post(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        StudentCommands::Get { id } => {
            let url = format!("{}/api/students/{}", base_url, id);
            let response = client.get(&url).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        StudentCommands::Checkout { id } => {
            let url = format!("{}/api/students/{}", base_url, id);
            let response = client.delete(&url).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
    }
    Ok(())
}

fn handle_allocate(client: &Client, base_url: &str, department: String) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/allocate", base_url);
    let body = serde_json::json!({ "department": department });
    let response = client.post(&url).json(&body).send()?;
    let result: serde_json::Value = response.json()?;
    println!("{}", serde_json::to_string_pretty(&result)?);
    Ok(())
}

fn handle_swap(client: &Client, base_url: &str, cmd: SwapCommands) -> Result<(), Box<dyn std::error::Error>> {
    match cmd {
        SwapCommands::List { status } => {
            let mut url = format!("{}/api/swap-requests", base_url);
            if let Some(st) = status {
                url = format!("{}?status={}", url, st);
            }
            let response = client.get(&url).send()?;
            let requests: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&requests)?);
        }
        SwapCommands::Create { requester_id, target_id } => {
            let url = format!("{}/api/swap-requests", base_url);
            let body = serde_json::json!({
                "requester_id": requester_id,
                "target_id": target_id
            });
            let response = client.post(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        SwapCommands::Get { id } => {
            let url = format!("{}/api/swap-requests/{}", base_url, id);
            let response = client.get(&url).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        SwapCommands::Respond { id, accept } => {
            let url = format!("{}/api/swap-requests/{}/respond", base_url, id);
            let body = serde_json::json!({ "accept": accept });
            let response = client.put(&url).json(&body).send()?;
            let result: serde_json::Value = response.json()?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
    }
    Ok(())
}

fn handle_stats(client: &Client, base_url: &str) -> Result<(), Box<dyn std::error::Error>> {
    let url = format!("{}/api/stats/available-beds", base_url);
    let response = client.get(&url).send()?;
    let stats: serde_json::Value = response.json()?;
    println!("{}", serde_json::to_string_pretty(&stats)?);
    Ok(())
}
