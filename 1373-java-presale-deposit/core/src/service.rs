use crate::errors::{PresaleError, PresaleResult};
use crate::models::{
    ActivityStatus, ActivitySummary, Order, OrderDetail, OrderStatus, PresaleActivity,
    PresaleConfig, Product,
};
use crate::store::InMemoryStore;
use chrono::{DateTime, Utc};
use uuid::Uuid;

pub const FINAL_PAYMENT_DEFAULT_DEADLINE_HOURS: i64 = 72;

pub struct PresaleService {
    store: InMemoryStore,
}

impl PresaleService {
    pub fn new() -> Self {
        Self {
            store: InMemoryStore::new(),
        }
    }

    pub fn store(&self) -> &InMemoryStore {
        &self.store
    }

    pub fn create_product(&self, name: String, original_price: u64) -> Product {
        let product = Product {
            id: Uuid::new_v4(),
            name,
            original_price,
        };
        self.store.add_product(product.clone());
        product
    }

    pub fn get_product(&self, id: Uuid) -> PresaleResult<Product> {
        self.store
            .get_product(id)
            .ok_or(PresaleError::ProductNotFound(id))
    }

    pub fn list_products(&self) -> Vec<Product> {
        self.store.list_products()
    }

    pub fn create_activity(
        &self,
        product_id: Uuid,
        config: PresaleConfig,
        max_participants: u32,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
        final_payment_deadline_hours: Option<i64>,
    ) -> PresaleResult<PresaleActivity> {
        if start_time >= end_time {
            return Err(PresaleError::InvalidConfig(
                "Start time must be before end time".to_string(),
            ));
        }
        if max_participants == 0 {
            return Err(PresaleError::InvalidConfig(
                "Max participants must be greater than 0".to_string(),
            ));
        }
        if config.inflation_rate == 0 {
            return Err(PresaleError::InvalidConfig(
                "Inflation rate must be greater than 0".to_string(),
            ));
        }
        if config.deposit_amount == 0 {
            return Err(PresaleError::InvalidConfig(
                "Deposit amount must be greater than 0".to_string(),
            ));
        }

        let product = self.get_product(product_id)?;

        let deadline_hours = final_payment_deadline_hours.unwrap_or(FINAL_PAYMENT_DEFAULT_DEADLINE_HOURS);
        if deadline_hours <= 0 {
            return Err(PresaleError::InvalidConfig(
                "Final payment deadline hours must be greater than 0".to_string(),
            ));
        }

        let activity = PresaleActivity::new(
            &product,
            config,
            max_participants,
            start_time,
            end_time,
            deadline_hours * 3600,
        );

        self.store.add_activity(activity.clone());
        Ok(activity)
    }

    pub fn start_activity(&self, activity_id: Uuid) -> PresaleResult<()> {
        let activity = self
            .store
            .get_activity(activity_id)
            .ok_or(PresaleError::ActivityNotFound(activity_id))?;

        if activity.status != ActivityStatus::Draft {
            return Err(PresaleError::InvalidActivityStatus(activity.status));
        }

        self.store.update_activity(activity_id, |a| {
            a.status = ActivityStatus::Active;
        })
    }

    pub fn end_activity(&self, activity_id: Uuid) -> PresaleResult<()> {
        let activity = self
            .store
            .get_activity(activity_id)
            .ok_or(PresaleError::ActivityNotFound(activity_id))?;

        if activity.status != ActivityStatus::Active {
            return Err(PresaleError::InvalidActivityStatus(activity.status));
        }

        self.store.update_activity(activity_id, |a| {
            a.status = ActivityStatus::Ended;
        })
    }

    pub fn cancel_activity(&self, activity_id: Uuid, now: DateTime<Utc>) -> PresaleResult<()> {
        let activity = self
            .store
            .get_activity(activity_id)
            .ok_or(PresaleError::ActivityNotFound(activity_id))?;

        if activity.status == ActivityStatus::Cancelled {
            return Ok(());
        }

        self.store.update_activity(activity_id, |a| {
            a.status = ActivityStatus::Cancelled;
        })?;

        let orders = self.store.list_orders_by_activity(activity_id);
        for order in orders {
            if order.status == OrderStatus::DepositPaid {
                self.store.update_order(order.id, |o| {
                    o.status = OrderStatus::Refunded;
                    o.closed_time = Some(now);
                })?;
            }
        }

        Ok(())
    }

    pub fn get_activity(&self, activity_id: Uuid) -> PresaleResult<ActivitySummary> {
        self.store
            .get_activity(activity_id)
            .map(|a| ActivitySummary::from(&a))
            .ok_or(PresaleError::ActivityNotFound(activity_id))
    }

    pub fn list_activities(&self) -> Vec<ActivitySummary> {
        self.store
            .list_activities()
            .iter()
            .map(ActivitySummary::from)
            .collect()
    }

    pub fn pay_deposit(
        &self,
        user_id: String,
        activity_id: Uuid,
        now: DateTime<Utc>,
    ) -> PresaleResult<Order> {
        let mut activity = self
            .store
            .get_activity(activity_id)
            .ok_or(PresaleError::ActivityNotFound(activity_id))?;

        if !activity.is_active(now) {
            return Err(PresaleError::ActivityNotActive);
        }

        if !activity.has_capacity() {
            return Err(PresaleError::ActivityFull);
        }

        if self.store.has_user_ordered_activity(&user_id, activity_id) {
            return Err(PresaleError::DuplicateOrder);
        }

        let order = Order::new(user_id, &activity, now);

        self.store.add_order(order.clone())?;

        self.store.update_activity(activity_id, |a| {
            a.current_participants += 1;
        })?;

        activity.current_participants += 1;
        if !activity.has_capacity() {
            self.store.update_activity(activity_id, |a| {
                if a.current_participants >= a.max_participants {
                    a.status = ActivityStatus::Ended;
                }
            })?;
        }

        Ok(order)
    }

    pub fn pay_final(
        &self,
        order_id: Uuid,
        paid_amount: u64,
        now: DateTime<Utc>,
    ) -> PresaleResult<Order> {
        let order = self
            .store
            .get_order(order_id)
            .ok_or(PresaleError::OrderNotFound(order_id))?;

        if order.status != OrderStatus::DepositPaid {
            return Err(PresaleError::InvalidOrderStatus(order.status));
        }

        if order.is_final_payment_overdue(now) {
            return Err(PresaleError::FinalPaymentOverdue);
        }

        if paid_amount != order.final_amount {
            return Err(PresaleError::FinalPaymentAmountMismatch);
        }

        self.store.update_order(order_id, |o| {
            o.status = OrderStatus::FinalPaid;
            o.final_paid_time = Some(now);
        })?;

        self.store.get_order(order_id).ok_or(PresaleError::OrderNotFound(order_id))
    }

    pub fn get_order(&self, order_id: Uuid) -> PresaleResult<OrderDetail> {
        self.store
            .get_order(order_id)
            .map(|o| OrderDetail::from(&o))
            .ok_or(PresaleError::OrderNotFound(order_id))
    }

    pub fn list_orders_by_user(&self, user_id: &str) -> Vec<OrderDetail> {
        self.store
            .list_orders_by_user(user_id)
            .iter()
            .map(OrderDetail::from)
            .collect()
    }

    pub fn process_overdue_orders(&self, now: DateTime<Utc>) -> Vec<Uuid> {
        let orders = self.store.list_orders();
        let mut expired = Vec::new();

        for order in orders {
            if order.status == OrderStatus::DepositPaid && order.is_final_payment_overdue(now) {
                let _ = self.store.update_order(order.id, |o| {
                    o.status = OrderStatus::Expired;
                    o.closed_time = Some(now);
                });
                expired.push(order.id);
            }
        }

        expired
    }

    pub fn complete_order(&self, order_id: Uuid, now: DateTime<Utc>) -> PresaleResult<()> {
        let order = self
            .store
            .get_order(order_id)
            .ok_or(PresaleError::OrderNotFound(order_id))?;

        if order.status != OrderStatus::FinalPaid {
            return Err(PresaleError::InvalidOrderStatus(order.status));
        }

        self.store.update_order(order_id, |o| {
            o.status = OrderStatus::Completed;
            o.closed_time = Some(now);
        })
    }
}

impl Default for PresaleService {
    fn default() -> Self {
        Self::new()
    }
}
