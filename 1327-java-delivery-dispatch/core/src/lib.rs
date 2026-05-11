use serde::{Deserialize, Serialize};
use std::collections::VecDeque;
use std::sync::Arc;
use tokio::sync::Mutex;
use uuid::Uuid;
use chrono::{DateTime, Utc};

pub const DEFAULT_MAX_ORDERS_PER_RIDER: usize = 3;
pub const DEFAULT_MAX_DETOUR_DISTANCE_KM: f64 = 2.0;

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct Location {
    pub latitude: f64,
    pub longitude: f64,
}

impl Location {
    pub fn distance_to(&self, other: &Location) -> f64 {
        let lat1 = self.latitude.to_radians();
        let lat2 = other.latitude.to_radians();
        let delta_lat = (other.latitude - self.latitude).to_radians();
        let delta_lng = (other.longitude - self.longitude).to_radians();

        let a = (delta_lat / 2.0).sin().powi(2)
            + lat1.cos() * lat2.cos() * (delta_lng / 2.0).sin().powi(2);
        let c = 2.0 * a.sqrt().atan2((1.0 - a).sqrt());

        6371.0 * c
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum RiderStatus {
    Idle,
    Delivering,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Rider {
    pub id: Uuid,
    pub name: String,
    pub location: Location,
    pub status: RiderStatus,
    pub current_orders: Vec<Uuid>,
}

impl Rider {
    pub fn new(name: String, location: Location) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            location,
            status: RiderStatus::Idle,
            current_orders: Vec::new(),
        }
    }

    pub fn can_accept_order(&self, max_orders: usize) -> bool {
        self.current_orders.len() < max_orders
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Pending,
    Assigned,
    InProgress,
    Delivered,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub merchant_id: Uuid,
    pub merchant_location: Location,
    pub customer_location: Location,
    pub status: OrderStatus,
    pub assigned_rider_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
}

impl Order {
    pub fn new(
        merchant_id: Uuid,
        merchant_location: Location,
        customer_location: Location,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            merchant_id,
            merchant_location,
            customer_location,
            status: OrderStatus::Pending,
            assigned_rider_id: None,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Merchant {
    pub id: Uuid,
    pub name: String,
    pub location: Location,
}

impl Merchant {
    pub fn new(name: String, location: Location) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            location,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DispatchConfig {
    pub max_orders_per_rider: usize,
    pub max_detour_distance_km: f64,
}

impl Default for DispatchConfig {
    fn default() -> Self {
        Self {
            max_orders_per_rider: DEFAULT_MAX_ORDERS_PER_RIDER,
            max_detour_distance_km: DEFAULT_MAX_DETOUR_DISTANCE_KM,
        }
    }
}

#[derive(Debug, Clone)]
pub struct DispatchState {
    pub riders: Arc<Mutex<std::collections::HashMap<Uuid, Rider>>>,
    pub merchants: Arc<Mutex<std::collections::HashMap<Uuid, Merchant>>>,
    pub orders: Arc<Mutex<std::collections::HashMap<Uuid, Order>>>,
    pub waiting_queue: Arc<Mutex<VecDeque<Uuid>>>,
    pub config: DispatchConfig,
}

impl DispatchState {
    pub fn new() -> Self {
        Self::with_config(DispatchConfig::default())
    }

    pub fn with_config(config: DispatchConfig) -> Self {
        Self {
            riders: Arc::new(Mutex::new(std::collections::HashMap::new())),
            merchants: Arc::new(Mutex::new(std::collections::HashMap::new())),
            orders: Arc::new(Mutex::new(std::collections::HashMap::new())),
            waiting_queue: Arc::new(Mutex::new(VecDeque::new())),
            config,
        }
    }

    pub async fn add_rider(&self, name: String, location: Location) -> Rider {
        let rider = Rider::new(name, location);
        let rider_id = rider.id;
        let mut riders = self.riders.lock().await;
        riders.insert(rider_id, rider.clone());
        rider
    }

    pub async fn get_rider(&self, rider_id: Uuid) -> Option<Rider> {
        let riders = self.riders.lock().await;
        riders.get(&rider_id).cloned()
    }

    pub async fn list_riders(&self) -> Vec<Rider> {
        let riders = self.riders.lock().await;
        riders.values().cloned().collect()
    }

    pub async fn update_rider_location(&self, rider_id: Uuid, location: Location) -> Option<Rider> {
        let mut riders = self.riders.lock().await;
        if let Some(rider) = riders.get_mut(&rider_id) {
            rider.location = location;
            return Some(rider.clone());
        }
        None
    }

    pub async fn add_merchant(&self, name: String, location: Location) -> Merchant {
        let merchant = Merchant::new(name, location);
        let merchant_id = merchant.id;
        let mut merchants = self.merchants.lock().await;
        merchants.insert(merchant_id, merchant.clone());
        merchant
    }

    pub async fn get_merchant(&self, merchant_id: Uuid) -> Option<Merchant> {
        let merchants = self.merchants.lock().await;
        merchants.get(&merchant_id).cloned()
    }

    pub async fn list_merchants(&self) -> Vec<Merchant> {
        let merchants = self.merchants.lock().await;
        merchants.values().cloned().collect()
    }

    pub async fn create_order(&self, merchant_id: Uuid) -> Option<Order> {
        let merchants = self.merchants.lock().await;
        let merchant = merchants.get(&merchant_id)?;
        let order = Order::new(merchant_id, merchant.location, Location {
            latitude: merchant.location.latitude + 0.01,
            longitude: merchant.location.longitude + 0.01,
        });
        drop(merchants);

        let order_id = order.id;
        let mut orders = self.orders.lock().await;
        orders.insert(order_id, order.clone());
        drop(orders);

        self.try_dispatch_order(order_id).await;
        Some(order)
    }

    pub async fn get_order(&self, order_id: Uuid) -> Option<Order> {
        let orders = self.orders.lock().await;
        orders.get(&order_id).cloned()
    }

    pub async fn list_orders(&self) -> Vec<Order> {
        let orders = self.orders.lock().await;
        orders.values().cloned().collect()
    }

    pub async fn complete_order(&self, order_id: Uuid) -> Option<Order> {
        let mut orders = self.orders.lock().await;
        let order = orders.get_mut(&order_id)?;
        if order.status != OrderStatus::Assigned && order.status != OrderStatus::InProgress {
            return None;
        }

        let rider_id = order.assigned_rider_id?;
        order.status = OrderStatus::Delivered;
        let order_clone = order.clone();
        drop(orders);

        let mut riders = self.riders.lock().await;
        if let Some(rider) = riders.get_mut(&rider_id) {
            if let Some(idx) = rider.current_orders.iter().position(|id| *id == order_id) {
                rider.current_orders.remove(idx);
            }
            if rider.current_orders.is_empty() {
                rider.status = RiderStatus::Idle;
            }
        }
        drop(riders);

        self.process_waiting_queue_for_rider(rider_id).await;
        Some(order_clone)
    }

    async fn try_dispatch_order(&self, order_id: Uuid) {
        let orders = self.orders.lock().await;
        let order = orders.get(&order_id).cloned();
        drop(orders);

        let order = match order {
            Some(o) => o,
            None => return,
        };

        let riders = self.riders.lock().await;
        let riders_list: Vec<Rider> = riders.values().cloned().collect();
        drop(riders);

        let idle_riders: Vec<Rider> = riders_list
            .iter()
            .filter(|r| r.status == RiderStatus::Idle && r.can_accept_order(self.config.max_orders_per_rider))
            .cloned()
            .collect();

        if !idle_riders.is_empty() {
            let nearest = idle_riders
                .into_iter()
                .min_by(|a, b| {
                    let dist_a = a.location.distance_to(&order.merchant_location);
                    let dist_b = b.location.distance_to(&order.merchant_location);
                    dist_a.partial_cmp(&dist_b).unwrap_or(std::cmp::Ordering::Equal)
                })
                .unwrap();

            self.assign_order(order_id, nearest.id).await;
            return;
        }

        let busy_riders: Vec<Rider> = riders_list
            .iter()
            .filter(|r| r.status == RiderStatus::Delivering && r.can_accept_order(self.config.max_orders_per_rider))
            .cloned()
            .collect();

        for rider in busy_riders {
            let distance = rider.location.distance_to(&order.merchant_location);
            if distance <= self.config.max_detour_distance_km {
                self.assign_order(order_id, rider.id).await;
                return;
            }
        }

        let mut queue = self.waiting_queue.lock().await;
        if !queue.iter().any(|id| *id == order_id) {
            queue.push_back(order_id);
        }
    }

    async fn assign_order(&self, order_id: Uuid, rider_id: Uuid) {
        let mut riders = self.riders.lock().await;
        let rider = match riders.get_mut(&rider_id) {
            Some(r) => r,
            None => return,
        };
        rider.current_orders.push(order_id);
        rider.status = RiderStatus::Delivering;
        drop(riders);

        let mut orders = self.orders.lock().await;
        if let Some(order) = orders.get_mut(&order_id) {
            order.status = OrderStatus::Assigned;
            order.assigned_rider_id = Some(rider_id);
        }
    }

    async fn process_waiting_queue_for_rider(&self, rider_id: Uuid) {
        let queue = self.waiting_queue.lock().await;
        let next_order_id = queue.front().cloned();
        drop(queue);

        let next_order_id = match next_order_id {
            Some(id) => id,
            None => return,
        };

        let orders = self.orders.lock().await;
        let order = orders.get(&next_order_id).cloned();
        drop(orders);

        if order.is_none() {
            return;
        }

        let riders = self.riders.lock().await;
        let rider = riders.get(&rider_id).cloned();
        drop(riders);

        if let Some(rider) = rider {
            if rider.can_accept_order(self.config.max_orders_per_rider) {
                let mut queue = self.waiting_queue.lock().await;
                queue.pop_front();
                drop(queue);
                self.assign_order(next_order_id, rider_id).await;
            }
        }
    }

    pub async fn get_waiting_queue(&self) -> Vec<Uuid> {
        let queue = self.waiting_queue.lock().await;
        queue.iter().cloned().collect()
    }
}

impl Default for DispatchState {
    fn default() -> Self {
        Self::new()
    }
}
