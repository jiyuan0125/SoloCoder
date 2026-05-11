use std::collections::HashMap;
use parking_lot::RwLock;
use uuid::Uuid;
use crate::models::*;

#[derive(Default)]
pub struct InMemoryStore {
    pickup_points: RwLock<HashMap<Uuid, PickupPoint>>,
    products: RwLock<HashMap<Uuid, Product>>,
    orders: RwLock<HashMap<Uuid, Order>>,
    sorting_lists: RwLock<HashMap<Uuid, SortingList>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }
    
    pub fn create_pickup_point(&self, point: PickupPoint) {
        self.pickup_points.write().insert(point.id, point);
    }
    
    pub fn get_pickup_point(&self, id: &Uuid) -> Option<PickupPoint> {
        self.pickup_points.read().get(id).cloned()
    }
    
    pub fn list_pickup_points(&self) -> Vec<PickupPoint> {
        self.pickup_points.read().values().cloned().collect()
    }
    
    pub fn create_product(&self, product: Product) {
        self.products.write().insert(product.id, product);
    }
    
    pub fn get_product(&self, id: &Uuid) -> Option<Product> {
        self.products.read().get(id).cloned()
    }
    
    pub fn list_products_by_pickup_point(&self, pickup_point_id: &Uuid) -> Vec<Product> {
        self.products
            .read()
            .values()
            .filter(|p| p.pickup_point_id == *pickup_point_id)
            .cloned()
            .collect()
    }
    
    pub fn update_product(&self, product: Product) {
        self.products.write().insert(product.id, product);
    }
    
    pub fn create_order(&self, order: Order) {
        self.orders.write().insert(order.id, order);
    }
    
    pub fn get_order(&self, id: &Uuid) -> Option<Order> {
        self.orders.read().get(id).cloned()
    }
    
    pub fn update_order(&self, order: Order) {
        self.orders.write().insert(order.id, order);
    }
    
    pub fn list_orders_by_pickup_point(&self, pickup_point_id: &Uuid) -> Vec<Order> {
        self.orders
            .read()
            .values()
            .filter(|o| o.pickup_point_id == *pickup_point_id)
            .cloned()
            .collect()
    }
    
    pub fn list_orders_by_user(&self, user_id: &Uuid) -> Vec<Order> {
        self.orders
            .read()
            .values()
            .filter(|o| o.user_id == *user_id)
            .cloned()
            .collect()
    }
    
    pub fn create_sorting_list(&self, list: SortingList) {
        self.sorting_lists.write().insert(list.id, list);
    }
    
    pub fn get_sorting_list(&self, id: &Uuid) -> Option<SortingList> {
        self.sorting_lists.read().get(id).cloned()
    }
    
    pub fn update_sorting_list(&self, list: SortingList) {
        self.sorting_lists.write().insert(list.id, list);
    }
    
    pub fn list_sorting_lists_by_pickup_point(&self, pickup_point_id: &Uuid) -> Vec<SortingList> {
        self.sorting_lists
            .read()
            .values()
            .filter(|l| l.pickup_point_id == *pickup_point_id)
            .cloned()
            .collect()
    }
    
    pub fn list_all_orders(&self) -> Vec<Order> {
        self.orders.read().values().cloned().collect()
    }
}
