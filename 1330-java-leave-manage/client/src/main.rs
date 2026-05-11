use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::Serialize;

#[derive(Parser, Debug)]
#[command(name = "leave_client")]
#[command(about = "假期管理系统命令行客户端")]
struct Args {
    #[arg(long, env = "LEAVE_SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand, Debug)]
enum Command {
    #[command(about = "健康检查")]
    Health,

    #[command(about = "添加员工")]
    AddEmployee {
        #[arg(short, long)]
        name: String,
        #[arg(short, long)]
        hire_date: String,
    },

    #[command(about = "列出所有员工")]
    ListEmployees,

    #[command(about = "查看员工信息")]
    GetEmployee {
        #[arg(short, long)]
        id: String,
    },

    #[command(about = "查看员工假期余额")]
    Balance {
        #[arg(short, long)]
        employee_id: String,
    },

    #[command(about = "申请假期")]
    Apply {
        #[arg(short, long)]
        employee_id: String,
        #[arg(short, long, help = "annual/personal/sick/compensatory")]
        leave_type: String,
        #[arg(short, long)]
        start_date: String,
        #[arg(short, long)]
        end_date: String,
        #[arg(short, long)]
        proof: Option<String>,
    },

    #[command(about = "撤销假期（仅未来的假期）")]
    Cancel {
        #[arg(short, long)]
        request_id: String,
    },

    #[command(about = "查看员工的假期申请记录")]
    ListEmployeeLeaves {
        #[arg(short, long)]
        employee_id: String,
    },

    #[command(about = "查看所有假期申请")]
    ListAllLeaves,

    #[command(about = "查看单个假期申请详情")]
    GetLeave {
        #[arg(short, long)]
        id: String,
    },

    #[command(about = "为员工添加假期余额（事假/病假/调休）")]
    AddBalance {
        #[arg(short, long)]
        employee_id: String,
        #[arg(short, long)]
        leave_type: String,
        #[arg(short, long)]
        days: u32,
    },

    #[command(about = "初始化员工某年度的年假")]
    InitAnnualLeave {
        #[arg(short, long)]
        employee_id: String,
        #[arg(short, long)]
        year: i32,
    },

    #[command(about = "执行年底年假结转（最多结转3天）")]
    Carryover {
        #[arg(short, long)]
        employee_id: String,
        #[arg(short, long)]
        year: i32,
    },
}

#[derive(Debug, Serialize)]
struct CreateEmployeeReq {
    name: String,
    hire_date: String,
}

#[derive(Debug, Serialize)]
struct ApplyLeaveReq {
    employee_id: String,
    leave_type: String,
    start_date: String,
    end_date: String,
    sick_leave_proof: Option<String>,
}

#[derive(Debug, Serialize)]
struct AddBalanceReq {
    employee_id: String,
    leave_type: String,
    days: u32,
}

#[derive(Debug, Serialize)]
struct InitAnnualReq {
    employee_id: String,
    year: i32,
}

#[derive(Debug, Serialize)]
struct CarryoverReq {
    employee_id: String,
    year: i32,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server_url.trim_end_matches('/').to_string();

    match args.command {
        Command::Health => {
            let resp = client.get(format!("{}/health", base_url)).send().await?;
            print_response(resp).await;
        }

        Command::AddEmployee { name, hire_date } => {
            let req = CreateEmployeeReq { name, hire_date };
            let resp = client
                .post(format!("{}/employees", base_url))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::ListEmployees => {
            let resp = client.get(format!("{}/employees", base_url)).send().await?;
            print_response(resp).await;
        }

        Command::GetEmployee { id } => {
            let resp = client.get(format!("{}/employees/{}", base_url, id)).send().await?;
            print_response(resp).await;
        }

        Command::Balance { employee_id } => {
            let resp = client
                .get(format!("{}/employees/{}/balance", base_url, employee_id))
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::Apply {
            employee_id,
            leave_type,
            start_date,
            end_date,
            proof,
        } => {
            let req = ApplyLeaveReq {
                employee_id,
                leave_type,
                start_date,
                end_date,
                sick_leave_proof: proof,
            };
            let resp = client
                .post(format!("{}/leaves", base_url))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::Cancel { request_id } => {
            let resp = client
                .post(format!("{}/leaves/{}/cancel", base_url, request_id))
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::ListEmployeeLeaves { employee_id } => {
            let resp = client
                .get(format!("{}/employees/{}/leaves", base_url, employee_id))
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::ListAllLeaves => {
            let resp = client.get(format!("{}/leaves", base_url)).send().await?;
            print_response(resp).await;
        }

        Command::GetLeave { id } => {
            let resp = client.get(format!("{}/leaves/{}", base_url, id)).send().await?;
            print_response(resp).await;
        }

        Command::AddBalance {
            employee_id,
            leave_type,
            days,
        } => {
            let req = AddBalanceReq {
                employee_id,
                leave_type,
                days,
            };
            let resp = client
                .post(format!("{}/balance/add", base_url))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::InitAnnualLeave { employee_id, year } => {
            let req = InitAnnualReq { employee_id, year };
            let resp = client
                .post(format!("{}/annual-leave/initialize", base_url))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }

        Command::Carryover { employee_id, year } => {
            let req = CarryoverReq { employee_id, year };
            let resp = client
                .post(format!("{}/annual-leave/carryover", base_url))
                .json(&req)
                .send()
                .await?;
            print_response(resp).await;
        }
    }

    Ok(())
}

async fn print_response(resp: reqwest::Response) {
    let status = resp.status();
    let body = resp.text().await.unwrap_or_default();

    println!("状态: {}", status);

    if let Ok(json) = serde_json::from_str::<serde_json::Value>(&body) {
        println!("{}", serde_json::to_string_pretty(&json).unwrap_or_else(|_| body));
    } else {
        println!("{}", body);
    }
}
