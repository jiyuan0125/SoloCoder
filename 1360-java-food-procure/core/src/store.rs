use std::sync::{Arc, Mutex};
use crate::models::*;
use uuid::Uuid;

pub type SharedStore = Arc<Mutex<InMemoryStore>>;

#[derive(Debug, Clone, Default)]
pub struct InMemoryStore {
    pub suppliers: Vec<Supplier>,
    pub ingredients: Vec<Ingredient>,
    pub stores: Vec<Store>,
    pub purchase_requests: Vec<PurchaseRequest>,
    pub quotations: Vec<Quotation>,
    pub purchase_orders: Vec<PurchaseOrder>,
    pub store_monthly_stats: Vec<StoreMonthlyStats>,
    pub in_progress_locks: Vec<InProgressLock>,
    pub merged_order_pools: Vec<MergedOrderPool>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn supplier_by_id(&self, id: Uuid) -> Option<Supplier> {
        self.suppliers.iter().find(|s| s.id == id).cloned()
    }

    pub fn ingredient_by_id(&self, id: Uuid) -> Option<Ingredient> {
        self.ingredients.iter().find(|i| i.id == id).cloned()
    }

    pub fn store_by_id(&self, id: Uuid) -> Option<Store> {
        self.stores.iter().find(|s| s.id == id).cloned()
    }

    pub fn purchase_request_by_id(&self, id: Uuid) -> Option<PurchaseRequest> {
        self.purchase_requests.iter().find(|p| p.id == id).cloned()
    }

    pub fn quotations_by_request(&self, request_id: Uuid) -> Vec<Quotation> {
        self.quotations
            .iter()
            .filter(|q| q.purchase_request_id == request_id)
            .cloned()
            .collect()
    }

    pub fn quotation_by_id(&self, id: Uuid) -> Option<Quotation> {
        self.quotations.iter().find(|q| q.id == id).cloned()
    }

    pub fn purchase_order_by_id(&self, id: Uuid) -> Option<PurchaseOrder> {
        self.purchase_orders.iter().find(|o| o.id == id).cloned()
    }

    pub fn get_store_monthly_stats(&self, store_id: Uuid, month: &str) -> Option<StoreMonthlyStats> {
        self.store_monthly_stats
            .iter()
            .find(|s| s.store_id == store_id && s.month == month)
            .cloned()
    }

    pub fn get_or_create_store_monthly_stats(&mut self, store_id: Uuid, month: &str) -> StoreMonthlyStats {
        if let Some(stats) = self.get_store_monthly_stats(store_id, month) {
            return stats;
        }
        let stats = StoreMonthlyStats {
            store_id,
            month: month.to_string(),
            total_purchase_amount: 0.0,
            emergency_purchase_amount: 0.0,
            emergency_purchase_count: 0,
        };
        self.store_monthly_stats.push(stats.clone());
        stats
    }

    pub fn update_store_monthly_stats<F: FnOnce(&mut StoreMonthlyStats)>(&mut self, store_id: Uuid, month: &str, f: F) {
        let mut stats = self.get_or_create_store_monthly_stats(store_id, month);
        f(&mut stats);
        if let Some(existing) = self.store_monthly_stats.iter_mut().find(|s| s.store_id == store_id && s.month == month) {
            *existing = stats;
        }
    }
}

pub fn create_shared_store() -> SharedStore {
    Arc::new(Mutex::new(InMemoryStore::new()))
}
