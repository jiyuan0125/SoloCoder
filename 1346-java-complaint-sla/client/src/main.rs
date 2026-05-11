use chrono::{DateTime, Local};
use clap::{Parser, Subcommand};
use colored::*;
use reqwest::blocking::Client;
use uuid::Uuid;

use complaint_core::{
    Complaint, Handler, CreateComplaintRequest, RespondToComplaintRequest,
    ProposeSolutionRequest, CustomerFeedbackRequest, FollowUpFeedbackRequest, Severity, Status,
};

#[derive(Parser, Debug)]
#[command(author, version, about = "客诉SLA时效管控系统 - 命令行客户端", long_about = None)]
struct Cli {
    #[arg(long, env = "COMPLAINT_SERVER_URL", default_value = "http://127.0.0.1:8100")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Health,
    Handlers,
    List,
    Get { id: Uuid },
    Create {
        #[arg(long)]
        title: String,
        #[arg(long)]
        description: String,
        #[arg(long)]
        customer_name: String,
        #[arg(long)]
        customer_contact: String,
        #[arg(long, value_parser = parse_severity)]
        severity: Severity,
        #[arg(long)]
        handler_id: Option<String>,
    },
    Respond {
        id: Uuid,
        #[arg(long)]
        handler_id: String,
        #[arg(long)]
        response: String,
    },
    Solution {
        id: Uuid,
        #[arg(long)]
        handler_id: String,
        #[arg(long)]
        solution: String,
    },
    Feedback {
        id: Uuid,
        #[arg(long)]
        accepted: bool,
        #[arg(long)]
        comments: Option<String>,
    },
    FollowUp {
        id: Uuid,
        #[arg(long)]
        satisfied: bool,
        #[arg(long)]
        comments: Option<String>,
    },
}

fn parse_severity(s: &str) -> Result<Severity, String> {
    match s.to_lowercase().as_str() {
        "general" | "一般" | "1" => Ok(Severity::General),
        "serious" | "严重" | "2" => Ok(Severity::Serious),
        "urgent" | "紧急" | "3" => Ok(Severity::Urgent),
        _ => Err(format!("无效的严重程度: {}，有效值: general/serious/urgent 或 一般/严重/紧急", s)),
    }
}

struct ApiClient {
    base_url: String,
    client: Client,
}

impl ApiClient {
    fn new(base_url: String) -> Self {
        Self {
            base_url,
            client: Client::new(),
        }
    }

    fn url(&self, path: &str) -> String {
        format!("{}{}", self.base_url, path)
    }

    fn health(&self) -> Result<String, reqwest::Error> {
        self.client.get(self.url("/health")).send()?.text()
    }

    fn list_handlers(&self) -> Result<Vec<Handler>, reqwest::Error> {
        self.client.get(self.url("/handlers")).send()?.json()
    }

    fn list_complaints(&self) -> Result<Vec<Complaint>, reqwest::Error> {
        self.client.get(self.url("/complaints")).send()?.json()
    }

    fn get_complaint(&self, id: Uuid) -> Result<Complaint, reqwest::Error> {
        self.client.get(self.url(&format!("/complaints/{}", id))).send()?.json()
    }

    fn create_complaint(&self, req: CreateComplaintRequest) -> Result<Complaint, reqwest::Error> {
        self.client
            .post(self.url("/complaints"))
            .json(&req)
            .send()?
            .json()
    }

    fn respond(&self, id: Uuid, req: RespondToComplaintRequest) -> Result<Complaint, reqwest::Error> {
        self.client
            .post(self.url(&format!("/complaints/{}/respond", id)))
            .json(&req)
            .send()?
            .json()
    }

    fn propose_solution(&self, id: Uuid, req: ProposeSolutionRequest) -> Result<Complaint, reqwest::Error> {
        self.client
            .post(self.url(&format!("/complaints/{}/solution", id)))
            .json(&req)
            .send()?
            .json()
    }

    fn customer_feedback(&self, id: Uuid, req: CustomerFeedbackRequest) -> Result<Complaint, reqwest::Error> {
        self.client
            .post(self.url(&format!("/complaints/{}/feedback", id)))
            .json(&req)
            .send()?
            .json()
    }

    fn follow_up(&self, id: Uuid, req: FollowUpFeedbackRequest) -> Result<Complaint, reqwest::Error> {
        self.client
            .post(self.url(&format!("/complaints/{}/followup", id)))
            .json(&req)
            .send()?
            .json()
    }
}

fn format_datetime(dt: DateTime<chrono::Utc>) -> String {
    let local: DateTime<Local> = dt.with_timezone(&Local);
    local.format("%Y-%m-%d %H:%M:%S").to_string()
}

fn severity_color(s: Severity) -> ColoredString {
    match s {
        Severity::General => s.as_str().green(),
        Severity::Serious => s.as_str().yellow(),
        Severity::Urgent => s.as_str().red().bold(),
    }
}

fn status_color(s: Status) -> ColoredString {
    match s {
        Status::Pending => s.as_str().yellow(),
        Status::Processing => s.as_str().cyan(),
        Status::AwaitingFeedback => s.as_str().magenta(),
        Status::Reopened => s.as_str().bright_yellow(),
        Status::Escalated => s.as_str().red().bold(),
        Status::Closed => s.as_str().green(),
    }
}

fn print_complaint_short(c: &Complaint) {
    let overdue = if c.is_overdue(chrono::Utc::now()) {
        " [超时]".red().bold()
    } else {
        "".normal()
    };
    println!(
        "{} {} [{}] {} - {} (截止: {}){}",
        c.id.to_string().dimmed(),
        severity_color(c.severity),
        status_color(c.status),
        c.title.bold(),
        c.customer_name,
        format_datetime(c.deadline),
        overdue,
    );
}

fn print_complaint_detail(c: &Complaint) {
    println!("{}", "=".repeat(80));
    println!("投诉ID: {}", c.id);
    println!("标题: {}", c.title.bold());
    println!("客户: {} ({})", c.customer_name, c.customer_contact);
    println!("严重程度: {}", severity_color(c.severity));
    println!(
        "原始等级: {} | 状态: {}",
        c.original_severity.as_str(),
        status_color(c.status)
    );
    println!(
        "处理人: {} | 重开次数: {}",
        c.handler_id.as_deref().unwrap_or("未分配"),
        c.reopen_count
    );
    println!("创建时间: {}", format_datetime(c.created_at));
    println!("截止时间: {}", format_datetime(c.deadline));
    if let Some(rt) = c.response_time {
        println!("首次响应: {}", format_datetime(rt));
    }
    if let Some(ct) = c.closed_at {
        println!("结案时间: {}", format_datetime(ct));
    }
    println!();
    println!("描述: {}", c.description);
    println!();
    
    if !c.escalations.is_empty() {
        println!("{}", "升级记录:".bold());
        for e in &c.escalations {
            let from = e.from_severity.map(|s| s.as_str()).unwrap_or("-");
            let to = e.to_severity.map(|s| s.as_str()).unwrap_or("上级主管");
            println!("  [{}] {} -> {} - {}", format_datetime(e.escalated_at), from, to, e.reason);
        }
        println!();
    }

    if !c.follow_ups.is_empty() {
        println!("{}", "回访记录:".bold());
        for f in &c.follow_ups {
            let satisfaction = if f.satisfaction { "满意".green() } else { "不满意".red() };
            let comments = f.comments.as_deref().unwrap_or("");
            println!("  [{}] {} - {}", format_datetime(f.created_at), satisfaction, comments);
        }
        println!();
    }

    println!("{}", "历史记录:".bold());
    for h in &c.history {
        let details = h.details.as_deref().unwrap_or("");
        println!("  [{}] {}: {} ({})", format_datetime(h.timestamp), h.action, details, h.actor);
    }
    println!("{}", "=".repeat(80));
}

fn print_handlers(handlers: &[Handler]) {
    println!("{}", "处理人列表:".bold());
    println!("{:-<60}", "");
    println!("{:<15} {:<20} {:<10}", "ID", "姓名", "级别");
    println!("{:-<60}", "");
    for h in handlers {
        println!("{:<15} {:<20} Lv.{}", h.id, h.name, h.level);
    }
}

fn main() {
    let cli = Cli::parse();
    let api = ApiClient::new(cli.server);

    match cli.command {
        Commands::Health => {
            match api.health() {
                Ok(status) => println!("服务状态: {}", status.green()),
                Err(e) => eprintln!("{}: {}", "连接失败".red(), e),
            }
        }
        Commands::Handlers => {
            match api.list_handlers() {
                Ok(handlers) => print_handlers(&handlers),
                Err(e) => eprintln!("{}: {}", "获取处理人列表失败".red(), e),
            }
        }
        Commands::List => {
            match api.list_complaints() {
                Ok(complaints) => {
                    if complaints.is_empty() {
                        println!("{}", "暂无投诉".dimmed());
                    } else {
                        println!("共 {} 条投诉", complaints.len().to_string().bold());
                        for c in &complaints {
                            print_complaint_short(c);
                        }
                    }
                }
                Err(e) => eprintln!("{}: {}", "获取投诉列表失败".red(), e),
            }
        }
        Commands::Get { id } => {
            match api.get_complaint(id) {
                Ok(complaint) => print_complaint_detail(&complaint),
                Err(e) => eprintln!("{}: {}", "获取投诉详情失败".red(), e),
            }
        }
        Commands::Create {
            title,
            description,
            customer_name,
            customer_contact,
            severity,
            handler_id,
        } => {
            let req = CreateComplaintRequest {
                title,
                description,
                customer_name,
                customer_contact,
                severity,
                handler_id,
            };
            match api.create_complaint(req) {
                Ok(c) => {
                    println!("{}", "投诉创建成功!".green().bold());
                    print_complaint_detail(&c);
                }
                Err(e) => eprintln!("{}: {}", "创建投诉失败".red(), e),
            }
        }
        Commands::Respond { id, handler_id, response } => {
            let req = RespondToComplaintRequest { handler_id, response };
            match api.respond(id, req) {
                Ok(c) => {
                    println!("{}", "响应成功!".green().bold());
                    print_complaint_short(&c);
                }
                Err(e) => eprintln!("{}: {}", "响应投诉失败".red(), e),
            }
        }
        Commands::Solution { id, handler_id, solution } => {
            let req = ProposeSolutionRequest { handler_id, solution };
            match api.propose_solution(id, req) {
                Ok(c) => {
                    println!("{}", "方案提交成功!".green().bold());
                    print_complaint_short(&c);
                }
                Err(e) => eprintln!("{}: {}", "提交方案失败".red(), e),
            }
        }
        Commands::Feedback { id, accepted, comments } => {
            let req = CustomerFeedbackRequest { accepted, comments };
            match api.customer_feedback(id, req) {
                Ok(c) => {
                    let msg = if accepted { "客户接受方案，投诉已结案".green() } else { "客户不接受方案，投诉已重新打开".yellow() };
                    println!("{}", msg.bold());
                    print_complaint_short(&c);
                }
                Err(e) => eprintln!("{}: {}", "提交反馈失败".red(), e),
            }
        }
        Commands::FollowUp { id, satisfied, comments } => {
            let req = FollowUpFeedbackRequest { satisfied, comments };
            match api.follow_up(id, req) {
                Ok(c) => {
                    let msg = if satisfied { "回访满意".green() } else { "回访不满意，投诉已重新打开并升级".yellow() };
                    println!("{}", msg.bold());
                    print_complaint_short(&c);
                }
                Err(e) => eprintln!("{}: {}", "提交回访反馈失败".red(), e),
            }
        }
    }
}
