use crate::errors::VisitorError;
use crate::models::{
    ApprovalDecision, PasscodeVerification, VisitorRecord, VisitorRegistration, VisitorStatus,
};
use chrono::{Duration, Utc};
use rand::Rng;
use std::collections::HashMap;
use uuid::Uuid;

pub const PASSCODE_VALIDITY_HOURS: i64 = 4;
const PASSCODE_LENGTH: usize = 6;
const MAX_PASSCODE_RETRIES: usize = 100;

pub struct VisitorService {
    visitors: HashMap<Uuid, VisitorRecord>,
    passcodes: HashMap<String, Uuid>,
}

impl VisitorService {
    pub fn new() -> Self {
        Self {
            visitors: HashMap::new(),
            passcodes: HashMap::new(),
        }
    }

    pub fn register_visitor(
        &mut self,
        registration: VisitorRegistration,
    ) -> Result<VisitorRecord, VisitorError> {
        self.validate_registration(&registration)?;
        let record = VisitorRecord::new(registration);
        self.visitors.insert(record.id, record.clone());
        Ok(record)
    }

    pub fn update_visitor(
        &mut self,
        id: Uuid,
        registration: VisitorRegistration,
    ) -> Result<VisitorRecord, VisitorError> {
        self.validate_registration(&registration)?;
        let record = self
            .visitors
            .get_mut(&id)
            .ok_or(VisitorError::VisitorNotFound)?;
        record.update_from_registration(registration);
        Ok(record.clone())
    }

    pub fn approve_visitor(
        &mut self,
        id: Uuid,
        decision: ApprovalDecision,
    ) -> Result<VisitorRecord, VisitorError> {
        let exists = self.visitors.contains_key(&id);
        if !exists {
            return Err(VisitorError::VisitorNotFound);
        }

        let status = self.visitors[&id].status;
        if status != VisitorStatus::PendingApproval {
            return Err(VisitorError::InvalidStatusTransition);
        }

        if decision.approved {
            let passcode = self.generate_unique_passcode()?;
            let expires_at = Utc::now() + Duration::hours(PASSCODE_VALIDITY_HOURS);

            let record = self.visitors.get_mut(&id).unwrap();
            record.passcode = Some(passcode.clone());
            record.passcode_expires_at = Some(expires_at);
            record.status = VisitorStatus::Approved;
            record.updated_at = Utc::now();

            self.passcodes.insert(passcode, id);
        } else {
            let record = self.visitors.get_mut(&id).unwrap();
            record.status = VisitorStatus::Rejected;
            record.updated_at = Utc::now();
        }

        Ok(self.visitors[&id].clone())
    }

    pub fn verify_passcode(
        &mut self,
        verification: PasscodeVerification,
    ) -> Result<VisitorRecord, VisitorError> {
        self.expire_passcodes();

        let visitor_id = self
            .passcodes
            .get(&verification.passcode)
            .ok_or(VisitorError::PasscodeInvalid)?;

        let record = self
            .visitors
            .get_mut(visitor_id)
            .ok_or(VisitorError::VisitorNotFound)?;

        if record.passcode_used {
            return Err(VisitorError::PasscodeAlreadyUsed);
        }

        let expires_at = record
            .passcode_expires_at
            .ok_or(VisitorError::PasscodeInvalid)?;

        if Utc::now() > expires_at {
            return Err(VisitorError::PasscodeExpired);
        }

        record.passcode_used = true;
        record.status = VisitorStatus::Visiting;
        record.arrival_time = Some(Utc::now());
        record.updated_at = Utc::now();

        self.passcodes.remove(&verification.passcode);

        Ok(record.clone())
    }

    pub fn sign_out(&mut self, id: Uuid) -> Result<VisitorRecord, VisitorError> {
        let record = self
            .visitors
            .get_mut(&id)
            .ok_or(VisitorError::VisitorNotFound)?;

        if record.status != VisitorStatus::Visiting {
            return Err(VisitorError::InvalidStatusTransition);
        }

        record.status = VisitorStatus::Completed;
        record.departure_time = Some(Utc::now());
        record.updated_at = Utc::now();

        Ok(record.clone())
    }

    pub fn get_visitor(&self, id: Uuid) -> Option<VisitorRecord> {
        self.visitors.get(&id).cloned()
    }

    pub fn get_all_visitors(&self) -> Vec<VisitorRecord> {
        self.visitors.values().cloned().collect()
    }

    pub fn get_visitors_by_host(&self, host: &str) -> Vec<VisitorRecord> {
        self.visitors
            .values()
            .filter(|v| v.host == host)
            .cloned()
            .collect()
    }

    pub fn expire_passcodes(&mut self) {
        let now = Utc::now();
        let expired_passcodes: Vec<String> = self
            .passcodes
            .keys()
            .filter(|passcode| {
                self.passcodes
                    .get(*passcode)
                    .and_then(|id| self.visitors.get(id))
                    .and_then(|v| v.passcode_expires_at)
                    .map(|expires_at| now > expires_at)
                    .unwrap_or(false)
            })
            .cloned()
            .collect();

        for passcode in expired_passcodes {
            if let Some(visitor_id) = self.passcodes.remove(&passcode) {
                if let Some(record) = self.visitors.get_mut(&visitor_id) {
                    if record.status == VisitorStatus::Approved && !record.passcode_used {
                        record.status = VisitorStatus::PasscodeExpired;
                        record.updated_at = Utc::now();
                    }
                }
            }
        }

        self.mark_not_signed_out();
    }

    pub fn mark_not_signed_out(&mut self) {
        let now = Utc::now();
        for record in self.visitors.values_mut() {
            if record.status == VisitorStatus::Visiting {
                if let Some(expires_at) = record.passcode_expires_at {
                    if now > expires_at {
                        record.status = VisitorStatus::NotSignedOut;
                        record.updated_at = now;
                    }
                }
            }
        }
    }

    fn generate_unique_passcode(&self) -> Result<String, VisitorError> {
        let mut rng = rand::thread_rng();

        for _ in 0..MAX_PASSCODE_RETRIES {
            let passcode: String = (0..PASSCODE_LENGTH)
                .map(|_| rng.gen_range(0..10).to_string())
                .collect();

            if !self.passcodes.contains_key(&passcode) {
                return Ok(passcode);
            }
        }

        Err(VisitorError::PasscodeGenerationFailed)
    }

    fn validate_registration(
        &self,
        registration: &VisitorRegistration,
    ) -> Result<(), VisitorError> {
        if registration.name.trim().is_empty() {
            return Err(VisitorError::InvalidInput("姓名不能为空".to_string()));
        }
        if registration.phone.trim().is_empty() {
            return Err(VisitorError::InvalidInput("手机号不能为空".to_string()));
        }
        if registration.id_card.trim().is_empty() {
            return Err(VisitorError::InvalidInput("身份证号不能为空".to_string()));
        }
        if registration.purpose.trim().is_empty() {
            return Err(VisitorError::InvalidInput("来访事由不能为空".to_string()));
        }
        if registration.host.trim().is_empty() {
            return Err(VisitorError::InvalidInput("被访人不能为空".to_string()));
        }
        Ok(())
    }
}

impl Default for VisitorService {
    fn default() -> Self {
        Self::new()
    }
}
