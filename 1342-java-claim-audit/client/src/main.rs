use std::str::FromStr;

use clap::{Parser, Subcommand};
use reqwest::Client;
use uuid::Uuid;

use claim_audit_core::{
    ApiResponse, ApproveRequest, AuditStage, Claim, CreateClaimRequest, FinalRejectRequest,
    ReturnRequest, ResubmitRequest, UpdateClaimRequest,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:8100")]
    server_url: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    Create {
        #[arg(short, long)]
        employee_id: String,
        #[arg(short, long)]
        title: String,
        #[arg(short, long)]
        description: String,
        #[arg(short, long)]
        amount: f64,
    },
    List,
    Get {
        #[arg(short, long)]
        id: Uuid,
    },
    Update {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        title: String,
        #[arg(short, long)]
        description: String,
        #[arg(short, long)]
        amount: f64,
    },
    Approve {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        operator: String,
    },
    Return {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long, value_parser = parse_audit_stage)]
        target_stage: AuditStage,
        #[arg(short, long)]
        reason: String,
        #[arg(short, long)]
        operator: String,
    },
    FinalReject {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        reason: String,
        #[arg(short, long)]
        operator: String,
    },
    Resubmit {
        #[arg(short, long)]
        id: Uuid,
        #[arg(short, long)]
        employee_id: String,
    },
}

fn parse_audit_stage(s: &str) -> Result<AuditStage, String> {
    AuditStage::from_str(s).map_err(|e| {
        format!(
            "{}. Valid values: Submitted, InitialReview, ReReview, FinalReview, Approved, FinalRejected",
            e
        )
    })
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    let client = Client::new();
    let base_url = cli.server_url.trim_end_matches('/');

    match cli.command {
        Commands::Create {
            employee_id,
            title,
            description,
            amount,
        } => {
            let req = CreateClaimRequest {
                employee_id,
                title,
                description,
                amount,
            };
            let resp: ApiResponse<Claim> = client
                .post(format!("{}/claims", base_url))
                .json(&req)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
        Commands::List => {
            let resp: ApiResponse<Vec<Claim>> = client
                .get(format!("{}/claims", base_url))
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_list_response(&resp);
        }
        Commands::Get { id } => {
            let resp: ApiResponse<Claim> = client
                .get(format!("{}/claims/{}", base_url, id))
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
        Commands::Update {
            id,
            title,
            description,
            amount,
        } => {
            let req = UpdateClaimRequest {
                title,
                description,
                amount,
            };
            let resp: ApiResponse<Claim> = client
                .put(format!("{}/claims/{}", base_url, id))
                .json(&req)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
        Commands::Approve { id, operator } => {
            let req = ApproveRequest { operator };
            let resp: ApiResponse<Claim> = client
                .post(format!("{}/claims/{}/approve", base_url, id))
                .json(&req)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
        Commands::Return {
            id,
            target_stage,
            reason,
            operator,
        } => {
            let req = ReturnRequest {
                target_stage,
                reason,
                operator,
            };
            let resp: ApiResponse<Claim> = client
                .post(format!("{}/claims/{}/return", base_url, id))
                .json(&req)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
        Commands::FinalReject {
            id,
            reason,
            operator,
        } => {
            let req = FinalRejectRequest { reason, operator };
            let resp: ApiResponse<Claim> = client
                .post(format!("{}/claims/{}/final-reject", base_url, id))
                .json(&req)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
        Commands::Resubmit { id, employee_id } => {
            let req = ResubmitRequest { employee_id };
            let resp: ApiResponse<Claim> = client
                .post(format!("{}/claims/{}/resubmit", base_url, id))
                .json(&req)
                .send()
                .await
                .unwrap()
                .json()
                .await
                .unwrap();
            print_response(&resp);
        }
    }
}

fn print_response(resp: &ApiResponse<Claim>) {
    println!("Success: {}", resp.success);
    println!("Message: {}", resp.message);
    if let Some(claim) = &resp.data {
        print_claim(claim);
    }
}

fn print_list_response(resp: &ApiResponse<Vec<Claim>>) {
    println!("Success: {}", resp.success);
    println!("Message: {}", resp.message);
    if let Some(claims) = &resp.data {
        println!("Claim count: {}", claims.len());
        for claim in claims {
            println!("---");
            print_claim(claim);
        }
    }
}

fn print_claim(claim: &Claim) {
    println!("ID: {}", claim.id);
    println!("Employee: {}", claim.employee_id);
    println!("Title: {}", claim.title);
    println!("Amount: {}", claim.amount);
    println!(
        "Stage: {} ({})",
        claim.current_stage.display_name(),
        stage_to_str(claim.current_stage)
    );
    println!(
        "Priority: {} ({})",
        claim.priority.display_name(),
        priority_to_str(claim.priority)
    );
    println!("Submitted at: {}", claim.submitted_at);
    println!("Stage start: {}", claim.stage_start_time);
    if let Some(deadline) = claim.stage_deadline {
        println!("Deadline: {}", deadline);
    }
    println!("Escalated: {}", claim.escalated_to_manager);
    println!("History:");
    for record in &claim.operation_history {
        println!(
            "  [{}] {} - {} by {}",
            record.timestamp,
            record.stage.display_name(),
            record.action,
            record.operator
        );
        if let Some(reason) = &record.reason {
            println!("    Reason: {}", reason);
        }
    }
}

fn stage_to_str(stage: AuditStage) -> &'static str {
    match stage {
        AuditStage::Submitted => "Submitted",
        AuditStage::InitialReview => "InitialReview",
        AuditStage::ReReview => "ReReview",
        AuditStage::FinalReview => "FinalReview",
        AuditStage::Approved => "Approved",
        AuditStage::FinalRejected => "FinalRejected",
    }
}

fn priority_to_str(priority: claim_audit_core::PriorityLevel) -> &'static str {
    match priority {
        claim_audit_core::PriorityLevel::Normal => "Normal",
        claim_audit_core::PriorityLevel::Serious => "Serious",
        claim_audit_core::PriorityLevel::Urgent => "Urgent",
    }
}
