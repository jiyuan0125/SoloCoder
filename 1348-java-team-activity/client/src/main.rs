use chrono::{DateTime, Utc};
use clap::{Parser, Subcommand};
use activity_core::{ActivityStatus, CreateActivityRequest, RegistrationStatus};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(name = "team-activity-cli")]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Department {
        #[command(subcommand)]
        dept_cmd: DepartmentCommands,
    },
    Activity {
        #[command(subcommand)]
        activity_cmd: ActivityCommands,
    },
}

#[derive(Subcommand, Debug)]
enum DepartmentCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        budget: f64,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ActivityCommands {
    Create {
        #[arg(short, long)]
        department_id: Uuid,
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        description: Option<String>,
        #[arg(short, long)]
        cost_per_person: f64,
        #[arg(short, long)]
        max_participants: usize,
        #[arg(short = 's', long)]
        start_time: DateTime<Utc>,
        #[arg(short = 'e', long)]
        end_time: DateTime<Utc>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Publish {
        #[arg(short, long)]
        id: Uuid,
    },
    Cancel {
        #[arg(short, long)]
        id: Uuid,
    },
    UpdateLimit {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        new_limit: usize,
    },
    Register {
        #[arg(short, long)]
        activity_id: Uuid,
        #[arg(short, long)]
        user_id: String,
    },
    CancelRegistration {
        #[arg(short, long)]
        activity_id: Uuid,
        #[arg(short, long)]
        user_id: String,
    },
    CloseRegistration {
        #[arg(short, long)]
        id: Uuid,
    },
    Settle {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, num_args = 1.., value_delimiter = ',')]
        participants: Vec<String>,
    },
}

#[derive(Debug, Serialize)]
struct CreateDepartmentRequest {
    name: String,
    annual_budget: f64,
}

#[derive(Debug, Serialize)]
struct RegisterRequest {
    user_id: String,
}

#[derive(Debug, Serialize)]
struct UpdateLimitRequest {
    new_limit: usize,
}

#[derive(Debug, Serialize)]
struct SettleRequest {
    actual_participants: Vec<String>,
}

#[derive(Debug, Deserialize)]
struct DepartmentResponse {
    id: Uuid,
    name: String,
    annual_budget: f64,
    remaining_budget: f64,
}

#[derive(Debug, Deserialize)]
struct ActivityResponse {
    id: Uuid,
    department_id: Uuid,
    name: String,
    description: String,
    cost_per_person: f64,
    max_participants: usize,
    start_time: DateTime<Utc>,
    end_time: DateTime<Utc>,
    status: ActivityStatus,
}

type ApiResult<T> = Result<T, Box<dyn std::error::Error>>;

#[tokio::main]
async fn main() -> ApiResult<()> {
    let args = Args::parse();
    let client = Client::new();

    match args.command {
        Commands::Department { dept_cmd } => match dept_cmd {
            DepartmentCommands::Create { name, budget } => {
                create_department(&client, &args.server, name, budget).await?
            }
            DepartmentCommands::List => list_departments(&client, &args.server).await?,
            DepartmentCommands::Get { id } => get_department(&client, &args.server, id).await?,
        },
        Commands::Activity { activity_cmd } => match activity_cmd {
            ActivityCommands::Create {
                department_id,
                name,
                description,
                cost_per_person,
                max_participants,
                start_time,
                end_time,
            } => {
                create_activity(
                    &client,
                    &args.server,
                    department_id,
                    name,
                    description,
                    cost_per_person,
                    max_participants,
                    start_time,
                    end_time,
                )
                .await?
            }
            ActivityCommands::List => list_activities(&client, &args.server).await?,
            ActivityCommands::Get { id } => get_activity(&client, &args.server, id).await?,
            ActivityCommands::Publish { id } => publish_activity(&client, &args.server, id).await?,
            ActivityCommands::Cancel { id } => cancel_activity(&client, &args.server, id).await?,
            ActivityCommands::UpdateLimit { id, new_limit } => {
                update_limit(&client, &args.server, id, new_limit).await?
            }
            ActivityCommands::Register {
                activity_id,
                user_id,
            } => register(&client, &args.server, activity_id, user_id).await?,
            ActivityCommands::CancelRegistration {
                activity_id,
                user_id,
            } => cancel_registration(&client, &args.server, activity_id, user_id).await?,
            ActivityCommands::CloseRegistration { id } => {
                close_registration(&client, &args.server, id).await?
            }
            ActivityCommands::Settle { id, participants } => {
                settle(&client, &args.server, id, participants).await?
            }
        },
    }

    Ok(())
}

async fn create_department(
    client: &Client,
    server: &str,
    name: String,
    budget: f64,
) -> ApiResult<()> {
    let req = CreateDepartmentRequest {
        name,
        annual_budget: budget,
    };

    let resp = client
        .post(format!("{}/api/departments", server))
        .json(&req)
        .send()
        .await?;

    if resp.status().is_success() {
        let dept: DepartmentResponse = resp.json().await?;
        println!(
            "Created department: {} ({}) - Budget: {:.2}, Remaining: {:.2}",
            dept.name, dept.id, dept.annual_budget, dept.remaining_budget
        );
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn list_departments(client: &Client, server: &str) -> ApiResult<()> {
    let resp = client
        .get(format!("{}/api/departments", server))
        .send()
        .await?;

    if resp.status().is_success() {
        let depts: Vec<DepartmentResponse> = resp.json().await?;
        println!("Departments:");
        for dept in depts {
            println!(
                "  {} ({}) - Budget: {:.2}, Remaining: {:.2}",
                dept.name, dept.id, dept.annual_budget, dept.remaining_budget
            );
        }
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn get_department(client: &Client, server: &str, id: Uuid) -> ApiResult<()> {
    let resp = client
        .get(format!("{}/api/departments/{}", server, id))
        .send()
        .await?;

    if resp.status().is_success() {
        let dept: DepartmentResponse = resp.json().await?;
        println!(
            "Department: {} ({})",
            dept.name, dept.id
        );
        println!("  Annual Budget: {:.2}", dept.annual_budget);
        println!("  Remaining Budget: {:.2}", dept.remaining_budget);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn create_activity(
    client: &Client,
    server: &str,
    department_id: Uuid,
    name: String,
    description: Option<String>,
    cost_per_person: f64,
    max_participants: usize,
    start_time: DateTime<Utc>,
    end_time: DateTime<Utc>,
) -> ApiResult<()> {
    let req = CreateActivityRequest {
        department_id,
        name,
        description: description.unwrap_or_default(),
        cost_per_person,
        max_participants,
        start_time,
        end_time,
    };

    let resp = client
        .post(format!("{}/api/activities", server))
        .json(&req)
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        print_activity(&activity);
        println!("\nRun 'activity publish --id {}' to publish it", activity.id);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn list_activities(client: &Client, server: &str) -> ApiResult<()> {
    let resp = client
        .get(format!("{}/api/activities", server))
        .send()
        .await?;

    if resp.status().is_success() {
        let activities: Vec<ActivityResponse> = resp.json().await?;
        println!("Activities:");
        for activity in activities {
            print_activity(&activity);
            println!("---");
        }
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn get_activity(client: &Client, server: &str, id: Uuid) -> ApiResult<()> {
    let resp = client
        .get(format!("{}/api/activities/{}", server, id))
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        print_activity(&activity);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn publish_activity(client: &Client, server: &str, id: Uuid) -> ApiResult<()> {
    let resp = client
        .post(format!("{}/api/activities/{}/publish", server, id))
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        println!("Activity published!");
        print_activity(&activity);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn cancel_activity(client: &Client, server: &str, id: Uuid) -> ApiResult<()> {
    let resp = client
        .post(format!("{}/api/activities/{}/cancel", server, id))
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        println!("Activity cancelled!");
        print_activity(&activity);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn update_limit(
    client: &Client,
    server: &str,
    id: Uuid,
    new_limit: usize,
) -> ApiResult<()> {
    let req = UpdateLimitRequest { new_limit };

    let resp = client
        .put(format!("{}/api/activities/{}/limit", server, id))
        .json(&req)
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        println!("Registration limit updated!");
        print_activity(&activity);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn register(
    client: &Client,
    server: &str,
    activity_id: Uuid,
    user_id: String,
) -> ApiResult<()> {
    let req = RegisterRequest { user_id };

    let resp = client
        .post(format!("{}/api/activities/{}/register", server, activity_id))
        .json(&req)
        .send()
        .await?;

    if resp.status().is_success() {
        let status: RegistrationStatus = resp.json().await?;
        match status {
            RegistrationStatus::Registered => println!("Successfully registered!"),
            RegistrationStatus::Waitlisted => println!("Activity is full, added to waitlist"),
            _ => println!("Registration status: {:?}", status),
        }
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn cancel_registration(
    client: &Client,
    server: &str,
    activity_id: Uuid,
    user_id: String,
) -> ApiResult<()> {
    let resp = client
        .delete(format!(
            "{}/api/activities/{}/register/{}",
            server, activity_id, user_id
        ))
        .send()
        .await?;

    if resp.status().is_success() {
        println!("Registration cancelled successfully");
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn close_registration(client: &Client, server: &str, id: Uuid) -> ApiResult<()> {
    let resp = client
        .post(format!("{}/api/activities/{}/close-registration", server, id))
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        println!("Registration closed!");
        print_activity(&activity);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

async fn settle(
    client: &Client,
    server: &str,
    id: Uuid,
    participants: Vec<String>,
) -> ApiResult<()> {
    let req = SettleRequest {
        actual_participants: participants,
    };

    let resp = client
        .post(format!("{}/api/activities/{}/settle", server, id))
        .json(&req)
        .send()
        .await?;

    if resp.status().is_success() {
        let activity: ActivityResponse = resp.json().await?;
        println!("Activity settled!");
        print_activity(&activity);
    } else {
        println!("Error: {}", resp.text().await?);
    }

    Ok(())
}

fn print_activity(activity: &ActivityResponse) {
    println!("Activity: {} ({})", activity.name, activity.id);
    println!("  Department: {}", activity.department_id);
    println!("  Description: {}", activity.description);
    println!("  Cost per person: {:.2}", activity.cost_per_person);
    println!("  Max participants: {}", activity.max_participants);
    println!(
        "  Estimated total: {:.2}",
        activity.cost_per_person * activity.max_participants as f64
    );
    println!("  Status: {:?}", activity.status);
    println!("  Start: {}", activity.start_time);
    println!("  End: {}", activity.end_time);
}
