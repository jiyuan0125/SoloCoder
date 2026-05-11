use std::collections::HashMap;
use std::sync::RwLock;

use chrono::Utc;
use uuid::Uuid;

use crate::models::*;
use crate::errors::*;

pub struct InMemoryRepository {
    products: RwLock<HashMap<Uuid, Product>>,
    activities: RwLock<HashMap<Uuid, BargainActivity>>,
    records: RwLock<HashMap<Uuid, BargainRecord>>,
    users: RwLock<HashMap<Uuid, User>>,
    user_activity_index: RwLock<HashMap<(Uuid, Uuid), Uuid>>,
    activity_records_index: RwLock<HashMap<Uuid, Vec<Uuid>>>,
    user_bargain_index: RwLock<HashMap<(Uuid, Uuid), Uuid>>,
}

impl InMemoryRepository {
    pub fn new() -> Self {
        Self {
            products: RwLock::new(HashMap::new()),
            activities: RwLock::new(HashMap::new()),
            records: RwLock::new(HashMap::new()),
            users: RwLock::new(HashMap::new()),
            user_activity_index: RwLock::new(HashMap::new()),
            activity_records_index: RwLock::new(HashMap::new()),
            user_bargain_index: RwLock::new(HashMap::new()),
        }
    }
    
    pub fn create_product(&self, req: CreateProductRequest) -> Result<Product> {
        if req.floor_price >= req.original_price {
            return Err(BargainError::InvalidPrice);
        }
        
        let product = Product {
            id: Uuid::new_v4(),
            name: req.name,
            original_price: round_price(req.original_price),
            floor_price: round_price(req.floor_price),
            created_at: Utc::now(),
        };
        
        self.products.write().unwrap().insert(product.id, product.clone());
        Ok(product)
    }
    
    pub fn get_product(&self, id: Uuid) -> Result<Product> {
        self.products.read().unwrap()
            .get(&id)
            .cloned()
            .ok_or(BargainError::ProductNotFound)
    }
    
    pub fn get_all_products(&self) -> Vec<Product> {
        self.products.read().unwrap().values().cloned().collect()
    }
    
    pub fn create_user(&self, name: String) -> User {
        let user = User {
            id: Uuid::new_v4(),
            name,
            created_at: Utc::now(),
        };
        self.users.write().unwrap().insert(user.id, user.clone());
        user
    }
    
    pub fn get_user(&self, id: Uuid) -> Option<User> {
        self.users.read().unwrap().get(&id).cloned()
    }
    
    pub fn has_active_activity(&self, user_id: Uuid, product_id: Uuid) -> bool {
        let key = (user_id, product_id);
        if let Some(activity_id) = self.user_activity_index.read().unwrap().get(&key) {
            if let Some(activity) = self.activities.read().unwrap().get(activity_id) {
                return activity.status == BargainStatus::Active || 
                       activity.status == BargainStatus::Success;
            }
        }
        false
    }
    
    pub fn create_activity(&self, product: &Product, user_id: Uuid, duration_hours: i64) -> Result<BargainActivity> {
        let key = (user_id, product.id);
        if self.has_active_activity(user_id, product.id) {
            return Err(BargainError::DuplicateActivity);
        }
        
        let now = Utc::now();
        let activity = BargainActivity {
            id: Uuid::new_v4(),
            product_id: product.id,
            initiator_id: user_id,
            current_price: product.original_price,
            original_price: product.original_price,
            floor_price: product.floor_price,
            status: BargainStatus::Active,
            created_at: now,
            expires_at: now + chrono::Duration::hours(duration_hours),
            bargain_count: 0,
        };
        
        self.activities.write().unwrap().insert(activity.id, activity.clone());
        self.user_activity_index.write().unwrap().insert(key, activity.id);
        self.activity_records_index.write().unwrap().insert(activity.id, Vec::new());
        
        Ok(activity)
    }
    
    pub fn get_activity(&self, id: Uuid) -> Result<BargainActivity> {
        self.activities.read().unwrap()
            .get(&id)
            .cloned()
            .ok_or(BargainError::ActivityNotFound)
    }
    
    pub fn update_activity(&self, activity: BargainActivity) -> Result<()> {
        if !self.activities.read().unwrap().contains_key(&activity.id) {
            return Err(BargainError::ActivityNotFound);
        }
        self.activities.write().unwrap().insert(activity.id, activity);
        Ok(())
    }
    
    pub fn has_user_bargained(&self, activity_id: Uuid, user_id: Uuid) -> bool {
        self.user_bargain_index.read().unwrap()
            .contains_key(&(activity_id, user_id))
    }
    
    pub fn add_bargain_record(&self, record: BargainRecord) {
        self.records.write().unwrap().insert(record.id, record.clone());
        self.activity_records_index.write().unwrap()
            .entry(record.activity_id)
            .or_default()
            .push(record.id);
        self.user_bargain_index.write().unwrap()
            .insert((record.activity_id, record.user_id), record.id);
    }
    
    pub fn get_records_for_activity(&self, activity_id: Uuid) -> Vec<BargainRecord> {
        let record_ids = self.activity_records_index.read().unwrap()
            .get(&activity_id)
            .cloned()
            .unwrap_or_default();
        
        let records = self.records.read().unwrap();
        record_ids.iter()
            .filter_map(|id| records.get(id))
            .cloned()
            .collect()
    }
}

impl Default for InMemoryRepository {
    fn default() -> Self {
        Self::new()
    }
}

fn round_price(price: f64) -> f64 {
    (price * 100.0).round() / 100.0
}
