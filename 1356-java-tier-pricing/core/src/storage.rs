use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;

use crate::models::*;

#[derive(Clone, Default)]
pub struct InMemoryStorage {
    products: Arc<RwLock<HashMap<String, Product>>>,
    tier_prices: Arc<RwLock<HashMap<String, TierPrice>>>,
    members: Arc<RwLock<HashMap<String, Member>>>,
    member_discounts: Arc<RwLock<HashMap<String, MemberDiscount>>>,
    special_prices: Arc<RwLock<HashMap<String, SpecialPrice>>>,
    full_discounts: Arc<RwLock<HashMap<String, FullDiscount>>>,
    approvals: Arc<RwLock<HashMap<String, ApprovalRequest>>>,
    orders: Arc<RwLock<HashMap<String, Order>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn save_product(&self, product: Product) {
        self.products.write().insert(product.id.clone(), product);
    }

    pub fn get_product(&self, id: &str) -> Option<Product> {
        self.products.read().get(id).cloned()
    }

    pub fn get_all_products(&self) -> Vec<Product> {
        self.products.read().values().cloned().collect()
    }

    pub fn save_tier_price(&self, tier_price: TierPrice) {
        self.tier_prices.write().insert(tier_price.id.clone(), tier_price);
    }

    pub fn get_tier_prices_by_product(&self, product_id: &str) -> Vec<TierPrice> {
        self.tier_prices
            .read()
            .values()
            .filter(|tp| tp.product_id == product_id)
            .cloned()
            .collect()
    }

    pub fn get_tier_price(&self, id: &str) -> Option<TierPrice> {
        self.tier_prices.read().get(id).cloned()
    }

    pub fn save_member(&self, member: Member) {
        self.members.write().insert(member.id.clone(), member);
    }

    pub fn get_member(&self, id: &str) -> Option<Member> {
        self.members.read().get(id).cloned()
    }

    pub fn get_all_members(&self) -> Vec<Member> {
        self.members.read().values().cloned().collect()
    }

    pub fn save_member_discount(&self, discount: MemberDiscount) {
        self.member_discounts.write().insert(discount.id.clone(), discount);
    }

    pub fn get_member_discounts(&self, member_type: MemberType, product_id: Option<&str>) -> Vec<MemberDiscount> {
        let discounts = self.member_discounts.read();
        discounts
            .values()
            .filter(|d| {
                d.member_type == member_type
                    && match (product_id, &d.product_id) {
                        (None, None) => true,
                        (Some(pid), Some(dpid)) => pid == dpid,
                        (_, _) => false,
                    }
            })
            .cloned()
            .collect()
    }

    pub fn save_special_price(&self, special_price: SpecialPrice) {
        self.special_prices.write().insert(special_price.id.clone(), special_price);
    }

    pub fn get_special_prices_by_product(&self, product_id: &str) -> Vec<SpecialPrice> {
        self.special_prices
            .read()
            .values()
            .filter(|sp| sp.product_id == product_id)
            .cloned()
            .collect()
    }

    pub fn get_special_price(&self, id: &str) -> Option<SpecialPrice> {
        self.special_prices.read().get(id).cloned()
    }

    pub fn save_full_discount(&self, discount: FullDiscount) {
        self.full_discounts.write().insert(discount.id.clone(), discount);
    }

    pub fn get_all_full_discounts(&self) -> Vec<FullDiscount> {
        self.full_discounts.read().values().cloned().collect()
    }

    pub fn save_approval(&self, approval: ApprovalRequest) {
        self.approvals.write().insert(approval.id.clone(), approval);
    }

    pub fn get_approval(&self, id: &str) -> Option<ApprovalRequest> {
        self.approvals.read().get(id).cloned()
    }

    pub fn get_all_approvals(&self) -> Vec<ApprovalRequest> {
        self.approvals.read().values().cloned().collect()
    }

    pub fn save_order(&self, order: Order) {
        self.orders.write().insert(order.id.clone(), order);
    }

    pub fn get_order(&self, id: &str) -> Option<Order> {
        self.orders.read().get(id).cloned()
    }

    pub fn get_all_orders(&self) -> Vec<Order> {
        self.orders.read().values().cloned().collect()
    }

    pub fn update_stock(&self, product_id: &str, quantity: i64) -> bool {
        let mut products = self.products.write();
        if let Some(product) = products.get_mut(product_id) {
            if product.stock >= quantity {
                product.stock -= quantity;
                product.updated_at = chrono::Utc::now();
                true
            } else {
                false
            }
        } else {
            false
        }
    }
}
