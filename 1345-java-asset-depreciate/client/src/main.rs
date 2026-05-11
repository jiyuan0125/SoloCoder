use clap::{Parser, Subcommand};
use reqwest::blocking::Client;
use std::collections::HashMap;

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
    Employee {
        #[command(subcommand)]
        action: EmployeeCommands,
    },
    Category {
        #[command(subcommand)]
        action: CategoryCommands,
    },
    Asset {
        #[command(subcommand)]
        action: AssetCommands,
    },
    Application {
        #[command(subcommand)]
        action: ApplicationCommands,
    },
    Return {
        #[command(subcommand)]
        action: ReturnCommands,
    },
    Transfer {
        #[command(subcommand)]
        action: TransferCommands,
    },
    History {
        asset_id: String,
    },
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    List,
    Create { name: String, role: String },
    Get { id: String },
}

#[derive(Subcommand, Debug)]
enum CategoryCommands {
    List,
    Create {
        name: String,
        #[arg(short = 'l', long = "life-months")]
        useful_life_months: u32,
        #[arg(short = 'm', long = "max")]
        max_per_employee: u32,
    },
    Get { id: String },
}

#[derive(Subcommand, Debug)]
enum AssetCommands {
    List,
    Details,
    Create {
        name: String,
        category_id: String,
        #[arg(short = 'p', long = "price")]
        purchase_price: f64,
        #[arg(short = 'd', long = "date")]
        purchase_date: String,
    },
    Get { id: String },
    Detail { id: String },
}

#[derive(Subcommand, Debug)]
enum ApplicationCommands {
    List,
    Apply {
        asset_id: String,
        applicant_id: String,
        reason: String,
    },
    Approve { application_id: String, approver_id: String },
    Reject { application_id: String, approver_id: String },
    Get { id: String },
}

#[derive(Subcommand, Debug)]
enum ReturnCommands {
    Do {
        asset_id: String,
        returned_by: String,
        verified_by: String,
        #[arg(short, long, default_value = "good")]
        condition: String,
        #[arg(short, long)]
        damage_note: Option<String>,
    },
    List { asset_id: String },
}

#[derive(Subcommand, Debug)]
enum TransferCommands {
    List,
    Initiate {
        asset_id: String,
        from_employee_id: String,
        to_employee_id: String,
    },
    ConfirmFrom { transfer_id: String, employee_id: String },
    ConfirmTo { transfer_id: String, employee_id: String },
    Get { id: String },
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/');

    match cli.command {
        Commands::Employee { action } => match action {
            EmployeeCommands::List => {
                let resp = client.get(format!("{}/employees", base_url)).send()?;
                print_json(resp.text()?);
            }
            EmployeeCommands::Create { name, role } => {
                let body = HashMap::from([("name", name), ("role", role)]);
                let resp = client.post(format!("{}/employees", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            EmployeeCommands::Get { id } => {
                let resp = client.get(format!("{}/employees/{}", base_url, id)).send()?;
                print_json(resp.text()?);
            }
        },
        Commands::Category { action } => match action {
            CategoryCommands::List => {
                let resp = client.get(format!("{}/categories", base_url)).send()?;
                print_json(resp.text()?);
            }
            CategoryCommands::Create {
                name,
                useful_life_months,
                max_per_employee,
            } => {
                let body = HashMap::from([
                    ("name", name),
                    ("useful_life_months", useful_life_months.to_string()),
                    ("max_per_employee", max_per_employee.to_string()),
                ]);
                let resp = client.post(format!("{}/categories", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            CategoryCommands::Get { id } => {
                let resp = client.get(format!("{}/categories/{}", base_url, id)).send()?;
                print_json(resp.text()?);
            }
        },
        Commands::Asset { action } => match action {
            AssetCommands::List => {
                let resp = client.get(format!("{}/assets", base_url)).send()?;
                print_json(resp.text()?);
            }
            AssetCommands::Details => {
                let resp = client.get(format!("{}/assets/details", base_url)).send()?;
                print_json(resp.text()?);
            }
            AssetCommands::Create {
                name,
                category_id,
                purchase_price,
                purchase_date,
            } => {
                let body = HashMap::from([
                    ("name", name),
                    ("category_id", category_id),
                    ("purchase_price", purchase_price.to_string()),
                    ("purchase_date", purchase_date),
                ]);
                let resp = client.post(format!("{}/assets", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            AssetCommands::Get { id } => {
                let resp = client.get(format!("{}/assets/{}", base_url, id)).send()?;
                print_json(resp.text()?);
            }
            AssetCommands::Detail { id } => {
                let resp = client.get(format!("{}/assets/{}/detail", base_url, id)).send()?;
                print_json(resp.text()?);
            }
        },
        Commands::Application { action } => match action {
            ApplicationCommands::List => {
                let resp = client.get(format!("{}/applications", base_url)).send()?;
                print_json(resp.text()?);
            }
            ApplicationCommands::Apply {
                asset_id,
                applicant_id,
                reason,
            } => {
                let body = HashMap::from([
                    ("asset_id", asset_id),
                    ("applicant_id", applicant_id),
                    ("reason", reason),
                ]);
                let resp = client.post(format!("{}/applications", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            ApplicationCommands::Approve {
                application_id,
                approver_id,
            } => {
                let body = HashMap::from([
                    ("application_id", application_id),
                    ("approver_id", approver_id),
                ]);
                let resp = client.post(format!("{}/applications/approve", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            ApplicationCommands::Reject {
                application_id,
                approver_id,
            } => {
                let body = HashMap::from([
                    ("application_id", application_id),
                    ("approver_id", approver_id),
                ]);
                let resp = client.post(format!("{}/applications/reject", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            ApplicationCommands::Get { id } => {
                let resp = client.get(format!("{}/applications/{}", base_url, id)).send()?;
                print_json(resp.text()?);
            }
        },
        Commands::Return { action } => match action {
            ReturnCommands::Do {
                asset_id,
                returned_by,
                verified_by,
                condition,
                damage_note,
            } => {
                let mut body: HashMap<String, String> = HashMap::new();
                body.insert("asset_id".to_string(), asset_id);
                body.insert("returned_by".to_string(), returned_by);
                body.insert("verified_by".to_string(), verified_by);
                body.insert("condition".to_string(), condition);
                if let Some(note) = damage_note {
                    body.insert("damage_note".to_string(), note);
                }
                let resp = client.post(format!("{}/returns", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            ReturnCommands::List { asset_id } => {
                let resp = client.get(format!("{}/returns/{}", base_url, asset_id)).send()?;
                print_json(resp.text()?);
            }
        },
        Commands::Transfer { action } => match action {
            TransferCommands::List => {
                let resp = client.get(format!("{}/transfers", base_url)).send()?;
                print_json(resp.text()?);
            }
            TransferCommands::Initiate {
                asset_id,
                from_employee_id,
                to_employee_id,
            } => {
                let body = HashMap::from([
                    ("asset_id", asset_id),
                    ("from_employee_id", from_employee_id),
                    ("to_employee_id", to_employee_id),
                ]);
                let resp = client.post(format!("{}/transfers", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            TransferCommands::ConfirmFrom {
                transfer_id,
                employee_id,
            } => {
                let body = HashMap::from([
                    ("transfer_id", transfer_id),
                    ("employee_id", employee_id),
                ]);
                let resp = client.post(format!("{}/transfers/confirm-from", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            TransferCommands::ConfirmTo {
                transfer_id,
                employee_id,
            } => {
                let body = HashMap::from([
                    ("transfer_id", transfer_id),
                    ("employee_id", employee_id),
                ]);
                let resp = client.post(format!("{}/transfers/confirm-to", base_url)).json(&body).send()?;
                print_json(resp.text()?);
            }
            TransferCommands::Get { id } => {
                let resp = client.get(format!("{}/transfers/{}", base_url, id)).send()?;
                print_json(resp.text()?);
            }
        },
        Commands::History { asset_id } => {
            let resp = client.get(format!("{}/history/{}", base_url, asset_id)).send()?;
            print_json(resp.text()?);
        }
    }

    Ok(())
}

fn print_json(text: String) {
    match serde_json::from_str::<serde_json::Value>(&text) {
        Ok(v) => println!("{}", serde_json::to_string_pretty(&v).unwrap_or(text)),
        Err(_) => println!("{}", text),
    }
}
