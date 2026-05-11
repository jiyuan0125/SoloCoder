use clap::{Parser, Subcommand};
use reqwest::Client;
use review_core::models::{
    CreateFollowUpRequest, CreateInitialReviewRequest, CreateReplyRequest, DeleteReviewRequest,
    ReplyTarget,
};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "SERVER_URL", default_value = "http://127.0.0.1:3000")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    CreateInitialReview {
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        customer_id: String,
        #[arg(long)]
        rating: u32,
        #[arg(long)]
        content: String,
    },
    CreateFollowUp {
        #[arg(long)]
        review_id: Uuid,
        #[arg(long)]
        customer_id: String,
        #[arg(long)]
        content: String,
    },
    CreateReply {
        #[arg(long)]
        review_id: Uuid,
        #[arg(long)]
        target_type: String,
        #[arg(long)]
        content: String,
    },
    CreateSupplementReply {
        #[arg(long)]
        review_id: Uuid,
        #[arg(long)]
        target_type: String,
        #[arg(long)]
        content: String,
    },
    DeleteReview {
        #[arg(long)]
        review_id: Uuid,
        #[arg(long)]
        customer_id: String,
    },
    ListReviews {
        #[arg(long)]
        product_id: String,
        #[arg(long)]
        rating: Option<u32>,
        #[arg(long)]
        has_follow_up: Option<bool>,
        #[arg(long)]
        has_reply: Option<bool>,
        #[arg(long)]
        as_merchant: Option<bool>,
    },
    GetReview {
        #[arg(long)]
        review_id: Uuid,
        #[arg(long)]
        as_merchant: Option<bool>,
    },
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();
    let client = Client::new();
    let base_url = args.server.trim_end_matches('/').to_string();

    match args.command {
        Commands::CreateInitialReview {
            product_id,
            customer_id,
            rating,
            content,
        } => {
            let req = CreateInitialReviewRequest {
                product_id,
                customer_id,
                rating,
                content,
            };
            let url = format!("{}/api/reviews", base_url);
            let response = client.post(&url).json(&req).send().await?;
            print_response(response).await;
        }
        Commands::CreateFollowUp {
            review_id,
            customer_id,
            content,
        } => {
            let req = CreateFollowUpRequest {
                review_id,
                customer_id,
                content,
            };
            let url = format!("{}/api/reviews/follow-up", base_url);
            let response = client.post(&url).json(&req).send().await?;
            print_response(response).await;
        }
        Commands::CreateReply {
            review_id,
            target_type,
            content,
        } => {
            let target = parse_target_type(&target_type)?;
            let req = CreateReplyRequest {
                review_id,
                target_type: target,
                content,
            };
            let url = format!("{}/api/reviews/reply", base_url);
            let response = client.post(&url).json(&req).send().await?;
            print_response(response).await;
        }
        Commands::CreateSupplementReply {
            review_id,
            target_type,
            content,
        } => {
            let target = parse_target_type(&target_type)?;
            let req = CreateReplyRequest {
                review_id,
                target_type: target,
                content,
            };
            let url = format!("{}/api/reviews/supplement-reply", base_url);
            let response = client.post(&url).json(&req).send().await?;
            print_response(response).await;
        }
        Commands::DeleteReview {
            review_id,
            customer_id,
        } => {
            let req = DeleteReviewRequest {
                review_id,
                customer_id,
            };
            let url = format!("{}/api/reviews", base_url);
            let response = client.delete(&url).json(&req).send().await?;
            if response.status().is_success() {
                println!("评价已删除");
            } else {
                print_response(response).await;
            }
        }
        Commands::ListReviews {
            product_id,
            rating,
            has_follow_up,
            has_reply,
            as_merchant,
        } => {
            let mut url = format!("{}/api/reviews/product/{}", base_url, product_id);
            let mut params = Vec::new();
            if let Some(r) = rating {
                params.push(format!("rating={}", r));
            }
            if let Some(hf) = has_follow_up {
                params.push(format!("has_follow_up={}", hf));
            }
            if let Some(hr) = has_reply {
                params.push(format!("has_reply={}", hr));
            }
            if let Some(am) = as_merchant {
                params.push(format!("as_merchant={}", am));
            }
            if !params.is_empty() {
                url.push('?');
                url.push_str(&params.join("&"));
            }
            let response = client.get(&url).send().await?;
            print_response(response).await;
        }
        Commands::GetReview {
            review_id,
            as_merchant,
        } => {
            let mut url = format!("{}/api/reviews/{}", base_url, review_id);
            if let Some(am) = as_merchant {
                url.push_str(&format!("?as_merchant={}", am));
            }
            let response = client.get(&url).send().await?;
            print_response(response).await;
        }
    }

    Ok(())
}

fn parse_target_type(s: &str) -> Result<ReplyTarget, String> {
    match s.to_lowercase().as_str() {
        "initial" | "initialreview" => Ok(ReplyTarget::InitialReview),
        "followup" | "follow-up" | "followupreview" | "follow-upreview" => {
            Ok(ReplyTarget::FollowUpReview)
        }
        _ => Err(format!("无效的 target_type: {}, 请使用 'initial' 或 'followup'", s)),
    }
}

async fn print_response(response: reqwest::Response) {
    let status = response.status();
    let body = response.text().await.unwrap_or_default();
    
    println!("状态码: {}", status);
    if !body.is_empty() {
        match serde_json::from_str::<serde_json::Value>(&body) {
            Ok(json) => println!("{}", serde_json::to_string_pretty(&json).unwrap_or(body)),
            Err(_) => println!("{}", body),
        }
    }
}
