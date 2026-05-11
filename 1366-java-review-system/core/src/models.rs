use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum UserRole {
    Customer,
    Merchant,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum Rating {
    One,
    Two,
    Three,
    Four,
    Five,
}

impl Rating {
    pub fn from_u32(value: u32) -> Option<Self> {
        match value {
            1 => Some(Rating::One),
            2 => Some(Rating::Two),
            3 => Some(Rating::Three),
            4 => Some(Rating::Four),
            5 => Some(Rating::Five),
            _ => None,
        }
    }

    pub fn to_u32(&self) -> u32 {
        match self {
            Rating::One => 1,
            Rating::Two => 2,
            Rating::Three => 3,
            Rating::Four => 4,
            Rating::Five => 5,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InitialReview {
    pub rating: Rating,
    pub content: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FollowUpReview {
    pub content: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MerchantReply {
    pub content: String,
    pub is_supplement: bool,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Review {
    pub id: Uuid,
    pub product_id: String,
    pub customer_id: String,
    pub initial: InitialReview,
    pub follow_up: Option<FollowUpReview>,
    pub replies: Vec<MerchantReply>,
    pub deleted_by_customer: bool,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReviewFilter {
    pub rating: Option<Rating>,
    pub has_follow_up: Option<bool>,
    pub has_reply: Option<bool>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateInitialReviewRequest {
    pub product_id: String,
    pub customer_id: String,
    pub rating: u32,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateFollowUpRequest {
    pub review_id: Uuid,
    pub customer_id: String,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateReplyRequest {
    pub review_id: Uuid,
    pub target_type: ReplyTarget,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ReplyTarget {
    InitialReview,
    FollowUpReview,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SupplementReplyRequest {
    pub review_id: Uuid,
    pub target_type: ReplyTarget,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeleteReviewRequest {
    pub review_id: Uuid,
    pub customer_id: String,
}
