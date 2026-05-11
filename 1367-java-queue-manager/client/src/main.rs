use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    ListWindows,
    GetWindow { id: u32 },
    Generate {
        #[arg(short, long)]
        service_type: String,
        #[arg(short, long, default_value_t = false)]
        vip: bool,
    },
    GetTicket { display: String },
    QueueLength { service_type: String },
    Call { window_id: u32 },
    Recall { window_id: u32 },
    Arrived { window_id: u32 },
    Complete { window_id: u32 },
    CloseWindow { window_id: u32 },
    OpenWindow { window_id: u32 },
}

#[derive(Debug, Serialize, Deserialize)]
struct GenerateTicketRequest {
    service_type: String,
    is_vip: bool,
}

fn main() {
    let cli = Cli::parse();
    
    match &cli.command {
        Commands::Health => {
            let url = format!("{}/health", cli.server);
            match ureq::get(&url).call() {
                Ok(res) => {
                    println!("Status: {}", res.status());
                    if let Ok(body) = res.into_string() {
                        println!("Response: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::ListWindows => {
            let url = format!("{}/windows", cli.server);
            match ureq::get(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Windows: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::GetWindow { id } => {
            let url = format!("{}/windows/{}", cli.server, id);
            match ureq::get(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Window: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::Generate { service_type, vip } => {
            let url = format!("{}/tickets/generate", cli.server);
            let req = GenerateTicketRequest {
                service_type: service_type.clone(),
                is_vip: *vip,
            };
            
            match ureq::post(&url).send_json(&req) {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Generated ticket: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::GetTicket { display } => {
            let url = format!("{}/tickets/{}", cli.server, display);
            match ureq::get(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Ticket: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::QueueLength { service_type } => {
            let url = format!("{}/queues/{}", cli.server, service_type);
            match ureq::get(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Queue length: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::Call { window_id } => {
            let url = format!("{}/windows/{}/call", cli.server, window_id);
            match ureq::post(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Called: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::Recall { window_id } => {
            let url = format!("{}/windows/{}/recall", cli.server, window_id);
            match ureq::post(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Recalled: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::Arrived { window_id } => {
            let url = format!("{}/windows/{}/arrived", cli.server, window_id);
            match ureq::post(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Arrived: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::Complete { window_id } => {
            let url = format!("{}/windows/{}/complete", cli.server, window_id);
            match ureq::post(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Completed: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::CloseWindow { window_id } => {
            let url = format!("{}/windows/{}/close", cli.server, window_id);
            match ureq::post(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Closed: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
        
        Commands::OpenWindow { window_id } => {
            let url = format!("{}/windows/{}/open", cli.server, window_id);
            match ureq::post(&url).call() {
                Ok(res) => {
                    if let Ok(body) = res.into_string() {
                        println!("Opened: {}", body);
                    }
                }
                Err(e) => eprintln!("Error: {}", e),
            }
        }
    }
}
