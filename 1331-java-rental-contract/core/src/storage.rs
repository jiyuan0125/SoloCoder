use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use super::models::*;
use super::error::RentalError;
use uuid::Uuid;

pub trait Storage: Clone + Send + Sync + 'static {
    fn get_all_properties(&self) -> Vec<Property>;
    fn get_property(&self, id: &Uuid) -> Option<Property>;
    fn get_property_by_no(&self, property_no: &str) -> Option<Property>;
    fn create_property(&self, property: Property) -> Result<Property, RentalError>;
    
    fn get_all_tenants(&self) -> Vec<Tenant>;
    fn get_tenant(&self, id: &Uuid) -> Option<Tenant>;
    fn get_tenant_by_phone(&self, phone: &str) -> Option<Tenant>;
    fn create_tenant(&self, tenant: Tenant) -> Result<Tenant, RentalError>;
    
    fn get_all_contracts(&self) -> Vec<Contract>;
    fn get_contract(&self, id: &Uuid) -> Option<Contract>;
    fn get_contracts_by_property(&self, property_id: &Uuid) -> Vec<Contract>;
    fn get_active_contract_by_property(&self, property_id: &Uuid) -> Option<Contract>;
    fn create_contract(&self, contract: Contract) -> Result<Contract, RentalError>;
    fn update_contract(&self, contract: Contract) -> Result<Contract, RentalError>;
    
    fn get_all_checkout_records(&self) -> Vec<CheckoutRecord>;
    fn get_checkout_record(&self, id: &Uuid) -> Option<CheckoutRecord>;
    fn get_checkout_record_by_contract(&self, contract_id: &Uuid) -> Option<CheckoutRecord>;
    fn create_checkout_record(&self, record: CheckoutRecord) -> Result<CheckoutRecord, RentalError>;
}

#[derive(Clone, Default)]
pub struct InMemoryStorage {
    properties: Arc<RwLock<HashMap<Uuid, Property>>>,
    tenants: Arc<RwLock<HashMap<Uuid, Tenant>>>,
    contracts: Arc<RwLock<HashMap<Uuid, Contract>>>,
    checkout_records: Arc<RwLock<HashMap<Uuid, CheckoutRecord>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self::default()
    }
}

impl Storage for InMemoryStorage {
    fn get_all_properties(&self) -> Vec<Property> {
        self.properties
            .read()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    fn get_property(&self, id: &Uuid) -> Option<Property> {
        self.properties
            .read()
            .unwrap()
            .get(id)
            .cloned()
    }

    fn get_property_by_no(&self, property_no: &str) -> Option<Property> {
        self.properties
            .read()
            .unwrap()
            .values()
            .find(|p| p.property_no == property_no)
            .cloned()
    }

    fn create_property(&self, property: Property) -> Result<Property, RentalError> {
        let mut properties = self.properties.write().unwrap();
        if properties.values().any(|p| p.property_no == property.property_no) {
            return Err(RentalError::PropertyNoExists(property.property_no.clone()));
        }
        properties.insert(property.id, property.clone());
        Ok(property)
    }

    fn get_all_tenants(&self) -> Vec<Tenant> {
        self.tenants
            .read()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    fn get_tenant(&self, id: &Uuid) -> Option<Tenant> {
        self.tenants
            .read()
            .unwrap()
            .get(id)
            .cloned()
    }

    fn get_tenant_by_phone(&self, phone: &str) -> Option<Tenant> {
        self.tenants
            .read()
            .unwrap()
            .values()
            .find(|t| t.phone == phone)
            .cloned()
    }

    fn create_tenant(&self, tenant: Tenant) -> Result<Tenant, RentalError> {
        let mut tenants = self.tenants.write().unwrap();
        if tenants.values().any(|t| t.phone == tenant.phone) {
            return Err(RentalError::TenantPhoneExists(tenant.phone.clone()));
        }
        tenants.insert(tenant.id, tenant.clone());
        Ok(tenant)
    }

    fn get_all_contracts(&self) -> Vec<Contract> {
        self.contracts
            .read()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    fn get_contract(&self, id: &Uuid) -> Option<Contract> {
        self.contracts
            .read()
            .unwrap()
            .get(id)
            .cloned()
    }

    fn get_contracts_by_property(&self, property_id: &Uuid) -> Vec<Contract> {
        self.contracts
            .read()
            .unwrap()
            .values()
            .filter(|c| c.property_id == *property_id)
            .cloned()
            .collect()
    }

    fn get_active_contract_by_property(&self, property_id: &Uuid) -> Option<Contract> {
        self.contracts
            .read()
            .unwrap()
            .values()
            .find(|c| c.property_id == *property_id && c.status == ContractStatus::Active)
            .cloned()
    }

    fn create_contract(&self, contract: Contract) -> Result<Contract, RentalError> {
        let mut contracts = self.contracts.write().unwrap();
        contracts.insert(contract.id, contract.clone());
        Ok(contract)
    }

    fn update_contract(&self, contract: Contract) -> Result<Contract, RentalError> {
        let mut contracts = self.contracts.write().unwrap();
        if !contracts.contains_key(&contract.id) {
            return Err(RentalError::ContractNotFound(contract.id.to_string()));
        }
        contracts.insert(contract.id, contract.clone());
        Ok(contract)
    }

    fn get_all_checkout_records(&self) -> Vec<CheckoutRecord> {
        self.checkout_records
            .read()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    fn get_checkout_record(&self, id: &Uuid) -> Option<CheckoutRecord> {
        self.checkout_records
            .read()
            .unwrap()
            .get(id)
            .cloned()
    }

    fn get_checkout_record_by_contract(&self, contract_id: &Uuid) -> Option<CheckoutRecord> {
        self.checkout_records
            .read()
            .unwrap()
            .values()
            .find(|r| r.contract_id == *contract_id)
            .cloned()
    }

    fn create_checkout_record(&self, record: CheckoutRecord) -> Result<CheckoutRecord, RentalError> {
        let mut records = self.checkout_records.write().unwrap();
        records.insert(record.id, record.clone());
        Ok(record)
    }
}
