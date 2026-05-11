use clap::{Parser, Subcommand};
use performance_core::*;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(long, default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,

    CreateDepartment {
        id: String,
        name: String,
    },
    ListDepartments,
    GetDepartment {
        id: String,
    },

    CreateEmployee {
        id: String,
        name: String,
        #[arg(long)]
        manager_id: Option<String>,
        #[arg(long)]
        base_performance: f64,
        #[arg(long, value_delimiter = ',', num_args = 1..)]
        departments: Vec<String>,
    },
    ListEmployees,
    GetEmployee {
        id: String,
    },

    CreateCycle {
        id: String,
        name: String,
        #[arg(long)]
        start_date: String,
        #[arg(long)]
        end_date: String,
    },
    ListCycles,
    GetCycle {
        id: String,
    },

    SubmitSelf {
        employee_id: String,
        cycle_id: String,
        score: f64,
    },
    SubmitManager {
        employee_id: String,
        cycle_id: String,
        manager_id: String,
        score: f64,
    },
    SubmitPeer {
        reviewer_id: String,
        reviewee_id: String,
        cycle_id: String,
        score: f64,
    },

    Results {
        cycle_id: String,
    },
    DepartmentResults {
        cycle_id: String,
        dept_id: String,
    },
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateDepartmentReq {
    id: String,
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateEmployeeReq {
    id: String,
    name: String,
    manager_id: Option<String>,
    base_performance: f64,
    department_assignments: Vec<DepartmentAssignment>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateCycleReq {
    id: String,
    name: String,
    start_date: String,
    end_date: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SubmitSelfReq {
    employee_id: String,
    cycle_id: String,
    score: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct SubmitManagerReq {
    employee_id: String,
    cycle_id: String,
    manager_id: String,
    score: f64,
}

#[derive(Debug, Serialize, Deserialize)]
struct SubmitPeerReq {
    reviewer_id: String,
    reviewee_id: String,
    cycle_id: String,
    score: f64,
}

fn print_json(value: &Value) {
    println!("{}", serde_json::to_string_pretty(value).unwrap());
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/').to_string();

    match cli.command {
        Commands::Health => {
            let resp = client.get(format!("{}/api/health", base_url)).send().await?;
            println!("{}", resp.text().await?);
        }

        Commands::CreateDepartment { id, name } => {
            let req = CreateDepartmentReq { id, name };
            let resp = client
                .post(format!("{}/api/departments", base_url))
                .json(&req)
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::ListDepartments => {
            let resp = client.get(format!("{}/api/departments", base_url)).send().await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::GetDepartment { id } => {
            let resp = client
                .get(format!("{}/api/departments/{}", base_url, id))
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }

        Commands::CreateEmployee {
            id,
            name,
            manager_id,
            base_performance,
            departments,
        } => {
            let department_assignments = departments
                .into_iter()
                .map(|dept_id| DepartmentAssignment {
                    department_id: dept_id,
                    start_date: "1970-01-01".to_string(),
                    end_date: None,
                })
                .collect();

            let req = CreateEmployeeReq {
                id,
                name,
                manager_id,
                base_performance,
                department_assignments,
            };
            let resp = client
                .post(format!("{}/api/employees", base_url))
                .json(&req)
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::ListEmployees => {
            let resp = client.get(format!("{}/api/employees", base_url)).send().await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::GetEmployee { id } => {
            let resp = client
                .get(format!("{}/api/employees/{}", base_url, id))
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }

        Commands::CreateCycle {
            id,
            name,
            start_date,
            end_date,
        } => {
            let req = CreateCycleReq {
                id,
                name,
                start_date,
                end_date,
            };
            let resp = client
                .post(format!("{}/api/cycles", base_url))
                .json(&req)
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::ListCycles => {
            let resp = client.get(format!("{}/api/cycles", base_url)).send().await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::GetCycle { id } => {
            let resp = client
                .get(format!("{}/api/cycles/{}", base_url, id))
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }

        Commands::SubmitSelf {
            employee_id,
            cycle_id,
            score,
        } => {
            let req = SubmitSelfReq {
                employee_id,
                cycle_id,
                score,
            };
            let resp = client
                .post(format!("{}/api/self-assessment", base_url))
                .json(&req)
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::SubmitManager {
            employee_id,
            cycle_id,
            manager_id,
            score,
        } => {
            let req = SubmitManagerReq {
                employee_id,
                cycle_id,
                manager_id,
                score,
            };
            let resp = client
                .post(format!("{}/api/manager-review", base_url))
                .json(&req)
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::SubmitPeer {
            reviewer_id,
            reviewee_id,
            cycle_id,
            score,
        } => {
            let req = SubmitPeerReq {
                reviewer_id,
                reviewee_id,
                cycle_id,
                score,
            };
            let resp = client
                .post(format!("{}/api/peer-review", base_url))
                .json(&req)
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }

        Commands::Results { cycle_id } => {
            let resp = client
                .get(format!("{}/api/cycles/{}/results", base_url, cycle_id))
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
        Commands::DepartmentResults { cycle_id, dept_id } => {
            let resp = client
                .get(format!(
                    "{}/api/cycles/{}/departments/{}/results",
                    base_url, cycle_id, dept_id
                ))
                .send()
                .await?;
            let body: Value = resp.json().await?;
            print_json(&body);
        }
    }

    Ok(())
}
