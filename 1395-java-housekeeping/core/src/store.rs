use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::models::{Aunt, Customer, Order, RefundRequest};
use crate::errors::{HousekeepingError, Result};

#[derive(Debug, Clone, Default)]
pub struct InMemoryStore {
    aunts: Arc<RwLock<HashMap<Uuid, Aunt>>>,
    customers: Arc<RwLock<HashMap<Uuid, Customer>>>,
    orders: Arc<RwLock<HashMap<Uuid, Order>>>,
    refund_requests: Arc<RwLock<HashMap<Uuid, RefundRequest>>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub async fn create_aunt(&self, mut aunt: Aunt) -> Result<Aunt> {
        if aunt.name.is_empty() {
            return Err(HousekeepingError::NameEmpty);
        }
        if aunt.phone.is_empty() {
            return Err(HousekeepingError::PhoneNumberEmpty);
        }
        if aunt.skills.is_empty() {
            return Err(HousekeepingError::AtLeastOneSkillRequired);
        }
        if aunt.available_times.is_empty() {
            return Err(HousekeepingError::AtLeastOneAvailableTimeRequired);
        }
        if aunt.hourly_rate == 0 {
            return Err(HousekeepingError::InvalidHourlyRate);
        }

        let mut aunts = self.aunts.write().await;
        aunt.id = Uuid::new_v4();
        aunts.insert(aunt.id, aunt.clone());
        Ok(aunt)
    }

    pub async fn get_aunt(&self, id: &Uuid) -> Result<Aunt> {
        let aunts = self.aunts.read().await;
        aunts.get(id)
            .cloned()
            .ok_or_else(|| HousekeepingError::AuntNotFound(id.to_string()))
    }

    pub async fn get_all_aunts(&self) -> Result<Vec<Aunt>> {
        let aunts = self.aunts.read().await;
        Ok(aunts.values().cloned().collect())
    }

    pub async fn update_aunt(&self, id: &Uuid, updated: Aunt) -> Result<Aunt> {
        let mut aunts = self.aunts.write().await;
        let existing = aunts.get_mut(id)
            .ok_or_else(|| HousekeepingError::AuntNotFound(id.to_string()))?;
        
        existing.name = updated.name;
        existing.phone = updated.phone;
        existing.skills = updated.skills;
        existing.hourly_rate = updated.hourly_rate;
        existing.available_times = updated.available_times;
        existing.updated_at = chrono::Local::now();
        
        Ok(existing.clone())
    }

    pub async fn create_customer(&self, mut customer: Customer) -> Result<Customer> {
        if customer.name.is_empty() {
            return Err(HousekeepingError::NameEmpty);
        }
        if customer.phone.is_empty() {
            return Err(HousekeepingError::PhoneNumberEmpty);
        }
        if customer.address.is_empty() {
            return Err(HousekeepingError::AddressEmpty);
        }

        let mut customers = self.customers.write().await;
        customer.id = Uuid::new_v4();
        customers.insert(customer.id, customer.clone());
        Ok(customer)
    }

    pub async fn get_customer(&self, id: &Uuid) -> Result<Customer> {
        let customers = self.customers.read().await;
        customers.get(id)
            .cloned()
            .ok_or_else(|| HousekeepingError::CustomerNotFound(id.to_string()))
    }

    pub async fn get_all_customers(&self) -> Result<Vec<Customer>> {
        let customers = self.customers.read().await;
        Ok(customers.values().cloned().collect())
    }

    pub async fn create_order(&self, order: Order) -> Result<Order> {
        let mut orders = self.orders.write().await;
        orders.insert(order.id, order.clone());
        Ok(order)
    }

    pub async fn get_order(&self, id: &Uuid) -> Result<Order> {
        let orders = self.orders.read().await;
        orders.get(id)
            .cloned()
            .ok_or_else(|| HousekeepingError::OrderNotFound(id.to_string()))
    }

    pub async fn get_all_orders(&self) -> Result<Vec<Order>> {
        let orders = self.orders.read().await;
        Ok(orders.values().cloned().collect())
    }

    pub async fn update_order(&self, order: Order) -> Result<Order> {
        let mut orders = self.orders.write().await;
        let existing = orders.get_mut(&order.id)
            .ok_or_else(|| HousekeepingError::OrderNotFound(order.id.to_string()))?;
        
        *existing = order;
        Ok(existing.clone())
    }

    pub async fn get_aunt_orders(&self, aunt_id: &Uuid) -> Result<Vec<Order>> {
        let orders = self.orders.read().await;
        Ok(orders.values()
            .filter(|o| o.aunt_id.as_ref() == Some(aunt_id))
            .cloned()
            .collect())
    }

    pub async fn get_customer_orders(&self, customer_id: &Uuid) -> Result<Vec<Order>> {
        let orders = self.orders.read().await;
        Ok(orders.values()
            .filter(|o| o.customer_id == *customer_id)
            .cloned()
            .collect())
    }

    pub async fn create_refund_request(&self, refund: RefundRequest) -> Result<RefundRequest> {
        let mut refunds = self.refund_requests.write().await;
        refunds.insert(refund.id, refund.clone());
        Ok(refund)
    }

    pub async fn get_refund_request(&self, id: &Uuid) -> Result<RefundRequest> {
        let refunds = self.refund_requests.read().await;
        refunds.get(id)
            .cloned()
            .ok_or_else(|| HousekeepingError::RefundRequestNotFound(id.to_string()))
    }

    pub async fn get_all_refund_requests(&self) -> Result<Vec<RefundRequest>> {
        let refunds = self.refund_requests.read().await;
        Ok(refunds.values().cloned().collect())
    }

    pub async fn update_refund_request(&self, refund: RefundRequest) -> Result<RefundRequest> {
        let mut refunds = self.refund_requests.write().await;
        let existing = refunds.get_mut(&refund.id)
            .ok_or_else(|| HousekeepingError::RefundRequestNotFound(refund.id.to_string()))?;
        
        *existing = refund;
        Ok(existing.clone())
    }
}
