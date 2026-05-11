use chrono::Local;
use uuid::Uuid;

use crate::errors::AssetError;
use crate::models::*;
use crate::repository::InMemoryRepository;
use crate::utils::today;

pub struct AssetService {
    repo: InMemoryRepository,
}

impl AssetService {
    pub fn new(repo: InMemoryRepository) -> Self {
        Self { repo }
    }

    pub fn create_employee(&self, name: String, role: UserRole) -> Employee {
        let employee = Employee::new(name, role);
        self.repo.create_employee(employee.clone());
        employee
    }

    pub fn get_employee(&self, id: &str) -> Result<Employee, AssetError> {
        self.repo
            .get_employee(id)
            .ok_or_else(|| AssetError::EmployeeNotFound(id.to_string()))
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.repo.list_employees()
    }

    pub fn create_category(
        &self,
        name: String,
        useful_life_months: u32,
        max_per_employee: u32,
    ) -> AssetCategory {
        let category = AssetCategory::new(name, useful_life_months, max_per_employee);
        self.repo.create_category(category.clone());
        category
    }

    pub fn get_category(&self, id: &str) -> Result<AssetCategory, AssetError> {
        self.repo
            .get_category(id)
            .ok_or_else(|| AssetError::CategoryNotFound(id.to_string()))
    }

    pub fn list_categories(&self) -> Vec<AssetCategory> {
        self.repo.list_categories()
    }

    pub fn create_asset(
        &self,
        name: String,
        category_id: String,
        purchase_price: f64,
        purchase_date: chrono::NaiveDate,
    ) -> Result<Asset, AssetError> {
        self.get_category(&category_id)?;
        let asset = Asset::new(name, category_id, purchase_price, purchase_date);
        self.repo.create_asset(asset.clone());
        Ok(asset)
    }

    pub fn get_asset(&self, id: &str) -> Result<Asset, AssetError> {
        self.repo
            .get_asset(id)
            .ok_or_else(|| AssetError::AssetNotFound(id.to_string()))
    }

    pub fn get_asset_detail(&self, id: &str) -> Result<AssetDetail, AssetError> {
        let asset = self.get_asset(id)?;
        let category = self.get_category(&asset.category_id)?;
        let (accumulated_depreciation, current_net_value) =
            asset.calculate_depreciation(category.useful_life_months, today());

        let current_holder_name = asset
            .current_holder_id
            .as_ref()
            .and_then(|h| self.repo.get_employee(h))
            .map(|e| e.name);

        Ok(AssetDetail {
            asset,
            category_name: category.name,
            accumulated_depreciation,
            current_net_value,
            current_holder_name,
        })
    }

    pub fn list_assets(&self) -> Vec<Asset> {
        self.repo.list_assets()
    }

    pub fn list_asset_details(&self) -> Vec<AssetDetail> {
        self.repo
            .list_assets()
            .into_iter()
            .filter_map(|a| self.get_asset_detail(&a.id).ok())
            .collect()
    }

    pub fn apply_for_borrow(
        &self,
        asset_id: String,
        applicant_id: String,
        reason: String,
    ) -> Result<BorrowApplication, AssetError> {
        let asset = self.get_asset(&asset_id)?;
        let _applicant = self.get_employee(&applicant_id)?;
        let category = self.get_category(&asset.category_id)?;

        if !asset.is_available() {
            return Err(AssetError::AssetNotAvailable(asset.id));
        }

        let current_count = self
            .repo
            .count_assets_by_holder_and_category(&applicant_id, &category.id);
        if current_count >= category.max_per_employee {
            return Err(AssetError::CategoryLimitExceeded {
                category: category.name,
                limit: category.max_per_employee,
            });
        }

        let request_id = Uuid::new_v4().to_string();
        if !self.repo.try_lock_asset(&asset_id, &request_id) {
            return Err(AssetError::ConcurrentReservationFailed);
        }

        let application = BorrowApplication::new(asset_id.clone(), applicant_id, reason);
        self.repo.create_application(application.clone());
        self.repo.release_asset_lock(&asset_id);

        Ok(application)
    }

    pub fn approve_application(
        &self,
        application_id: String,
        approver_id: String,
    ) -> Result<BorrowApplication, AssetError> {
        let approver = self.get_employee(&approver_id)?;
        if !approver.is_admin() {
            return Err(AssetError::NotAnAdmin);
        }

        let mut application = self
            .repo
            .get_application(&application_id)
            .ok_or_else(|| AssetError::ApplicationNotFound(application_id.clone()))?;

        if application.status != ApplicationStatus::Pending {
            return Err(AssetError::ApplicationAlreadyProcessed);
        }

        let asset = self.get_asset(&application.asset_id)?;
        let applicant = self.get_employee(&application.applicant_id)?;
        let category = self.get_category(&asset.category_id)?;

        if !asset.is_available() {
            return Err(AssetError::AssetNotAvailable(asset.id));
        }

        let current_count =
            self.repo
                .count_assets_by_holder_and_category(&applicant.id, &category.id);
        if current_count >= category.max_per_employee {
            return Err(AssetError::CategoryLimitExceeded {
                category: category.name,
                limit: category.max_per_employee,
            });
        }

        application.status = ApplicationStatus::Approved;
        application.approved_by = Some(approver_id);
        application.processed_at = Some(Local::now());
        self.repo.update_application(application.clone());

        let mut asset = asset;
        asset.status = AssetStatus::InUse;
        asset.current_holder_id = Some(applicant.id.clone());
        self.repo.update_asset(asset.clone());

        self.repo.add_history(ResponsibilityHistory::new(
            asset.id.clone(),
            applicant.id,
            "领用".to_string(),
            Some(format!("由 {} 审批", approver.name)),
        ));

        Ok(application)
    }

    pub fn reject_application(
        &self,
        application_id: String,
        approver_id: String,
    ) -> Result<BorrowApplication, AssetError> {
        let approver = self.get_employee(&approver_id)?;
        if !approver.is_admin() {
            return Err(AssetError::NotAnAdmin);
        }

        let mut application = self
            .repo
            .get_application(&application_id)
            .ok_or_else(|| AssetError::ApplicationNotFound(application_id.clone()))?;

        if application.status != ApplicationStatus::Pending {
            return Err(AssetError::ApplicationAlreadyProcessed);
        }

        application.status = ApplicationStatus::Rejected;
        application.approved_by = Some(approver_id);
        application.processed_at = Some(Local::now());
        self.repo.update_application(application.clone());

        Ok(application)
    }

    pub fn list_applications(&self) -> Vec<BorrowApplication> {
        self.repo.list_applications()
    }

    pub fn get_application(&self, id: &str) -> Result<BorrowApplication, AssetError> {
        self.repo
            .get_application(id)
            .ok_or_else(|| AssetError::ApplicationNotFound(id.to_string()))
    }

    pub fn return_asset(
        &self,
        asset_id: String,
        returned_by: String,
        verified_by: String,
        condition: ReturnCondition,
        damage_note: Option<String>,
    ) -> Result<Asset, AssetError> {
        let verifier = self.get_employee(&verified_by)?;
        if !verifier.is_admin() {
            return Err(AssetError::ReturnNeedsAdmin);
        }

        let mut asset = self.get_asset(&asset_id)?;

        if asset.current_holder_id.as_deref() != Some(&returned_by) {
            return Err(AssetError::NotAssetOwner);
        }

        if asset.status != AssetStatus::InUse {
            return Err(AssetError::InvalidOperation);
        }

        let record = ReturnRecord {
            id: Uuid::new_v4().to_string(),
            asset_id: asset_id.clone(),
            returned_by: returned_by.clone(),
            verified_by: verified_by.clone(),
            condition,
            damage_note: damage_note.clone(),
            returned_at: Local::now(),
        };
        self.repo.create_return_record(record);

        let history_note = match condition {
            ReturnCondition::Good => format!("归还完好，由 {} 确认", verifier.name),
            ReturnCondition::Damaged => format!(
                "归还损坏: {}，由 {} 确认",
                damage_note.unwrap_or_else(|| "未备注".to_string()),
                verifier.name
            ),
        };

        self.repo.add_history(ResponsibilityHistory::new(
            asset_id.clone(),
            returned_by,
            "归还".to_string(),
            Some(history_note),
        ));

        asset.status = AssetStatus::Available;
        asset.current_holder_id = None;
        self.repo.update_asset(asset.clone());

        Ok(asset)
    }

    pub fn list_return_records(&self, asset_id: &str) -> Vec<ReturnRecord> {
        self.repo.list_return_records(asset_id)
    }

    pub fn initiate_transfer(
        &self,
        asset_id: String,
        from_employee_id: String,
        to_employee_id: String,
    ) -> Result<TransferRequest, AssetError> {
        let asset = self.get_asset(&asset_id)?;
        if asset.current_holder_id.as_deref() != Some(&from_employee_id) {
            return Err(AssetError::NotAssetOwner);
        }

        let _to_employee = self.get_employee(&to_employee_id)?;

        if asset.status != AssetStatus::InUse {
            return Err(AssetError::InvalidOperation);
        }

        let transfer = TransferRequest::new(asset_id, from_employee_id, to_employee_id);
        self.repo.create_transfer(transfer.clone());

        Ok(transfer)
    }

    pub fn confirm_transfer_by_from(
        &self,
        transfer_id: String,
        from_employee_id: String,
    ) -> Result<TransferRequest, AssetError> {
        let mut transfer = self
            .repo
            .get_transfer(&transfer_id)
            .ok_or_else(|| AssetError::InternalError("转移记录不存在".to_string()))?;

        if transfer.from_employee_id != from_employee_id {
            return Err(AssetError::InternalError("不是转出方".to_string()));
        }

        if transfer.is_completed() {
            return Err(AssetError::ApplicationAlreadyProcessed);
        }

        transfer.from_confirmed = true;
        self.try_complete_transfer(&mut transfer)
    }

    pub fn confirm_transfer_by_to(
        &self,
        transfer_id: String,
        to_employee_id: String,
    ) -> Result<TransferRequest, AssetError> {
        let mut transfer = self
            .repo
            .get_transfer(&transfer_id)
            .ok_or_else(|| AssetError::InternalError("转移记录不存在".to_string()))?;

        if transfer.to_employee_id != to_employee_id {
            return Err(AssetError::InternalError("不是转入方".to_string()));
        }

        if transfer.is_completed() {
            return Err(AssetError::ApplicationAlreadyProcessed);
        }

        transfer.to_confirmed = true;
        self.try_complete_transfer(&mut transfer)
    }

    fn try_complete_transfer(&self, transfer: &mut TransferRequest) -> Result<TransferRequest, AssetError> {
        if transfer.is_completed() {
            let mut asset = self.get_asset(&transfer.asset_id)?;
            let from_emp = self.get_employee(&transfer.from_employee_id)?;
            let to_emp = self.get_employee(&transfer.to_employee_id)?;

            self.repo.add_history(ResponsibilityHistory::new(
                asset.id.clone(),
                transfer.from_employee_id.clone(),
                "转出".to_string(),
                Some(format!("转移给 {}", to_emp.name)),
            ));

            self.repo.add_history(ResponsibilityHistory::new(
                asset.id.clone(),
                transfer.to_employee_id.clone(),
                "转入".to_string(),
                Some(format!("从 {} 转移", from_emp.name)),
            ));

            asset.current_holder_id = Some(transfer.to_employee_id.clone());
            self.repo.update_asset(asset);

            transfer.completed_at = Some(Local::now());
        }

        self.repo.update_transfer(transfer.clone());
        Ok(transfer.clone())
    }

    pub fn list_transfers(&self) -> Vec<TransferRequest> {
        self.repo.list_transfers()
    }

    pub fn get_transfer(&self, id: &str) -> Result<TransferRequest, AssetError> {
        self.repo
            .get_transfer(id)
            .ok_or_else(|| AssetError::InternalError("转移记录不存在".to_string()))
    }

    pub fn list_history_by_asset(&self, asset_id: &str) -> Vec<ResponsibilityHistory> {
        self.repo.list_history_by_asset(asset_id)
    }
}
