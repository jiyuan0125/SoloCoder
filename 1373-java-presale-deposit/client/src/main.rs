mod api;

use clap::{Parser, Subcommand};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "PRESALE_SERVER", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    CreateProduct {
        #[arg(long)]
        name: String,
        #[arg(long)]
        price: u64,
    },
    ListProducts,
    GetProduct {
        #[arg(long)]
        id: Uuid,
    },
    CreateActivity {
        #[arg(long)]
        product_id: Uuid,
        #[arg(long)]
        deposit: u64,
        #[arg(long)]
        inflation: u32,
        #[arg(long)]
        max_participants: u32,
        #[arg(long)]
        start: DateTime<Utc>,
        #[arg(long)]
        end: DateTime<Utc>,
        #[arg(long)]
        deadline_hours: Option<i64>,
    },
    ListActivities,
    GetActivity {
        #[arg(long)]
        id: Uuid,
    },
    StartActivity {
        #[arg(long)]
        id: Uuid,
    },
    EndActivity {
        #[arg(long)]
        id: Uuid,
    },
    CancelActivity {
        #[arg(long)]
        id: Uuid,
    },
    PayDeposit {
        #[arg(long)]
        user_id: String,
        #[arg(long)]
        activity_id: Uuid,
    },
    PayFinal {
        #[arg(long)]
        order_id: Uuid,
        #[arg(long)]
        amount: u64,
    },
    GetOrder {
        #[arg(long)]
        id: Uuid,
    },
    ListUserOrders {
        #[arg(long)]
        user_id: String,
    },
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    let client = api::PresaleApiClient::new(&cli.server);

    match cli.command {
        Commands::Health => {
            let result = client.health().await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::CreateProduct { name, price } => {
            let result = client.create_product(&name, price).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::ListProducts => {
            let result = client.list_products().await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::GetProduct { id } => {
            let result = client.get_product(id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::CreateActivity {
            product_id,
            deposit,
            inflation,
            max_participants,
            start,
            end,
            deadline_hours,
        } => {
            let result = client
                .create_activity(
                    product_id,
                    deposit,
                    inflation,
                    max_participants,
                    start,
                    end,
                    deadline_hours,
                )
                .await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::ListActivities => {
            let result = client.list_activities().await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::GetActivity { id } => {
            let result = client.get_activity(id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::StartActivity { id } => {
            let result = client.start_activity(id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::EndActivity { id } => {
            let result = client.end_activity(id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::CancelActivity { id } => {
            let result = client.cancel_activity(id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::PayDeposit {
            user_id,
            activity_id,
        } => {
            let result = client.pay_deposit(&user_id, activity_id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::PayFinal { order_id, amount } => {
            let result = client.pay_final(order_id, amount).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::GetOrder { id } => {
            let result = client.get_order(id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
        Commands::ListUserOrders { user_id } => {
            let result = client.list_user_orders(&user_id).await?;
            println!("{}", serde_json::to_string_pretty(&result)?);
        }
    }

    Ok(())
}
