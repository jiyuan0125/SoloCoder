mod handlers;
mod state;

use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing_subscriber::EnvFilter;

use archive_core::{Archive, ArchiveService, ClassificationLevel, InMemoryRepository, User, UserRole};

use state::AppState;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value = "3000")]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "0.0.0.0")]
    host: String,
}

fn init_sample_data(service: &ArchiveService<InMemoryRepository>) {
    let user1 = User::new("user-1", "张三", "zhangsan@example.com", "研发部", UserRole::User);
    let user2 = User::new("user-2", "李四", "lisi@example.com", "市场部", UserRole::User);
    let approver1 = User::new("approver-1", "王经理", "wangmgr@example.com", "档案部", UserRole::Approver);
    let approver2 = User::new("approver-2", "李总监", "lidir@example.com", "管理层", UserRole::Approver);
    let admin = User::new("admin", "管理员", "admin@example.com", "管理部", UserRole::Admin);

    service.create_user(user1);
    service.create_user(user2);
    service.create_user(approver1);
    service.create_user(approver2);
    service.create_user(admin);

    let archive1 = Archive::new("archive-001", "2024年度财务报告", "公司年度财务汇总报告", ClassificationLevel::Internal);
    let archive2 = Archive::new("archive-002", "员工手册", "公司规章制度和员工福利说明", ClassificationLevel::Public);
    let archive3 = Archive::new("archive-003", "核心技术架构文档", "公司核心系统的技术架构和设计文档", ClassificationLevel::Confidential);
    let archive4 = Archive::new("archive-004", "2024年度战略规划", "公司未来发展战略规划", ClassificationLevel::Internal);
    let archive5 = Archive::new("archive-005", "客户合作协议", "重要客户的合作协议副本", ClassificationLevel::Confidential);

    service.create_archive(archive1);
    service.create_archive(archive2);
    service.create_archive(archive3);
    service.create_archive(archive4);
    service.create_archive(archive5);
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let args = Args::parse();

    let repo = InMemoryRepository::new();
    let service = ArchiveService::new(repo);
    init_sample_data(&service);

    let state = AppState {
        service: Arc::new(Mutex::new(service)),
    };

    let app = handlers::create_router(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("无效的地址");

    tracing::info!("档案借阅管理系统服务端启动在 {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
