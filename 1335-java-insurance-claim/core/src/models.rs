use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Policy {
    pub id: String,
    pub policy_number: String,
    pub holder_name: String,
    pub deductible: Decimal,
    pub policy_year: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct Party {
    pub id: String,
    pub name: String,
    pub policy_id: String,
    pub liability_ratio: Decimal,
    pub own_loss: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct PartySettlement {
    pub party_id: String,
    pub party_name: String,
    pub liability_ratio: Decimal,
    pub own_loss: Decimal,
    pub assumed_amount: Decimal,
    pub deductible_applied: Decimal,
    pub payout_amount: Decimal,
    pub self_bear_amount: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct SettlementSummary {
    pub claim_id: String,
    pub total_loss: Decimal,
    pub party_settlements: Vec<PartySettlement>,
    pub total_payout: Decimal,
    pub total_self_bear: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum ClaimStatus {
    Open,
    Closed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claim {
    pub id: String,
    pub case_number: String,
    pub total_loss: Decimal,
    pub parties: Vec<Party>,
    pub status: ClaimStatus,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub settlement: Option<SettlementSummary>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateClaimRequest {
    pub case_number: String,
    pub total_loss: Decimal,
    pub parties: Vec<CreatePartyRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePartyRequest {
    pub name: String,
    pub policy_id: String,
    pub liability_ratio: Decimal,
    pub own_loss: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePolicyRequest {
    pub policy_number: String,
    pub holder_name: String,
    pub deductible: Decimal,
    pub policy_year: i32,
}

impl Default for ClaimStatus {
    fn default() -> Self {
        ClaimStatus::Open
    }
}

impl Claim {
    pub fn new(case_number: String, total_loss: Decimal, parties: Vec<Party>) -> Self {
        Claim {
            id: Uuid::new_v4().to_string(),
            case_number,
            total_loss,
            parties,
            status: ClaimStatus::Open,
            created_at: Utc::now(),
            closed_at: None,
            settlement: None,
        }
    }

    pub fn is_closed(&self) -> bool {
        matches!(self.status, ClaimStatus::Closed)
    }
}

impl Policy {
    pub fn new(policy_number: String, holder_name: String, deductible: Decimal, policy_year: i32) -> Self {
        Policy {
            id: Uuid::new_v4().to_string(),
            policy_number,
            holder_name,
            deductible,
            policy_year,
        }
    }
}

impl Party {
    pub fn new(name: String, policy_id: String, liability_ratio: Decimal, own_loss: Decimal) -> Self {
        Party {
            id: Uuid::new_v4().to_string(),
            name,
            policy_id,
            liability_ratio,
            own_loss,
        }
    }
}
