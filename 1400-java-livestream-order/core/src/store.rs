use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use chrono::Utc;
use uuid::Uuid;
use crate::models::{Order, Product, User, LiveStatus};
use crate::errors::SystemError;

#[derive(Debug, Clone)]
pub struct InMemoryStore {
    inner: Arc<Mutex<StoreState>>,
}

#[derive(Debug)]
struct StoreState {
    products: HashMap<Uuid, Product>,
    orders: HashMap<Uuid, Order>,
    users: HashMap<Uuid, User>,
    live_status: LiveStatus,
}

impl InMemoryStore {
    pub fn new() -> Self {
        InMemoryStore {
            inner: Arc::new(Mutex::new(StoreState {
                products: HashMap::new(),
                orders: HashMap::new(),
                users: HashMap::new(),
                live_status: LiveStatus::NotStarted,
            })),
        }
    }

    pub fn get_product(&self, id: Uuid) -> Result<Product, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.products.get(&id).cloned().ok_or_else(|| SystemError::ProductNotFound(id.to_string()))
    }

    pub fn list_products(&self) -> Result<Vec<Product>, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        Ok(state.products.values().cloned().collect())
    }

    pub fn insert_product(&self, product: Product) -> Result<(), SystemError> {
        let mut state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.products.insert(product.id, product);
        Ok(())
    }

    pub fn update_product<F>(&self, id: Uuid, f: F) -> Result<Product, SystemError>
    where
        F: FnOnce(&mut Product),
    {
        let mut state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        let product = state.products.get_mut(&id).ok_or_else(|| SystemError::ProductNotFound(id.to_string()))?;
        f(product);
        product.updated_at = Utc::now();
        Ok(product.clone())
    }

    pub fn get_order(&self, id: Uuid) -> Result<Order, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.orders.get(&id).cloned().ok_or_else(|| SystemError::OrderNotFound(id.to_string()))
    }

    pub fn list_orders(&self) -> Result<Vec<Order>, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        Ok(state.orders.values().cloned().collect())
    }

    pub fn list_orders_by_user(&self, user_id: Uuid) -> Result<Vec<Order>, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        Ok(state.orders.values().filter(|o| o.user_id == user_id).cloned().collect())
    }

    pub fn list_orders_by_product(&self, product_id: Uuid) -> Result<Vec<Order>, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        Ok(state.orders.values().filter(|o| o.product_id == product_id).cloned().collect())
    }

    pub fn insert_order(&self, order: Order) -> Result<(), SystemError> {
        let mut state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.orders.insert(order.id, order);
        Ok(())
    }

    pub fn update_order<F>(&self, id: Uuid, f: F) -> Result<Order, SystemError>
    where
        F: FnOnce(&mut Order),
    {
        let mut state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        let order = state.orders.get_mut(&id).ok_or_else(|| SystemError::OrderNotFound(id.to_string()))?;
        f(order);
        Ok(order.clone())
    }

    pub fn get_user(&self, id: Uuid) -> Result<User, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.users.get(&id).cloned().ok_or_else(|| SystemError::UserNotFound(id.to_string()))
    }

    pub fn list_users(&self) -> Result<Vec<User>, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        Ok(state.users.values().cloned().collect())
    }

    pub fn insert_user(&self, user: User) -> Result<(), SystemError> {
        let mut state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.users.insert(user.id, user);
        Ok(())
    }

    pub fn get_live_status(&self) -> Result<LiveStatus, SystemError> {
        let state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        Ok(state.live_status)
    }

    pub fn set_live_status(&self, status: LiveStatus) -> Result<(), SystemError> {
        let mut state = self.inner.lock().map_err(|e| SystemError::Internal(e.to_string()))?;
        state.live_status = status;
        Ok(())
    }
}

impl Default for InMemoryStore {
    fn default() -> Self {
        Self::new()
    }
}
