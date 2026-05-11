use anyhow::{Context, Result};
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:8080")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    User {
        #[command(subcommand)]
        action: UserCommands,
    },
    Course {
        #[command(subcommand)]
        action: CourseCommands,
    },
    Assignment {
        #[command(subcommand)]
        action: AssignmentCommands,
    },
    Submission {
        #[command(subcommand)]
        action: SubmissionCommands,
    },
    Stats {
        #[command(subcommand)]
        action: StatsCommands,
    },
}

#[derive(Subcommand, Debug)]
enum UserCommands {
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        role: String,
    },
}

#[derive(Subcommand, Debug)]
enum CourseCommands {
    List,
    Get {
        #[arg(long)]
        id: String,
    },
    Create {
        #[arg(long)]
        name: String,
        #[arg(long, num_args = 0..)]
        teachers: Vec<String>,
    },
    AddStudent {
        #[arg(long)]
        course_id: String,
        #[arg(long)]
        student_id: String,
    },
}

#[derive(Subcommand, Debug)]
enum AssignmentCommands {
    List {
        #[arg(long)]
        course_id: String,
    },
    Get {
        #[arg(long)]
        id: String,
    },
    Create {
        #[arg(long)]
        course_id: String,
        #[arg(long)]
        title: String,
        #[arg(long)]
        teacher_id: String,
        #[arg(long, num_args = 0..)]
        questions: Vec<String>,
    },
}

#[derive(Subcommand, Debug)]
enum SubmissionCommands {
    List {
        #[arg(long)]
        assignment_id: String,
    },
    Get {
        #[arg(long)]
        id: String,
    },
    Submit {
        #[arg(long)]
        assignment_id: String,
        #[arg(long)]
        student_id: String,
        #[arg(long, num_args = 0..)]
        answers: Vec<String>,
    },
    Grade {
        #[arg(long)]
        submission_id: String,
        #[arg(long)]
        teacher_id: String,
        #[arg(long, num_args = 0..)]
        scores: Vec<String>,
    },
    Appeal {
        #[arg(long)]
        submission_id: String,
        #[arg(long)]
        student_id: String,
    },
    GradeAppeal {
        #[arg(long)]
        submission_id: String,
        #[arg(long)]
        teacher_id: String,
        #[arg(long, num_args = 0..)]
        scores: Vec<String>,
    },
}

#[derive(Subcommand, Debug)]
enum StatsCommands {
    Course {
        #[arg(long)]
        course_id: String,
    },
    Distribution {
        #[arg(long)]
        course_id: String,
    },
    Trends {
        #[arg(long)]
        course_id: String,
    },
    Workload,
}

#[derive(Debug, Serialize)]
struct CreateUserRequest {
    name: String,
    role: String,
}

#[derive(Debug, Serialize)]
struct CreateCourseRequest {
    name: String,
    teacher_ids: Vec<String>,
}

#[derive(Debug, Serialize)]
struct AddStudentRequest {
    student_id: String,
}

#[derive(Debug, Serialize)]
struct CreateQuestion {
    question_type: String,
    title: String,
    max_score: u32,
    answer: Option<String>,
}

#[derive(Debug, Serialize)]
struct CreateAssignmentRequest {
    title: String,
    teacher_id: String,
    questions: Vec<CreateQuestion>,
}

#[derive(Debug, Serialize, Deserialize)]
struct AnswerDto {
    question_id: String,
    answer: String,
}

#[derive(Debug, Serialize)]
struct SubmitAssignmentRequest {
    student_id: String,
    answers: Vec<AnswerDto>,
}

#[derive(Debug, Serialize)]
struct GradeSubjectiveRequest {
    teacher_id: String,
    scores: HashMap<String, u32>,
}

#[derive(Debug, Serialize)]
struct AppealRequest {
    student_id: String,
}

fn parse_questions(inputs: &[String]) -> Result<Vec<CreateQuestion>> {
    let mut questions = Vec::new();
    for input in inputs {
        let parts: Vec<&str> = input.split('|').collect();
        if parts.len() < 3 {
            anyhow::bail!("Question format must be: type|title|max_score|answer(optional)");
        }
        let q_type = parts[0].to_string();
        let title = parts[1].to_string();
        let max_score: u32 = parts[2].parse().context("Invalid max score")?;
        let answer = if parts.len() > 3 {
            Some(parts[3].to_string())
        } else {
            None
        };
        questions.push(CreateQuestion {
            question_type: q_type,
            title,
            max_score,
            answer,
        });
    }
    Ok(questions)
}

fn parse_answers(inputs: &[String]) -> Result<Vec<AnswerDto>> {
    let mut answers = Vec::new();
    for input in inputs {
        let parts: Vec<&str> = input.splitn(2, '|').collect();
        if parts.len() != 2 {
            anyhow::bail!("Answer format must be: question_id|answer");
        }
        answers.push(AnswerDto {
            question_id: parts[0].to_string(),
            answer: parts[1].to_string(),
        });
    }
    Ok(answers)
}

fn parse_scores(inputs: &[String]) -> Result<HashMap<String, u32>> {
    let mut scores = HashMap::new();
    for input in inputs {
        let parts: Vec<&str> = input.splitn(2, '|').collect();
        if parts.len() != 2 {
            anyhow::bail!("Score format must be: question_id|score");
        }
        let score: u32 = parts[1].parse().context("Invalid score")?;
        scores.insert(parts[0].to_string(), score);
    }
    Ok(scores)
}

async fn print_response(res: reqwest::Response) -> Result<()> {
    let status = res.status();
    let body: serde_json::Value = res.json().await?;
    println!("Status: {}", status);
    println!("{}", serde_json::to_string_pretty(&body)?);
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();

    match args.command {
        Commands::User { action } => match action {
            UserCommands::List => {
                let res = client.get(format!("{}/users", base_url)).send().await?;
                print_response(res).await?;
            }
            UserCommands::Get { id } => {
                let res = client.get(format!("{}/users/{}", base_url, id)).send().await?;
                print_response(res).await?;
            }
            UserCommands::Create { name, role } => {
                let req = CreateUserRequest { name, role };
                let res = client
                    .post(format!("{}/users", base_url))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
        },
        Commands::Course { action } => match action {
            CourseCommands::List => {
                let res = client.get(format!("{}/courses", base_url)).send().await?;
                print_response(res).await?;
            }
            CourseCommands::Get { id } => {
                let res = client.get(format!("{}/courses/{}", base_url, id)).send().await?;
                print_response(res).await?;
            }
            CourseCommands::Create { name, teachers } => {
                let req = CreateCourseRequest {
                    name,
                    teacher_ids: teachers,
                };
                let res = client
                    .post(format!("{}/courses", base_url))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
            CourseCommands::AddStudent { course_id, student_id } => {
                let req = AddStudentRequest { student_id };
                let res = client
                    .post(format!("{}/courses/{}/students", base_url, course_id))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
        },
        Commands::Assignment { action } => match action {
            AssignmentCommands::List { course_id } => {
                let res = client
                    .get(format!("{}/courses/{}/assignments", base_url, course_id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            AssignmentCommands::Get { id } => {
                let res = client
                    .get(format!("{}/assignments/{}", base_url, id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            AssignmentCommands::Create {
                course_id,
                title,
                teacher_id,
                questions,
            } => {
                let questions = parse_questions(&questions)?;
                let req = CreateAssignmentRequest {
                    title,
                    teacher_id,
                    questions,
                };
                let res = client
                    .post(format!("{}/courses/{}/assignments", base_url, course_id))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
        },
        Commands::Submission { action } => match action {
            SubmissionCommands::List { assignment_id } => {
                let res = client
                    .get(format!("{}/assignments/{}/submissions", base_url, assignment_id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            SubmissionCommands::Get { id } => {
                let res = client
                    .get(format!("{}/submissions/{}", base_url, id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            SubmissionCommands::Submit {
                assignment_id,
                student_id,
                answers,
            } => {
                let answers = parse_answers(&answers)?;
                let req = SubmitAssignmentRequest {
                    student_id,
                    answers,
                };
                let res = client
                    .post(format!("{}/assignments/{}/submissions", base_url, assignment_id))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
            SubmissionCommands::Grade {
                submission_id,
                teacher_id,
                scores,
            } => {
                let scores = parse_scores(&scores)?;
                let req = GradeSubjectiveRequest {
                    teacher_id,
                    scores,
                };
                let res = client
                    .post(format!("{}/submissions/{}/grade", base_url, submission_id))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
            SubmissionCommands::Appeal {
                submission_id,
                student_id,
            } => {
                let req = AppealRequest { student_id };
                let res = client
                    .post(format!("{}/submissions/{}/appeal", base_url, submission_id))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
            SubmissionCommands::GradeAppeal {
                submission_id,
                teacher_id,
                scores,
            } => {
                let scores = parse_scores(&scores)?;
                let req = GradeSubjectiveRequest {
                    teacher_id,
                    scores,
                };
                let res = client
                    .post(format!("{}/submissions/{}/grade-appeal", base_url, submission_id))
                    .json(&req)
                    .send()
                    .await?;
                print_response(res).await?;
            }
        },
        Commands::Stats { action } => match action {
            StatsCommands::Course { course_id } => {
                let res = client
                    .get(format!("{}/courses/{}/statistics", base_url, course_id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            StatsCommands::Distribution { course_id } => {
                let res = client
                    .get(format!("{}/courses/{}/distribution", base_url, course_id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            StatsCommands::Trends { course_id } => {
                let res = client
                    .get(format!("{}/courses/{}/trends", base_url, course_id))
                    .send()
                    .await?;
                print_response(res).await?;
            }
            StatsCommands::Workload => {
                let res = client.get(format!("{}/teachers/workload", base_url)).send().await?;
                print_response(res).await?;
            }
        },
    }

    Ok(())
}
