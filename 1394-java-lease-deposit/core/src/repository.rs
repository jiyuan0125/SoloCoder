use std::collections::HashMap;
use std::sync::Mutex;
use uuid::Uuid;
use chrono::Local;
use rust_decimal::Decimal;

use crate::error::LeaseError;
use crate::models::{Tenant, Room, Contract, ContractStatus, ContractType};

#[derive(Default)]
pub struct InMemoryRepository {
    tenants: Mutex<HashMap<Uuid, Tenant>>,
    rooms: Mutex<HashMap<Uuid, Room>>,
    contracts: Mutex<HashMap<Uuid, Contract>>,
}

impl InMemoryRepository {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn create_tenant(
        &self,
        name: String,
        phone: String,
        id_card: Option<String>,
    ) -> Result<Tenant, LeaseError> {
        let tenant = Tenant {
            id: Uuid::new_v4(),
            name,
            phone,
            id_card,
            created_at: Local::now(),
        };

        let mut tenants = self.tenants.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock tenants: {}", e))
        })?;

        tenants.insert(tenant.id, tenant.clone());
        Ok(tenant)
    }

    pub fn get_tenant(&self, id: Uuid) -> Result<Option<Tenant>, LeaseError> {
        let tenants = self.tenants.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock tenants: {}", e))
        })?;

        Ok(tenants.get(&id).cloned())
    }

    pub fn list_tenants(&self) -> Result<Vec<Tenant>, LeaseError> {
        let tenants = self.tenants.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock tenants: {}", e))
        })?;

        Ok(tenants.values().cloned().collect())
    }

    pub fn create_room(
        &self,
        room_number: String,
        area: Decimal,
        default_monthly_rent: Decimal,
    ) -> Result<Room, LeaseError> {
        let now = Local::now();
        let room = Room {
            id: Uuid::new_v4(),
            room_number: room_number.clone(),
            area,
            default_monthly_rent,
            is_available: true,
            created_at: now,
            updated_at: now,
        };

        let mut rooms = self.rooms.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock rooms: {}", e))
        })?;

        if rooms.values().any(|r| r.room_number == room_number) {
            return Err(LeaseError::RoomAlreadyExists(room_number));
        }

        rooms.insert(room.id, room.clone());
        Ok(room)
    }

    pub fn get_room(&self, id: Uuid) -> Result<Option<Room>, LeaseError> {
        let rooms = self.rooms.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock rooms: {}", e))
        })?;

        Ok(rooms.get(&id).cloned())
    }

    pub fn get_room_by_number(&self, room_number: &str) -> Result<Option<Room>, LeaseError> {
        let rooms = self.rooms.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock rooms: {}", e))
        })?;

        Ok(rooms.values().find(|r| r.room_number == room_number).cloned())
    }

    pub fn list_rooms(&self) -> Result<Vec<Room>, LeaseError> {
        let rooms = self.rooms.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock rooms: {}", e))
        })?;

        Ok(rooms.values().cloned().collect())
    }

    pub fn update_room(&self, room: Room) -> Result<Room, LeaseError> {
        let mut rooms = self.rooms.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock rooms: {}", e))
        })?;

        if !rooms.contains_key(&room.id) {
            return Err(LeaseError::RoomNotFound(room.id.to_string()));
        }

        let updated_room = Room {
            updated_at: Local::now(),
            ..room
        };

        rooms.insert(updated_room.id, updated_room.clone());
        Ok(updated_room)
    }

    pub fn create_contract(
        &self,
        room_id: Uuid,
        tenant_id: Uuid,
        tenant_name: String,
        start_date: chrono::NaiveDate,
        end_date: chrono::NaiveDate,
        monthly_rent: Decimal,
        deposit_amount: Decimal,
    ) -> Result<Contract, LeaseError> {
        let now = Local::now();
        let contract = Contract {
            id: Uuid::new_v4(),
            room_id,
            tenant_id,
            tenant_name,
            start_date,
            end_date,
            monthly_rent,
            deposit_amount,
            deposit_paid: false,
            first_month_rent_paid: false,
            status: ContractStatus::Pending,
            contract_type: ContractType::FixedTerm,
            created_at: now,
            updated_at: now,
        };

        let mut contracts = self.contracts.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock contracts: {}", e))
        })?;

        contracts.insert(contract.id, contract.clone());
        Ok(contract)
    }

    pub fn get_contract(&self, id: Uuid) -> Result<Option<Contract>, LeaseError> {
        let contracts = self.contracts.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock contracts: {}", e))
        })?;

        Ok(contracts.get(&id).cloned())
    }

    pub fn list_contracts(&self) -> Result<Vec<Contract>, LeaseError> {
        let contracts = self.contracts.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock contracts: {}", e))
        })?;

        Ok(contracts.values().cloned().collect())
    }

    pub fn update_contract(&self, contract: Contract) -> Result<Contract, LeaseError> {
        let mut contracts = self.contracts.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock contracts: {}", e))
        })?;

        if !contracts.contains_key(&contract.id) {
            return Err(LeaseError::ContractNotFound(contract.id.to_string()));
        }

        let updated_contract = Contract {
            updated_at: Local::now(),
            ..contract
        };

        contracts.insert(updated_contract.id, updated_contract.clone());
        Ok(updated_contract)
    }

    pub fn get_active_contract_for_room(&self, room_id: Uuid) -> Result<Option<Contract>, LeaseError> {
        let contracts = self.contracts.lock().map_err(|e| {
            LeaseError::InternalError(format!("Failed to lock contracts: {}", e))
        })?;

        Ok(contracts.values().find(|c| {
            c.room_id == room_id && (c.status == ContractStatus::Active || c.status == ContractStatus::MonthToMonth)
        }).cloned())
    }
}
