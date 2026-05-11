use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::de::DeserializeOwned;
use serde::Serialize;
use thesis_core::*;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "THESIS_SERVER", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Status,
    AdvancePhase,
    CreateStudent {
        name: String,
        #[arg(long)]
        student_no: String,
    },
    ListStudents,
    GetStudent {
        id: StudentId,
    },
    SubmitPrefs {
        student_id: StudentId,
        #[arg(num_args = 1..=3)]
        advisors: Vec<AdvisorId>,
    },
    ModifyPrefs {
        student_id: StudentId,
        #[arg(num_args = 1..=3)]
        advisors: Vec<AdvisorId>,
    },
    CreateAdvisor {
        name: String,
        #[arg(long)]
        capacity: u32,
    },
    ListAdvisors,
    GetAdvisor {
        id: AdvisorId,
    },
    Pool {
        advisor_id: AdvisorId,
    },
    Select {
        advisor_id: AdvisorId,
        student_id: StudentId,
        #[arg(long)]
        accept: bool,
    },
    CreateAppeal {
        student_id: StudentId,
        reason: String,
    },
    ListAppeals,
    Assign {
        student_id: StudentId,
        advisor_id: AdvisorId,
        reason: String,
        #[arg(long, default_value = "admin")]
        operator: String,
    },
    Changes {
        #[arg(long)]
        student_id: Option<StudentId>,
    },
    Results,
}

fn prefs_from_vec(v: Vec<AdvisorId>) -> [Option<AdvisorId>; 3] {
    let mut arr = [None, None, None];
    for (i, id) in v.into_iter().enumerate().take(3) {
        arr[i] = Some(id);
    }
    arr
}

struct ApiClient {
    base: String,
    client: Client,
}

impl ApiClient {
    fn new(base: String) -> Self {
        Self {
            base,
            client: Client::new(),
        }
    }

    async fn get<T: DeserializeOwned>(&self, path: &str) -> anyhow::Result<T> {
        let url = format!("{}{}", self.base, path);
        let resp = self.client.get(&url).send().await?;
        let api_resp: ApiResponse<T> = resp.json().await?;
        if api_resp.success {
            Ok(api_resp.data.unwrap())
        } else {
            Err(anyhow::anyhow!(api_resp.error.unwrap_or_default()))
        }
    }

    async fn post<B: Serialize, T: DeserializeOwned>(&self, path: &str, body: B) -> anyhow::Result<T> {
        let url = format!("{}{}", self.base, path);
        let resp = self.client.post(&url).json(&body).send().await?;
        let api_resp: ApiResponse<T> = resp.json().await?;
        if api_resp.success {
            Ok(api_resp.data.unwrap())
        } else {
            Err(anyhow::anyhow!(api_resp.error.unwrap_or_default()))
        }
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();
    let client = ApiClient::new(args.server);

    match args.command {
        Commands::Status => {
            let status: SystemStatus = client.get("/status").await?;
            println!("{:#?}", status);
        }
        Commands::AdvancePhase => {
            let phase: SystemPhase = client.post::<_, SystemPhase>("/phase/advance", ()).await?;
            println!("Advanced to phase: {:?}", phase);
        }
        Commands::CreateStudent { name, student_no } => {
            let req = CreateStudentRequest { name, student_no };
            let student: Student = client.post("/students", req).await?;
            println!("Created student: {:#?}", student);
        }
        Commands::ListStudents => {
            let students: Vec<Student> = client.get("/students").await?;
            for s in students {
                println!(
                    "{} ({}) - {:?} | Prefs: {:?}",
                    s.name, s.student_no, s.status, s.preferences
                );
            }
        }
        Commands::GetStudent { id } => {
            let student: Student = client.get(&format!("/students/{}", id)).await?;
            println!("{:#?}", student);
        }
        Commands::SubmitPrefs {
            student_id,
            advisors,
        } => {
            let req = SubmitPreferencesRequest {
                preferences: prefs_from_vec(advisors),
            };
            let student: Student = client
                .post(&format!("/students/{}/submit", student_id), req)
                .await?;
            println!("Submitted preferences for: {}", student.name);
        }
        Commands::ModifyPrefs {
            student_id,
            advisors,
        } => {
            let req = ModifyPreferencesRequest {
                preferences: prefs_from_vec(advisors),
            };
            let student: Student = client
                .post(&format!("/students/{}/modify", student_id), req)
                .await?;
            println!("Modified preferences for: {}", student.name);
        }
        Commands::CreateAdvisor { name, capacity } => {
            let req = CreateAdvisorRequest { name, capacity };
            let advisor: Advisor = client.post("/advisors", req).await?;
            println!("Created advisor: {:#?}", advisor);
        }
        Commands::ListAdvisors => {
            let advisors: Vec<Advisor> = client.get("/advisors").await?;
            for a in advisors {
                println!(
                    "{} (ID: {}) | Capacity: {}/{}",
                    a.name,
                    a.id,
                    a.accepted.len(),
                    a.capacity
                );
            }
        }
        Commands::GetAdvisor { id } => {
            let advisor: Advisor = client.get(&format!("/advisors/{}", id)).await?;
            println!("{:#?}", advisor);
        }
        Commands::Pool { advisor_id } => {
            let students: Vec<Student> =
                client.get(&format!("/advisors/{}/pool", advisor_id)).await?;
            if students.is_empty() {
                println!("No eligible students in pool.");
            } else {
                for s in students {
                    println!("{} ({}) - {:?}", s.name, s.student_no, s.status);
                }
            }
        }
        Commands::Select {
            advisor_id,
            student_id,
            accept,
        } => {
            let req = AdvisorSelectRequest {
                student_id,
                accept,
            };
            let result: String = client
                .post(&format!("/advisors/{}/select", advisor_id), req)
                .await?;
            println!("{}", result);
        }
        Commands::CreateAppeal { student_id, reason } => {
            let req = CreateAppealRequest { student_id, reason };
            let appeal: AppealRecord = client.post("/appeals", req).await?;
            println!("Created appeal: {:#?}", appeal);
        }
        Commands::ListAppeals => {
            let appeals: Vec<AppealRecord> = client.get("/appeals").await?;
            for a in appeals {
                println!(
                    "Student {}: {} (resolved: {})",
                    a.student_id, a.reason, a.resolved
                );
            }
        }
        Commands::Assign {
            student_id,
            advisor_id,
            reason,
            operator,
        } => {
            let req = ManualAssignRequest {
                student_id,
                advisor_id,
                reason,
                operator,
            };
            let result: String = client.post("/assign", req).await?;
            println!("{}", result);
        }
        Commands::Changes { student_id } => {
            let path = match student_id {
                Some(id) => format!("/changes/{}", id),
                None => "/changes".into(),
            };
            let logs: Vec<ChangeLog> = client.get(&path).await?;
            for log in logs {
                println!(
                    "[{}] {} -> {:?} (by: {})",
                    log.timestamp, log.student_id, log.change_type, log.operator
                );
                if let Some(r) = log.reason {
                    println!("  Reason: {}", r);
                }
            }
        }
        Commands::Results => {
            let results: Vec<MatchResult> = client.get("/results").await?;
            for r in results {
                let advisor_name = r.advisor.map(|a| a.name).unwrap_or_else(|| "Unassigned".into());
                println!(
                    "{} ({}) -> {} | Status: {:?}",
                    r.student.name, r.student.student_no, advisor_name, r.student.status
                );
            }
        }
    }

    Ok(())
}
