use anyhow::Result;
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::Deserialize;
use travel_planner_core::{
    Attraction, Transportation, Itinerary, NewAttraction, NewTransportation,
    NewItinerary, NewItineraryNode, AddItineraryNodesRequest, TransportationType
};
use chrono::{DateTime, NaiveTime, Utc};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    Attraction {
        #[command(subcommand)]
        action: AttractionCommands,
    },
    Transportation {
        #[command(subcommand)]
        action: TransportationCommands,
    },
    Itinerary {
        #[command(subcommand)]
        action: ItineraryCommands,
    },
}

#[derive(Subcommand, Debug)]
enum AttractionCommands {
    List,
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        description: Option<String>,
        #[arg(short, long)]
        duration: u32,
        #[arg(short, long)]
        open: String,
        #[arg(short, long)]
        close: String,
    },
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum TransportationCommands {
    List,
    Create {
        #[arg(short = 'f', long)]
        from: Uuid,
        #[arg(short = 't', long)]
        to: Uuid,
        #[arg(short, long)]
        type_: String,
        #[arg(short, long)]
        duration: u32,
    },
}

#[derive(Subcommand, Debug)]
enum ItineraryCommands {
    List,
    Create {
        #[arg(short, long)]
        name: String,
    },
    Get {
        id: Uuid,
    },
    AddNode {
        #[arg(short, long)]
        itinerary: Uuid,
        #[arg(short, long)]
        attraction: Uuid,
        #[arg(short, long)]
        arrival: String,
        #[arg(short, long)]
        departure: String,
    },
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
    details: Vec<String>,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Health => {
            let response = client.get(&format!("{}/health", base_url))
                .send()
                .await?;
            println!("{}", response.text().await?);
        }
        Commands::Attraction { action } => {
            handle_attraction_commands(&client, &base_url, action).await?;
        }
        Commands::Transportation { action } => {
            handle_transportation_commands(&client, &base_url, action).await?;
        }
        Commands::Itinerary { action } => {
            handle_itinerary_commands(&client, &base_url, action).await?;
        }
    }

    Ok(())
}

async fn handle_attraction_commands(client: &Client, base_url: &str, action: AttractionCommands) -> Result<()> {
    match action {
        AttractionCommands::List => {
            let attractions: Vec<Attraction> = client
                .get(&format!("{}/attractions", base_url))
                .send()
                .await?
                .json()
                .await?;
            for a in attractions {
                println!("ID: {}\n  Name: {}\n  Duration: {} min\n  Open: {} - {}\n",
                    a.id, a.name, a.suggested_duration_minutes, a.opening_time, a.closing_time);
            }
        }
        AttractionCommands::Create { name, description, duration, open, close } => {
            let open_time = NaiveTime::parse_from_str(&open, "%H:%M")?;
            let close_time = NaiveTime::parse_from_str(&close, "%H:%M")?;

            let new_attraction = NewAttraction {
                name,
                description,
                suggested_duration_minutes: duration,
                opening_time: open_time,
                closing_time: close_time,
            };

            let response = client
                .post(&format!("{}/attractions", base_url))
                .json(&new_attraction)
                .send()
                .await?;

            let status = response.status();
            if status.is_success() {
                let attraction: Attraction = response.json().await?;
                println!("Created attraction:");
                println!("  ID: {}\n  Name: {}\n  Duration: {} min\n  Open: {} - {}",
                    attraction.id, attraction.name, attraction.suggested_duration_minutes,
                    attraction.opening_time, attraction.closing_time);
            } else {
                let err: ErrorResponse = response.json().await?;
                eprintln!("Error: {}", err.error);
                for detail in err.details {
                    eprintln!("  - {}", detail);
                }
            }
        }
        AttractionCommands::Get { id } => {
            let response = client
                .get(&format!("{}/attractions/{}", base_url, id))
                .send()
                .await?;

            if response.status().is_success() {
                let attraction: Attraction = response.json().await?;
                println!("ID: {}\nName: {}\nDescription: {:?}\nDuration: {} min\nOpen: {} - {}",
                    attraction.id, attraction.name, attraction.description,
                    attraction.suggested_duration_minutes, attraction.opening_time, attraction.closing_time);
            } else {
                eprintln!("Attraction not found: {}", id);
            }
        }
    }
    Ok(())
}

async fn handle_transportation_commands(client: &Client, base_url: &str, action: TransportationCommands) -> Result<()> {
    match action {
        TransportationCommands::List => {
            let transportations: Vec<Transportation> = client
                .get(&format!("{}/transportations", base_url))
                .send()
                .await?
                .json()
                .await?;
            for t in transportations {
                println!("From: {} -> To: {}\n  Type: {:?}\n  Duration: {} min\n",
                    t.from_attraction_id, t.to_attraction_id, t.transport_type, t.duration_minutes);
            }
        }
        TransportationCommands::Create { from, to, type_, duration } => {
            let transport_type = match type_.to_lowercase().as_str() {
                "walk" => TransportationType::Walk,
                "bus" => TransportationType::Bus,
                "taxi" => TransportationType::Taxi,
                "subway" => TransportationType::Subway,
                "car" => TransportationType::Car,
                _ => {
                    eprintln!("Invalid transportation type: {}. Valid types: walk, bus, taxi, subway, car", type_);
                    return Ok(());
                }
            };

            let new_transport = NewTransportation {
                from_attraction_id: from,
                to_attraction_id: to,
                transport_type,
                duration_minutes: duration,
            };

            let response = client
                .post(&format!("{}/transportations", base_url))
                .json(&new_transport)
                .send()
                .await?;

            let status = response.status();
            if status.is_success() {
                let transport: Transportation = response.json().await?;
                println!("Created transportation:");
                println!("  From: {} -> To: {}\n  Type: {:?}\n  Duration: {} min",
                    transport.from_attraction_id, transport.to_attraction_id,
                    transport.transport_type, transport.duration_minutes);
            } else {
                let err: ErrorResponse = response.json().await?;
                eprintln!("Error: {}", err.error);
                for detail in err.details {
                    eprintln!("  - {}", detail);
                }
            }
        }
    }
    Ok(())
}

async fn handle_itinerary_commands(client: &Client, base_url: &str, action: ItineraryCommands) -> Result<()> {
    match action {
        ItineraryCommands::List => {
            let itineraries: Vec<Itinerary> = client
                .get(&format!("{}/itineraries", base_url))
                .send()
                .await?
                .json()
                .await?;
            for i in itineraries {
                println!("ID: {}\n  Name: {}\n  Nodes: {}\n",
                    i.id, i.name, i.nodes.len());
            }
        }
        ItineraryCommands::Create { name } => {
            let new_itinerary = NewItinerary { name };

            let response = client
                .post(&format!("{}/itineraries", base_url))
                .json(&new_itinerary)
                .send()
                .await?;

            let itinerary: Itinerary = response.json().await?;
            println!("Created itinerary:");
            println!("  ID: {}\n  Name: {}", itinerary.id, itinerary.name);
        }
        ItineraryCommands::Get { id } => {
            let response = client
                .get(&format!("{}/itineraries/{}", base_url, id))
                .send()
                .await?;

            if response.status().is_success() {
                let itinerary: Itinerary = response.json().await?;
                println!("ID: {}\nName: {}\nNodes:", itinerary.id, itinerary.name);
                for (idx, node) in itinerary.nodes.iter().enumerate() {
                    println!("  {}. Attraction: {}", idx + 1, node.attraction_id);
                    println!("     Arrival: {}", node.arrival_time);
                    println!("     Departure: {}", node.departure_time);
                }
            } else {
                eprintln!("Itinerary not found: {}", id);
            }
        }
        ItineraryCommands::AddNode { itinerary, attraction, arrival, departure } => {
            let arrival_time: DateTime<Utc> = DateTime::parse_from_rfc3339(&arrival)?.with_timezone(&Utc);
            let departure_time: DateTime<Utc> = DateTime::parse_from_rfc3339(&departure)?.with_timezone(&Utc);

            let request = AddItineraryNodesRequest {
                itinerary_id: itinerary,
                nodes: vec![NewItineraryNode {
                    attraction_id: attraction,
                    arrival_time,
                    departure_time,
                }],
            };

            let response = client
                .post(&format!("{}/itineraries/{}", base_url, itinerary))
                .json(&request)
                .send()
                .await?;

            let status = response.status();
            if status.is_success() {
                let itinerary: Itinerary = response.json().await?;
                println!("Updated itinerary:");
                println!("  ID: {}\n  Name: {}\n  Nodes: {}", itinerary.id, itinerary.name, itinerary.nodes.len());
            } else {
                let err: ErrorResponse = response.json().await?;
                eprintln!("Error: {}", err.error);
                for detail in err.details {
                    eprintln!("  - {}", detail);
                }
            }
        }
    }
    Ok(())
}
