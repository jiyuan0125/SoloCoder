use chrono::{DateTime, Utc};
use uuid::Uuid;
use crate::{
    Contract, ContractError, ContractStatus, ApprovalRequest, ApprovalStatus,
    ApprovalType, ReminderInfo, CreateContractRequest,
    RenewalRequest, PriceAdjustmentRequest, calculate_approval_level_for_renewal,
    needs_approval_for_price_change, ContractStore,
};

pub struct ContractService {
    store: ContractStore,
}

impl ContractService {
    pub fn new(store: ContractStore) -> Self {
        ContractService { store }
    }

    pub fn create_contract(&self, req: CreateContractRequest) -> Result<Contract, ContractError> {
        if req.amount <= 0.0 {
            return Err(ContractError::Validation("合同金额必须大于0".to_string()));
        }

        if let Some(parent_id) = req.parent_id {
            let parent = self.store.get_contract(&parent_id)
                .ok_or_else(|| ContractError::NotFound(format!("父合同不存在: {}", parent_id)))?;
            if !parent.is_framework {
                return Err(ContractError::Validation("子合同的父合同必须是框架合同".to_string()));
            }
            if req.is_framework {
                return Err(ContractError::Validation("子合同不能是框架合同".to_string()));
            }
        }

        let contract = Contract::new(req);
        self.store.insert_contract(contract.clone());
        Ok(contract)
    }

    pub fn get_contract(&self, id: &Uuid) -> Result<Contract, ContractError> {
        self.store.get_contract(id)
            .ok_or_else(|| ContractError::NotFound(id.to_string()))
    }

    pub fn get_all_contracts(&self) -> Vec<Contract> {
        self.store.get_all_contracts()
    }

    pub fn activate_contract(&self, id: &Uuid) -> Result<Contract, ContractError> {
        let mut contract = self.get_contract(id)?;
        
        if !contract.status.can_transition_to(ContractStatus::Active) {
            return Err(ContractError::InvalidStateTransition {
                current: contract.status.name().to_string(),
                target: ContractStatus::Active.name().to_string(),
            });
        }

        contract.status = ContractStatus::Active;
        contract.updated_at = Utc::now();
        self.store.update_contract(contract.clone());
        Ok(contract)
    }

    pub fn request_renewal(&self, id: &Uuid, req: RenewalRequest) -> Result<ApprovalRequest, ContractError> {
        let mut contract = self.get_contract(id)?;

        if !contract.status.can_renew() {
            return Err(ContractError::InvalidOperation(
                format!("合同状态 {} 不能续签", contract.status.name())
            ));
        }

        if contract.status == ContractStatus::Terminated {
            return Err(ContractError::InvalidOperation("已终止的合同不能续签".to_string()));
        }

        if req.new_amount <= 0.0 {
            return Err(ContractError::Validation("续签金额必须大于0".to_string()));
        }

        let approval_level = calculate_approval_level_for_renewal(contract.amount, req.new_amount);
        
        let approval = ApprovalRequest {
            id: Uuid::new_v4(),
            contract_id: *id,
            approval_type: ApprovalType::Renewal,
            level: approval_level,
            status: ApprovalStatus::Pending,
            old_amount: contract.amount,
            new_amount: req.new_amount,
            created_at: Utc::now(),
            decided_at: None,
            reason: None,
        };

        contract.status = ContractStatus::RenewalPending;
        contract.updated_at = Utc::now();
        self.store.update_contract(contract);
        self.store.insert_approval(approval.clone());

        Ok(approval)
    }

    pub fn approve_renewal(&self, approval_id: &Uuid, end_date: DateTime<Utc>) -> Result<Contract, ContractError> {
        let mut approval = self.store.get_approval(approval_id)
            .ok_or_else(|| ContractError::ApprovalNotFound(approval_id.to_string()))?;

        if approval.status != ApprovalStatus::Pending {
            return Err(ContractError::InvalidOperation("审批已处理".to_string()));
        }

        let mut contract = self.get_contract(&approval.contract_id)?;

        approval.status = ApprovalStatus::Approved;
        approval.decided_at = Some(Utc::now());
        self.store.update_approval(approval.clone());

        contract.amount = contract.pending_adjustment
            .map(|p| p.new_amount)
            .unwrap_or(approval.new_amount);
        contract.effective_amount = contract.amount;
        contract.end_date = end_date;
        contract.status = ContractStatus::Active;
        contract.updated_at = Utc::now();
        contract.pending_adjustment = None;

        self.store.update_contract(contract.clone());
        Ok(contract)
    }

    pub fn reject_renewal(&self, approval_id: &Uuid, reason: String) -> Result<(), ContractError> {
        let mut approval = self.store.get_approval(approval_id)
            .ok_or_else(|| ContractError::ApprovalNotFound(approval_id.to_string()))?;

        if approval.status != ApprovalStatus::Pending {
            return Err(ContractError::InvalidOperation("审批已处理".to_string()));
        }

        let mut contract = self.get_contract(&approval.contract_id)?;

        approval.status = ApprovalStatus::Rejected;
        approval.decided_at = Some(Utc::now());
        approval.reason = Some(reason);
        self.store.update_approval(approval);

        contract.status = ContractStatus::Active;
        contract.updated_at = Utc::now();
        contract.pending_adjustment = None;
        self.store.update_contract(contract);

        Ok(())
    }

    pub fn adjust_price(&self, id: &Uuid, req: PriceAdjustmentRequest) -> Result<Option<ApprovalRequest>, ContractError> {
        let mut contract = self.get_contract(id)?;

        if contract.status != ContractStatus::Active {
            return Err(ContractError::InvalidOperation(
                format!("只有生效中的合同才能调整价格，当前状态: {}", contract.status.name())
            ));
        }

        if contract.pending_adjustment.is_some() {
            return Err(ContractError::PriceAdjustmentPending);
        }

        if req.new_amount <= 0.0 {
            return Err(ContractError::Validation("新金额必须大于0".to_string()));
        }

        let approval_type = if req.new_amount > contract.amount {
            ApprovalType::PriceIncrease
        } else {
            ApprovalType::PriceDecrease
        };

        if let Some(level) = needs_approval_for_price_change(contract.amount, req.new_amount) {
            let approval = ApprovalRequest {
                id: Uuid::new_v4(),
                contract_id: *id,
                approval_type,
                level,
                status: ApprovalStatus::Pending,
                old_amount: contract.amount,
                new_amount: req.new_amount,
                created_at: Utc::now(),
                decided_at: None,
                reason: None,
            };

            contract.pending_adjustment = Some(crate::PendingPriceAdjustment {
                approval_id: approval.id,
                new_amount: req.new_amount,
            });
            contract.updated_at = Utc::now();
            self.store.update_contract(contract);
            self.store.insert_approval(approval.clone());

            Ok(Some(approval))
        } else {
            contract.amount = req.new_amount;
            contract.effective_amount = req.new_amount;
            contract.updated_at = Utc::now();
            self.store.update_contract(contract.clone());
            Ok(None)
        }
    }

    pub fn approve_price_adjustment(&self, approval_id: &Uuid) -> Result<Contract, ContractError> {
        let mut approval = self.store.get_approval(approval_id)
            .ok_or_else(|| ContractError::ApprovalNotFound(approval_id.to_string()))?;

        if approval.status != ApprovalStatus::Pending {
            return Err(ContractError::InvalidOperation("审批已处理".to_string()));
        }

        let mut contract = self.get_contract(&approval.contract_id)?;

        approval.status = ApprovalStatus::Approved;
        approval.decided_at = Some(Utc::now());
        self.store.update_approval(approval);

        contract.amount = contract.pending_adjustment
            .map(|p| p.new_amount)
            .unwrap_or(contract.amount);
        contract.effective_amount = contract.amount;
        contract.pending_adjustment = None;
        contract.updated_at = Utc::now();
        self.store.update_contract(contract.clone());

        Ok(contract)
    }

    pub fn reject_price_adjustment(&self, approval_id: &Uuid, reason: String) -> Result<(), ContractError> {
        let mut approval = self.store.get_approval(approval_id)
            .ok_or_else(|| ContractError::ApprovalNotFound(approval_id.to_string()))?;

        if approval.status != ApprovalStatus::Pending {
            return Err(ContractError::InvalidOperation("审批已处理".to_string()));
        }

        let mut contract = self.get_contract(&approval.contract_id)?;

        approval.status = ApprovalStatus::Rejected;
        approval.decided_at = Some(Utc::now());
        approval.reason = Some(reason);
        self.store.update_approval(approval);

        contract.pending_adjustment = None;
        contract.updated_at = Utc::now();
        self.store.update_contract(contract);

        Ok(())
    }

    pub fn terminate_contract(&self, id: &Uuid) -> Result<Contract, ContractError> {
        let mut contract = self.get_contract(id)?;

        if !contract.status.can_transition_to(ContractStatus::Terminated) {
            return Err(ContractError::InvalidStateTransition {
                current: contract.status.name().to_string(),
                target: ContractStatus::Terminated.name().to_string(),
            });
        }

        contract.status = ContractStatus::Terminated;
        contract.updated_at = Utc::now();
        self.store.update_contract(contract.clone());
        Ok(contract)
    }

    pub fn close_contract(&self, id: &Uuid) -> Result<Contract, ContractError> {
        let contract = self.get_contract(id)?;

        if contract.is_framework {
            let sub_contracts = self.store.get_sub_contracts(id);
            let has_active_subs = sub_contracts.iter().any(|c| {
                matches!(c.status, ContractStatus::Active | ContractStatus::Expiring | ContractStatus::RenewalPending)
            });
            if has_active_subs {
                return Err(ContractError::HasActiveSubContracts);
            }
        }

        let mut contract = contract;
        if !contract.status.can_close() {
            return Err(ContractError::InvalidStateTransition {
                current: contract.status.name().to_string(),
                target: ContractStatus::Closed.name().to_string(),
            });
        }

        contract.status = ContractStatus::Closed;
        contract.updated_at = Utc::now();
        self.store.update_contract(contract.clone());
        Ok(contract)
    }

    pub fn get_expiring_contracts_reminders(&self) -> Vec<ReminderInfo> {
        let now = Utc::now();
        let config = self.store.get_config();
        let contracts = self.get_all_contracts();

        contracts
            .iter()
            .filter(|c| matches!(c.status, ContractStatus::Active | ContractStatus::Expiring))
            .filter_map(|c| {
                let days = c.days_until_expiry(now);
                if days < 0 {
                    return None;
                }
                c.get_reminder_level(&config.reminder_config, now).map(|level| {
                    ReminderInfo {
                        contract_id: c.id,
                        contract_name: c.name.clone(),
                        days_until_expiry: days,
                        level,
                        recipients: c.get_reminder_recipients(&config.reminder_config, level),
                        end_date: c.end_date,
                    }
                })
            })
            .collect()
    }

    pub fn get_config(&self) -> crate::SystemConfig {
        self.store.get_config()
    }

    pub fn update_config(&self, config: crate::SystemConfig) {
        self.store.update_config(config)
    }

    pub fn get_approval(&self, id: &Uuid) -> Result<ApprovalRequest, ContractError> {
        self.store.get_approval(id)
            .ok_or_else(|| ContractError::ApprovalNotFound(id.to_string()))
    }
}
