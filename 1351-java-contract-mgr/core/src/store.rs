use std::collections::HashMap;
use std::sync::Mutex;
use uuid::Uuid;
use crate::{Contract, ApprovalRequest, SystemConfig};

pub struct ContractStore {
    contracts: Mutex<HashMap<Uuid, Contract>>,
    approvals: Mutex<HashMap<Uuid, ApprovalRequest>>,
    config: Mutex<SystemConfig>,
}

impl ContractStore {
    pub fn new() -> Self {
        ContractStore {
            contracts: Mutex::new(HashMap::new()),
            approvals: Mutex::new(HashMap::new()),
            config: Mutex::new(SystemConfig::default()),
        }
    }

    pub fn insert_contract(&self, contract: Contract) {
        let mut contracts = self.contracts.lock().unwrap();
        contracts.insert(contract.id, contract);
    }

    pub fn get_contract(&self, id: &Uuid) -> Option<Contract> {
        let contracts = self.contracts.lock().unwrap();
        contracts.get(id).cloned()
    }

    pub fn update_contract(&self, contract: Contract) {
        let mut contracts = self.contracts.lock().unwrap();
        contracts.insert(contract.id, contract);
    }

    pub fn get_all_contracts(&self) -> Vec<Contract> {
        let contracts = self.contracts.lock().unwrap();
        contracts.values().cloned().collect()
    }

    pub fn get_sub_contracts(&self, parent_id: &Uuid) -> Vec<Contract> {
        let contracts = self.contracts.lock().unwrap();
        contracts
            .values()
            .filter(|c| c.parent_id == Some(*parent_id))
            .cloned()
            .collect()
    }

    pub fn insert_approval(&self, approval: ApprovalRequest) {
        let mut approvals = self.approvals.lock().unwrap();
        approvals.insert(approval.id, approval);
    }

    pub fn get_approval(&self, id: &Uuid) -> Option<ApprovalRequest> {
        let approvals = self.approvals.lock().unwrap();
        approvals.get(id).cloned()
    }

    pub fn update_approval(&self, approval: ApprovalRequest) {
        let mut approvals = self.approvals.lock().unwrap();
        approvals.insert(approval.id, approval);
    }

    pub fn get_config(&self) -> SystemConfig {
        self.config.lock().unwrap().clone()
    }

    pub fn update_config(&self, config: SystemConfig) {
        let mut c = self.config.lock().unwrap();
        *c = config;
    }
}

impl Default for ContractStore {
    fn default() -> Self {
        Self::new()
    }
}
