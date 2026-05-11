use std::str::FromStr;

use clap::{Parser, Subcommand};
use chrono::{DateTime, Utc, NaiveDateTime};
use reqwest::Client;
use rust_decimal::Decimal;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about = "宴会预订管理系统客户端", long_about = None)]
struct Cli {
    #[arg(long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    MenuSets,
    MenuItems,
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        phone: String,
        #[arg(long)]
        banquet_type: String,
        #[arg(long)]
        menu_set_id: String,
        #[arg(long)]
        tables: u32,
        #[arg(long)]
        event_date: String,
        #[arg(long)]
        deposit: String,
    },
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    AdjustTables {
        #[arg(long)]
        id: String,
        #[arg(long)]
        tables: u32,
    },
    SubstituteDish {
        #[arg(long)]
        id: String,
        #[arg(long)]
        original: String,
        #[arg(long)]
        replacement: String,
    },
    Complete {
        #[arg(long)]
        id: String,
        #[arg(long)]
        actual_tables: u32,
        #[arg(long, value_parser = ["pack", "refund"])]
        unused: String,
    },
    Cancel {
        #[arg(long)]
        id: String,
    },
    Summary {
        #[arg(long)]
        id: String,
    },
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::MenuSets => list_menu_sets(&client, &base_url).await,
        Commands::MenuItems => list_menu_items(&client, &base_url).await,
        Commands::Create { name, phone, banquet_type, menu_set_id, tables, event_date, deposit } => {
            create_booking(&client, &base_url, name, phone, banquet_type, menu_set_id, tables, event_date, deposit).await
        }
        Commands::List => list_bookings(&client, &base_url).await,
        Commands::Get { id } => get_booking(&client, &base_url, id).await,
        Commands::AdjustTables { id, tables } => adjust_tables(&client, &base_url, id, tables).await,
        Commands::SubstituteDish { id, original, replacement } => substitute_dish(&client, &base_url, id, original, replacement).await,
        Commands::Complete { id, actual_tables, unused } => complete_booking(&client, &base_url, id, actual_tables, unused).await,
        Commands::Cancel { id } => cancel_booking(&client, &base_url, id).await,
        Commands::Summary { id } => get_summary(&client, &base_url, id).await,
    }
}

async fn list_menu_sets(client: &Client, base_url: &str) {
    match client.get(format!("{}/menu-sets", base_url)).send().await {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("菜单套系列表：");
                if let Some(arr) = json.as_array() {
                    for set in arr {
                        println!("  ID: {}", set["id"].as_str().unwrap());
                        println!("  名称: {}", set["name"].as_str().unwrap());
                        println!("  类型: {:?}", set["banquet_type"]);
                        println!("  每桌价格: {}", set["price_per_table"]);
                        println!("  菜品数量: {}", set["items"].as_array().map(|a| a.len()).unwrap_or(0));
                        println!("  ---");
                    }
                }
            } else {
                eprintln!("请求失败: {}", resp.status());
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn list_menu_items(client: &Client, base_url: &str) {
    match client.get(format!("{}/menu-items", base_url)).send().await {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("菜品列表：");
                if let Some(arr) = json.as_array() {
                    for item in arr {
                        println!("  ID: {}", item["id"].as_str().unwrap());
                        println!("  名称: {}", item["name"].as_str().unwrap());
                        println!("  分类: {}", item["category"].as_str().unwrap());
                        println!("  价格: {}", item["price"]);
                        println!("  ---");
                    }
                }
            } else {
                eprintln!("请求失败: {}", resp.status());
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn create_booking(
    client: &Client,
    base_url: &str,
    name: String,
    phone: String,
    banquet_type: String,
    menu_set_id: String,
    tables: u32,
    event_date: String,
    deposit: String,
) {
    let banquet_type = match parse_banquet_type(&banquet_type) {
        Some(t) => t,
        None => {
            eprintln!("无效的宴会类型，支持: wedding/birthday/graduation/business");
            return;
        }
    };

    let menu_set_id = match Uuid::parse_str(&menu_set_id) {
        Ok(id) => id,
        Err(_) => {
            eprintln!("无效的菜单套系ID");
            return;
        }
    };

    let event_date = match parse_datetime(&event_date) {
        Some(dt) => dt,
        None => {
            eprintln!("无效的日期格式，请使用: YYYY-MM-DD HH:MM:SS");
            return;
        }
    };

    let deposit = match Decimal::from_str(&deposit) {
        Ok(d) => d,
        Err(_) => {
            eprintln!("无效的定金金额");
            return;
        }
    };

    let body = serde_json::json!({
        "customer_name": name,
        "customer_phone": phone,
        "banquet_type": banquet_type,
        "menu_set_id": menu_set_id,
        "booked_tables": tables,
        "event_date": event_date,
        "deposit": deposit,
    });

    match client.post(format!("{}/bookings", base_url))
        .json(&body)
        .send()
        .await
    {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("预订成功！");
                println!("预订ID: {}", json["id"].as_str().unwrap());
                println!("客户: {}", json["customer_name"].as_str().unwrap());
                println!("宴会类型: {:?}", json["banquet_type"]);
                println!("预订桌数: {}", json["booked_tables"]);
                println!("宴会日期: {}", json["event_date"]);
                println!("定金: {}", json["deposit"]);
            } else {
                let status = resp.status();
                let text = resp.text().await.unwrap_or_default();
                eprintln!("请求失败: {} - {}", status, text);
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn list_bookings(client: &Client, base_url: &str) {
    match client.get(format!("{}/bookings", base_url)).send().await {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("预订列表：");
                if let Some(arr) = json.as_array() {
                    if arr.is_empty() {
                        println!("  暂无预订");
                    } else {
                        for booking in arr {
                            println!("  ID: {}", booking["id"].as_str().unwrap());
                            println!("  客户: {}", booking["customer_name"].as_str().unwrap());
                            println!("  宴会类型: {:?}", booking["banquet_type"]);
                            println!("  桌数: {}", booking["booked_tables"]);
                            println!("  日期: {}", booking["event_date"]);
                            println!("  状态: {:?}", booking["status"]);
                            println!("  ---");
                        }
                    }
                }
            } else {
                eprintln!("请求失败: {}", resp.status());
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn get_booking(client: &Client, base_url: &str, id: String) {
    match client.get(format!("{}/bookings/{}", base_url, id)).send().await {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("预订详情：");
                println!("  ID: {}", json["id"].as_str().unwrap());
                println!("  客户: {}", json["customer_name"].as_str().unwrap());
                println!("  电话: {}", json["customer_phone"].as_str().unwrap());
                println!("  宴会类型: {:?}", json["banquet_type"]);
                println!("  菜单套系ID: {}", json["menu_set_id"].as_str().unwrap());
                println!("  预订桌数: {}", json["booked_tables"]);
                println!("  实际桌数: {}", json["actual_tables"]);
                println!("  宴会日期: {}", json["event_date"]);
                println!("  定金: {}", json["deposit"]);
                println!("  状态: {:?}", json["status"]);
                println!("  违约金: {}", json["penalty_amount"]);
                println!("  退款金额: {}", json["refund_amount"]);
                if let Some(subs) = json["substitutions"].as_array() {
                    if !subs.is_empty() {
                        println!("  菜品替换:");
                        for sub in subs {
                            println!("    {} -> {}", sub["original_name"], sub["replacement_name"]);
                        }
                    }
                }
            } else {
                eprintln!("请求失败: {}", resp.status());
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn adjust_tables(client: &Client, base_url: &str, id: String, tables: u32) {
    let body = serde_json::json!({
        "booking_id": id,
        "new_tables": tables,
    });

    match client.put(format!("{}/bookings/{}/adjust-tables", base_url, id))
        .json(&body)
        .send()
        .await
    {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("调桌成功！");
                println!("当前预订桌数: {}", json["booked_tables"]);
                println!("违约金: {}", json["penalty_amount"]);
            } else {
                let status = resp.status();
                let text = resp.text().await.unwrap_or_default();
                eprintln!("请求失败: {} - {}", status, text);
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn substitute_dish(client: &Client, base_url: &str, id: String, original: String, replacement: String) {
    let body = serde_json::json!({
        "booking_id": id,
        "original_item_id": original,
        "replacement_item_id": replacement,
    });

    match client.put(format!("{}/bookings/{}/substitute-dish", base_url, id))
        .json(&body)
        .send()
        .await
    {
        Ok(resp) => {
            if resp.status().is_success() {
                println!("菜品替换成功！");
            } else {
                let status = resp.status();
                let text = resp.text().await.unwrap_or_default();
                eprintln!("请求失败: {} - {}", status, text);
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn complete_booking(client: &Client, base_url: &str, id: String, actual_tables: u32, unused: String) {
    let unused_option = if unused == "pack" { "Pack" } else { "Refund" };
    
    let body = serde_json::json!({
        "booking_id": id,
        "actual_tables": actual_tables,
        "unused_option": unused_option,
    });

    match client.put(format!("{}/bookings/{}/complete", base_url, id))
        .json(&body)
        .send()
        .await
    {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("预订完成！");
                println!("实际桌数: {}", json["actual_tables"]);
                println!("退款金额: {}", json["refund_amount"]);
            } else {
                let status = resp.status();
                let text = resp.text().await.unwrap_or_default();
                eprintln!("请求失败: {} - {}", status, text);
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn cancel_booking(client: &Client, base_url: &str, id: String) {
    match client.delete(format!("{}/bookings/{}", base_url, id))
        .send()
        .await
    {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("预订已取消！");
                println!("退款金额: {}", json["refund_amount"]);
            } else {
                let status = resp.status();
                let text = resp.text().await.unwrap_or_default();
                eprintln!("请求失败: {} - {}", status, text);
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

async fn get_summary(client: &Client, base_url: &str, id: String) {
    match client.get(format!("{}/bookings/{}/summary", base_url, id)).send().await {
        Ok(resp) => {
            if resp.status().is_success() {
                let json: serde_json::Value = resp.json().await.unwrap();
                println!("预订费用汇总：");
                println!("  预订ID: {}", json["booking_id"].as_str().unwrap());
                println!("  客户: {}", json["customer_name"].as_str().unwrap());
                println!("  宴会类型: {}", json["banquet_type"].as_str().unwrap());
                println!("  预订桌数: {}", json["booked_tables"]);
                println!("  宴会日期: {}", json["event_date"]);
                println!("  状态: {:?}", json["status"]);
                println!("  总金额: {}", json["total_amount"]);
                println!("  定金: {}", json["deposit"]);
                println!("  违约金: {}", json["penalty_amount"]);
                println!("  退款金额: {}", json["refund_amount"]);
                println!("  应付余额: {}", json["balance"]);
            } else {
                eprintln!("请求失败: {}", resp.status());
            }
        }
        Err(e) => eprintln!("连接失败: {}", e),
    }
}

fn parse_banquet_type(s: &str) -> Option<String> {
    match s.to_lowercase().as_str() {
        "wedding" | "婚宴" => Some("Wedding".to_string()),
        "birthday" | "寿宴" => Some("Birthday".to_string()),
        "graduation" | "升学宴" => Some("Graduation".to_string()),
        "business" | "商务宴" => Some("Business".to_string()),
        _ => None,
    }
}

fn parse_datetime(s: &str) -> Option<DateTime<Utc>> {
    let formats = [
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%dT%H:%M:%S",
        "%Y-%m-%d",
    ];

    for format in &formats {
        if let Ok(dt) = NaiveDateTime::parse_from_str(s, format) {
            return Some(DateTime::from_naive_utc_and_offset(dt, Utc));
        }
        if let Ok(date) = chrono::NaiveDate::parse_from_str(s, "%Y-%m-%d") {
            let dt = date.and_hms_opt(18, 0, 0).unwrap();
            return Some(DateTime::from_naive_utc_and_offset(dt, Utc));
        }
    }
    None
}
