use clap::{Parser, Subcommand};
use league_core::{MatchResult, StandingsEntry, Match};
use reqwest::Client;
use chrono::{DateTime, Utc};

#[derive(Parser, Debug)]
#[command(name = "league-cli", about = "足球联赛积分排名系统命令行工具")]
struct Cli {
    #[arg(long, env = "SERVER_URL", default_value = "http://localhost:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    #[command(about = "查看积分榜")]
    Standings,

    #[command(about = "添加比赛结果")]
    AddMatch {
        #[arg(long)]
        home: String,
        #[arg(long)]
        away: String,
        #[arg(long)]
        home_goals: i32,
        #[arg(long)]
        away_goals: i32,
        #[arg(long)]
        round: i32,
        #[arg(long, default_value_t = Utc::now())]
        date: DateTime<Utc>,
    },

    #[command(about = "查看比赛记录")]
    Matches,

    #[command(about = "添加球队")]
    AddTeam {
        #[arg(long)]
        name: String,
    },

    #[command(about = "查看所有球队")]
    Teams,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let http_client = Client::new();

    match cli.command {
        Commands::Standings => {
            let standings: Vec<StandingsEntry> = http_client
                .get(format!("{}/standings", cli.server))
                .send()
                .await?
                .json()
                .await?;
            print_standings(&standings);
        }
        Commands::AddMatch {
            home,
            away,
            home_goals,
            away_goals,
            round,
            date,
        } => {
            let match_result = MatchResult {
                home_team: home,
                away_team: away,
                home_goals,
                away_goals,
                round,
                date,
            };
            let response = http_client
                .post(format!("{}/matches", cli.server))
                .json(&match_result)
                .send()
                .await?;
            if response.status().is_success() {
                println!("比赛结果已添加");
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error.get("error").unwrap_or(&serde_json::json!("未知错误")));
                std::process::exit(1);
            }
        }
        Commands::Matches => {
            let matches: Vec<Match> = http_client
                .get(format!("{}/matches", cli.server))
                .send()
                .await?
                .json()
                .await?;
            print_matches(&matches);
        }
        Commands::AddTeam { name } => {
            let response = http_client
                .post(format!("{}/teams", cli.server))
                .json(&name)
                .send()
                .await?;
            if response.status().is_success() {
                println!("球队 {} 已添加", name);
            } else {
                let error: serde_json::Value = response.json().await?;
                eprintln!("错误: {}", error.get("error").unwrap_or(&serde_json::json!("未知错误")));
                std::process::exit(1);
            }
        }
        Commands::Teams => {
            let teams: Vec<String> = http_client
                .get(format!("{}/teams", cli.server))
                .send()
                .await?
                .json()
                .await?;
            println!("参赛球队:");
            for team in teams {
                println!("  - {}", team);
            }
        }
    }

    Ok(())
}

fn print_standings(standings: &[StandingsEntry]) {
    println!(
        "{:<5} {:<20} {:<6} {:<4} {:<4} {:<4} {:<6} {:<6} {:<6} {:<6}",
        "排名", "球队", "场次", "胜", "平", "负", "进球", "失球", "净胜球", "积分"
    );
    println!("{:-<80}", "");

    for entry in standings {
        println!(
            "{:<5} {:<20} {:<6} {:<4} {:<4} {:<4} {:<6} {:<6} {:<6} {:<6}",
            entry.rank,
            entry.team,
            entry.played,
            entry.won,
            entry.drawn,
            entry.lost,
            entry.goals_for,
            entry.goals_against,
            entry.goal_difference,
            entry.points
        );
    }
}

fn print_matches(matches: &[Match]) {
    if matches.is_empty() {
        println!("暂无比赛记录");
        return;
    }

    println!("{:<6} {:<20} {:<4} - {:<4} {:<20}", "轮次", "主队", "比分", "", "客队");
    println!("{:-<60}", "");

    for m in matches {
        println!(
            "{:<6} {:<20} {:<4} - {:<4} {:<20}",
            m.round, m.home_team, m.home_goals, m.away_goals, m.away_team
        );
    }
}
