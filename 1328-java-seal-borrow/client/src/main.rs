use chrono::{DateTime, Utc};
use clap::{Parser, Subcommand};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use seal_borrow_core::{
    SealStatus, SealType,
    CreateSealRequest, CreateBorrowRequest, ApproveRequest, RejectRequest,
    RenewRequest, ReturnRequest, UpdateSealStatusRequest, CreateEmployeeRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about = "印章借用管理系统命令行客户端", long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://localhost:8080")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Employee {
        #[command(subcommand)]
        command: EmployeeCommands,
    },
    Seal {
        #[command(subcommand)]
        command: SealCommands,
    },
    Borrow {
        #[command(subcommand)]
        command: BorrowCommands,
    },
    Reminder {
        #[command(subcommand)]
        command: ReminderCommands,
    },
}

#[derive(Subcommand, Debug)]
enum EmployeeCommands {
    Create {
        name: String,
        email: String,
    },
    List,
    Get {
        id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum SealCommands {
    Create {
        name: String,
        #[arg(value_enum)]
        seal_type: SealTypeArg,
        custodian_id: Uuid,
    },
    List,
    Get {
        id: Uuid,
    },
    UpdateStatus {
        seal_id: Uuid,
        #[arg(value_enum)]
        status: SealStatusArg,
    },
}

#[derive(Subcommand, Debug)]
enum BorrowCommands {
    Create {
        seal_id: Uuid,
        borrower_id: Uuid,
        reason: String,
        expected_return_date: String,
    },
    List,
    Get {
        id: Uuid,
    },
    ListBySeal {
        seal_id: Uuid,
    },
    ListByBorrower {
        borrower_id: Uuid,
    },
    Approve {
        request_id: Uuid,
        approver_id: Uuid,
    },
    Reject {
        request_id: Uuid,
        approver_id: Uuid,
        reason: String,
    },
    Renew {
        request_id: Uuid,
        borrower_id: Uuid,
        reason: String,
        new_expected_return_date: String,
    },
    ProcessRenewal {
        renewal_request_id: Uuid,
        #[arg(long)]
        approved: bool,
        approver_id: Uuid,
        #[arg(long)]
        reject_reason: Option<String>,
    },
    Return {
        request_id: Uuid,
    },
}

#[derive(Subcommand, Debug)]
enum ReminderCommands {
    List,
    ListByRequest {
        request_id: Uuid,
    },
}

#[derive(clap::ValueEnum, Clone, Debug, Copy)]
enum SealTypeArg {
    Official,
    Finance,
    Contract,
    Legal,
    Other,
}

#[derive(clap::ValueEnum, Clone, Debug, Copy)]
enum SealStatusArg {
    InStock,
    Borrowed,
    Maintenance,
}

impl From<SealTypeArg> for SealType {
    fn from(arg: SealTypeArg) -> Self {
        match arg {
            SealTypeArg::Official => SealType::Official,
            SealTypeArg::Finance => SealType::Finance,
            SealTypeArg::Contract => SealType::Contract,
            SealTypeArg::Legal => SealType::Legal,
            SealTypeArg::Other => SealType::Other,
        }
    }
}

impl From<SealStatusArg> for SealStatus {
    fn from(arg: SealStatusArg) -> Self {
        match arg {
            SealStatusArg::InStock => SealStatus::InStock,
            SealStatusArg::Borrowed => SealStatus::Borrowed,
            SealStatusArg::Maintenance => SealStatus::Maintenance,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct ProcessRenewalRequest {
    renewal_request_id: Uuid,
    approved: bool,
    approver_id: Uuid,
    reject_reason: Option<String>,
}

struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        Self {
            client: Client::new(),
            base_url,
        }
    }

    async fn post<T: Serialize, R: for<'de> Deserialize<'de>>(&self, path: &str, body: &T) -> Result<R, String> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.post(&url)
            .json(body)
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "未知错误".to_string());
            return Err(format!("API错误: {}", error_text));
        }
        
        response.json().await.map_err(|e| format!("解析响应失败: {}", e))
    }

    async fn put<T: Serialize, R: for<'de> Deserialize<'de>>(&self, path: &str, body: &T) -> Result<R, String> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.put(&url)
            .json(body)
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "未知错误".to_string());
            return Err(format!("API错误: {}", error_text));
        }
        
        response.json().await.map_err(|e| format!("解析响应失败: {}", e))
    }

    async fn get<R: for<'de> Deserialize<'de>>(&self, path: &str) -> Result<R, String> {
        let url = format!("{}{}", self.base_url, path);
        let response = self.client.get(&url)
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;
        
        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "未知错误".to_string());
            return Err(format!("API错误: {}", error_text));
        }
        
        response.json().await.map_err(|e| format!("解析响应失败: {}", e))
    }
}

fn parse_datetime(s: &str) -> Result<DateTime<Utc>, String> {
    DateTime::parse_from_rfc3339(s)
        .map(|dt| dt.with_timezone(&Utc))
        .or_else(|_| {
            DateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S %z")
                .map(|dt| dt.with_timezone(&Utc))
        })
        .map_err(|e| format!("日期时间格式错误: {}. 请使用 RFC3339 格式，例如: 2024-01-15T10:00:00Z", e))
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let api_client = ApiClient::new(cli.server_url);

    let result = match cli.command {
        Commands::Employee { command } => handle_employee_commands(&api_client, command).await,
        Commands::Seal { command } => handle_seal_commands(&api_client, command).await,
        Commands::Borrow { command } => handle_borrow_commands(&api_client, command).await,
        Commands::Reminder { command } => handle_reminder_commands(&api_client, command).await,
    };

    match result {
        Ok(output) => println!("{}", output),
        Err(e) => eprintln!("错误: {}", e),
    }
}

async fn handle_employee_commands(api: &ApiClient, command: EmployeeCommands) -> Result<String, String> {
    match command {
        EmployeeCommands::Create { name, email } => {
            let req = CreateEmployeeRequest { name, email };
            let employee: serde_json::Value = api.post("/employees", &req).await?;
            Ok(format!("创建员工成功:\n{}", serde_json::to_string_pretty(&employee).unwrap()))
        }
        EmployeeCommands::List => {
            let employees: serde_json::Value = api.get("/employees").await?;
            Ok(format!("员工列表:\n{}", serde_json::to_string_pretty(&employees).unwrap()))
        }
        EmployeeCommands::Get { id } => {
            let employee: serde_json::Value = api.get(&format!("/employees/{}", id)).await?;
            Ok(format!("员工信息:\n{}", serde_json::to_string_pretty(&employee).unwrap()))
        }
    }
}

async fn handle_seal_commands(api: &ApiClient, command: SealCommands) -> Result<String, String> {
    match command {
        SealCommands::Create { name, seal_type, custodian_id } => {
            let req = CreateSealRequest {
                name,
                seal_type: seal_type.into(),
                custodian_id,
            };
            let seal: serde_json::Value = api.post("/seals", &req).await?;
            Ok(format!("创建印章成功:\n{}", serde_json::to_string_pretty(&seal).unwrap()))
        }
        SealCommands::List => {
            let seals: serde_json::Value = api.get("/seals").await?;
            Ok(format!("印章列表:\n{}", serde_json::to_string_pretty(&seals).unwrap()))
        }
        SealCommands::Get { id } => {
            let seal: serde_json::Value = api.get(&format!("/seals/{}", id)).await?;
            Ok(format!("印章信息:\n{}", serde_json::to_string_pretty(&seal).unwrap()))
        }
        SealCommands::UpdateStatus { seal_id, status } => {
            let req = UpdateSealStatusRequest {
                seal_id,
                status: status.into(),
            };
            let seal: serde_json::Value = api.put("/seals/status", &req).await?;
            Ok(format!("更新印章状态成功:\n{}", serde_json::to_string_pretty(&seal).unwrap()))
        }
    }
}

async fn handle_borrow_commands(api: &ApiClient, command: BorrowCommands) -> Result<String, String> {
    match command {
        BorrowCommands::Create { seal_id, borrower_id, reason, expected_return_date } => {
            let expected_return = parse_datetime(&expected_return_date)?;
            let req = CreateBorrowRequest {
                seal_id,
                borrower_id,
                reason,
                expected_return_date: expected_return,
            };
            let request: serde_json::Value = api.post("/borrow-requests", &req).await?;
            Ok(format!("创建借用申请成功:\n{}", serde_json::to_string_pretty(&request).unwrap()))
        }
        BorrowCommands::List => {
            let requests: serde_json::Value = api.get("/borrow-requests").await?;
            Ok(format!("借用申请列表:\n{}", serde_json::to_string_pretty(&requests).unwrap()))
        }
        BorrowCommands::Get { id } => {
            let request: serde_json::Value = api.get(&format!("/borrow-requests/{}", id)).await?;
            Ok(format!("借用申请信息:\n{}", serde_json::to_string_pretty(&request).unwrap()))
        }
        BorrowCommands::ListBySeal { seal_id } => {
            let requests: serde_json::Value = api.get(&format!("/borrow-requests/seal/{}", seal_id)).await?;
            Ok(format!("印章借用记录:\n{}", serde_json::to_string_pretty(&requests).unwrap()))
        }
        BorrowCommands::ListByBorrower { borrower_id } => {
            let requests: serde_json::Value = api.get(&format!("/borrow-requests/borrower/{}", borrower_id)).await?;
            Ok(format!("借用人借用记录:\n{}", serde_json::to_string_pretty(&requests).unwrap()))
        }
        BorrowCommands::Approve { request_id, approver_id } => {
            let req = ApproveRequest { request_id, approver_id };
            let request: serde_json::Value = api.post("/borrow-requests/approve", &req).await?;
            Ok(format!("审批通过:\n{}", serde_json::to_string_pretty(&request).unwrap()))
        }
        BorrowCommands::Reject { request_id, approver_id, reason } => {
            let req = RejectRequest { request_id, approver_id, reason };
            let request: serde_json::Value = api.post("/borrow-requests/reject", &req).await?;
            Ok(format!("审批拒绝:\n{}", serde_json::to_string_pretty(&request).unwrap()))
        }
        BorrowCommands::Renew { request_id, borrower_id, reason, new_expected_return_date } => {
            let new_return = parse_datetime(&new_expected_return_date)?;
            let req = RenewRequest {
                request_id,
                borrower_id,
                reason,
                new_expected_return_date: new_return,
            };
            let request: serde_json::Value = api.post("/borrow-requests/renew", &req).await?;
            Ok(format!("续借申请提交成功:\n{}", serde_json::to_string_pretty(&request).unwrap()))
        }
        BorrowCommands::ProcessRenewal { renewal_request_id, approved, approver_id, reject_reason } => {
            let req = ProcessRenewalRequest {
                renewal_request_id,
                approved,
                approver_id,
                reject_reason,
            };
            let request: serde_json::Value = api.post("/borrow-requests/process-renewal", &req).await?;
            let action = if approved { "通过" } else { "拒绝" };
            Ok(format!("续借审批{}:\n{}", action, serde_json::to_string_pretty(&request).unwrap()))
        }
        BorrowCommands::Return { request_id } => {
            let req = ReturnRequest { request_id };
            let request: serde_json::Value = api.post("/borrow-requests/return", &req).await?;
            Ok(format!("归还成功:\n{}", serde_json::to_string_pretty(&request).unwrap()))
        }
    }
}

async fn handle_reminder_commands(api: &ApiClient, command: ReminderCommands) -> Result<String, String> {
    match command {
        ReminderCommands::List => {
            let reminders: serde_json::Value = api.get("/reminders").await?;
            Ok(format!("催还记录列表:\n{}", serde_json::to_string_pretty(&reminders).unwrap()))
        }
        ReminderCommands::ListByRequest { request_id } => {
            let reminders: serde_json::Value = api.get(&format!("/reminders/request/{}", request_id)).await?;
            Ok(format!("借用申请催还记录:\n{}", serde_json::to_string_pretty(&reminders).unwrap()))
        }
    }
}
