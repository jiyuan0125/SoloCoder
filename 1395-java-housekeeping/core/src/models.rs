use chrono::{DateTime, Datelike, Local, NaiveDate, NaiveTime, Weekday};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::BTreeSet;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize, PartialOrd, Ord)]
pub enum SkillType {
    Cleaning,
    Cooking,
    Nanny,
    ElderCare,
    PetCare,
    Tutoring,
}

impl SkillType {
    pub fn as_str(&self) -> &'static str {
        match self {
            SkillType::Cleaning => "保洁",
            SkillType::Cooking => "做饭",
            SkillType::Nanny => "月嫂",
            SkillType::ElderCare => "老人护理",
            SkillType::PetCare => "宠物照料",
            SkillType::Tutoring => "家教",
        }
    }

    pub fn all() -> Vec<SkillType> {
        vec![
            SkillType::Cleaning,
            SkillType::Cooking,
            SkillType::Nanny,
            SkillType::ElderCare,
            SkillType::PetCare,
            SkillType::Tutoring,
        ]
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum OrderStatus {
    Pending,
    Confirmed,
    InService,
    Completed,
    Cancelled,
    RefundRequested,
    RefundApproved,
    RefundRejected,
}

impl OrderStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            OrderStatus::Pending => "待确认",
            OrderStatus::Confirmed => "已确认",
            OrderStatus::InService => "服务中",
            OrderStatus::Completed => "已完成",
            OrderStatus::Cancelled => "已取消",
            OrderStatus::RefundRequested => "退款申请中",
            OrderStatus::RefundApproved => "退款已批准",
            OrderStatus::RefundRejected => "退款已拒绝",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct WeeklyAvailableTime {
    pub weekday: Weekday,
    pub start_time: NaiveTime,
    pub end_time: NaiveTime,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Aunt {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
    pub skills: BTreeSet<SkillType>,
    pub hourly_rate: u32,
    pub available_times: Vec<WeeklyAvailableTime>,
    pub created_at: DateTime<Local>,
    pub updated_at: DateTime<Local>,
}

impl Aunt {
    pub fn new(
        name: String,
        phone: String,
        skills: BTreeSet<SkillType>,
        hourly_rate: u32,
        available_times: Vec<WeeklyAvailableTime>,
    ) -> Self {
        let now = Local::now();
        Self {
            id: Uuid::new_v4(),
            name,
            phone,
            skills,
            hourly_rate,
            available_times,
            created_at: now,
            updated_at: now,
        }
    }

    pub fn has_skill(&self, skill: SkillType) -> bool {
        self.skills.contains(&skill)
    }

    pub fn is_available_at(&self, date: NaiveDate, start_time: NaiveTime, end_time: NaiveTime) -> bool {
        let weekday = date.weekday();
        self.available_times.iter().any(|at| {
            at.weekday == weekday && start_time >= at.start_time && end_time <= at.end_time
        })
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Customer {
    pub id: Uuid,
    pub name: String,
    pub phone: String,
    pub address: String,
    pub created_at: DateTime<Local>,
}

impl Customer {
    pub fn new(name: String, phone: String, address: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            phone,
            address,
            created_at: Local::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimeSlot {
    pub date: NaiveDate,
    pub start_time: NaiveTime,
    pub end_time: NaiveTime,
}

impl TimeSlot {
    pub fn new(date: NaiveDate, start_time: NaiveTime, end_time: NaiveTime) -> Self {
        Self {
            date,
            start_time,
            end_time,
        }
    }

    pub fn duration_minutes(&self) -> i64 {
        (self.end_time - self.start_time).num_minutes()
    }

    pub fn billable_hours(&self) -> u32 {
        let minutes = self.duration_minutes();
        if minutes <= 0 {
            return 0;
        }
        ((minutes as f64) / 60.0).ceil() as u32
    }

    pub fn overlaps_with(&self, other: &TimeSlot) -> bool {
        if self.date != other.date {
            return false;
        }
        !(self.end_time <= other.start_time || self.start_time >= other.end_time)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RefundRequest {
    pub id: Uuid,
    pub order_id: Uuid,
    pub customer_id: Uuid,
    pub reason: String,
    pub requested_percentage: u32,
    pub approved_percentage: Option<u32>,
    pub created_at: DateTime<Local>,
    pub reviewed_at: Option<DateTime<Local>>,
}

impl RefundRequest {
    pub fn new(
        order_id: Uuid,
        customer_id: Uuid,
        reason: String,
        requested_percentage: u32,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            order_id,
            customer_id,
            reason,
            requested_percentage,
            approved_percentage: None,
            created_at: Local::now(),
            reviewed_at: None,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub customer_id: Uuid,
    pub aunt_id: Option<Uuid>,
    pub skill_type: SkillType,
    pub time_slot: TimeSlot,
    pub status: OrderStatus,
    pub hourly_rate: u32,
    pub billable_hours: u32,
    pub total_amount: u32,
    pub refund_amount: Option<u32>,
    pub refund_request: Option<RefundRequest>,
    pub created_at: DateTime<Local>,
    pub updated_at: DateTime<Local>,
}

impl Order {
    pub fn new(
        customer_id: Uuid,
        aunt_id: Option<Uuid>,
        skill_type: SkillType,
        time_slot: TimeSlot,
        hourly_rate: u32,
    ) -> Self {
        let billable_hours = time_slot.billable_hours();
        let total_amount = billable_hours * hourly_rate;
        let now = Local::now();
        Self {
            id: Uuid::new_v4(),
            customer_id,
            aunt_id,
            skill_type,
            time_slot,
            status: OrderStatus::Pending,
            hourly_rate,
            billable_hours,
            total_amount,
            refund_amount: None,
            refund_request: None,
            created_at: now,
            updated_at: now,
        }
    }

    pub fn actual_amount(&self) -> u32 {
        self.total_amount - self.refund_amount.unwrap_or(0)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuntRecommendation {
    pub aunt: Aunt,
    pub hourly_rate: u32,
    pub is_available: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MonthlyEarnings {
    pub aunt_id: Uuid,
    pub year: i32,
    pub month: u32,
    pub total_billable_hours: u32,
    pub total_earnings: u32,
    pub order_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateAuntRequest {
    pub name: String,
    pub phone: String,
    pub skills: Vec<SkillType>,
    pub hourly_rate: u32,
    pub available_times: Vec<WeeklyAvailableTime>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateCustomerRequest {
    pub name: String,
    pub phone: String,
    pub address: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub customer_id: Uuid,
    pub skill_type: SkillType,
    pub time_slot: TimeSlot,
    pub aunt_id: Option<Uuid>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestRefundRequest {
    pub order_id: Uuid,
    pub customer_id: Uuid,
    pub reason: String,
    pub requested_percentage: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApproveRefundRequest {
    pub refund_request_id: Uuid,
    pub approved_percentage: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RejectRefundRequest {
    pub refund_request_id: Uuid,
    pub reason: String,
}
