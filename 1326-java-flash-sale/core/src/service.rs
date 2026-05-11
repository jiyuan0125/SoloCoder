use std::collections::HashMap;
use std::sync::Arc;

use chrono::Utc;
use tokio::sync::Mutex;

use crate::error::{FlashSaleError, Result};
use crate::models::{ActivityStatistics, FlashSaleActivity, Order, OrderStatus};

const ORDER_TIMEOUT_SECONDS: u32 = 300;

pub struct FlashSaleService {
    activities: Mutex<HashMap<String, FlashSaleActivity>>,
    orders: Mutex<HashMap<String, Order>>,
    stock_tracker: Mutex<HashMap<String, u32>>,
    user_activity_tracker: Mutex<HashMap<(String, String), bool>>,
}

impl FlashSaleService {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            activities: Mutex::new(HashMap::new()),
            orders: Mutex::new(HashMap::new()),
            stock_tracker: Mutex::new(HashMap::new()),
            user_activity_tracker: Mutex::new(HashMap::new()),
        })
    }

    pub async fn create_activity(
        &self,
        product_name: String,
        flash_price: f64,
        stock: u32,
        start_time: chrono::DateTime<Utc>,
        end_time: chrono::DateTime<Utc>,
    ) -> FlashSaleActivity {
        let activity = FlashSaleActivity::new(product_name, flash_price, stock, start_time, end_time);
        
        let mut activities = self.activities.lock().await;
        let mut stock_tracker = self.stock_tracker.lock().await;
        
        activities.insert(activity.id.clone(), activity.clone());
        stock_tracker.insert(activity.id.clone(), stock);
        
        activity
    }

    pub async fn get_activity(&self, activity_id: &str) -> Result<FlashSaleActivity> {
        let activities = self.activities.lock().await;
        activities
            .get(activity_id)
            .cloned()
            .ok_or(FlashSaleError::ActivityNotFound)
    }

    pub async fn list_activities(&self) -> Vec<FlashSaleActivity> {
        let activities = self.activities.lock().await;
        activities.values().cloned().collect()
    }

    pub async fn create_order(&self, activity_id: String, user_id: String) -> Result<Order> {
        let now = Utc::now();
        let mut orders = self.orders.lock().await;
        let mut stock_tracker = self.stock_tracker.lock().await;
        let user_tracker = self.user_activity_tracker.lock().await;
        let activities = self.activities.lock().await;

        let activity = activities
            .get(&activity_id)
            .cloned()
            .ok_or(FlashSaleError::ActivityNotFound)?;

        if now < activity.start_time {
            return Err(FlashSaleError::ActivityNotStarted);
        }

        if now > activity.end_time {
            return Err(FlashSaleError::ActivityEnded);
        }

        if user_tracker.get(&(user_id.clone(), activity_id.clone())).copied().unwrap_or(false) {
            return Err(FlashSaleError::UserAlreadyPurchased);
        }

        let current_stock = stock_tracker
            .get(&activity_id)
            .copied()
            .ok_or(FlashSaleError::ActivityNotFound)?;

        if current_stock == 0 {
            return Err(FlashSaleError::OutOfStock);
        }

        let order = Order::new(activity_id.clone(), user_id.clone(), ORDER_TIMEOUT_SECONDS);
        orders.insert(order.id.clone(), order.clone());

        *stock_tracker.get_mut(&activity_id).unwrap() -= 1;

        Ok(order)
    }

    pub async fn pay_order(&self, order_id: String) -> Result<Order> {
        let mut orders = self.orders.lock().await;
        let mut user_tracker = self.user_activity_tracker.lock().await;

        let order = orders
            .get_mut(&order_id)
            .ok_or(FlashSaleError::OrderNotFound)?;

        if order.status != OrderStatus::Created {
            return Err(FlashSaleError::InvalidOrderStatus);
        }

        order.status = OrderStatus::Paid;

        let key = (order.user_id.clone(), order.activity_id.clone());
        user_tracker.insert(key, true);

        Ok(order.clone())
    }

    pub async fn cancel_order(&self, order_id: String) -> Result<Order> {
        let mut orders = self.orders.lock().await;
        let mut stock_tracker = self.stock_tracker.lock().await;

        let order = orders
            .get_mut(&order_id)
            .ok_or(FlashSaleError::OrderNotFound)?;

        if order.status != OrderStatus::Created {
            return Err(FlashSaleError::InvalidOrderStatus);
        }

        order.status = OrderStatus::Cancelled;

        if let Some(stock) = stock_tracker.get_mut(&order.activity_id) {
            *stock += 1;
        }

        Ok(order.clone())
    }

    pub async fn get_order(&self, order_id: &str) -> Result<Order> {
        let orders = self.orders.lock().await;
        orders
            .get(order_id)
            .cloned()
            .ok_or(FlashSaleError::OrderNotFound)
    }

    pub async fn list_orders(&self) -> Vec<Order> {
        let orders = self.orders.lock().await;
        orders.values().cloned().collect()
    }

    pub async fn process_expired_orders(&self) {
        let now = Utc::now();
        let mut orders = self.orders.lock().await;
        let mut stock_tracker = self.stock_tracker.lock().await;

        for order in orders.values_mut() {
            if order.status == OrderStatus::Created && now > order.expires_at {
                order.status = OrderStatus::Cancelled;
                if let Some(stock) = stock_tracker.get_mut(&order.activity_id) {
                    *stock += 1;
                }
            }
        }
    }

    pub async fn process_ended_activities(&self) {
        let now = Utc::now();
        let activities = self.activities.lock().await;
        let mut orders = self.orders.lock().await;

        let ended_activity_ids: Vec<String> = activities
            .values()
            .filter(|a| now > a.end_time)
            .map(|a| a.id.clone())
            .collect();

        for order in orders.values_mut() {
            if order.status == OrderStatus::Created && ended_activity_ids.contains(&order.activity_id) {
                order.status = OrderStatus::Ended;
            }
        }
    }

    pub async fn get_statistics(&self, activity_id: &str) -> Result<ActivityStatistics> {
        let activities = self.activities.lock().await;
        let orders = self.orders.lock().await;
        let stock_tracker = self.stock_tracker.lock().await;

        let activity = activities
            .get(activity_id)
            .cloned()
            .ok_or(FlashSaleError::ActivityNotFound)?;

        let mut total_orders = 0u32;
        let mut sold_count = 0u32;

        for order in orders.values() {
            if order.activity_id == activity_id {
                total_orders += 1;
                if order.status == OrderStatus::Paid {
                    sold_count += 1;
                }
            }
        }

        let available_stock = stock_tracker.get(activity_id).copied().unwrap_or(0);

        Ok(ActivityStatistics {
            activity_id: activity_id.to_string(),
            product_name: activity.product_name,
            total_orders,
            sold_count,
            available_stock,
        })
    }
}
