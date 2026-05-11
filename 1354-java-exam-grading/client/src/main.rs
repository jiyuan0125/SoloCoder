use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "GRADING_SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    
    CreateTeacher {
        #[arg(short, long)]
        name: String,
    },
    
    GetTeacher {
        #[arg(short, long)]
        id: Uuid,
    },
    
    CreateStudent {
        #[arg(short, long)]
        name: String,
        #[arg(long)]
        mentor: Option<Uuid>,
    },
    
    GetStudent {
        #[arg(short, long)]
        id: Uuid,
    },
    
    SubmitExam {
        #[arg(short, long)]
        student: Uuid,
        #[arg(short, long)]
        exam: Uuid,
        #[arg(short, long)]
        answers: String,
    },
    
    GradeObjective {
        #[arg(short, long)]
        id: Uuid,
    },
    
    AssignTasks {
        #[arg(short, long)]
        exam: Uuid,
    },
    
    ListTasks {
        #[arg(short, long)]
        teacher: Uuid,
    },
    
    SubmitTask {
        #[arg(short, long)]
        task: Uuid,
        #[arg(short, long)]
        score: u32,
        #[arg(long)]
        comments: Option<String>,
    },
    
    MarkAnomaly {
        #[arg(short, long)]
        exam: Uuid,
        #[arg(long)]
        r#type: String,
        #[arg(short, long)]
        reporter: Uuid,
        #[arg(short, long)]
        description: String,
        #[arg(long)]
        related: Option<Uuid>,
    },
    
    PublishExam {
        #[arg(short, long)]
        exam: Uuid,
    },
    
    GetExam {
        #[arg(short, long)]
        id: Uuid,
    },
    
    GetStudentExam {
        #[arg(short, long)]
        id: Uuid,
    },
    
    RequestReview {
        #[arg(short, long)]
        exam: Uuid,
        #[arg(short, long)]
        student: Uuid,
        #[arg(short, long)]
        reason: String,
    },
    
    ProcessReview {
        #[arg(short, long)]
        review: Uuid,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct IdResponse {
    id: Uuid,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/');

    match args.command {
        Commands::Health => {
            let resp = client.get(format!("{}/health", base_url)).send().await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::CreateTeacher { name } => {
            let body = serde_json::json!({ "name": name });
            let resp: IdResponse = client
                .post(format!("{}/teachers", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;
            println!("Teacher ID: {}", resp.id);
        }
        
        Commands::GetTeacher { id } => {
            let resp = client
                .get(format!("{}/teachers/{}", base_url, id))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::CreateStudent { name, mentor } => {
            let mut body = serde_json::json!({ "name": name });
            if let Some(m) = mentor {
                body["mentor_id"] = serde_json::json!(m);
            }
            let resp: IdResponse = client
                .post(format!("{}/students", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;
            println!("Student ID: {}", resp.id);
        }
        
        Commands::GetStudent { id } => {
            let resp = client
                .get(format!("{}/students/{}", base_url, id))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::SubmitExam { student, exam, answers } => {
            let answers_json: serde_json::Value = serde_json::from_str(&answers)?;
            let body = serde_json::json!({
                "student_id": student,
                "exam_id": exam,
                "answers": answers_json
            });
            let resp: IdResponse = client
                .post(format!("{}/student-exams", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;
            println!("Student Exam ID: {}", resp.id);
        }
        
        Commands::GradeObjective { id } => {
            let resp = client
                .post(format!("{}/student-exams/{}/grade-objective", base_url, id))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::AssignTasks { exam } => {
            let resp = client
                .post(format!("{}/exams/{}/assign-tasks", base_url, exam))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::ListTasks { teacher } => {
            let resp = client
                .get(format!("{}/teachers/{}/tasks", base_url, teacher))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::SubmitTask { task, score, comments } => {
            let mut body = serde_json::json!({ "score": score });
            if let Some(c) = comments {
                body["comments"] = serde_json::json!(c);
            }
            let resp = client
                .post(format!("{}/tasks/{}/submit", base_url, task))
                .json(&body)
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::MarkAnomaly { exam, r#type, reporter, description, related } => {
            let mut body = serde_json::json!({
                "anomaly_type": r#type,
                "reported_by": reporter,
                "description": description
            });
            if let Some(r) = related {
                body["related_student_exam_id"] = serde_json::json!(r);
            }
            let resp: IdResponse = client
                .post(format!("{}/student-exams/{}/anomaly", base_url, exam))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;
            println!("Anomaly ID: {}", resp.id);
        }
        
        Commands::PublishExam { exam } => {
            let resp = client
                .post(format!("{}/exams/{}/publish", base_url, exam))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::GetExam { id } => {
            let resp = client
                .get(format!("{}/exams/{}", base_url, id))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::GetStudentExam { id } => {
            let resp = client
                .get(format!("{}/student-exams/{}", base_url, id))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
        
        Commands::RequestReview { exam, student, reason } => {
            let body = serde_json::json!({
                "student_exam_id": exam,
                "requested_by": student,
                "reason": reason
            });
            let resp: IdResponse = client
                .post(format!("{}/reviews", base_url))
                .json(&body)
                .send()
                .await?
                .json()
                .await?;
            println!("Review Request ID: {}", resp.id);
        }
        
        Commands::ProcessReview { review } => {
            let resp = client
                .post(format!("{}/reviews/{}/process", base_url, review))
                .send()
                .await?;
            println!("{}", resp.text().await?);
        }
    }

    Ok(())
}
