use std::collections::BTreeSet;
use std::sync::Arc;
use tokio::sync::Mutex;
use uuid::Uuid;
use chrono::{Datelike, Local};

use crate::models::{
    Aunt, AuntRecommendation, CreateAuntRequest, CreateCustomerRequest, CreateOrderRequest,
    Customer, MonthlyEarnings, Order, OrderStatus, RefundRequest, RequestRefundRequest,
    ApproveRefundRequest, RejectRefundRequest, SkillType, TimeSlot,
};
use crate::errors::{HousekeepingError, Result};
use crate::store::InMemoryStore;

#[derive(Debug, Clone)]
pub struct BookingService {
    store: InMemoryStore,
    booking_lock: Arc<Mutex<()>>,
}

impl BookingService {
    pub fn new(store: InMemoryStore) -> Self {
        Self {
            store,
            booking_lock: Arc::new(Mutex::new(())),
        }
    }

    pub fn store(&self) -> &InMemoryStore {
        &self.store
    }

    pub async fn create_aunt(&self, req: CreateAuntRequest) -> Result<Aunt> {
        let skills: BTreeSet<SkillType> = req.skills.into_iter().collect();
        let aunt = Aunt::new(
            req.name,
            req.phone,
            skills,
            req.hourly_rate,
            req.available_times,
        );
        self.store.create_aunt(aunt).await
    }

    pub async fn get_aunt(&self, id: &Uuid) -> Result<Aunt> {
        self.store.get_aunt(id).await
    }

    pub async fn get_all_aunts(&self) -> Result<Vec<Aunt>> {
        self.store.get_all_aunts().await
    }

    pub async fn create_customer(&self, req: CreateCustomerRequest) -> Result<Customer> {
        let customer = Customer::new(req.name, req.phone, req.address);
        self.store.create_customer(customer).await
    }

    pub async fn get_customer(&self, id: &Uuid) -> Result<Customer> {
        self.store.get_customer(id).await
    }

    pub async fn get_all_customers(&self) -> Result<Vec<Customer>> {
        self.store.get_all_customers().await
    }

    pub async fn recommend_aunts(
        &self,
        skill_type: SkillType,
        time_slot: &TimeSlot,
    ) -> Result<Vec<AuntRecommendation>> {
        if time_slot.duration_minutes() <= 0 {
            return Err(HousekeepingError::InvalidTimeSlot);
        }

        let all_aunts = self.store.get_all_aunts().await?;
        let mut recommendations = Vec::new();

        for aunt in all_aunts {
            if !aunt.has_skill(skill_type) {
                continue;
            }

            let is_available = aunt.is_available_at(
                time_slot.date,
                time_slot.start_time,
                time_slot.end_time,
            );

            let is_conflict = if aunt.id == Uuid::nil() {
                false
            } else {
                self.check_aunt_time_conflict(&aunt.id, time_slot).await?
            };

            recommendations.push(AuntRecommendation {
                aunt: aunt.clone(),
                hourly_rate: aunt.hourly_rate,
                is_available: is_available && !is_conflict,
            });
        }

        recommendations.sort_by(|a, b| a.hourly_rate.cmp(&b.hourly_rate));
        Ok(recommendations)
    }

    async fn check_aunt_time_conflict(
        &self,
        aunt_id: &Uuid,
        time_slot: &TimeSlot,
    ) -> Result<bool> {
        let aunt_orders = self.store.get_aunt_orders(aunt_id).await?;
        for order in aunt_orders {
            match order.status {
                OrderStatus::Confirmed | OrderStatus::InService => {
                    if order.time_slot.overlaps_with(time_slot) {
                        return Ok(true);
                    }
                }
                _ => continue,
            }
        }
        Ok(false)
    }

    pub async fn create_order(&self, req: CreateOrderRequest) -> Result<Order> {
        if req.time_slot.duration_minutes() <= 0 {
            return Err(HousekeepingError::InvalidTimeSlot);
        }

        let _lock = self.booking_lock.lock().await;

        self.store.get_customer(&req.customer_id).await?;

        let aunt = match req.aunt_id {
            Some(aunt_id) => {
                let aunt = self.store.get_aunt(&aunt_id).await?;
                if !aunt.has_skill(req.skill_type) {
                    return Err(HousekeepingError::AuntNoSkill);
                }
                if !aunt.is_available_at(
                    req.time_slot.date,
                    req.time_slot.start_time,
                    req.time_slot.end_time,
                ) {
                    return Err(HousekeepingError::AuntNotAvailable);
                }
                if self.check_aunt_time_conflict(&aunt.id, &req.time_slot).await? {
                    return Err(HousekeepingError::AuntTimeConflict);
                }
                aunt
            }
            None => {
                let recommendations = self
                    .recommend_aunts(req.skill_type, &req.time_slot)
                    .await?;
                let available_aunt = recommendations
                    .into_iter()
                    .find(|r| r.is_available)
                    .ok_or(HousekeepingError::NoAvailableAunt)?;
                available_aunt.aunt
            }
        };

        let order = Order::new(
            req.customer_id,
            Some(aunt.id),
            req.skill_type,
            req.time_slot,
            aunt.hourly_rate,
        );

        self.store.create_order(order).await
    }

    pub async fn get_order(&self, id: &Uuid) -> Result<Order> {
        self.store.get_order(id).await
    }

    pub async fn get_all_orders(&self) -> Result<Vec<Order>> {
        self.store.get_all_orders().await
    }

    pub async fn confirm_order(&self, order_id: &Uuid) -> Result<Order> {
        let mut order = self.store.get_order(order_id).await?;
        if order.status != OrderStatus::Pending {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::Pending.as_str().to_string(),
            });
        }
        order.status = OrderStatus::Confirmed;
        order.updated_at = Local::now();
        self.store.update_order(order).await
    }

    pub async fn start_service(&self, order_id: &Uuid) -> Result<Order> {
        let mut order = self.store.get_order(order_id).await?;
        if order.status != OrderStatus::Confirmed {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::Confirmed.as_str().to_string(),
            });
        }
        order.status = OrderStatus::InService;
        order.updated_at = Local::now();
        self.store.update_order(order).await
    }

    pub async fn complete_order(&self, order_id: &Uuid) -> Result<Order> {
        let mut order = self.store.get_order(order_id).await?;
        if order.status != OrderStatus::InService {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::InService.as_str().to_string(),
            });
        }
        order.status = OrderStatus::Completed;
        order.updated_at = Local::now();
        self.store.update_order(order).await
    }

    pub async fn cancel_order(&self, order_id: &Uuid) -> Result<Order> {
        let mut order = self.store.get_order(order_id).await?;
        if order.status != OrderStatus::Pending {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::Pending.as_str().to_string(),
            });
        }
        order.status = OrderStatus::Cancelled;
        order.updated_at = Local::now();
        self.store.update_order(order).await
    }

    pub async fn request_refund(&self, req: RequestRefundRequest) -> Result<Order> {
        if req.requested_percentage > 50 {
            return Err(HousekeepingError::RefundPercentageExceeded);
        }
        if req.requested_percentage == 0 || req.requested_percentage > 100 {
            return Err(HousekeepingError::InvalidRefundPercentage);
        }

        let mut order = self.store.get_order(&req.order_id).await?;
        if order.customer_id != req.customer_id {
            return Err(HousekeepingError::RefundNotOwner);
        }
        if order.status != OrderStatus::Completed {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::Completed.as_str().to_string(),
            });
        }
        if order.refund_request.is_some() {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: "无退款申请".to_string(),
            });
        }

        let refund_request = RefundRequest::new(
            req.order_id,
            req.customer_id,
            req.reason,
            req.requested_percentage,
        );

        self.store.create_refund_request(refund_request.clone()).await?;
        order.refund_request = Some(refund_request);
        order.status = OrderStatus::RefundRequested;
        order.updated_at = Local::now();
        self.store.update_order(order).await
    }

    pub async fn approve_refund(&self, req: ApproveRefundRequest) -> Result<Order> {
        if req.approved_percentage > 50 {
            return Err(HousekeepingError::RefundPercentageExceeded);
        }
        if req.approved_percentage == 0 || req.approved_percentage > 100 {
            return Err(HousekeepingError::InvalidRefundPercentage);
        }

        let mut refund = self.store.get_refund_request(&req.refund_request_id).await?;
        let mut order = self.store.get_order(&refund.order_id).await?;

        if order.status != OrderStatus::RefundRequested {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::RefundRequested.as_str().to_string(),
            });
        }

        refund.approved_percentage = Some(req.approved_percentage);
        refund.reviewed_at = Some(Local::now());

        let refund_amount = (order.total_amount as f64 * req.approved_percentage as f64 / 100.0) as u32;
        order.refund_amount = Some(refund_amount);
        order.refund_request = Some(refund.clone());
        order.status = OrderStatus::RefundApproved;
        order.updated_at = Local::now();

        self.store.update_refund_request(refund).await?;
        self.store.update_order(order).await
    }

    pub async fn reject_refund(&self, req: RejectRefundRequest) -> Result<Order> {
        let mut refund = self.store.get_refund_request(&req.refund_request_id).await?;
        let mut order = self.store.get_order(&refund.order_id).await?;

        if order.status != OrderStatus::RefundRequested {
            return Err(HousekeepingError::InvalidOrderStatus {
                current: order.status.as_str().to_string(),
                expected: OrderStatus::RefundRequested.as_str().to_string(),
            });
        }

        refund.reviewed_at = Some(Local::now());

        order.refund_request = Some(refund.clone());
        order.status = OrderStatus::RefundRejected;
        order.updated_at = Local::now();

        self.store.update_refund_request(refund).await?;
        self.store.update_order(order).await
    }

    pub async fn get_monthly_earnings(
        &self,
        aunt_id: &Uuid,
        year: i32,
        month: u32,
    ) -> Result<MonthlyEarnings> {
        self.store.get_aunt(aunt_id).await?;

        let orders = self.store.get_aunt_orders(aunt_id).await?;
        
        let mut total_billable_hours = 0u32;
        let mut total_earnings = 0u32;
        let mut order_count = 0u32;

        for order in orders {
            if order.status == OrderStatus::Completed 
                || order.status == OrderStatus::RefundApproved 
                || order.status == OrderStatus::RefundRejected {
                let order_date = order.time_slot.date;
                if order_date.year() == year && order_date.month() == month {
                    total_billable_hours += order.billable_hours;
                    total_earnings += order.actual_amount();
                    order_count += 1;
                }
            }
        }

        Ok(MonthlyEarnings {
            aunt_id: *aunt_id,
            year,
            month,
            total_billable_hours,
            total_earnings,
            order_count,
        })
    }

    pub async fn get_all_refund_requests(&self) -> Result<Vec<RefundRequest>> {
        self.store.get_all_refund_requests().await
    }
}
