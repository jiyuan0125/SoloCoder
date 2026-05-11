use chrono::{DateTime, Utc, NaiveDate, Datelike};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Currency {
    CNY,
    USD,
    EUR,
    JPY,
}

impl Currency {
    pub fn precision(&self) -> u32 {
        match self {
            Currency::CNY | Currency::USD | Currency::EUR => 2,
            Currency::JPY => 0,
        }
    }

    pub fn symbol(&self) -> &'static str {
        match self {
            Currency::CNY => "CNY",
            Currency::USD => "USD",
            Currency::EUR => "EUR",
            Currency::JPY => "JPY",
        }
    }

    pub fn from_symbol(s: &str) -> Option<Self> {
        match s.to_uppercase().as_str() {
            "CNY" => Some(Currency::CNY),
            "USD" => Some(Currency::USD),
            "EUR" => Some(Currency::EUR),
            "JPY" => Some(Currency::JPY),
            _ => None,
        }
    }
}

impl fmt::Display for Currency {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.symbol())
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct CurrencyPair {
    pub base: Currency,
    pub quote: Currency,
}

impl CurrencyPair {
    pub fn new(base: Currency, quote: Currency) -> Self {
        Self { base, quote }
    }

    pub fn to_string(&self) -> String {
        format!("{}/{}", self.base, self.quote)
    }

    pub fn from_string(s: &str) -> Option<Self> {
        let parts: Vec<&str> = s.split('/').collect();
        if parts.len() != 2 {
            return None;
        }
        let base = Currency::from_symbol(parts[0])?;
        let quote = Currency::from_symbol(parts[1])?;
        Some(Self { base, quote })
    }
}

impl fmt::Display for CurrencyPair {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}/{}", self.base, self.quote)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderSide {
    Buy,
    Sell,
}

impl fmt::Display for OrderSide {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            OrderSide::Buy => write!(f, "买入"),
            OrderSide::Sell => write!(f, "卖出"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Pending,
    PendingApproval,
    Approved,
    Rejected,
    Settled,
}

impl fmt::Display for OrderStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            OrderStatus::Pending => write!(f, "待交割"),
            OrderStatus::PendingApproval => write!(f, "待审批"),
            OrderStatus::Approved => write!(f, "已批准"),
            OrderStatus::Rejected => write!(f, "已拒绝"),
            OrderStatus::Settled => write!(f, "已交割"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Account {
    pub id: String,
    pub name: String,
    pub is_reviewer: bool,
    pub created_at: DateTime<Utc>,
}

impl Account {
    pub fn new(name: String, is_reviewer: bool) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name,
            is_reviewer,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Balance {
    pub available: f64,
    pub frozen: f64,
}

impl Balance {
    pub fn new(initial: f64) -> Self {
        Self {
            available: initial,
            frozen: 0.0,
        }
    }

    pub fn total(&self) -> f64 {
        self.available + self.frozen
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: String,
    pub account_id: String,
    pub pair: CurrencyPair,
    pub side: OrderSide,
    pub amount: f64,
    pub rate: f64,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
    pub settlement_date: NaiveDate,
    pub usd_equivalent: f64,
    pub approval_comment: Option<String>,
}

impl Order {
    pub fn new(
        account_id: String,
        pair: CurrencyPair,
        side: OrderSide,
        amount: f64,
        rate: f64,
        usd_equivalent: f64,
    ) -> Self {
        let now = Utc::now();
        let settlement_date = Self::calculate_settlement_date(now.date_naive());
        
        let status = if usd_equivalent > 50000.0 {
            OrderStatus::PendingApproval
        } else {
            OrderStatus::Pending
        };

        Self {
            id: Uuid::new_v4().to_string(),
            account_id,
            pair,
            side,
            amount,
            rate,
            status,
            created_at: now,
            settlement_date,
            usd_equivalent,
            approval_comment: None,
        }
    }

    pub fn calculate_settlement_date(trade_date: NaiveDate) -> NaiveDate {
        let mut settlement = trade_date;
        let mut business_days = 0;
        
        while business_days < 2 {
            settlement = settlement.succ_opt().unwrap_or(settlement);
            let weekday = settlement.weekday();
            if weekday != chrono::Weekday::Sat && weekday != chrono::Weekday::Sun {
                business_days += 1;
            }
        }
        
        settlement
    }

    pub fn counter_amount(&self) -> f64 {
        self.amount * self.rate
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExchangeRate {
    pub pair: CurrencyPair,
    pub bid: f64,
    pub ask: f64,
    pub updated_at: DateTime<Utc>,
}

impl ExchangeRate {
    pub fn new(pair: CurrencyPair, bid: f64, ask: f64) -> Self {
        Self {
            pair,
            bid,
            ask,
            updated_at: Utc::now(),
        }
    }
}
