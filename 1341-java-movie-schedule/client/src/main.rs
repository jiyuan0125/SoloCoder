use std::env;
use movie_schedule_core::{
    Hall, Movie, Schedule, ScheduleCreate, ScheduleUpdate,
};
use chrono::{DateTime, Local, TimeZone};
use uuid::Uuid;
use reqwest::Client;
use serde::de::DeserializeOwned;

const DEFAULT_BASE_URL: &str = "http://127.0.0.1:8100";

struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    fn new() -> Self {
        let base_url = env::var("SERVER_URL").unwrap_or_else(|_| DEFAULT_BASE_URL.to_string());
        Self {
            client: Client::new(),
            base_url,
        }
    }

    async fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T, Box<dyn std::error::Error>> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.get(&url).send().await?;
        let data = response.json::<T>().await?;
        Ok(data)
    }

    async fn post<T: DeserializeOwned, B: serde::Serialize>(&self, path: &str, body: &B) -> Result<T, Box<dyn std::error::Error>> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.post(&url).json(body).send().await?;
        if !response.status().is_success() {
            let status = response.status();
            let error_text = response.text().await?;
            return Err(format!("请求失败 ({}): {}", status, error_text).into());
        }
        let data = response.json::<T>().await?;
        Ok(data)
    }

    async fn put<T: DeserializeOwned, B: serde::Serialize>(&self, path: &str, body: &B) -> Result<T, Box<dyn std::error::Error>> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.put(&url).json(body).send().await?;
        if !response.status().is_success() {
            let status = response.status();
            let error_text = response.text().await?;
            return Err(format!("请求失败 ({}): {}", status, error_text).into());
        }
        let data = response.json::<T>().await?;
        Ok(data)
    }

    async fn delete(&self, path: &str) -> Result<(), Box<dyn std::error::Error>> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.delete(&url).send().await?;
        if !response.status().is_success() {
            let status = response.status();
            let error_text = response.text().await?;
            return Err(format!("请求失败 ({}): {}", status, error_text).into());
        }
        Ok(())
    }
}

fn print_help() {
    println!("影院排片管理系统 - 命令行客户端");
    println!();
    println!("用法:");
    println!("  client halls                      列出所有影厅");
    println!("  client movies                     列出所有影片");
    println!("  client schedules                  列出所有排片");
    println!("  client create <hall_id> <movie_id> <start_time> 创建排片");
    println!("  client update <schedule_id> [--hall <hall_id>] [--movie <movie_id>] [--start <start_time>] 更新排片");
    println!("  client delete <schedule_id>       删除排片");
    println!("  client help                       显示帮助");
    println!();
    println!("时间格式示例: 2024-01-15 19:00");
    println!("环境变量: SERVER_URL (默认: http://127.0.0.1:8100)");
}

#[tokio::main]
async fn main() {
    let args: Vec<String> = env::args().collect();
    
    if args.len() < 2 {
        print_help();
        return;
    }

    let client = ApiClient::new();
    let command = args[1].as_str();

    match command {
        "halls" => {
            if let Err(e) = list_halls(&client).await {
                eprintln!("错误: {}", e);
            }
        }
        "movies" => {
            if let Err(e) = list_movies(&client).await {
                eprintln!("错误: {}", e);
            }
        }
        "schedules" => {
            if let Err(e) = list_schedules(&client).await {
                eprintln!("错误: {}", e);
            }
        }
        "create" => {
            if args.len() < 5 {
                eprintln!("错误: create命令需要 hall_id, movie_id, start_time 三个参数");
                return;
            }
            if let Err(e) = create_schedule(&client, &args[2], &args[3], &args[4..].join(" ")).await {
                eprintln!("错误: {}", e);
            }
        }
        "update" => {
            if args.len() < 3 {
                eprintln!("错误: update命令需要 schedule_id 参数");
                return;
            }
            if let Err(e) = update_schedule(&client, &args[2], &args[3..]).await {
                eprintln!("错误: {}", e);
            }
        }
        "delete" => {
            if args.len() < 3 {
                eprintln!("错误: delete命令需要 schedule_id 参数");
                return;
            }
            if let Err(e) = delete_schedule(&client, &args[2]).await {
                eprintln!("错误: {}", e);
            }
        }
        "help" | "--help" | "-h" => {
            print_help();
        }
        _ => {
            eprintln!("未知命令: {}", command);
            print_help();
        }
    }
}

async fn list_halls(client: &ApiClient) -> Result<(), Box<dyn std::error::Error>> {
    let halls: Vec<Hall> = client.get("/api/halls").await?;
    println!("\n影厅列表:");
    println!("{:-<80}", "");
    for hall in halls {
        println!(
            "ID: {}\n  名称: {}\n  类型: {}\n  座位数: {}\n  基础票价: ¥{:.0}\n",
            hall.id,
            hall.name,
            hall.hall_type.display_name(),
            hall.hall_type.seat_count(),
            hall.hall_type.base_price()
        );
    }
    Ok(())
}

async fn list_movies(client: &ApiClient) -> Result<(), Box<dyn std::error::Error>> {
    let movies: Vec<Movie> = client.get("/api/movies").await?;
    println!("\n影片列表:");
    println!("{:-<80}", "");
    for movie in movies {
        let hours = movie.duration_minutes / 60;
        let mins = movie.duration_minutes % 60;
        println!(
            "ID: {}\n  名称: {}\n  时长: {}小时{}分钟\n",
            movie.id,
            movie.title,
            hours,
            mins
        );
    }
    Ok(())
}

async fn list_schedules(client: &ApiClient) -> Result<(), Box<dyn std::error::Error>> {
    let schedules: Vec<Schedule> = client.get("/api/schedules").await?;
    let halls: Vec<Hall> = client.get("/api/halls").await?;
    let movies: Vec<Movie> = client.get("/api/movies").await?;
    
    let hall_map: std::collections::HashMap<_, _> = halls.into_iter().map(|h| (h.id, h)).collect();
    let movie_map: std::collections::HashMap<_, _> = movies.into_iter().map(|m| (m.id, m)).collect();

    println!("\n排片列表:");
    println!("{:-<100}", "");
    
    if schedules.is_empty() {
        println!("暂无排片");
    } else {
        for schedule in schedules {
            let hall = hall_map.get(&schedule.hall_id);
            let movie = movie_map.get(&schedule.movie_id);
            
            let duration = schedule.end_time.signed_duration_since(schedule.start_time);
            let hours = duration.num_hours();
            let mins = duration.num_minutes() % 60;
            
            println!(
                "ID: {}\n  影厅: {}\n  影片: {}\n  时间: {} - {} (时长: {}小时{}分钟)\n",
                schedule.id,
                hall.map(|h| h.name.as_str()).unwrap_or("未知影厅"),
                movie.map(|m| m.title.as_str()).unwrap_or("未知影片"),
                schedule.start_time.format("%Y-%m-%d %H:%M"),
                schedule.end_time.format("%Y-%m-%d %H:%M"),
                hours,
                mins
            );
        }
    }
    Ok(())
}

async fn create_schedule(
    client: &ApiClient,
    hall_id_str: &str,
    movie_id_str: &str,
    start_time_str: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let hall_id = Uuid::parse_str(hall_id_str)?;
    let movie_id = Uuid::parse_str(movie_id_str)?;
    
    let start_time = parse_datetime(start_time_str)?;
    
    let create = ScheduleCreate {
        hall_id,
        movie_id,
        start_time,
    };

    println!("\n正在创建排片...");
    let schedule: Schedule = client.post("/api/schedules", &create).await?;
    
    println!("排片创建成功！");
    let duration = schedule.end_time.signed_duration_since(schedule.start_time);
    let hours = duration.num_hours();
    let mins = duration.num_minutes() % 60;
    println!(
        "ID: {}\n时间: {} - {} (时长: {}小时{}分钟)",
        schedule.id,
        schedule.start_time.format("%Y-%m-%d %H:%M"),
        schedule.end_time.format("%Y-%m-%d %H:%M"),
        hours,
        mins
    );
    
    Ok(())
}

async fn update_schedule(
    client: &ApiClient,
    schedule_id_str: &str,
    args: &[String],
) -> Result<(), Box<dyn std::error::Error>> {
    let schedule_id = Uuid::parse_str(schedule_id_str)?;
    
    let mut update = ScheduleUpdate {
        hall_id: None,
        movie_id: None,
        start_time: None,
    };

    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "--hall" => {
                if i + 1 < args.len() {
                    update.hall_id = Some(Uuid::parse_str(&args[i + 1])?);
                    i += 2;
                } else {
                    return Err("--hall 参数缺少值".into());
                }
            }
            "--movie" => {
                if i + 1 < args.len() {
                    update.movie_id = Some(Uuid::parse_str(&args[i + 1])?);
                    i += 2;
                } else {
                    return Err("--movie 参数缺少值".into());
                }
            }
            "--start" => {
                if i + 1 < args.len() {
                    let time_str = format!("{} {}", args[i + 1], args.get(i + 2).unwrap_or(&"00:00".to_string()));
                    update.start_time = Some(parse_datetime(&time_str)?);
                    i += 3;
                } else {
                    return Err("--start 参数缺少值".into());
                }
            }
            _ => {
                i += 1;
            }
        }
    }

    println!("\n正在更新排片...");
    let schedule: Schedule = client.put(
        &format!("/api/schedules/{}", schedule_id),
        &update,
    ).await?;
    
    println!("排片更新成功！");
    let duration = schedule.end_time.signed_duration_since(schedule.start_time);
    let hours = duration.num_hours();
    let mins = duration.num_minutes() % 60;
    println!(
        "ID: {}\n时间: {} - {} (时长: {}小时{}分钟)",
        schedule.id,
        schedule.start_time.format("%Y-%m-%d %H:%M"),
        schedule.end_time.format("%Y-%m-%d %H:%M"),
        hours,
        mins
    );
    
    Ok(())
}

async fn delete_schedule(client: &ApiClient, schedule_id_str: &str) -> Result<(), Box<dyn std::error::Error>> {
    let schedule_id = Uuid::parse_str(schedule_id_str)?;
    
    println!("\n正在删除排片...");
    client.delete(&format!("/api/schedules/{}", schedule_id)).await?;
    println!("排片删除成功！");
    
    Ok(())
}

fn parse_datetime(s: &str) -> Result<DateTime<Local>, Box<dyn std::error::Error>> {
    let parsed = chrono::NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M")?;
    let local = Local.from_local_datetime(&parsed).single()
        .ok_or("无法解析本地时间")?;
    Ok(local)
}
