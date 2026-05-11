use archive_core::*;
use clap::{Subcommand, ValueEnum};
use serde::{Deserialize, Serialize};

use crate::api::ApiClient;

#[derive(Subcommand, Debug)]
pub enum UserSubCommand {
    #[command(about = "列出所有用户")]
    List,
    #[command(about = "获取用户详情")]
    Get { id: String },
    #[command(about = "创建用户")]
    Create {
        #[arg(long)]
        id: String,
        #[arg(long)]
        name: String,
        #[arg(long)]
        email: String,
        #[arg(long)]
        department: String,
        #[arg(long, value_enum)]
        role: UserRoleArg,
    },
}

#[derive(Subcommand, Debug)]
pub enum ArchiveSubCommand {
    #[command(about = "列出所有档案")]
    List,
    #[command(about = "获取档案详情")]
    Get { id: String },
    #[command(about = "创建档案")]
    Create {
        #[arg(long)]
        id: String,
        #[arg(long)]
        title: String,
        #[arg(long)]
        description: String,
        #[arg(long, value_enum)]
        classification: ClassificationLevelArg,
    },
}

#[derive(Subcommand, Debug)]
pub enum BorrowSubCommand {
    #[command(about = "列出所有借阅")]
    List,
    #[command(about = "获取借阅详情")]
    Get { id: String },
    #[command(about = "申请借阅")]
    Request {
        #[arg(long)]
        user_id: String,
        #[arg(long)]
        archive_id: String,
    },
    #[command(about = "续借")]
    Renew { id: String },
    #[command(about = "归还")]
    Return {
        id: String,
        #[arg(long)]
        checked_by: String,
        #[arg(long)]
        damaged: bool,
        #[arg(long)]
        damage_description: Option<String>,
        #[arg(long)]
        compensation: Option<f64>,
    },
}

#[derive(Subcommand, Debug)]
pub enum ApprovalSubCommand {
    #[command(about = "获取审批流程详情")]
    Get { id: String },
    #[command(about = "审批通过")]
    Approve {
        id: String,
        #[arg(long)]
        approver_id: String,
        #[arg(long)]
        step: u8,
        #[arg(long)]
        comment: Option<String>,
    },
    #[command(about = "审批拒绝")]
    Reject {
        id: String,
        #[arg(long)]
        approver_id: String,
        #[arg(long)]
        step: u8,
        #[arg(long)]
        comment: Option<String>,
    },
}

#[derive(Subcommand, Debug)]
pub enum ReservationSubCommand {
    #[command(about = "列出所有预约")]
    List,
    #[command(about = "获取预约详情")]
    Get { id: String },
    #[command(about = "预约档案")]
    Create {
        #[arg(long)]
        user_id: String,
        #[arg(long)]
        archive_id: String,
    },
    #[command(about = "领取预约的档案")]
    Claim { id: String },
    #[command(about = "取消预约")]
    Cancel { id: String },
}

#[derive(Subcommand, Debug)]
pub enum SystemSubCommand {
    #[command(about = "获取到期提醒（3天内到期）")]
    Reminders,
    #[command(about = "处理超期")]
    ProcessOverdue,
    #[command(about = "检查过期预约")]
    CheckExpiredReservations,
}

#[derive(ValueEnum, Clone, Debug)]
pub enum ClassificationLevelArg {
    Public,
    Internal,
    Confidential,
}

#[derive(ValueEnum, Clone, Debug)]
pub enum UserRoleArg {
    User,
    Approver,
    Admin,
}

impl From<ClassificationLevelArg> for ClassificationLevel {
    fn from(arg: ClassificationLevelArg) -> Self {
        match arg {
            ClassificationLevelArg::Public => ClassificationLevel::Public,
            ClassificationLevelArg::Internal => ClassificationLevel::Internal,
            ClassificationLevelArg::Confidential => ClassificationLevel::Confidential,
        }
    }
}

impl From<UserRoleArg> for UserRole {
    fn from(arg: UserRoleArg) -> Self {
        match arg {
            UserRoleArg::User => UserRole::User,
            UserRoleArg::Approver => UserRole::Approver,
            UserRoleArg::Admin => UserRole::Admin,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateUserRequest {
    id: String,
    name: String,
    email: String,
    department: String,
    role: UserRole,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateArchiveRequest {
    id: String,
    title: String,
    description: String,
    classification: ClassificationLevel,
}

#[derive(Debug, Serialize, Deserialize)]
struct BorrowRequest {
    user_id: String,
    archive_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApproveRequest {
    approver_id: String,
    step: u8,
    comment: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ReturnRequest {
    is_damaged: bool,
    damage_description: Option<String>,
    compensation_amount: Option<f64>,
    checked_by: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ReserveRequest {
    user_id: String,
    archive_id: String,
}

fn print_user(user: &User) {
    println!("ID: {}", user.id);
    println!("姓名: {}", user.name);
    println!("邮箱: {}", user.email);
    println!("部门: {}", user.department);
    println!("角色: {:?}", user.role);
    println!("最大借阅数: {}", user.max_borrow_count);
    println!("状态: {}", if user.is_suspended { "已暂停" } else { "正常" });
}

fn print_archive(archive: &Archive) {
    println!("ID: {}", archive.id);
    println!("标题: {}", archive.title);
    println!("描述: {}", archive.description);
    println!("密级: {}", archive.classification);
    println!("借阅期限: {}天", archive.classification.max_borrow_days());
    println!(
        "是否可带出: {}",
        if archive.classification.can_take_out() { "是" } else { "否" }
    );
    println!("状态: {}", if archive.is_borrowed { "已借出" } else { "可借阅" });
}

fn print_borrow(borrow: &Borrow) {
    println!("ID: {}", borrow.id);
    println!("档案ID: {}", borrow.archive_id);
    println!("用户ID: {}", borrow.user_id);
    println!("状态: {:?}", borrow.status);
    println!("借阅日期: {}", borrow.borrow_date.format("%Y-%m-%d %H:%M:%S"));
    println!("到期日期: {}", borrow.due_date.format("%Y-%m-%d %H:%M:%S"));
    println!("续借次数: {}", borrow.renewal_count);
    println!(
        "借阅方式: {}",
        if borrow.is_reading_room { "阅览室阅览" } else { "可带出" }
    );
    if let Some(damage) = &borrow.damage_report {
        println!("损坏报告: {}", damage.description);
        if let Some(amount) = damage.compensation_amount {
            println!("赔偿金额: {}元", amount);
        }
    }
}

fn print_reservation(reservation: &Reservation) {
    println!("ID: {}", reservation.id);
    println!("档案ID: {}", reservation.archive_id);
    println!("用户ID: {}", reservation.user_id);
    println!("状态: {:?}", reservation.status);
    println!("排队位置: {}", reservation.queue_position);
    if let Some(notified) = reservation.notified_at {
        println!("通知时间: {}", notified.format("%Y-%m-%d %H:%M:%S"));
    }
    if let Some(claimed) = reservation.claimed_at {
        println!("领取时间: {}", claimed.format("%Y-%m-%d %H:%M:%S"));
    }
}

pub async fn handle_user_subcommand(client: &ApiClient, cmd: UserSubCommand) -> Result<(), String> {
    match cmd {
        UserSubCommand::List => {
            let users: Vec<User> = client.get("/users").await?;
            for user in &users {
                print_user(user);
                println!("---");
            }
            println!("共 {} 个用户", users.len());
        }
        UserSubCommand::Get { id } => {
            let user: User = client.get(&format!("/users/{}", id)).await?;
            print_user(&user);
        }
        UserSubCommand::Create { id, name, email, department, role } => {
            let req = CreateUserRequest {
                id,
                name,
                email,
                department,
                role: role.into(),
            };
            let user: User = client.post("/users", &req).await?;
            println!("用户创建成功:");
            print_user(&user);
        }
    }
    Ok(())
}

pub async fn handle_archive_subcommand(client: &ApiClient, cmd: ArchiveSubCommand) -> Result<(), String> {
    match cmd {
        ArchiveSubCommand::List => {
            let archives: Vec<Archive> = client.get("/archives").await?;
            for archive in &archives {
                print_archive(archive);
                println!("---");
            }
            println!("共 {} 份档案", archives.len());
        }
        ArchiveSubCommand::Get { id } => {
            let archive: Archive = client.get(&format!("/archives/{}", id)).await?;
            print_archive(&archive);
        }
        ArchiveSubCommand::Create { id, title, description, classification } => {
            let req = CreateArchiveRequest {
                id,
                title,
                description,
                classification: classification.into(),
            };
            let archive: Archive = client.post("/archives", &req).await?;
            println!("档案创建成功:");
            print_archive(&archive);
        }
    }
    Ok(())
}

pub async fn handle_borrow_subcommand(client: &ApiClient, cmd: BorrowSubCommand) -> Result<(), String> {
    match cmd {
        BorrowSubCommand::List => {
            let borrows: Vec<Borrow> = client.get("/borrows").await?;
            for borrow in &borrows {
                print_borrow(borrow);
                println!("---");
            }
            println!("共 {} 条借阅记录", borrows.len());
        }
        BorrowSubCommand::Get { id } => {
            let borrow: Borrow = client.get(&format!("/borrows/{}", id)).await?;
            print_borrow(&borrow);
        }
        BorrowSubCommand::Request { user_id, archive_id } => {
            let req = BorrowRequest { user_id, archive_id };
            let borrow: Borrow = client.post("/borrows", &req).await?;
            println!("借阅申请创建成功:");
            print_borrow(&borrow);
        }
        BorrowSubCommand::Renew { id } => {
            let borrow: Borrow = client.put_empty(&format!("/borrows/{}/renew", id)).await?;
            println!("续借成功:");
            print_borrow(&borrow);
        }
        BorrowSubCommand::Return { id, checked_by, damaged, damage_description, compensation } => {
            let req = ReturnRequest {
                is_damaged: damaged,
                damage_description,
                compensation_amount: compensation,
                checked_by,
            };
            let borrow: Borrow = client.post(&format!("/borrows/{}/return", id), &req).await?;
            println!("归还成功:");
            print_borrow(&borrow);
        }
    }
    Ok(())
}

pub async fn handle_approval_subcommand(client: &ApiClient, cmd: ApprovalSubCommand) -> Result<(), String> {
    match cmd {
        ApprovalSubCommand::Get { id } => {
            let process: ApprovalProcess = client.get(&format!("/approvals/{}", id)).await?;
            println!("审批流程ID: {}", process.id);
            println!("借阅ID: {}", process.borrow_id);
            println!("档案ID: {}", process.archive_id);
            println!("申请人ID: {}", process.requester_id);
            println!("创建时间: {}", process.created_at.format("%Y-%m-%d %H:%M:%S"));
            println!("\n审批步骤:");
            for step in &process.steps {
                println!("  步骤{} - 审批人: {}, 状态: {:?}", step.step, step.approver_id, step.status);
                if let Some(comment) = &step.comment {
                    println!("    备注: {}", comment);
                }
                if let Some(time) = step.approved_at {
                    println!("    处理时间: {}", time.format("%Y-%m-%d %H:%M:%S"));
                }
            }
        }
        ApprovalSubCommand::Approve { id, approver_id, step, comment } => {
            let req = ApproveRequest {
                approver_id,
                step,
                comment,
            };
            let borrow: Borrow = client.post(&format!("/approvals/{}/approve", id), &req).await?;
            println!("审批通过:");
            print_borrow(&borrow);
        }
        ApprovalSubCommand::Reject { id, approver_id, step, comment } => {
            let req = ApproveRequest {
                approver_id,
                step,
                comment,
            };
            let borrow: Borrow = client.post(&format!("/approvals/{}/reject", id), &req).await?;
            println!("审批拒绝:");
            print_borrow(&borrow);
        }
    }
    Ok(())
}

pub async fn handle_reservation_subcommand(client: &ApiClient, cmd: ReservationSubCommand) -> Result<(), String> {
    match cmd {
        ReservationSubCommand::List => {
            let reservations: Vec<Reservation> = client.get("/reservations").await?;
            for reservation in &reservations {
                print_reservation(reservation);
                println!("---");
            }
            println!("共 {} 条预约记录", reservations.len());
        }
        ReservationSubCommand::Get { id } => {
            let reservation: Reservation = client.get(&format!("/reservations/{}", id)).await?;
            print_reservation(&reservation);
        }
        ReservationSubCommand::Create { user_id, archive_id } => {
            let req = ReserveRequest { user_id, archive_id };
            let reservation: Reservation = client.post("/reservations", &req).await?;
            println!("预约成功:");
            print_reservation(&reservation);
        }
        ReservationSubCommand::Claim { id } => {
            let borrow: Borrow = client.post(&format!("/reservations/{}/claim", id), &serde_json::json!({})).await?;
            println!("领取成功，借阅记录:");
            print_borrow(&borrow);
        }
        ReservationSubCommand::Cancel { id } => {
            let reservation: Reservation = client.delete(&format!("/reservations/{}/cancel", id)).await?;
            println!("取消预约成功:");
            print_reservation(&reservation);
        }
    }
    Ok(())
}

pub async fn handle_system_subcommand(client: &ApiClient, cmd: SystemSubCommand) -> Result<(), String> {
    match cmd {
        SystemSubCommand::Reminders => {
            let borrows: Vec<Borrow> = client.get("/reminders/due").await?;
            for borrow in &borrows {
                print_borrow(borrow);
                println!("---");
            }
            println!("共 {} 条即将到期的借阅记录", borrows.len());
        }
        SystemSubCommand::ProcessOverdue => {
            let overdues: Vec<OverdueInfo> = client.post("/overdue/process", &serde_json::json!({})).await?;
            for info in &overdues {
                println!("借阅ID: {}", info.borrow_id);
                println!("超期天数: {}", info.days_overdue);
                println!("处理阶段: {:?}", info.stage);
                if let Some(time) = info.notified_at {
                    println!("通知时间: {}", time.format("%Y-%m-%d %H:%M:%S"));
                }
                println!("---");
            }
            println!("共处理 {} 条超期记录", overdues.len());
        }
        SystemSubCommand::CheckExpiredReservations => {
            let expired: Vec<Reservation> = client.post("/reservations/check-expired", &serde_json::json!({})).await?;
            for reservation in &expired {
                print_reservation(reservation);
                println!("---");
            }
            println!("共 {} 条过期预约", expired.len());
        }
    }
    Ok(())
}
