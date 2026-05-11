use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use chrono::{Date, Utc};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "CHECKUP_SERVER", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(about = "项目管理")]
    Item {
        #[command(subcommand)]
        action: ItemCommands,
    },
    #[command(about = "套餐管理")]
    Package {
        #[command(subcommand)]
        action: PackageCommands,
    },
    #[command(about = "客户管理")]
    Customer {
        #[command(subcommand)]
        action: CustomerCommands,
    },
    #[command(about = "互斥规则管理")]
    MutexRule {
        #[command(subcommand)]
        action: MutexRuleCommands,
    },
    #[command(about = "预约管理")]
    Reservation {
        #[command(subcommand)]
        action: ReservationCommands,
    },
}

#[derive(Subcommand, Debug)]
enum ItemCommands {
    #[command(about = "创建项目")]
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        price: f64,
        #[arg(short = 'g', long, default_value = "none")]
        gender: String,
        #[arg(short, long, default_value_t = false)]
        contrast: bool,
        #[arg(long, default_value_t = false)]
        gastroscopy: bool,
        #[arg(long, default_value_t = false)]
        colonoscopy: bool,
        #[arg(long, default_value_t = false)]
        xray: bool,
        #[arg(long, default_value_t = false)]
        pregnant_forbidden: bool,
    },
    #[command(about = "列出所有项目")]
    List,
    #[command(about = "获取项目详情")]
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum PackageCommands {
    #[command(about = "创建套餐")]
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long, value_delimiter = ',')]
        items: Vec<Uuid>,
        #[arg(short = 'p', long)]
        price: f64,
    },
    #[command(about = "列出所有套餐")]
    List,
    #[command(about = "获取套餐详情")]
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum CustomerCommands {
    #[command(about = "创建客户")]
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        gender: String,
        #[arg(short = 'b', long)]
        birth_date: String,
        #[arg(long, default_value_t = false)]
        pregnant: bool,
    },
    #[command(about = "列出所有客户")]
    List,
    #[command(about = "获取客户详情")]
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum MutexRuleCommands {
    #[command(about = "创建互斥规则")]
    Create {
        #[arg(short, long)]
        rule_type: String,
        #[arg(short = '1', long)]
        item1: Uuid,
        #[arg(short = '2', long)]
        item2: Uuid,
        #[arg(short, long)]
        description: String,
    },
    #[command(about = "列出所有互斥规则")]
    List,
    #[command(about = "删除互斥规则")]
    Delete {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ReservationCommands {
    #[command(about = "创建预约")]
    Create {
        #[arg(short, long)]
        customer_id: Uuid,
        #[arg(short = 'a', long)]
        appointment_date: String,
        #[arg(short = 'p', long)]
        base_package: Option<Uuid>,
        #[arg(short, long, value_delimiter = ',', default_value = "")]
        add_items: Vec<Uuid>,
        #[arg(short, long, value_delimiter = ',', default_value = "")]
        remove_items: Vec<Uuid>,
    },
    #[command(about = "列出所有预约")]
    List,
    #[command(about = "获取预约详情")]
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct CheckupItem {
    id: Uuid,
    name: String,
    price: f64,
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server;

    match cli.command {
        Commands::Item { action } => handle_item(&client, &base_url, action).await,
        Commands::Package { action } => handle_package(&client, &base_url, action).await,
        Commands::Customer { action } => handle_customer(&client, &base_url, action).await,
        Commands::MutexRule { action } => handle_mutex_rule(&client, &base_url, action).await,
        Commands::Reservation { action } => handle_reservation(&client, &base_url, action).await,
    }
}

async fn handle_item(client: &Client, base_url: &str, action: ItemCommands) {
    match action {
        ItemCommands::Create { name, price, gender, contrast, gastroscopy, colonoscopy, xray, pregnant_forbidden } => {
            let gender_restriction = match gender.to_lowercase().as_str() {
                "male" => "MaleOnly",
                "female" => "FemaleOnly",
                _ => "None",
            };

            let body = serde_json::json!({
                "name": name,
                "price": price,
                "gender_restriction": gender_restriction,
                "has_contrast_agent": contrast,
                "is_gastroscopy": gastroscopy,
                "is_colonoscopy": colonoscopy,
                "is_xray": xray,
                "is_pregnant_forbidden": pregnant_forbidden,
            });

            let resp = client.post(&format!("{}/items", base_url))
                .json(&body)
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                println!("✓ 项目创建成功");
                let item: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&item).unwrap());
            } else {
                println!("✗ 创建失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
        ItemCommands::List => {
            let resp = client.get(&format!("{}/items", base_url))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let items: serde_json::Value = resp.json().await.unwrap();
                println!("项目列表:");
                println!("{}", serde_json::to_string_pretty(&items).unwrap());
            }
        }
        ItemCommands::Get { id } => {
            let resp = client.get(&format!("{}/items/{}", base_url, id))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let item: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&item).unwrap());
            } else {
                println!("✗ 获取失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
    }
}

async fn handle_package(client: &Client, base_url: &str, action: PackageCommands) {
    match action {
        PackageCommands::Create { name, items, price } => {
            let body = serde_json::json!({
                "name": name,
                "items": items,
                "package_price": price,
            });

            let resp = client.post(&format!("{}/packages", base_url))
                .json(&body)
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                println!("✓ 套餐创建成功");
                let pkg: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&pkg).unwrap());
            } else {
                println!("✗ 创建失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
        PackageCommands::List => {
            let resp = client.get(&format!("{}/packages", base_url))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let pkgs: serde_json::Value = resp.json().await.unwrap();
                println!("套餐列表:");
                println!("{}", serde_json::to_string_pretty(&pkgs).unwrap());
            }
        }
        PackageCommands::Get { id } => {
            let resp = client.get(&format!("{}/packages/{}", base_url, id))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let pkg: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&pkg).unwrap());
            } else {
                println!("✗ 获取失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
    }
}

async fn handle_customer(client: &Client, base_url: &str, action: CustomerCommands) {
    match action {
        CustomerCommands::Create { name, gender, birth_date, pregnant } => {
            let gender_enum = match gender.to_lowercase().as_str() {
                "male" => "Male",
                "female" => "Female",
                _ => "Unspecified",
            };

            let body = serde_json::json!({
                "name": name,
                "gender": gender_enum,
                "birth_date": birth_date,
                "is_pregnant": pregnant,
            });

            let resp = client.post(&format!("{}/customers", base_url))
                .json(&body)
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                println!("✓ 客户创建成功");
                let cust: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&cust).unwrap());
            } else {
                println!("✗ 创建失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
        CustomerCommands::List => {
            let resp = client.get(&format!("{}/customers", base_url))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let custs: serde_json::Value = resp.json().await.unwrap();
                println!("客户列表:");
                println!("{}", serde_json::to_string_pretty(&custs).unwrap());
            }
        }
        CustomerCommands::Get { id } => {
            let resp = client.get(&format!("{}/customers/{}", base_url, id))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let cust: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&cust).unwrap());
            } else {
                println!("✗ 获取失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
    }
}

async fn handle_mutex_rule(client: &Client, base_url: &str, action: MutexRuleCommands) {
    match action {
        MutexRuleCommands::Create { rule_type, item1, item2, description } => {
            let rule_type_enum = match rule_type.to_lowercase().as_str() {
                "hard" => "Hard",
                _ => "Warning",
            };

            let body = serde_json::json!({
                "rule_type": rule_type_enum,
                "item1": item1,
                "item2": item2,
                "description": description,
            });

            let resp = client.post(&format!("{}/mutex-rules", base_url))
                .json(&body)
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                println!("✓ 互斥规则创建成功");
                let rule: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&rule).unwrap());
            } else {
                println!("✗ 创建失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
        MutexRuleCommands::List => {
            let resp = client.get(&format!("{}/mutex-rules", base_url))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let rules: serde_json::Value = resp.json().await.unwrap();
                println!("互斥规则列表:");
                println!("{}", serde_json::to_string_pretty(&rules).unwrap());
            }
        }
        MutexRuleCommands::Delete { id } => {
            let resp = client.delete(&format!("{}/mutex-rules/{}", base_url, id))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                println!("✓ 互斥规则删除成功");
            } else {
                println!("✗ 删除失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
    }
}

async fn handle_reservation(client: &Client, base_url: &str, action: ReservationCommands) {
    match action {
        ReservationCommands::Create { customer_id, appointment_date, base_package, add_items, remove_items } => {
            let body = serde_json::json!({
                "customer_id": customer_id,
                "appointment_date": appointment_date,
                "base_package": base_package,
                "additional_items": add_items,
                "removed_items": remove_items,
            });

            let resp = client.post(&format!("{}/reservations", base_url))
                .json(&body)
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                println!("✓ 预约创建成功");
                let res: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            } else {
                println!("✗ 创建失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
        ReservationCommands::List => {
            let resp = client.get(&format!("{}/reservations", base_url))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let res: serde_json::Value = resp.json().await.unwrap();
                println!("预约列表:");
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            }
        }
        ReservationCommands::Get { id } => {
            let resp = client.get(&format!("{}/reservations/{}", base_url, id))
                .send()
                .await
                .unwrap();
            
            if resp.status().is_success() {
                let res: serde_json::Value = resp.json().await.unwrap();
                println!("{}", serde_json::to_string_pretty(&res).unwrap());
            } else {
                println!("✗ 获取失败: {}", resp.status());
                println!("{}", resp.text().await.unwrap());
            }
        }
    }
}
