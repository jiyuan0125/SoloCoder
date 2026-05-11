use chrono::{DateTime, Local, NaiveDate};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InvoiceStatus {
    Normal,
    Voided,
    Red冲ed,
}

impl std::fmt::Display for InvoiceStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            InvoiceStatus::Normal => write!(f, "normal"),
            InvoiceStatus::Voided => write!(f, "voided"),
            InvoiceStatus::Red冲ed => write!(f, "red冲ed"),
        }
    }
}

impl std::str::FromStr for InvoiceStatus {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "normal" => Ok(InvoiceStatus::Normal),
            "voided" => Ok(InvoiceStatus::Voided),
            "red冲ed" => Ok(InvoiceStatus::Red冲ed),
            _ => Err(format!("无效的发票状态: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Invoice {
    pub id: Uuid,
    pub invoice_code: String,
    pub invoice_number: String,
    pub issue_date: NaiveDate,
    pub buyer_name: String,
    pub amount: Decimal,
    pub tax: Decimal,
    pub tax_rate: Decimal,
    pub status: InvoiceStatus,
    pub created_at: DateTime<Local>,
    pub updated_at: DateTime<Local>,
    pub original_invoice_id: Option<Uuid>,
    pub red冲_invoice_id: Option<Uuid>,
    pub is_red冲: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateInvoiceRequest {
    pub invoice_code: String,
    pub invoice_number: String,
    pub issue_date: NaiveDate,
    pub buyer_name: String,
    pub amount: Decimal,
    pub tax: Option<Decimal>,
    pub tax_rate: Option<Decimal>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchInvoiceRequest {
    pub buyer_name: Option<String>,
    pub start_date: Option<NaiveDate>,
    pub end_date: Option<NaiveDate>,
    pub status: Option<InvoiceStatus>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Red冲InvoiceRequest {
    pub invoice_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VoidInvoiceRequest {
    pub invoice_id: Uuid,
}
