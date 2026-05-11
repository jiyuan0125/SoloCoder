use chrono::{Duration, Utc};
use uuid::Uuid;

use crate::error::RentalError;
use crate::models::*;
use crate::storage::Storage;

#[derive(Clone)]
pub struct RentalService<S: Storage> {
    storage: S,
}

impl<S: Storage> RentalService<S> {
    pub fn new(storage: S) -> Self {
        Self { storage }
    }

    pub fn list_properties(&self) -> Vec<Property> {
        self.storage.get_all_properties()
    }

    pub fn get_property(&self, id: &Uuid) -> Result<Property, RentalError> {
        self.storage
            .get_property(id)
            .ok_or_else(|| RentalError::PropertyNotFound(id.to_string()))
    }

    pub fn create_property(&self, req: CreatePropertyRequest) -> Result<Property, RentalError> {
        if req.area <= 0.0 {
            return Err(RentalError::InvalidDate("Area must be positive".to_string()));
        }
        if req.monthly_rent < 0.0 {
            return Err(RentalError::InvalidDate("Monthly rent cannot be negative".to_string()));
        }
        let property = Property::new(req.property_no, req.area, req.monthly_rent);
        self.storage.create_property(property)
    }

    pub fn list_tenants(&self) -> Vec<Tenant> {
        self.storage.get_all_tenants()
    }

    pub fn get_tenant(&self, id: &Uuid) -> Result<Tenant, RentalError> {
        self.storage
            .get_tenant(id)
            .ok_or_else(|| RentalError::TenantNotFound(id.to_string()))
    }

    pub fn create_tenant(&self, req: CreateTenantRequest) -> Result<Tenant, RentalError> {
        if req.name.trim().is_empty() {
            return Err(RentalError::InvalidDate("Tenant name cannot be empty".to_string()));
        }
        if req.phone.trim().is_empty() {
            return Err(RentalError::InvalidDate("Tenant phone cannot be empty".to_string()));
        }
        let tenant = Tenant::new(req.name, req.phone);
        self.storage.create_tenant(tenant)
    }

    pub fn list_contracts(&self) -> Vec<Contract> {
        self.storage.get_all_contracts()
    }

    pub fn get_contract(&self, id: &Uuid) -> Result<Contract, RentalError> {
        self.storage
            .get_contract(id)
            .ok_or_else(|| RentalError::ContractNotFound(id.to_string()))
    }

    pub fn get_expiring_contracts(&self, days: i64) -> Vec<Contract> {
        let today = Utc::now().date_naive();
        let future_date = today + Duration::days(days);
        
        self.storage
            .get_all_contracts()
            .into_iter()
            .filter(|c| {
                c.status == ContractStatus::Active
                    && c.end_date >= today
                    && c.end_date <= future_date
            })
            .collect()
    }

    pub fn create_contract(&self, req: CreateContractRequest) -> Result<Contract, RentalError> {
        if req.start_date >= req.end_date {
            return Err(RentalError::InvalidDate(
                "Start date must be before end date".to_string(),
            ));
        }
        
        if req.deposit_amount < 0.0 {
            return Err(RentalError::InvalidDate(
                "Deposit amount cannot be negative".to_string(),
            ));
        }

        let property = self.get_property(&req.property_id)?;
        let _tenant = self.get_tenant(&req.tenant_id)?;

        if let Some(_active) = self.storage.get_active_contract_by_property(&req.property_id) {
            return Err(RentalError::PropertyOccupied);
        }

        let monthly_rent = req.monthly_rent.unwrap_or(property.monthly_rent);
        if monthly_rent < 0.0 {
            return Err(RentalError::InvalidDate(
                "Monthly rent cannot be negative".to_string(),
            ));
        }

        let contract = Contract::new(
            req.tenant_id,
            req.property_id,
            req.start_date,
            req.end_date,
            monthly_rent,
            req.deposit_amount,
            None,
        );

        self.storage.create_contract(contract)
    }

    pub fn renew_contract(&self, req: RenewContractRequest) -> Result<Contract, RentalError> {
        let original = self.get_contract(&req.contract_id)?;
        
        if original.status != ContractStatus::Active {
            return Err(RentalError::ContractNotActive(original.id.to_string()));
        }

        let extend_years = req.extend_years.unwrap_or(1);
        if extend_years <= 0 {
            return Err(RentalError::InvalidDate(
                "Extend years must be positive".to_string(),
            ));
        }

        let new_start_date = original.end_date;
        let new_end_date = new_start_date
            .checked_add_signed(Duration::days(extend_years as i64 * 365))
            .ok_or_else(|| RentalError::InvalidDate("Invalid date calculation".to_string()))?;

        let new_monthly_rent = req.new_monthly_rent.unwrap_or(original.monthly_rent);
        if new_monthly_rent < 0.0 {
            return Err(RentalError::InvalidDate(
                "Monthly rent cannot be negative".to_string(),
            ));
        }

        let new_deposit = req.new_deposit_amount.unwrap_or(original.deposit_amount);
        if new_deposit < 0.0 {
            return Err(RentalError::InvalidDate(
                "Deposit amount cannot be negative".to_string(),
            ));
        }

        let new_contract = Contract::new(
            original.tenant_id,
            original.property_id,
            new_start_date,
            new_end_date,
            new_monthly_rent,
            new_deposit,
            Some(original.id),
        );

        let mut updated_original = original.clone();
        updated_original.status = ContractStatus::Expired;
        updated_original.updated_at = Utc::now();
        self.storage.update_contract(updated_original)?;

        self.storage.create_contract(new_contract)
    }

    pub fn checkout(&self, req: CheckoutRequest) -> Result<CheckoutRecord, RentalError> {
        let contract = self.get_contract(&req.contract_id)?;
        
        if contract.status != ContractStatus::Active {
            return Err(RentalError::ContractNotActive(contract.id.to_string()));
        }

        if req.damage_cost < 0.0 {
            return Err(RentalError::InvalidDamageCost);
        }

        let today = Utc::now().date_naive();
        let actual_end_date = req.actual_end_date.unwrap_or(today);

        if actual_end_date < contract.start_date {
            return Err(RentalError::InvalidDate(
                "Actual end date cannot be before start date".to_string(),
            ));
        }

        let (refund_amount, outstanding_debt) = if req.damage_cost >= contract.deposit_amount {
            (0.0, req.damage_cost - contract.deposit_amount)
        } else {
            (contract.deposit_amount - req.damage_cost, 0.0)
        };

        let checkout_record = CheckoutRecord {
            id: Uuid::new_v4(),
            contract_id: contract.id,
            actual_end_date,
            damage_cost: req.damage_cost,
            refund_amount,
            outstanding_debt,
            created_at: Utc::now(),
        };

        let mut updated_contract = contract.clone();
        updated_contract.end_date = actual_end_date;
        updated_contract.status = ContractStatus::Terminated;
        updated_contract.updated_at = Utc::now();
        
        self.storage.update_contract(updated_contract)?;
        self.storage.create_checkout_record(checkout_record.clone())?;

        Ok(checkout_record)
    }

    pub fn get_checkout_record(&self, contract_id: &Uuid) -> Result<CheckoutRecord, RentalError> {
        self.storage
            .get_checkout_record_by_contract(contract_id)
            .ok_or_else(|| RentalError::CheckoutRecordNotFound(contract_id.to_string()))
    }
}

impl<S: Storage> RentalService<S> {
    pub fn is_property_available(&self, property_id: &Uuid) -> bool {
        self.storage
            .get_active_contract_by_property(property_id)
            .is_none()
    }

    pub fn get_contracts_by_tenant(&self, tenant_id: &Uuid) -> Vec<Contract> {
        self.storage
            .get_all_contracts()
            .into_iter()
            .filter(|c| c.tenant_id == *tenant_id)
            .collect()
    }

    pub fn get_active_contracts(&self) -> Vec<Contract> {
        self.storage
            .get_all_contracts()
            .into_iter()
            .filter(|c| c.status == ContractStatus::Active)
            .collect()
    }
}
