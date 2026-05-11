use chrono::{DateTime, Utc, Duration};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

pub type Amount = u64;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum ActivityStatus {
    Draft,
    Active,
    Ended,
    Cancelled,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum OrderStatus {
    DepositPaid,
    FinalPaid,
    Completed,
    Expired,
    Refunded,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub id: Uuid,
    pub name: String,
    pub original_price: Amount,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PresaleConfig {
    pub deposit_amount: Amount,
    pub inflation_rate: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PresaleActivity {
    pub id: Uuid,
    pub product_id: Uuid,
    pub product_name: String,
    pub original_price: Amount,
    pub config: PresaleConfig,
    pub max_participants: u32,
    pub current_participants: u32,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub status: ActivityStatus,
    pub final_payment_deadline_offset_seconds: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub user_id: String,
    pub activity_id: Uuid,
    pub product_id: Uuid,
    pub product_name: String,
    pub original_price: Amount,
    pub deposit_amount: Amount,
    pub inflated_amount: Amount,
    pub final_amount: Amount,
    pub status: OrderStatus,
    pub deposit_paid_time: Option<DateTime<Utc>>,
    pub final_paid_time: Option<DateTime<Utc>>,
    pub final_payment_deadline: Option<DateTime<Utc>>,
    pub closed_time: Option<DateTime<Utc>>,
}

impl PresaleActivity {
    pub fn new(
        product: &Product,
        config: PresaleConfig,
        max_participants: u32,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
        final_payment_deadline_offset_seconds: i64,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            product_id: product.id,
            product_name: product.name.clone(),
            original_price: product.original_price,
            config,
            max_participants,
            current_participants: 0,
            start_time,
            end_time,
            status: ActivityStatus::Draft,
            final_payment_deadline_offset_seconds,
        }
    }

    pub fn inflated_amount(&self) -> Amount {
        self.config.deposit_amount.saturating_mul(self.config.inflation_rate as u64)
    }

    pub fn final_amount(&self) -> Amount {
        let inflated = self.inflated_amount();
        if inflated >= self.original_price {
            0
        } else {
            self.original_price - inflated
        }
    }

    pub fn has_capacity(&self) -> bool {
        self.current_participants < self.max_participants
    }

    pub fn is_active(&self, now: DateTime<Utc>) -> bool {
        self.status == ActivityStatus::Active && now >= self.start_time && now < self.end_time
    }
}

impl Order {
    pub fn new(
        user_id: String,
        activity: &PresaleActivity,
        deposit_paid_time: DateTime<Utc>,
    ) -> Self {
        let inflated_amount = activity.inflated_amount();
        let final_amount = activity.final_amount();

        let final_payment_deadline = activity
            .end_time
            .checked_add_signed(Duration::seconds(activity.final_payment_deadline_offset_seconds));

        Self {
            id: Uuid::new_v4(),
            user_id,
            activity_id: activity.id,
            product_id: activity.product_id,
            product_name: activity.product_name.clone(),
            original_price: activity.original_price,
            deposit_amount: activity.config.deposit_amount,
            inflated_amount,
            final_amount,
            status: OrderStatus::DepositPaid,
            deposit_paid_time: Some(deposit_paid_time),
            final_paid_time: None,
            final_payment_deadline,
            closed_time: None,
        }
    }

    pub fn is_final_payment_overdue(&self, now: DateTime<Utc>) -> bool {
        match self.final_payment_deadline {
            Some(deadline) => now > deadline,
            None => false,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ActivitySummary {
    pub id: Uuid,
    pub product_name: String,
    pub original_price: Amount,
    pub deposit_amount: Amount,
    pub inflated_amount: Amount,
    pub final_amount: Amount,
    pub inflation_rate: u32,
    pub max_participants: u32,
    pub current_participants: u32,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub status: ActivityStatus,
}

impl From<&PresaleActivity> for ActivitySummary {
    fn from(a: &PresaleActivity) -> Self {
        Self {
            id: a.id,
            product_name: a.product_name.clone(),
            original_price: a.original_price,
            deposit_amount: a.config.deposit_amount,
            inflated_amount: a.inflated_amount(),
            final_amount: a.final_amount(),
            inflation_rate: a.config.inflation_rate,
            max_participants: a.max_participants,
            current_participants: a.current_participants,
            start_time: a.start_time,
            end_time: a.end_time,
            status: a.status,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderDetail {
    pub id: Uuid,
    pub user_id: String,
    pub activity_id: Uuid,
    pub product_name: String,
    pub original_price: Amount,
    pub deposit_amount: Amount,
    pub inflated_amount: Amount,
    pub final_amount: Amount,
    pub status: OrderStatus,
    pub deposit_paid_time: Option<DateTime<Utc>>,
    pub final_paid_time: Option<DateTime<Utc>>,
    pub final_payment_deadline: Option<DateTime<Utc>>,
}

impl From<&Order> for OrderDetail {
    fn from(o: &Order) -> Self {
        Self {
            id: o.id,
            user_id: o.user_id.clone(),
            activity_id: o.activity_id,
            product_name: o.product_name.clone(),
            original_price: o.original_price,
            deposit_amount: o.deposit_amount,
            inflated_amount: o.inflated_amount,
            final_amount: o.final_amount,
            status: o.status,
            deposit_paid_time: o.deposit_paid_time,
            final_paid_time: o.final_paid_time,
            final_payment_deadline: o.final_payment_deadline,
        }
    }
}
