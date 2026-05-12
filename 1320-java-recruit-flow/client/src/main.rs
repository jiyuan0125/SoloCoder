use chrono::{DateTime, NaiveDateTime, TimeZone, Utc};
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::json;
use uuid::Uuid;

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
    Health,

    Candidate {
        #[command(subcommand)]
        cmd: CandidateCommands,
    },

    Interview {
        #[command(subcommand)]
        cmd: InterviewCommands,
    },

    Offer {
        #[command(subcommand)]
        cmd: OfferCommands,
    },
}

#[derive(Subcommand, Debug)]
enum CandidateCommands {
    Create {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        email: String,
        #[arg(short, long)]
        phone: Option<String>,
        #[arg(short, long)]
        resume: Option<String>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    InitialScreening {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, num_args = 0..=1, default_missing_value = "true", action = clap::ArgAction::Set)]
        passed: bool,
    },
    SecondScreening {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, value_parser = ["recommend", "pending", "reject"])]
        decision: String,
    },
    CheckExpired,
    ConfirmExpired {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, num_args = 0..=1, default_missing_value = "true", action = clap::ArgAction::Set)]
        confirmed: bool,
    },
}

#[derive(Subcommand, Debug)]
enum InterviewCommands {
    Schedule {
        #[arg(short, long)]
        candidate_id: Uuid,
        #[arg(long = "interview-type", value_parser = ["tech", "hr", "director"])]
        interview_type: String,
        #[arg(short, long)]
        round: u32,
        #[arg(short, long)]
        interviewer: String,
        #[arg(long = "start-time")]
        start_time: String,
        #[arg(short, long)]
        duration: Option<u32>,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    SubmitResult {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, num_args = 0..=1, default_missing_value = "true", action = clap::ArgAction::Set)]
        passed: bool,
        #[arg(short, long)]
        notes: Option<String>,
    },
}

#[derive(Subcommand, Debug)]
enum OfferCommands {
    Create {
        #[arg(short, long)]
        candidate_id: Uuid,
        #[arg(short, long)]
        salary: u32,
        #[arg(short, long)]
        position: String,
        #[arg(short, long)]
        department: String,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Submit {
        #[arg(short, long)]
        id: Uuid,
    },
    Approve {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, value_parser = ["hr", "vp", "ceo"])]
        approver: String,
    },
    Reject {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, value_parser = ["hr", "vp", "ceo"])]
        approver: String,
    },
    Revise {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        salary: u32,
        #[arg(short, long)]
        position: String,
        #[arg(short, long)]
        department: String,
    },
}

#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: String,
}

fn parse_datetime(s: &str) -> anyhow::Result<DateTime<Utc>> {
    if let Ok(dt) = NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S") {
        Ok(Utc.from_utc_datetime(&dt))
    } else if let Ok(dt) = DateTime::parse_from_rfc3339(s) {
        Ok(dt.with_timezone(&Utc))
    } else {
        Err(anyhow::anyhow!(
            "Invalid datetime format. Use 'YYYY-MM-DD HH:MM:SS' or RFC3339"
        ))
    }
}

fn print_json<T: Serialize>(data: &T) {
    println!("{}", serde_json::to_string_pretty(data).unwrap());
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();

    match args.command {
        Commands::Health => {
            let resp = client.get(format!("{}/health", base_url)).send().await?;
            let body: serde_json::Value = resp.json().await?;
            print_json(&body);
        }

        Commands::Candidate { cmd } => match cmd {
            CandidateCommands::Create {
                name,
                email,
                phone,
                resume,
            } => {
                let body = json!({
                    "name": name,
                    "email": email,
                    "phone": phone,
                    "resume": resume
                });
                let resp = client
                    .post(format!("{}/candidates", base_url))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            CandidateCommands::List => {
                let resp = client.get(format!("{}/candidates", base_url)).send().await?;
                handle_response(resp).await?;
            }
            CandidateCommands::Get { id } => {
                let resp = client
                    .get(format!("{}/candidates/{}", base_url, id))
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            CandidateCommands::InitialScreening { id, passed } => {
                let body = json!({ "passed": passed });
                let resp = client
                    .post(format!("{}/candidates/{}/initial-screening", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            CandidateCommands::SecondScreening { id, decision } => {
                let decision_val = match decision.as_str() {
                    "recommend" => "RecommendInterview",
                    "pending" => "Pending",
                    "reject" => "NotSuitable",
                    _ => unreachable!(),
                };
                let body = json!({ "decision": decision_val });
                let resp = client
                    .post(format!("{}/candidates/{}/second-screening", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            CandidateCommands::CheckExpired => {
                let resp = client
                    .post(format!("{}/candidates/check-expired", base_url))
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            CandidateCommands::ConfirmExpired { id, confirmed } => {
                let body = json!({ "confirmed": confirmed });
                let resp = client
                    .post(format!("{}/candidates/{}/confirm-expired", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
        },

        Commands::Interview { cmd } => match cmd {
            InterviewCommands::Schedule {
                candidate_id,
                interview_type,
                round,
                interviewer,
                start_time,
                duration,
            } => {
                let interview_type_val = match interview_type.as_str() {
                    "tech" => "Technical",
                    "hr" => "HR",
                    "director" => "Director",
                    _ => unreachable!(),
                };
                let start = parse_datetime(&start_time)?;
                let body = json!({
                    "interview_type": interview_type_val,
                    "round": round,
                    "interviewer": interviewer,
                    "start_time": start,
                    "duration_minutes": duration
                });
                let resp = client
                    .post(format!("{}/candidates/{}/interviews", base_url, candidate_id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            InterviewCommands::List => {
                let resp = client.get(format!("{}/interviews", base_url)).send().await?;
                handle_response(resp).await?;
            }
            InterviewCommands::Get { id } => {
                let resp = client
                    .get(format!("{}/interviews/{}", base_url, id))
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            InterviewCommands::SubmitResult { id, passed, notes } => {
                let body = json!({
                    "passed": passed,
                    "notes": notes
                });
                let resp = client
                    .post(format!("{}/interviews/{}/result", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
        },

        Commands::Offer { cmd } => match cmd {
            OfferCommands::Create {
                candidate_id,
                salary,
                position,
                department,
            } => {
                let body = json!({
                    "salary": salary,
                    "position": position,
                    "department": department
                });
                let resp = client
                    .post(format!("{}/candidates/{}/offers", base_url, candidate_id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            OfferCommands::List => {
                let resp = client.get(format!("{}/offers", base_url)).send().await?;
                handle_response(resp).await?;
            }
            OfferCommands::Get { id } => {
                let resp = client
                    .get(format!("{}/offers/{}", base_url, id))
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            OfferCommands::Submit { id } => {
                let resp = client
                    .post(format!("{}/offers/{}/submit", base_url, id))
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            OfferCommands::Approve { id, approver } => {
                let approver_val = match approver.as_str() {
                    "hr" => "HRDirector",
                    "vp" => "VP",
                    "ceo" => "CEO",
                    _ => unreachable!(),
                };
                let body = json!({ "approver_level": approver_val });
                let resp = client
                    .post(format!("{}/offers/{}/approve", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            OfferCommands::Reject { id, approver } => {
                let approver_val = match approver.as_str() {
                    "hr" => "HRDirector",
                    "vp" => "VP",
                    "ceo" => "CEO",
                    _ => unreachable!(),
                };
                let body = json!({ "approver_level": approver_val });
                let resp = client
                    .post(format!("{}/offers/{}/reject", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
            OfferCommands::Revise {
                id,
                salary,
                position,
                department,
            } => {
                let body = json!({
                    "salary": salary,
                    "position": position,
                    "department": department
                });
                let resp = client
                    .post(format!("{}/offers/{}/revise", base_url, id))
                    .json(&body)
                    .send()
                    .await?;
                handle_response(resp).await?;
            }
        },
    }

    Ok(())
}

async fn handle_response(resp: reqwest::Response) -> anyhow::Result<()> {
    let status = resp.status();
    if status.is_success() {
        let body: serde_json::Value = resp.json().await?;
        print_json(&body);
    } else {
        let body: ErrorResponse = resp.json().await?;
        eprintln!("Error ({}): {}", status, body.error);
        std::process::exit(1);
    }
    Ok(())
}
