use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::{HashMap, HashSet};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum QualificationLevel {
    C = 3,
    B = 2,
    A = 1,
}

impl QualificationLevel {
    pub fn meets_requirement(&self, required: &Self) -> bool {
        *self <= *required
    }
}

impl std::fmt::Display for QualificationLevel {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            QualificationLevel::A => write!(f, "A"),
            QualificationLevel::B => write!(f, "B"),
            QualificationLevel::C => write!(f, "C"),
        }
    }
}

impl std::str::FromStr for QualificationLevel {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "A" => Ok(QualificationLevel::A),
            "B" => Ok(QualificationLevel::B),
            "C" => Ok(QualificationLevel::C),
            _ => Err(format!("Invalid qualification level: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ProcurementStatus {
    Published,
    Bidding,
    Evaluating,
    Awarded,
    Cancelled,
}

impl std::fmt::Display for ProcurementStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ProcurementStatus::Published => write!(f, "发布中"),
            ProcurementStatus::Bidding => write!(f, "报价中"),
            ProcurementStatus::Evaluating => write!(f, "评标中"),
            ProcurementStatus::Awarded => write!(f, "已定标"),
            ProcurementStatus::Cancelled => write!(f, "已废标"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Supplier {
    pub id: Uuid,
    pub name: String,
    pub qualification: QualificationLevel,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Procurement {
    pub id: Uuid,
    pub title: String,
    pub description: String,
    pub required_qualification: Option<QualificationLevel>,
    pub publish_date: DateTime<Utc>,
    pub status: ProcurementStatus,
    pub invited_suppliers: HashSet<Uuid>,
    pub accepted_suppliers: HashSet<Uuid>,
    pub current_round: u32,
    pub max_rounds: u32,
    pub bids: HashMap<u32, HashMap<Uuid, Bid>>,
    pub withdrawn_suppliers: HashSet<Uuid>,
    pub winner_id: Option<Uuid>,
    pub winner_bid: Option<Bid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Bid {
    pub supplier_id: Uuid,
    pub procurement_id: Uuid,
    pub round: u32,
    pub price: f64,
    pub delivery_date: DateTime<Utc>,
    pub submitted_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnonymousBid {
    pub price: f64,
    pub delivery_date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoundResult {
    pub round: u32,
    pub bids: Vec<AnonymousBid>,
    pub lowest_price: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateProcurementRequest {
    pub title: String,
    pub description: String,
    pub required_qualification: Option<QualificationLevel>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InviteSuppliersRequest {
    pub procurement_id: Uuid,
    pub supplier_ids: Vec<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AcceptInvitationRequest {
    pub procurement_id: Uuid,
    pub supplier_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubmitBidRequest {
    pub procurement_id: Uuid,
    pub supplier_id: Uuid,
    pub price: f64,
    pub delivery_date: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WithdrawBidRequest {
    pub procurement_id: Uuid,
    pub supplier_id: Uuid,
}
