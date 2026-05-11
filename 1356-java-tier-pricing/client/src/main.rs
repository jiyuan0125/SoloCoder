use clap::{Parser, Subcommand};
use reqwest::Client;
use serde_json::{json, Value};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Product {
        #[command(subcommand)]
        action: ProductCommands,
    },
    Member {
        #[command(subcommand)]
        action: MemberCommands,
    },
    Price {
        #[command(subcommand)]
        action: PriceCommands,
    },
    Order {
        #[command(subcommand)]
        action: OrderCommands,
    },
    Approval {
        #[command(subcommand)]
        action: ApprovalCommands,
    },
}

#[derive(Subcommand, Debug)]
enum ProductCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        price: f64,
        #[arg(long, default_value_t = 100)]
        stock: i64,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    UpdatePrice {
        #[arg(long)]
        id: String,
        #[arg(long)]
        new_price: f64,
    },
}

#[derive(Subcommand, Debug)]
enum MemberCommands {
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        type_: String,
    },
    List,
}

#[derive(Subcommand, Debug)]
enum PriceCommands {
    CreateTier {
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        min_qty: i64,
        #[arg(long)]
        max_qty: Option<i64>,
        #[arg(long)]
        discount: f64,
        #[arg(long, default_value_t = 30)]
        days: i64,
    },
    CreateMemberDiscount {
        #[arg(long)]
        type_: String,
        #[arg(long)]
        product_id: Option<String>,
        #[arg(long)]
        discount: f64,
        #[arg(long, default_value_t = 30)]
        days: i64,
    },
    CreateSpecial {
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        price: f64,
        #[arg(long, default_value_t = 7)]
        days: i64,
    },
    CreateFullDiscount {
        #[arg(long)]
        threshold: f64,
        #[arg(long)]
        discount: f64,
        #[arg(long, default_value_t = 7)]
        days: i64,
    },
    Query {
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        quantity: i64,
        #[arg(long)]
        member_type: Option<String>,
    },
}

#[derive(Subcommand, Debug)]
enum OrderCommands {
    Create {
        #[arg(long)]
        member_id: Option<String>,
        #[arg(long, value_delimiter = ',')]
        items: Vec<String>,
    },
    Get {
        #[arg(long)]
        id: String,
    },
}

#[derive(Subcommand, Debug)]
enum ApprovalCommands {
    List,
    Approve {
        #[arg(long)]
        id: String,
    },
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url;

    match cli.command {
        Commands::Product { action } => match action {
            ProductCommands::Create { name, price, stock } => {
                let resp = client
                    .post(format!("{}/products", base_url))
                    .json(&json!({
                        "name": name,
                        "base_price": price,
                        "stock": stock,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            ProductCommands::List => {
                let resp = client.get(format!("{}/products", base_url)).send().await?;
                print_response(resp).await;
            }
            ProductCommands::Get { id } => {
                let resp = client.get(format!("{}/products/{}", base_url, id)).send().await?;
                print_response(resp).await;
            }
            ProductCommands::UpdatePrice { id, new_price } => {
                let resp = client
                    .put(format!("{}/products/{}/price", base_url, id))
                    .json(&json!({ "new_price": new_price }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
        },
        Commands::Member { action } => match action {
            MemberCommands::Create { name, type_ } => {
                let member_type = parse_member_type(&type_)?;
                let resp = client
                    .post(format!("{}/members", base_url))
                    .json(&json!({
                        "name": name,
                        "member_type": member_type,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            MemberCommands::List => {
                let resp = client.get(format!("{}/members", base_url)).send().await?;
                print_response(resp).await;
            }
        },
        Commands::Price { action } => match action {
            PriceCommands::CreateTier {
                product_id,
                min_qty,
                max_qty,
                discount,
                days,
            } => {
                let resp = client
                    .post(format!("{}/tier-prices", base_url))
                    .json(&json!({
                        "product_id": product_id,
                        "min_quantity": min_qty,
                        "max_quantity": max_qty,
                        "discount_percent": discount,
                        "days": days,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            PriceCommands::CreateMemberDiscount {
                type_,
                product_id,
                discount,
                days,
            } => {
                let member_type = parse_member_type(&type_)?;
                let resp = client
                    .post(format!("{}/member-discounts", base_url))
                    .json(&json!({
                        "member_type": member_type,
                        "product_id": product_id,
                        "discount_percent": discount,
                        "days": days,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            PriceCommands::CreateSpecial {
                product_id,
                price,
                days,
            } => {
                let resp = client
                    .post(format!("{}/special-prices", base_url))
                    .json(&json!({
                        "product_id": product_id,
                        "special_price": price,
                        "days": days,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            PriceCommands::CreateFullDiscount {
                threshold,
                discount,
                days,
            } => {
                let resp = client
                    .post(format!("{}/full-discounts", base_url))
                    .json(&json!({
                        "threshold": threshold,
                        "discount": discount,
                        "days": days,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            PriceCommands::Query {
                product_id,
                quantity,
                member_type,
            } => {
                let mt = match member_type {
                    Some(t) => Some(parse_member_type(&t)?),
                    None => None,
                };
                let resp = client
                    .post(format!("{}/price-query", base_url))
                    .json(&json!({
                        "product_id": product_id,
                        "quantity": quantity,
                        "member_type": mt,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
        },
        Commands::Order { action } => match action {
            OrderCommands::Create { member_id, items } => {
                let order_items = parse_order_items(&items)?;
                let resp = client
                    .post(format!("{}/orders", base_url))
                    .json(&json!({
                        "member_id": member_id,
                        "items": order_items,
                    }))
                    .send()
                    .await?;
                print_response(resp).await;
            }
            OrderCommands::Get { id } => {
                let resp = client.get(format!("{}/orders/{}", base_url, id)).send().await?;
                print_response(resp).await;
            }
        },
        Commands::Approval { action } => match action {
            ApprovalCommands::List => {
                let resp = client.get(format!("{}/approvals", base_url)).send().await?;
                print_response(resp).await;
            }
            ApprovalCommands::Approve { id } => {
                let resp = client
                    .post(format!("{}/approvals/{}/approve", base_url, id))
                    .send()
                    .await?;
                print_response(resp).await;
            }
        },
    }

    Ok(())
}

fn parse_member_type(s: &str) -> Result<&'static str, String> {
    match s.to_lowercase().as_str() {
        "normal" => Ok("Normal"),
        "gold" => Ok("Gold"),
        "platinum" => Ok("Platinum"),
        _ => Err(format!("无效的会员类型: {}", s)),
    }
}

fn parse_order_items(items: &[String]) -> Result<Vec<Value>, String> {
    let mut result = Vec::new();
    for item in items {
        let parts: Vec<&str> = item.split(':').collect();
        if parts.len() != 2 {
            return Err(format!("无效的订单项格式: {}, 应为 product_id:quantity", item));
        }
        let product_id = parts[0].to_string();
        let quantity = parts[1]
            .parse::<i64>()
            .map_err(|e| format!("无效的数量: {}", e))?;
        result.push(json!({
            "product_id": product_id,
            "quantity": quantity,
        }));
    }
    Ok(result)
}

async fn print_response(resp: reqwest::Response) {
    let status = resp.status();
    let text = resp.text().await.unwrap_or_default();
    
    println!("Status: {}", status);
    if !text.is_empty() {
        if let Ok(json) = serde_json::from_str::<Value>(&text) {
            println!("{}", serde_json::to_string_pretty(&json).unwrap());
        } else {
            println!("{}", text);
        }
    }
}
