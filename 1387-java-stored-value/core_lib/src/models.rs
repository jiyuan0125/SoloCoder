use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum MemberLevel {
    Regular,
    Silver,
    Gold,
    Diamond,
}

impl MemberLevel {
    pub fn discount(&self) -> f64 {
        match self {
            MemberLevel::Regular => 1.0,
            MemberLevel::Silver => 0.98,
            MemberLevel::Gold => 0.95,
            MemberLevel::Diamond => 0.90,
        }
    }
    
    pub fn name(&self) -> &'static str {
        match self {
            MemberLevel::Regular => "普通会员",
            MemberLevel::Silver => "银卡会员",
            MemberLevel::Gold => "金卡会员",
            MemberLevel::Diamond => "钻石会员",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Member {
    pub id: String,
    pub name: String,
    pub level: MemberLevel,
    pub principal_balance: u64,
    pub gift_balance: u64,
    pub points: u64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl Member {
    pub fn new(name: String, level: MemberLevel) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            level,
            principal_balance: 0,
            gift_balance: 0,
            points: 0,
            created_at: now,
            updated_at: now,
        }
    }
    
    pub fn total_balance(&self) -> u64 {
        self.principal_balance + self.gift_balance
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TransactionType {
    Recharge,
    Gift,
    Consume,
    Refund,
}

impl TransactionType {
    pub fn name(&self) -> &'static str {
        match self {
            TransactionType::Recharge => "充值",
            TransactionType::Gift => "赠送",
            TransactionType::Consume => "消费",
            TransactionType::Refund => "退款",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Transaction {
    pub id: String,
    pub member_id: String,
    pub transaction_type: TransactionType,
    pub original_price: u64,
    pub discounted_price: u64,
    pub principal_amount: i64,
    pub gift_amount: i64,
    pub points: i64,
    pub description: String,
    pub related_transaction_id: Option<String>,
    pub created_at: DateTime<Utc>,
}

impl Transaction {
    pub fn new(
        member_id: String,
        transaction_type: TransactionType,
        original_price: u64,
        discounted_price: u64,
        principal_amount: i64,
        gift_amount: i64,
        points: i64,
        description: String,
        related_transaction_id: Option<String>,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            member_id,
            transaction_type,
            original_price,
            discounted_price,
            principal_amount,
            gift_amount,
            points,
            description,
            related_transaction_id,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RechargeConfig {
    pub thresholds: Vec<(u64, u64)>,
}

impl Default for RechargeConfig {
    fn default() -> Self {
        Self {
            thresholds: vec![
                (100, 5),
                (300, 20),
                (500, 50),
                (1000, 150),
            ],
        }
    }
}

impl RechargeConfig {
    pub fn get_gift_amount(&self, recharge_amount: u64) -> u64 {
        let mut gift_amount = 0;
        for &(threshold, gift) in &self.thresholds {
            if recharge_amount >= threshold {
                gift_amount = gift;
            }
        }
        gift_amount
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConsumeBreakdown {
    pub gift_deducted: u64,
    pub principal_deducted: u64,
}
