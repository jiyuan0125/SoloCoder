use std::collections::HashMap;
use std::sync::Arc;
use chrono::{DateTime, Duration, Utc, NaiveTime};
use uuid::Uuid;
use crate::error::*;
use crate::models::*;
use crate::store::InMemoryStore;

#[derive(Clone)]
pub struct CommunityService {
    store: Arc<InMemoryStore>,
}

impl CommunityService {
    pub fn new(store: Arc<InMemoryStore>) -> Self {
        Self { store }
    }
    
    fn get_today_cut_off_time(&self, cut_off_time: NaiveTime) -> DateTime<Utc> {
        let now = Utc::now();
        let today = now.date_naive();
        let naive_datetime = today.and_time(cut_off_time);
        DateTime::from_naive_utc_and_offset(naive_datetime, *now.offset())
    }
    
    fn get_payment_deadline(
        &self,
        created_at: DateTime<Utc>,
        cut_off_time: DateTime<Utc>,
    ) -> DateTime<Utc> {
        let thirty_minutes_later = created_at + Duration::minutes(30);
        if thirty_minutes_later < cut_off_time {
            thirty_minutes_later
        } else {
            cut_off_time
        }
    }
    
    pub fn create_pickup_point(
        &self,
        name: String,
        cut_off_time: NaiveTime,
    ) -> PickupPoint {
        let point = PickupPoint {
            id: Uuid::new_v4(),
            name,
            cut_off_time,
            created_at: Utc::now(),
        };
        self.store.create_pickup_point(point.clone());
        point
    }
    
    pub fn get_pickup_point(&self, id: &Uuid) -> Result<PickupPoint, BusinessError> {
        self.store.get_pickup_point(id).ok_or(BusinessError::PickupPointNotFound)
    }
    
    pub fn list_pickup_points(&self) -> Vec<PickupPoint> {
        self.store.list_pickup_points()
    }
    
    pub fn create_product(
        &self,
        pickup_point_id: Uuid,
        name: String,
        unit_price: i64,
        stock: i64,
    ) -> Result<Product, BusinessError> {
        self.get_pickup_point(&pickup_point_id)?;
        let product = Product {
            id: Uuid::new_v4(),
            pickup_point_id,
            name,
            unit_price,
            stock,
            created_at: Utc::now(),
        };
        self.store.create_product(product.clone());
        Ok(product)
    }
    
    pub fn get_product(&self, id: &Uuid) -> Result<Product, BusinessError> {
        self.store.get_product(id).ok_or(BusinessError::ProductNotFound)
    }
    
    pub fn list_products_by_pickup_point(&self, pickup_point_id: &Uuid) -> Result<Vec<Product>, BusinessError> {
        self.get_pickup_point(pickup_point_id)?;
        Ok(self.store.list_products_by_pickup_point(pickup_point_id))
    }
    
    pub fn create_order(
        &self,
        pickup_point_id: Uuid,
        user_id: Uuid,
        items: Vec<CreateOrderItemRequest>,
    ) -> Result<Order, BusinessError> {
        if items.is_empty() {
            return Err(BusinessError::EmptyOrder);
        }
        
        let pickup_point = self.get_pickup_point(&pickup_point_id)?;
        let created_at = Utc::now();
        let cut_off_time = self.get_today_cut_off_time(pickup_point.cut_off_time);
        
        let cut_off_time = if created_at >= cut_off_time {
            self.get_today_cut_off_time(pickup_point.cut_off_time) + Duration::days(1)
        } else {
            cut_off_time
        };
        
        let product_ids: Vec<Uuid> = items.iter().map(|i| i.product_id).collect();
        let mut products_map = HashMap::new();
        for product_id in &product_ids {
            let product = self.get_product(product_id)?;
            if product.pickup_point_id != pickup_point_id {
                return Err(BusinessError::ProductNotInPickupPoint);
            }
            products_map.insert(*product_id, product);
        }
        
        let order_items: Vec<OrderItem> = items.iter().map(|item| {
            let product = products_map.get(&item.product_id).unwrap();
            let subtotal = product.unit_price * item.quantity;
            OrderItem {
                product_id: item.product_id,
                product_name: product.name.clone(),
                unit_price: product.unit_price,
                quantity: item.quantity,
                subtotal,
            }
        }).collect();
        
        let total_amount: i64 = order_items.iter().map(|i| i.subtotal).sum();
        
        self.lock_stock(&items)?;
        
        let order = Order {
            id: Uuid::new_v4(),
            pickup_point_id,
            user_id,
            items: order_items,
            total_amount,
            status: OrderStatus::Created,
            created_at,
            payment_deadline: self.get_payment_deadline(created_at, cut_off_time),
            cut_off_time,
            paid_at: None,
            picked_up_at: None,
        };
        
        self.store.create_order(order.clone());
        Ok(order)
    }
    
    fn lock_stock(&self, items: &[CreateOrderItemRequest]) -> Result<(), BusinessError> {
        for item in items {
            loop {
                let product = self.get_product(&item.product_id)?;
                if product.stock < item.quantity {
                    return Err(BusinessError::InsufficientStock {
                        available: product.stock,
                        requested: item.quantity,
                    });
                }
                let mut updated_product = product.clone();
                updated_product.stock -= item.quantity;
                
                let current = self.get_product(&item.product_id)?;
                if current.stock == product.stock {
                    self.store.update_product(updated_product);
                    break;
                }
            }
        }
        Ok(())
    }
    
    fn release_stock(&self, order: &Order) {
        for item in &order.items {
            let product = match self.get_product(&item.product_id) {
                Ok(p) => p,
                Err(_) => continue,
            };
            let mut updated_product = product;
            updated_product.stock += item.quantity;
            self.store.update_product(updated_product);
        }
    }
    
    pub fn get_order(&self, id: &Uuid) -> Result<Order, BusinessError> {
        self.store.get_order(id).ok_or(BusinessError::OrderNotFound)
    }
    
    pub fn list_orders_by_user(&self, user_id: &Uuid) -> Vec<Order> {
        self.store.list_orders_by_user(user_id)
    }
    
    pub fn pay_order(&self, order_id: Uuid) -> Result<Order, BusinessError> {
        let mut order = self.get_order(&order_id)?;
        let now = Utc::now();
        
        if order.status == OrderStatus::Paid {
            return Err(BusinessError::OrderAlreadyPaid);
        }
        
        if order.status != OrderStatus::Created {
            return Err(BusinessError::InvalidOperation(
                format!("Cannot pay order with status {:?}", order.status)
            ));
        }
        
        if now >= order.payment_deadline {
            return Err(BusinessError::PaymentDeadlinePassed);
        }
        
        order.status = OrderStatus::Paid;
        order.paid_at = Some(now);
        self.store.update_order(order.clone());
        Ok(order)
    }
    
    pub fn cancel_order(&self, order_id: Uuid) -> Result<Order, BusinessError> {
        let mut order = self.get_order(&order_id)?;
        
        if order.status != OrderStatus::Created {
            return Err(BusinessError::InvalidOperation(
                format!("Cannot cancel order with status {:?}", order.status)
            ));
        }
        
        order.status = OrderStatus::Cancelled;
        self.store.update_order(order.clone());
        self.release_stock(&order);
        Ok(order)
    }
    
    pub fn refund_order(&self, order_id: Uuid) -> Result<Order, BusinessError> {
        let mut order = self.get_order(&order_id)?;
        let now = Utc::now();
        
        if order.status != OrderStatus::Paid {
            return Err(BusinessError::InvalidOperation(
                format!("Cannot refund order with status {:?}", order.status)
            ));
        }
        
        if now >= order.cut_off_time {
            return Err(BusinessError::RefundAfterCutOffTime);
        }
        
        order.status = OrderStatus::Refunded;
        self.store.update_order(order.clone());
        self.release_stock(&order);
        Ok(order)
    }
    
    pub fn process_cut_off(&self, pickup_point_id: Uuid) -> Result<SortingList, BusinessError> {
        let pickup_point = self.get_pickup_point(&pickup_point_id)?;
        let now = Utc::now();
        let cut_off_time = self.get_today_cut_off_time(pickup_point.cut_off_time);
        
        let orders = self.store.list_orders_by_pickup_point(&pickup_point_id);
        let mut product_stats: HashMap<Uuid, (String, i64, Vec<Uuid>)> = HashMap::new();
        
        for mut order in orders {
            if order.cut_off_time <= now {
                if order.status == OrderStatus::Created {
                    order.status = OrderStatus::Cancelled;
                    self.store.update_order(order.clone());
                    self.release_stock(&order);
                } else if order.status == OrderStatus::Paid {
                    for item in &order.items {
                        let entry = product_stats
                            .entry(item.product_id)
                            .or_insert_with(|| (item.product_name.clone(), 0, Vec::new()));
                        entry.1 += item.quantity;
                        entry.2.push(order.id);
                    }
                }
            }
        }
        
        let sorting_items: Vec<SortingItem> = product_stats
            .into_iter()
            .map(|(product_id, (name, qty, order_ids))| SortingItem {
                product_id,
                product_name: name,
                total_quantity: qty,
                order_ids,
            })
            .collect();
        
        let sorting_list = SortingList {
            id: Uuid::new_v4(),
            pickup_point_id,
            batch_date: cut_off_time,
            items: sorting_items,
            created_at: now,
            confirmed_at: None,
        };
        
        self.store.create_sorting_list(sorting_list.clone());
        Ok(sorting_list)
    }
    
    pub fn get_sorting_list(&self, id: &Uuid) -> Result<SortingList, BusinessError> {
        self.store.get_sorting_list(id).ok_or(BusinessError::SortingListNotFound)
    }
    
    pub fn list_sorting_lists_by_pickup_point(&self, pickup_point_id: &Uuid) -> Result<Vec<SortingList>, BusinessError> {
        self.get_pickup_point(pickup_point_id)?;
        Ok(self.store.list_sorting_lists_by_pickup_point(pickup_point_id))
    }
    
    pub fn confirm_sorting_complete(&self, sorting_list_id: Uuid) -> Result<SortingList, BusinessError> {
        let mut list = self.get_sorting_list(&sorting_list_id)?;
        
        if list.confirmed_at.is_some() {
            return Err(BusinessError::InvalidOperation(
                "Sorting list already confirmed".to_string()
            ));
        }
        
        list.confirmed_at = Some(Utc::now());
        self.store.update_sorting_list(list.clone());
        
        let orders = self.store.list_orders_by_pickup_point(&list.pickup_point_id);
        for mut order in orders {
            if order.status == OrderStatus::Paid && order.cut_off_time <= list.batch_date {
                order.status = OrderStatus::ReadyForPickup;
                self.store.update_order(order);
            }
        }
        
        Ok(list)
    }
    
    pub fn pickup_order(&self, order_id: Uuid) -> Result<Order, BusinessError> {
        let mut order = self.get_order(&order_id)?;
        
        if order.status != OrderStatus::ReadyForPickup {
            return Err(BusinessError::InvalidOperation(
                format!("Cannot pickup order with status {:?}", order.status)
            ));
        }
        
        order.status = OrderStatus::PickedUp;
        order.picked_up_at = Some(Utc::now());
        self.store.update_order(order.clone());
        Ok(order)
    }
    
    pub fn clean_expired_orders(&self) -> Vec<Order> {
        let all_orders = self.store.list_all_orders();
        
        let mut cancelled = Vec::new();
        let now = Utc::now();
        
        for order in all_orders {
            if order.status == OrderStatus::Created && now >= order.payment_deadline {
                let mut updated = order.clone();
                updated.status = OrderStatus::Cancelled;
                self.store.update_order(updated);
                self.release_stock(&order);
                cancelled.push(order);
            }
        }
        
        cancelled
    }
}
