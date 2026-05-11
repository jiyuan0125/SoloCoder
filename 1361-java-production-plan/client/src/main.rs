use clap::{Parser, Subcommand};
use serde_json::Value;
use std::collections::HashMap;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_URL", default_value = "http://localhost:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Line {
        #[command(subcommand)]
        action: LineCommands,
    },
    Order {
        #[command(subcommand)]
        action: OrderCommands,
    },
    Maintenance {
        #[command(subcommand)]
        action: MaintenanceCommands,
    },
    Schedule {
        #[command(subcommand)]
        action: ScheduleCommands,
    },
}

#[derive(Subcommand, Debug)]
enum LineCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long, value_delimiter = ',')]
        devices: Vec<String>,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum OrderCommands {
    Create {
        #[arg(long)]
        id: String,
        #[arg(long)]
        name: String,
        #[arg(long, default_value = "normal")]
        priority: String,
        #[arg(long)]
        duration: u32,
        #[arg(long)]
        due_date: String,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum MaintenanceCommands {
    Create {
        #[arg(long)]
        id: String,
        #[arg(long)]
        device_id: String,
        #[arg(long)]
        start: String,
        #[arg(long)]
        end: String,
        #[arg(long)]
        description: Option<String>,
    },
    Update {
        #[arg(long)]
        id: String,
        #[arg(long)]
        device_id: String,
        #[arg(long)]
        start: String,
        #[arg(long)]
        end: String,
        #[arg(long)]
        description: Option<String>,
    },
    Delete {
        #[arg(long)]
        id: String,
    },
    List,
}

#[derive(Subcommand, Debug)]
enum ScheduleCommands {
    List,
    Line {
        #[arg(long)]
        line_id: String,
    },
    Reschedule,
}

fn main() {
    let args = Args::parse();
    let client = reqwest::blocking::Client::new();

    match args.command {
        Commands::Line { action } => handle_line(&client, &args.server, action),
        Commands::Order { action } => handle_order(&client, &args.server, action),
        Commands::Maintenance { action } => handle_maintenance(&client, &args.server, action),
        Commands::Schedule { action } => handle_schedule(&client, &args.server, action),
    }
}

fn handle_line(client: &reqwest::blocking::Client, server: &str, action: LineCommands) {
    match action {
        LineCommands::Create { name, devices } => {
            let device_list: Vec<HashMap<String, String>> = devices
                .iter()
                .enumerate()
                .map(|(i, d)| {
                    let mut map = HashMap::new();
                    map.insert("id".to_string(), format!("dev_{}", i));
                    map.insert("name".to_string(), d.clone());
                    map
                })
                .collect();

            let mut body = HashMap::new();
            body.insert("name".to_string(), Value::String(name));
            body.insert("devices".to_string(), serde_json::to_value(device_list).unwrap());

            let resp = client.post(format!("{}/lines", server))
                .json(&body)
                .send()
                .unwrap();
            
            print_response(resp);
        }
        LineCommands::List => {
            let resp = client.get(format!("{}/lines", server))
                .send()
                .unwrap();
            print_response(resp);
        }
        LineCommands::Get { id } => {
            let resp = client.get(format!("{}/lines/{}", server, id))
                .send()
                .unwrap();
            print_response(resp);
        }
    }
}

fn handle_order(client: &reqwest::blocking::Client, server: &str, action: OrderCommands) {
    match action {
        OrderCommands::Create { id, name, priority, duration, due_date } => {
            let mut body = HashMap::new();
            body.insert("id".to_string(), Value::String(id));
            body.insert("name".to_string(), Value::String(name));
            body.insert("priority".to_string(), Value::String(priority));
            body.insert("duration_minutes".to_string(), Value::Number(serde_json::Number::from(duration)));
            body.insert("due_date".to_string(), Value::String(due_date));

            let resp = client.post(format!("{}/orders", server))
                .json(&body)
                .send()
                .unwrap();
            print_response(resp);
        }
        OrderCommands::List => {
            let resp = client.get(format!("{}/orders", server))
                .send()
                .unwrap();
            print_response(resp);
        }
        OrderCommands::Get { id } => {
            let resp = client.get(format!("{}/orders/{}", server, id))
                .send()
                .unwrap();
            print_response(resp);
        }
    }
}

fn handle_maintenance(client: &reqwest::blocking::Client, server: &str, action: MaintenanceCommands) {
    match action {
        MaintenanceCommands::Create { id, device_id, start, end, description } => {
            let mut body = HashMap::new();
            body.insert("id".to_string(), Value::String(id));
            body.insert("device_id".to_string(), Value::String(device_id));
            body.insert("start_time".to_string(), Value::String(start));
            body.insert("end_time".to_string(), Value::String(end));
            if let Some(desc) = description {
                body.insert("description".to_string(), Value::String(desc));
            }

            let resp = client.post(format!("{}/maintenances", server))
                .json(&body)
                .send()
                .unwrap();
            print_response(resp);
        }
        MaintenanceCommands::Update { id, device_id, start, end, description } => {
            let mut body = HashMap::new();
            body.insert("device_id".to_string(), Value::String(device_id));
            body.insert("start_time".to_string(), Value::String(start));
            body.insert("end_time".to_string(), Value::String(end));
            if let Some(desc) = description {
                body.insert("description".to_string(), Value::String(desc));
            }

            let resp = client.put(format!("{}/maintenances/{}", server, id))
                .json(&body)
                .send()
                .unwrap();
            print_response(resp);
        }
        MaintenanceCommands::Delete { id } => {
            let resp = client.delete(format!("{}/maintenances/{}", server, id))
                .send()
                .unwrap();
            print_response(resp);
        }
        MaintenanceCommands::List => {
            let resp = client.get(format!("{}/maintenances", server))
                .send()
                .unwrap();
            print_response(resp);
        }
    }
}

fn handle_schedule(client: &reqwest::blocking::Client, server: &str, action: ScheduleCommands) {
    match action {
        ScheduleCommands::List => {
            let resp = client.get(format!("{}/schedule", server))
                .send()
                .unwrap();
            print_response(resp);
        }
        ScheduleCommands::Line { line_id } => {
            let resp = client.get(format!("{}/lines/{}/schedule", server, line_id))
                .send()
                .unwrap();
            print_response(resp);
        }
        ScheduleCommands::Reschedule => {
            let resp = client.post(format!("{}/schedule", server))
                .send()
                .unwrap();
            print_response(resp);
        }
    }
}

fn print_response(resp: reqwest::blocking::Response) {
    let status = resp.status();
    let text = resp.text().unwrap();
    
    if let Ok(json) = serde_json::from_str::<Value>(&text) {
        println!("Status: {}", status);
        println!("{}", serde_json::to_string_pretty(&json).unwrap());
    } else {
        println!("Status: {}", status);
        println!("{}", text);
    }
}
