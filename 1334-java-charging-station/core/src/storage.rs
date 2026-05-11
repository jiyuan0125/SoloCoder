use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Mutex;
use uuid::Uuid;

use crate::models::{Charger, Order, SystemState};

pub type SharedState = Arc<Mutex<SystemState>>;

pub fn create_shared_state() -> SharedState {
    Arc::new(Mutex::new(SystemState::new()))
}

pub async fn initialize_demo_data(state: SharedState) {
    let mut guard = state.lock().await;
    
    let chargers = vec![
        (7, "Station A - Charger 1"),
        (7, "Station A - Charger 2"),
        (22, "Station B - Fast Charger 1"),
        (22, "Station B - Fast Charger 2"),
        (11, "Station C - Medium Charger"),
    ];
    
    for (power, name) in chargers {
        let charger = Charger {
            id: Uuid::new_v4(),
            name: name.to_string(),
            power_kw: power,
            status: crate::models::ChargerStatus::Idle,
        };
        guard.chargers.insert(charger.id, charger);
    }
}

pub async fn get_charger(state: &SharedState, id: Uuid) -> Option<Charger> {
    let guard = state.lock().await;
    guard.chargers.get(&id).cloned()
}

pub async fn get_all_chargers(state: &SharedState) -> HashMap<Uuid, Charger> {
    let guard = state.lock().await;
    guard.chargers.clone()
}

pub async fn add_charger(state: &SharedState, charger: Charger) {
    let mut guard = state.lock().await;
    guard.chargers.insert(charger.id, charger);
}

pub async fn update_charger(state: &SharedState, charger: Charger) {
    let mut guard = state.lock().await;
    guard.chargers.insert(charger.id, charger);
}

pub async fn get_order(state: &SharedState, id: Uuid) -> Option<Order> {
    let guard = state.lock().await;
    guard.orders.get(&id).cloned()
}

pub async fn get_all_orders(state: &SharedState) -> HashMap<Uuid, Order> {
    let guard = state.lock().await;
    guard.orders.clone()
}

pub async fn add_order(state: &SharedState, order: Order) {
    let mut guard = state.lock().await;
    guard.orders.insert(order.id, order);
}

pub async fn update_order(state: &SharedState, order: Order) {
    let mut guard = state.lock().await;
    guard.orders.insert(order.id, order);
}
