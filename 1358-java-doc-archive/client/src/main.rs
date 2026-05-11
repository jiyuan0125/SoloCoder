mod api;
mod commands;

use clap::{Parser, Subcommand};

use commands::*;

#[derive(Parser, Debug)]
#[command(author, version, about = "档案借阅管理系统客户端", long_about = None)]
struct Cli {
    #[arg(short, long, env = "ARCHIVE_SERVER_URL", default_value = "http://localhost:3000")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(about = "用户管理命令")]
    User {
        #[command(subcommand)]
        cmd: UserSubCommand,
    },

    #[command(about = "档案管理命令")]
    Archive {
        #[command(subcommand)]
        cmd: ArchiveSubCommand,
    },

    #[command(about = "借阅管理命令")]
    Borrow {
        #[command(subcommand)]
        cmd: BorrowSubCommand,
    },

    #[command(about = "审批管理命令")]
    Approval {
        #[command(subcommand)]
        cmd: ApprovalSubCommand,
    },

    #[command(about = "预约管理命令")]
    Reservation {
        #[command(subcommand)]
        cmd: ReservationSubCommand,
    },

    #[command(about = "提醒和超期处理命令")]
    System {
        #[command(subcommand)]
        cmd: SystemSubCommand,
    },
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = api::ApiClient::new(&cli.server_url);

    let result = match cli.command {
        Commands::User { cmd } => handle_user_subcommand(&client, cmd).await,
        Commands::Archive { cmd } => handle_archive_subcommand(&client, cmd).await,
        Commands::Borrow { cmd } => handle_borrow_subcommand(&client, cmd).await,
        Commands::Approval { cmd } => handle_approval_subcommand(&client, cmd).await,
        Commands::Reservation { cmd } => handle_reservation_subcommand(&client, cmd).await,
        Commands::System { cmd } => handle_system_subcommand(&client, cmd).await,
    };

    if let Err(e) = result {
        eprintln!("错误: {}", e);
        std::process::exit(1);
    }
}
