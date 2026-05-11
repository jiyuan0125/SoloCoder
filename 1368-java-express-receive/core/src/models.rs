use chrono::{DateTime, Utc, NaiveDate};
use serde::{Serialize, Deserialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PackageStatus {
    InStock,
    PickedUp,
    Stranded,
    Returned,
    Abnormal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Package {
    pub id: Uuid,
    pub tracking_number: String,
    pub recipient_phone: String,
    pub recipient_name: Option<String>,
    pub pickup_code: String,
    pub courier_name: String,
    pub courier_phone: String,
    pub express_company: String,
    pub status: PackageStatus,
    pub created_at: DateTime<Utc>,
    pub picked_up_at: Option<DateTime<Utc>>,
    pub is_proxy: bool,
    pub proxy_phone: Option<String>,
    pub proxy_name: Option<String>,
    pub id_last_four: Option<String>,
    pub failed_attempts: u32,
    pub locked_until: Option<DateTime<Utc>>,
    pub stranded_reminded: bool,
    pub return_notified: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProxyRecord {
    pub id: Uuid,
    pub package_id: Uuid,
    pub proxy_phone: String,
    pub proxy_name: Option<String>,
    pub recipient_phone: String,
    pub notified_at: DateTime<Utc>,
    pub feedback: Option<ProxyFeedback>,
    pub feedback_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum ProxyFeedback {
    Confirmed,
    Rejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DailyCodeCounter {
    pub date: NaiveDate,
    pub counter: u32,
    pub letter_prefix: char,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Notification {
    pub id: Uuid,
    pub package_id: Uuid,
    pub recipient_phone: String,
    pub pickup_codes: Vec<String>,
    pub notification_type: NotificationType,
    pub sent_at: DateTime<Utc>,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum NotificationType {
    Arrival,
    Stranded,
    Return,
    ProxyRequested,
    ProxyConfirmed,
    ProxyRejected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InboundRequest {
    pub tracking_number: String,
    pub recipient_phone: String,
    pub recipient_name: Option<String>,
    pub courier_name: String,
    pub courier_phone: String,
    pub express_company: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PickupRequest {
    pub pickup_code: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ManualVerifyRequest {
    pub phone: String,
    pub id_last_four: String,
    pub package_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProxyRequest {
    pub package_id: Uuid,
    pub proxy_phone: String,
    pub proxy_name: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProxyFeedbackRequest {
    pub package_id: Uuid,
    pub feedback: ProxyFeedback,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InboundResult {
    pub package: Package,
    pub notification: Option<Notification>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PickupResult {
    pub package: Package,
    pub success: bool,
    pub message: String,
    pub locked: bool,
    pub locked_until: Option<DateTime<Utc>>,
}
