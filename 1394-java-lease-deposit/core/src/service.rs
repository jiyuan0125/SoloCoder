use uuid::Uuid;
use chrono::Local;
use rust_decimal::Decimal;

use crate::error::LeaseError;
use crate::models::{
    Tenant, Room, Contract, ContractStatus, ContractType,
    CreateTenantRequest, CreateRoomRequest, CreateContractRequest,
    CheckoutRequest, CheckoutResult, RenewalReminder, RenewContractRequest,
};
use crate::repository::InMemoryRepository;

pub struct LeaseService {
    repo: InMemoryRepository,
}

impl LeaseService {
    pub fn new() -> Self {
        Self {
            repo: InMemoryRepository::new(),
        }
    }

    pub fn create_tenant(&self, req: CreateTenantRequest) -> Result<Tenant, LeaseError> {
        if req.name.is_empty() {
            return Err(LeaseError::ValidationError("Tenant name cannot be empty".into()));
        }
        if req.phone.is_empty() {
            return Err(LeaseError::ValidationError("Tenant phone cannot be empty".into()));
        }

        self.repo.create_tenant(req.name, req.phone, req.id_card)
    }

    pub fn get_tenant(&self, id: Uuid) -> Result<Option<Tenant>, LeaseError> {
        self.repo.get_tenant(id)
    }

    pub fn list_tenants(&self) -> Result<Vec<Tenant>, LeaseError> {
        self.repo.list_tenants()
    }

    pub fn create_room(&self, req: CreateRoomRequest) -> Result<Room, LeaseError> {
        if req.room_number.is_empty() {
            return Err(LeaseError::ValidationError("Room number cannot be empty".into()));
        }
        if req.area <= Decimal::ZERO {
            return Err(LeaseError::ValidationError("Area must be positive".into()));
        }
        if req.default_monthly_rent <= Decimal::ZERO {
            return Err(LeaseError::ValidationError("Default monthly rent must be positive".into()));
        }

        self.repo.create_room(req.room_number, req.area, req.default_monthly_rent)
    }

    pub fn get_room(&self, id: Uuid) -> Result<Option<Room>, LeaseError> {
        self.repo.get_room(id)
    }

    pub fn get_room_by_number(&self, room_number: &str) -> Result<Option<Room>, LeaseError> {
        self.repo.get_room_by_number(room_number)
    }

    pub fn list_rooms(&self) -> Result<Vec<Room>, LeaseError> {
        self.repo.list_rooms()
    }

    pub fn create_contract(&self, req: CreateContractRequest) -> Result<Contract, LeaseError> {
        let room = self.repo.get_room(req.room_id)?
            .ok_or_else(|| LeaseError::RoomNotFound(req.room_id.to_string()))?;

        let tenant = self.repo.get_tenant(req.tenant_id)?
            .ok_or_else(|| LeaseError::ValidationError("Tenant not found".into()))?;

        if req.start_date >= req.end_date {
            return Err(LeaseError::ValidationError(
                "End date must be after start date".into(),
            ));
        }

        if let Some(_existing) = self.repo.get_active_contract_for_room(req.room_id)? {
            return Err(LeaseError::RoomHasActiveContract(req.room_id.to_string()));
        }

        let monthly_rent = req.monthly_rent.unwrap_or(room.default_monthly_rent);
        if monthly_rent <= Decimal::ZERO {
            return Err(LeaseError::ValidationError("Monthly rent must be positive".into()));
        }

        let deposit_amount = monthly_rent.checked_mul(Decimal::from(2))
            .ok_or_else(|| LeaseError::InvalidAmount("Overflow calculating deposit".into()))?;

        let contract = self.repo.create_contract(
            req.room_id,
            req.tenant_id,
            tenant.name.clone(),
            req.start_date,
            req.end_date,
            monthly_rent,
            deposit_amount,
        )?;

        Ok(contract)
    }

    pub fn get_contract(&self, id: Uuid) -> Result<Option<Contract>, LeaseError> {
        self.repo.get_contract(id)
    }

    pub fn list_contracts(&self) -> Result<Vec<Contract>, LeaseError> {
        self.repo.list_contracts()
    }

    pub fn process_checkin(&self, contract_id: Uuid) -> Result<Contract, LeaseError> {
        let contract = self.repo.get_contract(contract_id)?
            .ok_or_else(|| LeaseError::ContractNotFound(contract_id.to_string()))?;

        if contract.status != ContractStatus::Pending {
            return Err(LeaseError::ValidationError(
                "Contract is not pending check-in".into(),
            ));
        }

        let room = self.repo.get_room(contract.room_id)?
            .ok_or_else(|| LeaseError::RoomNotFound(contract.room_id.to_string()))?;

        let updated_room = Room {
            is_available: false,
            ..room
        };
        self.repo.update_room(updated_room)?;

        let updated_contract = Contract {
            status: ContractStatus::Active,
            deposit_paid: true,
            first_month_rent_paid: true,
            ..contract
        };

        self.repo.update_contract(updated_contract)
    }

    pub fn get_renewal_reminders(&self) -> Result<Vec<RenewalReminder>, LeaseError> {
        let today = Local::now().date_naive();
        let contracts = self.repo.list_contracts()?;
        let mut reminders = Vec::new();

        for contract in contracts {
            if contract.status == ContractStatus::Active && contract.contract_type == ContractType::FixedTerm {
                let days_remaining = (contract.end_date - today).num_days();
                if days_remaining > 0 && days_remaining <= 30 {
                    let room = self.repo.get_room(contract.room_id)?
                        .ok_or_else(|| LeaseError::RoomNotFound(contract.room_id.to_string()))?;

                    reminders.push(RenewalReminder {
                        contract_id: contract.id,
                        room_number: room.room_number,
                        tenant_name: contract.tenant_name,
                        end_date: contract.end_date,
                        days_remaining,
                    });
                }
            }
        }

        reminders.sort_by_key(|r| r.days_remaining);
        Ok(reminders)
    }

    pub fn process_expired_contract(&self, contract_id: Uuid) -> Result<Contract, LeaseError> {
        let contract = self.repo.get_contract(contract_id)?
            .ok_or_else(|| LeaseError::ContractNotFound(contract_id.to_string()))?;

        if contract.status != ContractStatus::Active {
            return Err(LeaseError::ContractNotActive);
        }

        if contract.contract_type != ContractType::FixedTerm {
            return Err(LeaseError::ContractAlreadyMonthToMonth);
        }

        let today = Local::now().date_naive();
        if today <= contract.end_date {
            return Err(LeaseError::ValidationError(
                "Contract has not expired yet".into(),
            ));
        }

        let updated_contract = Contract {
            contract_type: ContractType::MonthToMonth,
            status: ContractStatus::MonthToMonth,
            ..contract
        };

        self.repo.update_contract(updated_contract)
    }

    pub fn renew_contract(&self, req: RenewContractRequest) -> Result<Contract, LeaseError> {
        let contract = self.repo.get_contract(req.contract_id)?
            .ok_or_else(|| LeaseError::ContractNotFound(req.contract_id.to_string()))?;

        if contract.status != ContractStatus::Active {
            return Err(LeaseError::ContractNotActive);
        }

        if contract.contract_type != ContractType::FixedTerm {
            return Err(LeaseError::ContractNotFixedTerm);
        }

        if req.new_end_date <= contract.end_date {
            return Err(LeaseError::ValidationError(
                "New end date must be after current end date".into(),
            ));
        }

        let monthly_rent = req.new_monthly_rent.unwrap_or(contract.monthly_rent);
        if monthly_rent <= Decimal::ZERO {
            return Err(LeaseError::ValidationError("Monthly rent must be positive".into()));
        }

        let updated_contract = Contract {
            end_date: req.new_end_date,
            monthly_rent,
            ..contract
        };

        self.repo.update_contract(updated_contract)
    }

    pub fn process_checkout(&self, req: CheckoutRequest) -> Result<CheckoutResult, LeaseError> {
        let contract = self.repo.get_contract(req.contract_id)?
            .ok_or_else(|| LeaseError::ContractNotFound(req.contract_id.to_string()))?;

        if contract.status != ContractStatus::Active && contract.status != ContractStatus::MonthToMonth {
            return Err(LeaseError::ContractNotActive);
        }

        if req.checkout_date < contract.start_date {
            return Err(LeaseError::ValidationError(
                "Checkout date cannot be before contract start date".into(),
            ));
        }

        let damage_fee = req.damage_fee.unwrap_or(Decimal::ZERO);
        if damage_fee < Decimal::ZERO {
            return Err(LeaseError::ValidationError("Damage fee cannot be negative".into()));
        }

        let early_termination_fee = match contract.contract_type {
            ContractType::FixedTerm => {
                if req.checkout_date < contract.end_date {
                    contract.monthly_rent.checked_mul(Decimal::from(2))
                        .ok_or_else(|| LeaseError::InvalidAmount("Overflow calculating termination fee".into()))?
                } else {
                    Decimal::ZERO
                }
            }
            ContractType::MonthToMonth => Decimal::ZERO,
        };

        let total_deductions = early_termination_fee.checked_add(damage_fee)
            .ok_or_else(|| LeaseError::InvalidAmount("Overflow calculating total deductions".into()))?;

        let deposit_refund = if total_deductions < contract.deposit_amount {
            contract.deposit_amount.checked_sub(total_deductions)
                .ok_or_else(|| LeaseError::InvalidAmount("Overflow calculating refund".into()))?
        } else {
            Decimal::ZERO
        };

        let additional_payment = if total_deductions > contract.deposit_amount {
            total_deductions.checked_sub(contract.deposit_amount)
                .ok_or_else(|| LeaseError::InvalidAmount("Overflow calculating additional payment".into()))?
        } else {
            Decimal::ZERO
        };

        let room = self.repo.get_room(contract.room_id)?
            .ok_or_else(|| LeaseError::RoomNotFound(contract.room_id.to_string()))?;

        let updated_room = Room {
            is_available: true,
            ..room
        };
        self.repo.update_room(updated_room)?;

        let updated_contract = Contract {
            status: ContractStatus::Terminated,
            ..contract
        };
        self.repo.update_contract(updated_contract)?;

        Ok(CheckoutResult {
            contract_id: contract.id,
            room_id: contract.room_id,
            tenant_id: contract.tenant_id,
            checkout_date: req.checkout_date,
            total_damage_fee: damage_fee,
            early_termination_fee,
            deposit_refund,
            additional_payment,
        })
    }
}

impl Default for LeaseService {
    fn default() -> Self {
        Self::new()
    }
}
