use std::sync::Arc;

use chrono::{DateTime, Utc, Duration};

use crate::error::*;
use crate::models::*;
use crate::storage::InMemoryStorage;

pub const APPROVAL_THRESHOLD_INCREASE: f64 = 20.0;
pub const APPROVAL_THRESHOLD_DECREASE: f64 = 30.0;

#[derive(Clone)]
pub struct PricingService {
    storage: InMemoryStorage,
}

impl PricingService {
    pub fn new(storage: InMemoryStorage) -> Self {
        Self { storage }
    }

    pub fn query_price(
        &self,
        product_id: &str,
        quantity: i64,
        member_type: Option<MemberType>,
        query_time: DateTime<Utc>,
    ) -> Result<PriceResult> {
        let product = self
            .storage
            .get_product(product_id)
            .ok_or_else(|| SystemError::ProductNotFound(product_id.to_string()))?;

        if let Some(special_price) = self.get_active_special_price(product_id, query_time) {
            let total_price = special_price.special_price * quantity as f64;
            return Ok(PriceResult {
                product_id: product_id.to_string(),
                quantity,
                unit_price: special_price.special_price,
                total_price,
                price_type: PriceType::Special,
                rule_id: Some(special_price.id),
                is_special: true,
            });
        }

        let tier_price = self.get_best_tier_price(product_id, quantity, query_time);
        let tier_unit_price = tier_price.as_ref().map(|tp| {
            product.base_retail_price * (1.0 - tp.discount_percent / 100.0)
        });

        let member_discount = member_type.and_then(|mt| {
            self.get_best_member_discount(mt, Some(product_id), query_time)
        });
        let member_unit_price = member_discount.as_ref().map(|d| {
            product.base_retail_price * (1.0 - d.discount_percent / 100.0)
        });

        let mut unit_price = product.base_retail_price;
        let mut price_type = PriceType::Base;
        let mut rule_id = None;

        if let Some(tup) = tier_unit_price {
            unit_price = tup;
            price_type = PriceType::Tier;
            rule_id = tier_price.as_ref().map(|tp| tp.id.clone());
        }

        if let Some(mup) = member_unit_price {
            if mup < unit_price {
                unit_price = mup;
                price_type = PriceType::Member;
                rule_id = member_discount.as_ref().map(|d| d.id.clone());
            }
        }

        let total_price = unit_price * quantity as f64;

        Ok(PriceResult {
            product_id: product_id.to_string(),
            quantity,
            unit_price,
            total_price,
            price_type,
            rule_id,
            is_special: false,
        })
    }

    fn get_active_special_price(
        &self,
        product_id: &str,
        query_time: DateTime<Utc>,
    ) -> Option<SpecialPrice> {
        self.storage
            .get_special_prices_by_product(product_id)
            .into_iter()
            .filter(|sp| sp.is_valid_at(query_time))
            .max_by_key(|sp| sp.start_time)
    }

    fn get_best_tier_price(
        &self,
        product_id: &str,
        quantity: i64,
        query_time: DateTime<Utc>,
    ) -> Option<TierPrice> {
        self.storage
            .get_tier_prices_by_product(product_id)
            .into_iter()
            .filter(|tp| tp.is_valid_at(query_time))
            .filter(|tp| {
                quantity >= tp.min_quantity
                    && tp.max_quantity.map_or(true, |max| quantity <= max)
            })
            .max_by(|a, b| {
                a.discount_percent
                    .partial_cmp(&b.discount_percent)
                    .unwrap_or(std::cmp::Ordering::Equal)
            })
    }

    fn get_best_member_discount(
        &self,
        member_type: MemberType,
        product_id: Option<&str>,
        query_time: DateTime<Utc>,
    ) -> Option<MemberDiscount> {
        let product_specific = self
            .storage
            .get_member_discounts(member_type, product_id)
            .into_iter()
            .filter(|d| d.is_valid_at(query_time))
            .max_by(|a, b| {
                a.discount_percent
                    .partial_cmp(&b.discount_percent)
                    .unwrap_or(std::cmp::Ordering::Equal)
            });

        if product_specific.is_some() {
            return product_specific;
        }

        self.storage
            .get_member_discounts(member_type, None)
            .into_iter()
            .filter(|d| d.is_valid_at(query_time))
            .max_by(|a, b| {
                a.discount_percent
                    .partial_cmp(&b.discount_percent)
                    .unwrap_or(std::cmp::Ordering::Equal)
            })
    }

    pub fn calculate_full_discount(
        &self,
        eligible_amount: f64,
        query_time: DateTime<Utc>,
    ) -> FullDiscountResult {
        let best_discount = self
            .storage
            .get_all_full_discounts()
            .into_iter()
            .filter(|d| d.is_valid_at(query_time))
            .filter(|d| eligible_amount >= d.threshold)
            .max_by(|a, b| {
                a.discount
                    .partial_cmp(&b.discount)
                    .unwrap_or(std::cmp::Ordering::Equal)
            });

        if let Some(d) = best_discount {
            FullDiscountResult {
                applied: true,
                threshold: d.threshold,
                discount: d.discount,
                eligible_amount,
            }
        } else {
            FullDiscountResult {
                applied: false,
                threshold: 0.0,
                discount: 0.0,
                eligible_amount,
            }
        }
    }

    pub fn update_product_base_price(
        &self,
        product_id: &str,
        new_price: f64,
    ) -> Result<Option<ApprovalRequest>> {
        let mut product = self
            .storage
            .get_product(product_id)
            .ok_or_else(|| SystemError::ProductNotFound(product_id.to_string()))?;

        let old_price = product.base_retail_price;
        let change_percent = ((new_price - old_price).abs() / old_price) * 100.0;
        let change_type = if new_price > old_price {
            PriceChangeType::Increase
        } else {
            PriceChangeType::Decrease
        };

        let threshold = match change_type {
            PriceChangeType::Increase => APPROVAL_THRESHOLD_INCREASE,
            PriceChangeType::Decrease => APPROVAL_THRESHOLD_DECREASE,
        };

        if change_percent > threshold {
            let approval = ApprovalRequest::new(
                product_id.to_string(),
                ApprovalTargetType::ProductBasePrice,
                change_type,
                change_percent,
                old_price,
                new_price,
            );
            self.storage.save_approval(approval.clone());
            return Ok(Some(approval));
        }

        product.base_retail_price = new_price;
        product.updated_at = Utc::now();
        self.storage.save_product(product);

        Ok(None)
    }

    pub fn approve_price_change(&self, approval_id: &str) -> Result<()> {
        let mut approval = self
            .storage
            .get_approval(approval_id)
            .ok_or_else(|| SystemError::ApprovalNotFound(approval_id.to_string()))?;

        if !matches!(approval.status, ApprovalStatus::Pending) {
            return Err(SystemError::OperationDenied("审批已处理".to_string()));
        }

        match approval.target_type {
            ApprovalTargetType::ProductBasePrice => {
                let mut product = self
                    .storage
                    .get_product(&approval.target_id)
                    .ok_or_else(|| SystemError::ProductNotFound(approval.target_id.clone()))?;
                product.base_retail_price = approval.new_price;
                product.updated_at = Utc::now();
                self.storage.save_product(product);
            }
            _ => {
                return Err(SystemError::OperationDenied(
                    "暂不支持的审批类型".to_string(),
                ));
            }
        }

        approval.status = ApprovalStatus::Approved;
        approval.approved_at = Some(Utc::now());
        self.storage.save_approval(approval);

        Ok(())
    }

    pub fn create_product(&self, name: String, base_price: f64, stock: i64) -> Product {
        let product = Product::new(name, base_price, stock);
        self.storage.save_product(product.clone());
        product
    }

    pub fn create_tier_price(
        &self,
        product_id: String,
        min_qty: i64,
        max_qty: Option<i64>,
        discount_percent: f64,
        days: i64,
    ) -> TierPrice {
        let now = Utc::now();
        let end_time = now + Duration::days(days);
        let tier_price = TierPrice::new(
            product_id,
            min_qty,
            max_qty,
            discount_percent,
            now,
            end_time,
        );
        self.storage.save_tier_price(tier_price.clone());
        tier_price
    }

    pub fn create_member(&self, name: String, member_type: MemberType) -> Member {
        let member = Member::new(name, member_type);
        self.storage.save_member(member.clone());
        member
    }

    pub fn create_member_discount(
        &self,
        member_type: MemberType,
        product_id: Option<String>,
        discount_percent: f64,
        days: i64,
    ) -> MemberDiscount {
        let now = Utc::now();
        let end_time = now + Duration::days(days);
        let discount = MemberDiscount::new(
            member_type,
            product_id,
            discount_percent,
            now,
            end_time,
        );
        self.storage.save_member_discount(discount.clone());
        discount
    }

    pub fn create_special_price(
        &self,
        product_id: String,
        special_price: f64,
        days: i64,
    ) -> SpecialPrice {
        let now = Utc::now();
        let end_time = now + Duration::days(days);
        let sp = SpecialPrice::new(product_id, special_price, now, end_time);
        self.storage.save_special_price(sp.clone());
        sp
    }

    pub fn create_full_discount(
        &self,
        threshold: f64,
        discount: f64,
        days: i64,
    ) -> FullDiscount {
        let now = Utc::now();
        let end_time = now + Duration::days(days);
        let fd = FullDiscount::new(threshold, discount, now, end_time);
        self.storage.save_full_discount(fd.clone());
        fd
    }
}

#[derive(Clone)]
pub struct OrderService {
    storage: InMemoryStorage,
    pricing_service: PricingService,
    order_lock: Arc<parking_lot::Mutex<()>>,
}

impl OrderService {
    pub fn new(storage: InMemoryStorage, pricing_service: PricingService) -> Self {
        Self {
            storage,
            pricing_service,
            order_lock: Arc::new(parking_lot::Mutex::new(())),
        }
    }

    pub fn create_order(&self, request: CreateOrderRequest) -> Result<Order> {
        let _lock = self.order_lock.lock();
        let order_time = Utc::now();

        let member_type = request.member_id.as_ref().and_then(|mid| {
            self.storage.get_member(mid).map(|m| m.member_type)
        });

        let mut order_items = Vec::new();
        let mut subtotal = 0.0;
        let mut eligible_for_full_discount = 0.0;

        for item in &request.items {
            let price_result = self.pricing_service.query_price(
                &item.product_id,
                item.quantity,
                member_type,
                order_time,
            )?;

            let product = self
                .storage
                .get_product(&item.product_id)
                .ok_or_else(|| SystemError::ProductNotFound(item.product_id.clone()))?;

            if product.stock < item.quantity {
                return Err(SystemError::InsufficientStock(format!(
                    "商品 {} 库存不足",
                    product.name
                )));
            }

            order_items.push(OrderItem {
                product_id: item.product_id.clone(),
                quantity: item.quantity,
                unit_price: price_result.unit_price,
                total_price: price_result.total_price,
                price_type: price_result.price_type,
                is_special: price_result.is_special,
            });

            subtotal += price_result.total_price;

            if price_result.is_special {
                eligible_for_full_discount += price_result.total_price;
            } else {
                eligible_for_full_discount += price_result.total_price;
            }
        }

        let full_discount_result = self
            .pricing_service
            .calculate_full_discount(eligible_for_full_discount, order_time);

        let full_discount_amount = if full_discount_result.applied {
            let non_special_amount: f64 = order_items
                .iter()
                .filter(|item| !item.is_special)
                .map(|item| item.total_price)
                .sum();
            full_discount_result.discount.min(non_special_amount)
        } else {
            0.0
        };

        let total_amount = subtotal - full_discount_amount;

        for item in &order_items {
            if !self.storage.update_stock(&item.product_id, item.quantity) {
                return Err(SystemError::InsufficientStock(format!(
                    "商品库存不足"
                )));
            }
        }

        let order = Order::new(
            request.member_id.clone(),
            order_items,
            subtotal,
            full_discount_amount,
            total_amount,
        );

        self.storage.save_order(order.clone());

        Ok(order)
    }

    pub fn get_order(&self, order_id: &str) -> Result<Order> {
        self.storage
            .get_order(order_id)
            .ok_or_else(|| SystemError::OrderNotFound(order_id.to_string()))
    }
}
