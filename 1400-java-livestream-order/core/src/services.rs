use chrono::{Duration, Utc};
use uuid::Uuid;
use crate::models::*;
use crate::errors::SystemError;
use crate::store::InMemoryStore;

pub const LOCK_TIMEOUT_MINUTES: i64 = 15;

#[derive(Debug)]
pub struct LiveStreamService {
    store: InMemoryStore,
}

impl LiveStreamService {
    pub fn new(store: InMemoryStore) -> Self {
        LiveStreamService { store }
    }

    pub fn create_user(&self, request: CreateUserRequest) -> Result<User, SystemError> {
        let user = User {
            id: Uuid::new_v4(),
            name: request.name,
            created_at: Utc::now(),
        };
        self.store.insert_user(user.clone())?;
        Ok(user)
    }

    pub fn get_user(&self, user_id: Uuid) -> Result<User, SystemError> {
        self.store.get_user(user_id)
    }

    pub fn list_users(&self) -> Result<Vec<User>, SystemError> {
        self.store.list_users()
    }

    pub fn create_product(&self, request: CreateProductRequest) -> Result<Product, SystemError> {
        let now = Utc::now();
        let product = Product {
            id: Uuid::new_v4(),
            name: request.name,
            original_price: request.original_price,
            live_price: request.live_price,
            total_stock: request.total_stock,
            available_stock: request.total_stock,
            locked_stock: 0,
            purchase_limit: request.purchase_limit,
            is_on_sale: false,
            created_at: now,
            updated_at: now,
        };
        self.store.insert_product(product.clone())?;
        Ok(product)
    }

    pub fn get_product(&self, product_id: Uuid) -> Result<Product, SystemError> {
        self.store.get_product(product_id)
    }

    pub fn list_products(&self) -> Result<Vec<Product>, SystemError> {
        self.store.list_products()
    }

    pub fn update_product_stock(&self, request: UpdateProductStockRequest) -> Result<Product, SystemError> {
        self.store.update_product(request.product_id, |p| {
            p.total_stock += request.additional_stock;
            p.available_stock += request.additional_stock;
        })
    }

    pub fn set_product_sale_status(&self, request: SetProductSaleStatusRequest) -> Result<Product, SystemError> {
        self.store.update_product(request.product_id, |p| {
            p.is_on_sale = request.is_on_sale;
        })
    }

    pub fn start_live_stream(&self) -> Result<(), SystemError> {
        self.store.set_live_status(LiveStatus::Live)
    }

    pub fn end_live_stream(&self) -> Result<(), SystemError> {
        self.store.set_live_status(LiveStatus::Ended)?;
        self.release_all_pending_orders()
    }

    pub fn get_live_status(&self) -> Result<LiveStatus, SystemError> {
        self.store.get_live_status()
    }

    pub fn create_order(&self, request: CreateOrderRequest) -> Result<Order, SystemError> {
        let live_status = self.store.get_live_status()?;
        if live_status != LiveStatus::Live {
            return Err(SystemError::LiveStreamNotActive);
        }

        let user = self.store.get_user(request.user_id)?;
        let product = self.store.get_product(request.product_id)?;

        if !product.is_on_sale {
            return Err(SystemError::ProductNotOnSale);
        }

        if product.available_stock < request.quantity {
            return Err(SystemError::InsufficientStock {
                available: product.available_stock,
                requested: request.quantity,
            });
        }

        if let Some(limit) = product.purchase_limit {
            let user_purchased = self.get_user_purchased_quantity(user.id, product.id)?;
            if user_purchased + request.quantity > limit {
                return Err(SystemError::PurchaseLimitExceeded {
                    purchased: user_purchased,
                    limit,
                    requested: request.quantity,
                });
            }
        }

        self.store.update_product(request.product_id, |p| {
            p.available_stock -= request.quantity;
            p.locked_stock += request.quantity;
        })?;

        let now = Utc::now();
        let locked_until = now + Duration::minutes(LOCK_TIMEOUT_MINUTES);
        let total_amount = product.live_price * request.quantity as f64;

        let order = Order {
            id: Uuid::new_v4(),
            user_id: user.id,
            product_id: product.id,
            quantity: request.quantity,
            unit_price: product.live_price,
            total_amount,
            status: OrderStatus::PendingPayment,
            locked_until,
            created_at: now,
            paid_at: None,
            shipped_at: None,
            completed_at: None,
            cancelled_at: None,
            refunded_at: None,
            returned_at: None,
            refund_amount: None,
        };

        self.store.insert_order(order.clone())?;
        Ok(order)
    }

    pub fn pay_order(&self, request: PayOrderRequest) -> Result<Order, SystemError> {
        self.release_expired_locks()?;
        let order = self.store.get_order(request.order_id)?;

        match order.status {
            OrderStatus::PendingPayment => {
                if Utc::now() > order.locked_until {
                    return Err(SystemError::OrderLockExpired);
                }
            },
            OrderStatus::Paid => return Err(SystemError::OrderAlreadyPaid),
            OrderStatus::Cancelled => return Err(SystemError::OrderAlreadyCancelled),
            _ => return Err(SystemError::InvalidStatusTransition {
                from: format!("{:?}", order.status),
                to: "Paid".to_string(),
            }),
        }

        self.store.update_product(order.product_id, |p| {
            p.locked_stock -= order.quantity;
        })?;

        let updated = self.store.update_order(order.id, |o| {
            o.status = OrderStatus::Paid;
            o.paid_at = Some(Utc::now());
        })?;

        Ok(updated)
    }

    pub fn cancel_order(&self, request: CancelOrderRequest) -> Result<Order, SystemError> {
        self.release_expired_locks()?;
        let order = self.store.get_order(request.order_id)?;

        match order.status {
            OrderStatus::PendingPayment => {
                self.store.update_product(order.product_id, |p| {
                    p.locked_stock -= order.quantity;
                    p.available_stock += order.quantity;
                })?;
            },
            OrderStatus::Cancelled => return Err(SystemError::OrderAlreadyCancelled),
            _ => return Err(SystemError::InvalidStatusTransition {
                from: format!("{:?}", order.status),
                to: "Cancelled".to_string(),
            }),
        }

        let updated = self.store.update_order(order.id, |o| {
            o.status = OrderStatus::Cancelled;
            o.cancelled_at = Some(Utc::now());
        })?;

        Ok(updated)
    }

    pub fn ship_order(&self, order_id: Uuid) -> Result<Order, SystemError> {
        let order = self.store.get_order(order_id)?;

        match order.status {
            OrderStatus::Paid => {},
            OrderStatus::Shipped => return Err(SystemError::OrderAlreadyShipped),
            _ => return Err(SystemError::InvalidStatusTransition {
                from: format!("{:?}", order.status),
                to: "Shipped".to_string(),
            }),
        }

        let updated = self.store.update_order(order.id, |o| {
            o.status = OrderStatus::Shipped;
            o.shipped_at = Some(Utc::now());
        })?;

        Ok(updated)
    }

    pub fn complete_order(&self, order_id: Uuid) -> Result<Order, SystemError> {
        let order = self.store.get_order(order_id)?;

        match order.status {
            OrderStatus::Shipped => {},
            OrderStatus::Completed => return Err(SystemError::OrderAlreadyCompleted),
            _ => return Err(SystemError::InvalidStatusTransition {
                from: format!("{:?}", order.status),
                to: "Completed".to_string(),
            }),
        }

        let updated = self.store.update_order(order.id, |o| {
            o.status = OrderStatus::Completed;
            o.completed_at = Some(Utc::now());
        })?;

        Ok(updated)
    }

    pub fn refund_order(&self, request: RefundOrderRequest) -> Result<Order, SystemError> {
        let order = self.store.get_order(request.order_id)?;

        match order.status {
            OrderStatus::Paid | OrderStatus::Shipped | OrderStatus::Completed => {},
            OrderStatus::Refunded | OrderStatus::Returned => return Err(SystemError::OrderAlreadyRefunded),
            _ => return Err(SystemError::InvalidStatusTransition {
                from: format!("{:?}", order.status),
                to: if request.is_return { "Returned" } else { "Refunded" }.to_string(),
            }),
        }

        if request.is_return {
            self.store.update_product(order.product_id, |p| {
                p.available_stock += order.quantity;
                p.total_stock += order.quantity;
            })?;
        }

        let refund_amount = order.total_amount;
        let updated = self.store.update_order(order.id, |o| {
            o.status = if request.is_return { OrderStatus::Returned } else { OrderStatus::Refunded };
            o.refunded_at = Some(Utc::now());
            if request.is_return {
                o.returned_at = Some(Utc::now());
            }
            o.refund_amount = Some(refund_amount);
        })?;

        Ok(updated)
    }

    pub fn get_order(&self, order_id: Uuid) -> Result<Order, SystemError> {
        self.release_expired_locks()?;
        self.store.get_order(order_id)
    }

    pub fn list_orders(&self) -> Result<Vec<Order>, SystemError> {
        self.release_expired_locks()?;
        self.store.list_orders()
    }

    pub fn list_orders_by_user(&self, user_id: Uuid) -> Result<Vec<Order>, SystemError> {
        self.release_expired_locks()?;
        self.store.list_orders_by_user(user_id)
    }

    fn get_user_purchased_quantity(&self, user_id: Uuid, product_id: Uuid) -> Result<u32, SystemError> {
        let orders = self.store.list_orders_by_user(user_id)?;
        let quantity = orders.iter()
            .filter(|o| o.product_id == product_id)
            .filter(|o| matches!(o.status, OrderStatus::Paid | OrderStatus::Shipped | OrderStatus::Completed))
            .map(|o| o.quantity)
            .sum();
        Ok(quantity)
    }

    fn release_expired_locks(&self) -> Result<(), SystemError> {
        let orders = self.store.list_orders()?;
        let now = Utc::now();

        for order in orders {
            if order.status == OrderStatus::PendingPayment && now > order.locked_until {
                self.cancel_order(CancelOrderRequest { order_id: order.id })?;
            }
        }

        Ok(())
    }

    fn release_all_pending_orders(&self) -> Result<(), SystemError> {
        let orders = self.store.list_orders()?;

        for order in orders {
            if order.status == OrderStatus::PendingPayment {
                self.cancel_order(CancelOrderRequest { order_id: order.id })?;
            }
        }

        Ok(())
    }
}
