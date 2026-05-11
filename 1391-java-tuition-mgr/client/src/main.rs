use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:8601")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateStudent {
        name: String,
    },
    ListStudents,
    GetStudent {
        id: Uuid,
    },
    CreateCourse {
        name: String,
        #[arg(short, long)]
        capacity: u32,
        #[arg(short, long)]
        semester: String,
    },
    ListCourses {
        #[arg(short, long)]
        semester: String,
    },
    GetCourse {
        id: Uuid,
    },
    SetTuition {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        semester: String,
        #[arg(short, long)]
        amount: f64,
    },
    GetTuition {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        semester: String,
    },
    PayFirst {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        semester: String,
        #[arg(short, long)]
        amount: f64,
    },
    PaySecond {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        semester: String,
        #[arg(short, long)]
        amount: f64,
    },
    SetEnrollmentPeriod {
        #[arg(short, long)]
        semester: String,
        #[arg(short = 's', long)]
        start: String,
        #[arg(short = 'e', long)]
        end: String,
    },
    Enroll {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        course_id: Uuid,
        #[arg(short, long)]
        semester: String,
    },
    DropCourse {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        course_id: Uuid,
        #[arg(short, long)]
        semester: String,
    },
    ListEnrollments {
        #[arg(short, long)]
        student_id: Uuid,
        #[arg(short, long)]
        semester: String,
    },
    ListOwing {
        #[arg(short, long)]
        semester: String,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct Student {
    id: Uuid,
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct Course {
    id: Uuid,
    name: String,
    capacity: u32,
    semester: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CourseWithEnrollment {
    course: Course,
    enrolled_count: u32,
    is_full: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct TuitionRecord {
    student_id: Uuid,
    semester: String,
    total_amount: f64,
    first_installment: f64,
    second_installment: f64,
    is_paid: bool,
}

#[derive(Debug, Serialize, Deserialize)]
struct StudentOwingInfo {
    student: Student,
    semester: String,
    owed_amount: f64,
}

#[derive(Debug, Serialize)]
struct CreateStudentRequest {
    name: String,
}

#[derive(Debug, Serialize)]
struct CreateCourseRequest {
    name: String,
    capacity: u32,
    semester: String,
}

#[derive(Debug, Serialize)]
struct SetTuitionRequest {
    student_id: Uuid,
    semester: String,
    amount: f64,
}

#[derive(Debug, Serialize)]
struct PayRequest {
    student_id: Uuid,
    semester: String,
    amount: f64,
}

#[derive(Debug, Serialize)]
struct SetEnrollmentPeriodRequest {
    semester: String,
    start_time: chrono::DateTime<chrono::Utc>,
    end_time: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Serialize)]
struct EnrollRequest {
    student_id: Uuid,
    course_id: Uuid,
    semester: String,
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server.trim_end_matches('/');

    let result = match cli.command {
        Commands::CreateStudent { name } => {
            let req = CreateStudentRequest { name };
            let resp = client
                .post(format!("{}/students", base_url))
                .json(&req)
                .send()
                .await;
            handle_response::<Student>(resp).await
        }
        Commands::ListStudents => {
            let resp = client.get(format!("{}/students", base_url)).send().await;
            handle_response::<Vec<Student>>(resp).await
        }
        Commands::GetStudent { id } => {
            let resp = client
                .get(format!("{}/students/{}", base_url, id))
                .send()
                .await;
            handle_response::<Student>(resp).await
        }
        Commands::CreateCourse { name, capacity, semester } => {
            let req = CreateCourseRequest { name, capacity, semester };
            let resp = client
                .post(format!("{}/courses", base_url))
                .json(&req)
                .send()
                .await;
            handle_response::<Course>(resp).await
        }
        Commands::ListCourses { semester } => {
            let resp = client
                .get(format!("{}/courses?semester={}", base_url, semester))
                .send()
                .await;
            handle_response::<Vec<CourseWithEnrollment>>(resp).await
        }
        Commands::GetCourse { id } => {
            let resp = client
                .get(format!("{}/courses/{}", base_url, id))
                .send()
                .await;
            handle_response::<Course>(resp).await
        }
        Commands::SetTuition { student_id, semester, amount } => {
            let req = SetTuitionRequest { student_id, semester, amount };
            let resp = client
                .post(format!("{}/tuition", base_url))
                .json(&req)
                .send()
                .await;
            handle_response::<TuitionRecord>(resp).await
        }
        Commands::GetTuition { student_id, semester } => {
            let resp = client
                .get(format!("{}/tuition/{}/{}", base_url, student_id, semester))
                .send()
                .await;
            handle_response::<TuitionRecord>(resp).await
        }
        Commands::PayFirst { student_id, semester, amount } => {
            let req = PayRequest { student_id, semester, amount };
            let resp = client
                .post(format!("{}/pay/first", base_url))
                .json(&req)
                .send()
                .await;
            handle_response::<TuitionRecord>(resp).await
        }
        Commands::PaySecond { student_id, semester, amount } => {
            let req = PayRequest { student_id, semester, amount };
            let resp = client
                .post(format!("{}/pay/second", base_url))
                .json(&req)
                .send()
                .await;
            handle_response::<TuitionRecord>(resp).await
        }
        Commands::SetEnrollmentPeriod { semester, start, end } => {
            let start_time: chrono::DateTime<chrono::Utc> = start.parse().unwrap();
            let end_time: chrono::DateTime<chrono::Utc> = end.parse().unwrap();
            let req = SetEnrollmentPeriodRequest { semester, start_time, end_time };
            let resp = client
                .post(format!("{}/enrollment-period", base_url))
                .json(&req)
                .send()
                .await;
            match resp {
                Ok(r) if r.status().is_success() => Ok("Enrollment period set successfully".to_string()),
                Ok(r) => Err(format!("Error: {}", r.text().await.unwrap_or_default())),
                Err(e) => Err(format!("Request failed: {}", e)),
            }
        }
        Commands::Enroll { student_id, course_id, semester } => {
            let req = EnrollRequest { student_id, course_id, semester };
            let resp = client
                .post(format!("{}/enroll", base_url))
                .json(&req)
                .send()
                .await;
            handle_response::<serde_json::Value>(resp).await
        }
        Commands::DropCourse { student_id, course_id, semester } => {
            let resp = client
                .delete(format!("{}/enroll/{}/{}/{}", base_url, student_id, course_id, semester))
                .send()
                .await;
            match resp {
                Ok(r) if r.status().is_success() => Ok("Course dropped successfully".to_string()),
                Ok(r) => Err(format!("Error: {}", r.text().await.unwrap_or_default())),
                Err(e) => Err(format!("Request failed: {}", e)),
            }
        }
        Commands::ListEnrollments { student_id, semester } => {
            let resp = client
                .get(format!("{}/enrollments/{}/{}", base_url, student_id, semester))
                .send()
                .await;
            handle_response::<Vec<Course>>(resp).await
        }
        Commands::ListOwing { semester } => {
            let resp = client
                .get(format!("{}/owing/{}", base_url, semester))
                .send()
                .await;
            handle_response::<Vec<StudentOwingInfo>>(resp).await
        }
    };

    match result {
        Ok(msg) => println!("{}", msg),
        Err(e) => eprintln!("{}", e),
    }
}

async fn handle_response<T: for<'de> Deserialize<'de> + std::fmt::Debug + Serialize>(
    resp: Result<reqwest::Response, reqwest::Error>,
) -> Result<String, String> {
    match resp {
        Ok(response) => {
            let status = response.status();
            if status.is_success() {
                let body = response.text().await.map_err(|e| e.to_string())?;
                if body.is_empty() {
                    Ok("Success".to_string())
                } else {
                    let parsed: T = serde_json::from_str(&body).map_err(|e| e.to_string())?;
                    Ok(serde_json::to_string_pretty(&parsed).unwrap_or(format!("{:?}", parsed)))
                }
            } else {
                let body = response.text().await.unwrap_or_default();
                Err(format!("HTTP {}: {}", status, body))
            }
        }
        Err(e) => Err(format!("Request failed: {}", e)),
    }
}
