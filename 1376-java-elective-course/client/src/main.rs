use clap::{Parser, Subcommand};
use serde::Deserialize;
use serde_json::Value;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Courses,
    Course { id: String },
    Students,
    Student { id: String },
    Enroll { student_id: String, course_id: String },
    Withdraw { student_id: String, course_id: String },
    Enrollments { student_id: String },
    Transcript { student_id: String },
}

#[derive(Deserialize, Debug)]
struct ApiError {
    error: String,
}

async fn get_json(client: &reqwest::Client, url: &str) -> Result<Value, String> {
    let response = client.get(url).send().await.map_err(|e| e.to_string())?;
    
    if response.status().is_success() {
        response.json::<Value>().await.map_err(|e| e.to_string())
    } else {
        let error: ApiError = response.json().await.map_err(|e| e.to_string())?;
        Err(error.error)
    }
}

async fn post_json(client: &reqwest::Client, url: &str, body: Value) -> Result<Value, String> {
    let response = client.post(url).json(&body).send().await.map_err(|e| e.to_string())?;
    
    if response.status().is_success() {
        response.json::<Value>().await.map_err(|e| e.to_string())
    } else {
        let error: ApiError = response.json().await.map_err(|e| e.to_string())?;
        Err(error.error)
    }
}

fn print_courses(courses: &Value) {
    if let Some(arr) = courses.as_array() {
        println!("课程列表:");
        println!("{:-<60}", "");
        for course in arr {
            println!(
                "{} - {} (学分: {}, 容量: {})",
                course["id"].as_str().unwrap_or(""),
                course["name"].as_str().unwrap_or(""),
                course["credits"],
                course["capacity"]
            );
            if let Some(prereqs) = course["prerequisites"].as_array() {
                if !prereqs.is_empty() {
                    let prereq_names: Vec<_> = prereqs.iter().map(|p| p.as_str().unwrap_or("")).collect();
                    println!("  先修课程: {}", prereq_names.join(", "));
                }
            }
        }
    }
}

fn print_students(students: &Value) {
    if let Some(arr) = students.as_array() {
        println!("学生列表:");
        println!("{:-<60}", "");
        for student in arr {
            println!(
                "{} - {}",
                student["id"].as_str().unwrap_or(""),
                student["name"].as_str().unwrap_or("")
            );
        }
    }
}

fn print_enrollments(enrollments: &Value) {
    println!("选课结果:");
    println!("{:-<60}", "");
    println!("学生ID: {}", enrollments["student_id"].as_str().unwrap_or(""));
    
    if let Some(courses) = enrollments["enrolled_courses"].as_array() {
        if courses.is_empty() {
            println!("当前未选任何课程");
        } else {
            println!("已选课程:");
            for course in courses {
                println!(
                    "  {} - {} ({}学分)",
                    course["id"].as_str().unwrap_or(""),
                    course["name"].as_str().unwrap_or(""),
                    course["credits"]
                );
            }
        }
    }
    println!("\n总学分: {}", enrollments["total_credits"]);
}

fn print_transcript(transcript: &Value) {
    println!("成绩单:");
    println!("{:-<60}", "");
    println!("学生ID: {}", transcript["student_id"].as_str().unwrap_or(""));
    println!("");
    
    if let Some(courses) = transcript["courses"].as_array() {
        if courses.is_empty() {
            println!("暂无记录");
        } else {
            for entry in courses {
                match entry["type"].as_str() {
                    Some("completed") => {
                        println!(
                            "{} - {}: {}分 ({}学分)",
                            entry["course_id"].as_str().unwrap_or(""),
                            entry["course_name"].as_str().unwrap_or(""),
                            entry["score"],
                            entry["credits"]
                        );
                    }
                    Some("withdrawal") => {
                        println!(
                            "{} - {}: W (退课记录)",
                            entry["course_id"].as_str().unwrap_or(""),
                            entry["course_name"].as_str().unwrap_or("")
                        );
                    }
                    _ => {}
                }
            }
        }
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let client = reqwest::Client::new();

    match args.command {
        Commands::Courses => {
            let url = format!("{}/courses", args.server);
            match get_json(&client, &url).await {
                Ok(data) => print_courses(&data),
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Course { id } => {
            let url = format!("{}/courses/{}", args.server, id);
            match get_json(&client, &url).await {
                Ok(data) => {
                    println!("课程详情:");
                    println!("{:-<60}", "");
                    println!("ID: {}", data["id"].as_str().unwrap_or(""));
                    println!("名称: {}", data["name"].as_str().unwrap_or(""));
                    println!("学分: {}", data["credits"]);
                    println!("容量: {}", data["capacity"]);
                    if let Some(prereqs) = data["prerequisites"].as_array() {
                        let prereq_names: Vec<_> = prereqs.iter().map(|p| p.as_str().unwrap_or("")).collect();
                        println!("先修课程: {}", prereq_names.join(", "));
                    }
                }
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Students => {
            let url = format!("{}/students", args.server);
            match get_json(&client, &url).await {
                Ok(data) => print_students(&data),
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Student { id } => {
            let url = format!("{}/students/{}", args.server, id);
            match get_json(&client, &url).await {
                Ok(data) => {
                    println!("学生详情:");
                    println!("{:-<60}", "");
                    println!("ID: {}", data["id"].as_str().unwrap_or(""));
                    println!("姓名: {}", data["name"].as_str().unwrap_or(""));
                    if let Some(completed) = data["completed_courses"].as_array() {
                        if !completed.is_empty() {
                            println!("已完成课程:");
                            for c in completed {
                                println!(
                                    "  {} - {}分",
                                    c["course_id"].as_str().unwrap_or(""),
                                    c["score"]
                                );
                            }
                        }
                    }
                }
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Enroll { student_id, course_id } => {
            let url = format!("{}/enroll", args.server);
            let body = serde_json::json!({
                "student_id": student_id,
                "course_id": course_id,
            });
            match post_json(&client, &url, body).await {
                Ok(data) => {
                    println!("{}", data["message"].as_str().unwrap_or(""));
                }
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Withdraw { student_id, course_id } => {
            let url = format!("{}/withdraw", args.server);
            let body = serde_json::json!({
                "student_id": student_id,
                "course_id": course_id,
            });
            match post_json(&client, &url, body).await {
                Ok(data) => {
                    println!("{}", data["message"].as_str().unwrap_or(""));
                }
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Enrollments { student_id } => {
            let url = format!("{}/students/{}/enrollments", args.server, student_id);
            match get_json(&client, &url).await {
                Ok(data) => print_enrollments(&data),
                Err(e) => println!("错误: {}", e),
            }
        }
        Commands::Transcript { student_id } => {
            let url = format!("{}/students/{}/transcript", args.server, student_id);
            match get_json(&client, &url).await {
                Ok(data) => print_transcript(&data),
                Err(e) => println!("错误: {}", e),
            }
        }
    }
}
