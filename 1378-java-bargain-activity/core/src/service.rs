use chrono::Utc;
use uuid::Uuid;

use crate::models::*;
use crate::errors::*;
use crate::repository::InMemoryRepository;
use crate::pricing::{calculate_bargain_amount, round_price};

const DEFAULT_DURATION_HOURS: i64 = 48;

pub struct BargainService {
    repo: InMemoryRepository,
}

impl BargainService {
    pub fn new(repo: InMemoryRepository) -> Self {
        Self { repo }
    }
    
    pub fn create_product(&self, req: CreateProductRequest) -> Result<Product> {
        self.repo.create_product(req)
    }
    
    pub fn get_product(&self, id: Uuid) -> Result<Product> {
        self.repo.get_product(id)
    }
    
    pub fn list_products(&self) -> Vec<Product> {
        self.repo.get_all_products()
    }
    
    pub fn create_user(&self, name: String) -> User {
        self.repo.create_user(name)
    }
    
    pub fn get_user(&self, id: Uuid) -> Option<User> {
        self.repo.get_user(id)
    }
    
    pub fn start_bargain(&self, req: StartBargainRequest) -> Result<ActivityDetail> {
        let product = self.repo.get_product(req.product_id)?;
        
        let mut activity = self.repo.create_activity(&product, req.user_id, DEFAULT_DURATION_HOURS)?;
        
        let record = self.process_bargain(&mut activity, req.user_id)?;
        
        let records = vec![record];
        
        Ok(ActivityDetail {
            activity,
            product,
            records,
        })
    }
    
    pub fn process_bargain(&self, activity: &mut BargainActivity, user_id: Uuid) -> Result<BargainRecord> {
        self.validate_activity(activity)?;
        
        if self.repo.has_user_bargained(activity.id, user_id) {
            return Err(BargainError::AlreadyBargained);
        }
        
        let price_before = activity.current_price;
        let bargain_amount = calculate_bargain_amount(
            price_before,
            activity.floor_price,
            activity.bargain_count,
        );
        
        let price_after = round_price(price_before - bargain_amount);
        
        let record = BargainRecord {
            id: Uuid::new_v4(),
            activity_id: activity.id,
            user_id,
            bargain_amount: round_price(bargain_amount),
            price_before,
            price_after,
            created_at: Utc::now(),
        };
        
        activity.current_price = price_after;
        activity.bargain_count += 1;
        
        if price_after <= activity.floor_price {
            activity.current_price = activity.floor_price;
            activity.status = BargainStatus::Success;
        }
        
        self.repo.update_activity(activity.clone())?;
        self.repo.add_bargain_record(record.clone());
        
        Ok(record)
    }
    
    pub fn help_bargain(&self, req: BargainRequest) -> Result<ActivityDetail> {
        let mut activity = self.repo.get_activity(req.activity_id)?;
        let product = self.repo.get_product(activity.product_id)?;
        
        self.process_bargain(&mut activity, req.user_id)?;
        
        let records = self.repo.get_records_for_activity(activity.id);
        
        Ok(ActivityDetail {
            activity,
            product,
            records,
        })
    }
    
    pub fn get_activity_detail(&self, activity_id: Uuid) -> Result<ActivityDetail> {
        let mut activity = self.repo.get_activity(activity_id)?;
        
        if activity.status == BargainStatus::Active && Utc::now() > activity.expires_at {
            activity.status = BargainStatus::Failed;
            self.repo.update_activity(activity.clone())?;
        }
        
        let product = self.repo.get_product(activity.product_id)?;
        let records = self.repo.get_records_for_activity(activity.id);
        
        Ok(ActivityDetail {
            activity,
            product,
            records,
        })
    }
    
    pub fn purchase(&self, activity_id: Uuid) -> Result<ActivityDetail> {
        let mut activity = self.repo.get_activity(activity_id)?;
        
        if activity.status != BargainStatus::Active && activity.status != BargainStatus::Failed {
            return Err(BargainError::ActivityNotActive);
        }
        
        activity.status = BargainStatus::Purchased;
        self.repo.update_activity(activity.clone())?;
        
        let product = self.repo.get_product(activity.product_id)?;
        let records = self.repo.get_records_for_activity(activity.id);
        
        Ok(ActivityDetail {
            activity,
            product,
            records,
        })
    }
    
    pub fn abandon(&self, activity_id: Uuid) -> Result<ActivityDetail> {
        let mut activity = self.repo.get_activity(activity_id)?;
        
        if activity.status != BargainStatus::Active && activity.status != BargainStatus::Failed {
            return Err(BargainError::ActivityNotActive);
        }
        
        activity.status = BargainStatus::Abandoned;
        self.repo.update_activity(activity.clone())?;
        
        let product = self.repo.get_product(activity.product_id)?;
        let records = self.repo.get_records_for_activity(activity.id);
        
        Ok(ActivityDetail {
            activity,
            product,
            records,
        })
    }
    
    fn validate_activity(&self, activity: &BargainActivity) -> Result<()> {
        if activity.status != BargainStatus::Active {
            return Err(BargainError::ActivityNotActive);
        }
        
        if Utc::now() > activity.expires_at {
            return Err(BargainError::ActivityExpired);
        }
        
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_full_bargain_flow() {
        let repo = InMemoryRepository::new();
        let service = BargainService::new(repo);
        
        let product = service.create_product(CreateProductRequest {
            name: "iPhone".to_string(),
            original_price: 1000.0,
            floor_price: 500.0,
        }).unwrap();
        
        let user1 = service.create_user("Alice".to_string());
        let user2 = service.create_user("Bob".to_string());
        
        let detail = service.start_bargain(StartBargainRequest {
            product_id: product.id,
            user_id: user1.id,
        }).unwrap();
        
        assert_eq!(detail.activity.status, BargainStatus::Active);
        assert_eq!(detail.activity.bargain_count, 1);
        assert!(detail.activity.current_price < detail.activity.original_price);
        assert!(detail.activity.current_price >= detail.activity.floor_price);
        
        let detail2 = service.help_bargain(BargainRequest {
            activity_id: detail.activity.id,
            user_id: user2.id,
        }).unwrap();
        
        assert_eq!(detail2.activity.bargain_count, 2);
        assert!(detail2.activity.current_price < detail.activity.current_price);
        
        let result = service.help_bargain(BargainRequest {
            activity_id: detail.activity.id,
            user_id: user2.id,
        });
        
        assert!(result.is_err());
        match result {
            Err(BargainError::AlreadyBargained) => (),
            _ => panic!("Expected AlreadyBargained error"),
        }
    }
    
    #[test]
    fn test_duplicate_activity() {
        let repo = InMemoryRepository::new();
        let service = BargainService::new(repo);
        
        let product = service.create_product(CreateProductRequest {
            name: "iPhone".to_string(),
            original_price: 1000.0,
            floor_price: 500.0,
        }).unwrap();
        
        let user = service.create_user("Alice".to_string());
        
        service.start_bargain(StartBargainRequest {
            product_id: product.id,
            user_id: user.id,
        }).unwrap();
        
        let result = service.start_bargain(StartBargainRequest {
            product_id: product.id,
            user_id: user.id,
        });
        
        assert!(result.is_err());
        match result {
            Err(BargainError::DuplicateActivity) => (),
            _ => panic!("Expected DuplicateActivity error"),
        }
    }
}
