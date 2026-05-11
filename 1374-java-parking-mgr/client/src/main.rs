use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::json;

#[derive(Parser, Debug)]
#[command(name = "parking-cli", about = "Parking Management System CLI Client")]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(about = "List all parking spaces")]
    Spaces,

    #[command(about = "Get details of a specific space")]
    Space {
        #[arg(value_name = "SPACE_ID")]
        id: String,
    },

    #[command(about = "Add a new parking space")]
    AddSpace {
        #[arg(value_name = "SPACE_ID")]
        id: String,
        #[arg(value_name = "TYPE", help = "normal|charging|accessible")]
        space_type: String,
    },

    #[command(about = "List all users")]
    Users,

    #[command(about = "Get details of a specific user")]
    User {
        #[arg(value_name = "USER_ID")]
        id: String,
    },

    #[command(about = "Add a new user")]
    AddUser {
        #[arg(value_name = "USER_ID")]
        id: String,
        #[arg(value_name = "NAME")]
        name: String,
        #[arg(value_name = "PHONE")]
        phone: String,
    },

    #[command(about = "List all monthly cards")]
    MonthlyCards,

    #[command(about = "List monthly cards needing expiry reminders")]
    Reminders,

    #[command(about = "Add a new monthly card")]
    AddMonthlyCard {
        #[arg(value_name = "USER_ID")]
        user_id: String,
        #[arg(value_name = "PLATE")]
        plate: String,
        #[arg(value_name = "START", help = "RFC3339 format")]
        start: String,
        #[arg(value_name = "END", help = "RFC3339 format")]
        end: String,
        #[arg(short, long)]
        reserved_space: Option<String>,
    },

    #[command(about = "Enter parking")]
    Enter {
        #[arg(value_name = "PLATE")]
        plate: String,
        #[arg(value_name = "TYPE", help = "regular|electric|disabled")]
        vehicle_type: String,
        #[arg(short, long)]
        user: Option<String>,
    },

    #[command(about = "Exit parking")]
    Exit {
        #[arg(value_name = "PLATE")]
        plate: String,
    },

    #[command(about = "Get active parking record for a plate")]
    Active {
        #[arg(value_name = "PLATE")]
        plate: String,
    },

    #[command(about = "List all parking records")]
    Records,

    #[command(about = "Get details of a specific record")]
    Record {
        #[arg(value_name = "RECORD_ID")]
        id: String,
    },
}

#[derive(Serialize, Deserialize, Debug)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    let client = Client::new();

    match &cli.command {
        Commands::Spaces => list_spaces(&client, &cli.server).await?,
        Commands::Space { id } => get_space(&client, &cli.server, id).await?,
        Commands::AddSpace { id, space_type } => add_space(&client, &cli.server, id, space_type).await?,

        Commands::Users => list_users(&client, &cli.server).await?,
        Commands::User { id } => get_user(&client, &cli.server, id).await?,
        Commands::AddUser { id, name, phone } => add_user(&client, &cli.server, id, name, phone).await?,

        Commands::MonthlyCards => list_monthly_cards(&client, &cli.server).await?,
        Commands::Reminders => get_reminders(&client, &cli.server).await?,
        Commands::AddMonthlyCard { user_id, plate, start, end, reserved_space } => {
            add_monthly_card(&client, &cli.server, user_id, plate, start, end, reserved_space.as_deref()).await?
        }

        Commands::Enter { plate, vehicle_type, user } => enter_parking(&client, &cli.server, plate, vehicle_type, user.as_deref()).await?,
        Commands::Exit { plate } => exit_parking(&client, &cli.server, plate).await?,
        Commands::Active { plate } => get_active(&client, &cli.server, plate).await?,

        Commands::Records => list_records(&client, &cli.server).await?,
        Commands::Record { id } => get_record(&client, &cli.server, id).await?,
    }

    Ok(())
}

fn print_json<T: Serialize>(value: &T) {
    println!("{}", serde_json::to_string_pretty(value).unwrap());
}

fn handle_response<T: Serialize>(response: ApiResponse<T>) {
    if response.success {
        if let Some(data) = response.data {
            print_json(&data);
        } else {
            println!("Success");
        }
    } else {
        if let Some(err) = response.error {
            eprintln!("Error: {}", err);
        } else {
            eprintln!("Unknown error");
        }
        std::process::exit(1);
    }
}

async fn list_spaces(client: &Client, server: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<Vec<serde_json::Value>> = client
        .get(&format!("{}/spaces", server))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn get_space(client: &Client, server: &str, id: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .get(&format!("{}/spaces/{}", server, id))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn add_space(client: &Client, server: &str, id: &str, space_type: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .post(&format!("{}/spaces", server))
        .json(&json!({
            "id": id,
            "space_type": space_type
        }))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn list_users(client: &Client, server: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<Vec<serde_json::Value>> = client
        .get(&format!("{}/users", server))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn get_user(client: &Client, server: &str, id: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .get(&format!("{}/users/{}", server, id))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn add_user(client: &Client, server: &str, id: &str, name: &str, phone: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .post(&format!("{}/users", server))
        .json(&json!({
            "id": id,
            "name": name,
            "phone": phone
        }))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn list_monthly_cards(client: &Client, server: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<Vec<serde_json::Value>> = client
        .get(&format!("{}/monthly-cards", server))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn get_reminders(client: &Client, server: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<Vec<serde_json::Value>> = client
        .get(&format!("{}/monthly-cards/reminders", server))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn add_monthly_card(
    client: &Client,
    server: &str,
    user_id: &str,
    plate: &str,
    start: &str,
    end: &str,
    reserved_space: Option<&str>,
) -> anyhow::Result<()> {
    let mut body = json!({
        "user_id": user_id,
        "vehicle_plate": plate,
        "start_time": start,
        "end_time": end
    });

    if let Some(space) = reserved_space {
        body["reserved_space_id"] = json!(space);
    }

    let resp: ApiResponse<serde_json::Value> = client
        .post(&format!("{}/monthly-cards", server))
        .json(&body)
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn enter_parking(
    client: &Client,
    server: &str,
    plate: &str,
    vehicle_type: &str,
    user: Option<&str>,
) -> anyhow::Result<()> {
    let mut body = json!({
        "plate_number": plate,
        "vehicle_type": vehicle_type
    });

    if let Some(uid) = user {
        body["user_id"] = json!(uid);
    }

    let resp: ApiResponse<serde_json::Value> = client
        .post(&format!("{}/parking/enter", server))
        .json(&body)
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn exit_parking(client: &Client, server: &str, plate: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .post(&format!("{}/parking/exit", server))
        .json(&json!({
            "plate_number": plate
        }))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn get_active(client: &Client, server: &str, plate: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .get(&format!("{}/parking/active/{}", server, plate))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn list_records(client: &Client, server: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<Vec<serde_json::Value>> = client
        .get(&format!("{}/records", server))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}

async fn get_record(client: &Client, server: &str, id: &str) -> anyhow::Result<()> {
    let resp: ApiResponse<serde_json::Value> = client
        .get(&format!("{}/records/{}", server, id))
        .send()
        .await?
        .json()
        .await?;
    handle_response(resp);
    Ok(())
}
