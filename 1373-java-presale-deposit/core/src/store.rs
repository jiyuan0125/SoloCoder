use crate::models::{Order, PresaleActivity, Product};
use crate::errors::PresaleResult;
use parking_lot::RwLock;
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Default)]
pub struct InMemoryStore {
    products: RwLock<HashMap<Uuid, Product>>,
    activities: RwLock<HashMap<Uuid, PresaleActivity>>,
    orders: RwLock<HashMap<Uuid, Order>>,
    user_activity_orders: RwLock<HashMap<(String, Uuid), Uuid>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add_product(&self, product: Product) {
        self.products.write().insert(product.id, product);
    }

    pub fn get_product(&self, id: Uuid) -> Option<Product> {
        self.products.read().get(&id).cloned()
    }

    pub fn list_products(&self) -> Vec<Product> {
        self.products.read().values().cloned().collect()
    }

    pub fn add_activity(&self, activity: PresaleActivity) {
        self.activities.write().insert(activity.id, activity);
    }

    pub fn get_activity(&self, id: Uuid) -> Option<PresaleActivity> {
        self.activities.read().get(&id).cloned()
    }

    pub fn list_activities(&self) -> Vec<PresaleActivity> {
        self.activities.read().values().cloned().collect()
    }

    pub fn update_activity<F: FnOnce(&mut PresaleActivity)>(
        &self,
        id: Uuid,
        updater: F,
    ) -> PresaleResult<()> {
        let mut activities = self.activities.write();
        if let Some(activity) = activities.get_mut(&id) {
            updater(activity);
            Ok(())
        } else {
            Err(crate::errors::PresaleError::ActivityNotFound(id))
        }
    }

    pub fn add_order(&self, order: Order) -> PresaleResult<()> {
        let key = (order.user_id.clone(), order.activity_id);
        let mut user_activity_map = self.user_activity_orders.write();

        if user_activity_map.contains_key(&key) {
            return Err(crate::errors::PresaleError::DuplicateOrder);
        }

        user_activity_map.insert(key, order.id);
        self.orders.write().insert(order.id, order);
        Ok(())
    }

    pub fn get_order(&self, id: Uuid) -> Option<Order> {
        self.orders.read().get(&id).cloned()
    }

    pub fn list_orders(&self) -> Vec<Order> {
        self.orders.read().values().cloned().collect()
    }

    pub fn list_orders_by_user(&self, user_id: &str) -> Vec<Order> {
        self.orders
            .read()
            .values()
            .filter(|o| o.user_id == user_id)
            .cloned()
            .collect()
    }

    pub fn list_orders_by_activity(&self, activity_id: Uuid) -> Vec<Order> {
        self.orders
            .read()
            .values()
            .filter(|o| o.activity_id == activity_id)
            .cloned()
            .collect()
    }

    pub fn update_order<F: FnOnce(&mut Order)>(
        &self,
        id: Uuid,
        updater: F,
    ) -> PresaleResult<()> {
        let mut orders = self.orders.write();
        if let Some(order) = orders.get_mut(&id) {
            updater(order);
            Ok(())
        } else {
            Err(crate::errors::PresaleError::OrderNotFound(id))
        }
    }

    pub fn has_user_ordered_activity(&self, user_id: &str, activity_id: Uuid) -> bool {
        self.user_activity_orders
            .read()
            .contains_key(&(user_id.to_string(), activity_id))
    }
}
