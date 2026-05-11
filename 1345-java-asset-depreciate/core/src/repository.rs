use std::collections::HashMap;
use std::sync::Mutex;

use crate::models::*;

pub struct InMemoryRepository {
    employees: Mutex<HashMap<String, Employee>>,
    categories: Mutex<HashMap<String, AssetCategory>>,
    assets: Mutex<HashMap<String, Asset>>,
    applications: Mutex<HashMap<String, BorrowApplication>>,
    return_records: Mutex<HashMap<String, ReturnRecord>>,
    transfers: Mutex<HashMap<String, TransferRequest>>,
    history: Mutex<HashMap<String, ResponsibilityHistory>>,
    asset_locks: Mutex<HashMap<String, String>>,
}

impl InMemoryRepository {
    pub fn new() -> Self {
        Self {
            employees: Mutex::new(HashMap::new()),
            categories: Mutex::new(HashMap::new()),
            assets: Mutex::new(HashMap::new()),
            applications: Mutex::new(HashMap::new()),
            return_records: Mutex::new(HashMap::new()),
            transfers: Mutex::new(HashMap::new()),
            history: Mutex::new(HashMap::new()),
            asset_locks: Mutex::new(HashMap::new()),
        }
    }

    pub fn create_employee(&self, employee: Employee) {
        self.employees
            .lock()
            .unwrap()
            .insert(employee.id.clone(), employee);
    }

    pub fn get_employee(&self, id: &str) -> Option<Employee> {
        self.employees.lock().unwrap().get(id).cloned()
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.employees.lock().unwrap().values().cloned().collect()
    }

    pub fn create_category(&self, category: AssetCategory) {
        self.categories
            .lock()
            .unwrap()
            .insert(category.id.clone(), category);
    }

    pub fn get_category(&self, id: &str) -> Option<AssetCategory> {
        self.categories.lock().unwrap().get(id).cloned()
    }

    pub fn get_category_by_name(&self, name: &str) -> Option<AssetCategory> {
        self.categories
            .lock()
            .unwrap()
            .values()
            .find(|c| c.name == name)
            .cloned()
    }

    pub fn list_categories(&self) -> Vec<AssetCategory> {
        self.categories.lock().unwrap().values().cloned().collect()
    }

    pub fn create_asset(&self, asset: Asset) {
        self.assets
            .lock()
            .unwrap()
            .insert(asset.id.clone(), asset);
    }

    pub fn get_asset(&self, id: &str) -> Option<Asset> {
        self.assets.lock().unwrap().get(id).cloned()
    }

    pub fn update_asset(&self, asset: Asset) {
        self.assets
            .lock()
            .unwrap()
            .insert(asset.id.clone(), asset);
    }

    pub fn list_assets(&self) -> Vec<Asset> {
        self.assets.lock().unwrap().values().cloned().collect()
    }

    pub fn list_assets_by_holder(&self, holder_id: &str) -> Vec<Asset> {
        self.assets
            .lock()
            .unwrap()
            .values()
            .filter(|a| a.current_holder_id.as_deref() == Some(holder_id))
            .cloned()
            .collect()
    }

    pub fn list_assets_by_category(&self, category_id: &str) -> Vec<Asset> {
        self.assets
            .lock()
            .unwrap()
            .values()
            .filter(|a| a.category_id == category_id)
            .cloned()
            .collect()
    }

    pub fn count_assets_by_holder_and_category(&self, holder_id: &str, category_id: &str) -> u32 {
        self.assets
            .lock()
            .unwrap()
            .values()
            .filter(|a| {
                a.current_holder_id.as_deref() == Some(holder_id)
                    && a.category_id == category_id
                    && a.status == AssetStatus::InUse
            })
            .count() as u32
    }

    pub fn try_lock_asset(&self, asset_id: &str, request_id: &str) -> bool {
        let mut locks = self.asset_locks.lock().unwrap();
        if locks.contains_key(asset_id) {
            return false;
        }
        locks.insert(asset_id.to_string(), request_id.to_string());
        true
    }

    pub fn release_asset_lock(&self, asset_id: &str) {
        self.asset_locks.lock().unwrap().remove(asset_id);
    }

    pub fn create_application(&self, application: BorrowApplication) {
        self.applications
            .lock()
            .unwrap()
            .insert(application.id.clone(), application);
    }

    pub fn get_application(&self, id: &str) -> Option<BorrowApplication> {
        self.applications.lock().unwrap().get(id).cloned()
    }

    pub fn update_application(&self, application: BorrowApplication) {
        self.applications
            .lock()
            .unwrap()
            .insert(application.id.clone(), application);
    }

    pub fn list_applications(&self) -> Vec<BorrowApplication> {
        self.applications.lock().unwrap().values().cloned().collect()
    }

    pub fn list_applications_by_applicant(&self, applicant_id: &str) -> Vec<BorrowApplication> {
        self.applications
            .lock()
            .unwrap()
            .values()
            .filter(|a| a.applicant_id == applicant_id)
            .cloned()
            .collect()
    }

    pub fn create_return_record(&self, record: ReturnRecord) {
        self.return_records
            .lock()
            .unwrap()
            .insert(record.id.clone(), record);
    }

    pub fn list_return_records(&self, asset_id: &str) -> Vec<ReturnRecord> {
        self.return_records
            .lock()
            .unwrap()
            .values()
            .filter(|r| r.asset_id == asset_id)
            .cloned()
            .collect()
    }

    pub fn create_transfer(&self, transfer: TransferRequest) {
        self.transfers
            .lock()
            .unwrap()
            .insert(transfer.id.clone(), transfer);
    }

    pub fn get_transfer(&self, id: &str) -> Option<TransferRequest> {
        self.transfers.lock().unwrap().get(id).cloned()
    }

    pub fn update_transfer(&self, transfer: TransferRequest) {
        self.transfers
            .lock()
            .unwrap()
            .insert(transfer.id.clone(), transfer);
    }

    pub fn list_transfers(&self) -> Vec<TransferRequest> {
        self.transfers.lock().unwrap().values().cloned().collect()
    }

    pub fn add_history(&self, history: ResponsibilityHistory) {
        self.history
            .lock()
            .unwrap()
            .insert(history.id.clone(), history);
    }

    pub fn list_history_by_asset(&self, asset_id: &str) -> Vec<ResponsibilityHistory> {
        let mut list: Vec<_> = self
            .history
            .lock()
            .unwrap()
            .values()
            .filter(|h| h.asset_id == asset_id)
            .cloned()
            .collect();
        list.sort_by(|a, b| a.action_at.cmp(&b.action_at));
        list
    }
}

impl Default for InMemoryRepository {
    fn default() -> Self {
        Self::new()
    }
}
