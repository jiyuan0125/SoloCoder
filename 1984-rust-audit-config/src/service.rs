use time::OffsetDateTime;
use uuid::Uuid;

use crate::errors::ConfigError;
use crate::models::{AuditLog, ChangeRequest, ChangeRequestStatus, DiffItem, OperationType, SubmitRequest};
use crate::store::MemoryStore;
use crate::validators::{to_config_value, validate_config_changes};

#[derive(Clone)]
pub struct ConfigService {
    store: MemoryStore,
}

impl ConfigService {
    pub fn new(store: MemoryStore) -> Self {
        ConfigService { store }
    }

    fn create_audit_log(
        &self,
        operator: &str,
        op_type: OperationType,
        namespace: &str,
        change_id: Option<Uuid>,
        diffs: Vec<DiffItem>,
        description: String,
    ) -> AuditLog {
        AuditLog {
            id: Uuid::new_v4(),
            timestamp: OffsetDateTime::now_utc(),
            operator: operator.to_string(),
            operation_type: op_type,
            namespace: namespace.to_string(),
            change_request_id: change_id,
            diffs,
            description,
        }
    }

    pub fn submit_change(&self, req: SubmitRequest) -> Result<ChangeRequest, ConfigError> {
        if self.store.get_namespace(&req.namespace).is_none() {
            return Err(ConfigError::NamespaceNotFound(req.namespace.clone()));
        }
        
        validate_config_changes(&req.changes)?;

        let current_configs = self.store.get_current_configs(&req.namespace)
            .unwrap_or_default();
        
        let mut diffs = Vec::new();
        for change in &req.changes {
            let before = current_configs.get(&change.key).map(|c| c.config.clone());
            let after = to_config_value(change);
            let key = format!("{}/{}", req.namespace, change.key);
            
            if before != Some(after.clone()) {
                diffs.push(DiffItem {
                    key,
                    before,
                    after: Some(after),
                });
            }
        }

        let change_request = ChangeRequest {
            id: Uuid::new_v4(),
            namespace: req.namespace.clone(),
            submitter: req.submitter.clone(),
            submitted_at: OffsetDateTime::now_utc(),
            diffs: diffs.clone(),
            status: ChangeRequestStatus::Pending,
            approver: None,
            approved_at: None,
            canary_percent: Some(req.canary_percent),
            canary_started_at: None,
            canary_duration_secs: req.canary_duration_secs,
        };

        self.store.insert_change_request(change_request.clone());

        let log = self.create_audit_log(
            &req.submitter,
            OperationType::Submit,
            &req.namespace,
            Some(change_request.id),
            diffs,
            format!("Submitted change request {}", change_request.id),
        );
        self.store.record_audit_log(log);

        Ok(change_request)
    }

    pub fn approve_change(&self, id: &Uuid, approver: &str) -> Result<ChangeRequest, ConfigError> {
        let mut request = self.store.get_change_request(id)
            .ok_or_else(|| ConfigError::ChangeRequestNotFound(id.to_string()))?;

        if request.status != ChangeRequestStatus::Pending {
            return Err(ConfigError::InvalidStatusTransition(
                format!("Cannot approve request in {:?} state", request.status)
            ));
        }

        if !self.store.is_approver(&request.namespace, approver) {
            return Err(ConfigError::PermissionDenied(
                format!("{} is not an approver for namespace {}", approver, request.namespace)
            ));
        }

        request.status = ChangeRequestStatus::Approved;
        request.approver = Some(approver.to_string());
        request.approved_at = Some(OffsetDateTime::now_utc());

        self.store.update_change_request(id, request.clone());

        let log = self.create_audit_log(
            approver,
            OperationType::Approve,
            &request.namespace,
            Some(request.id),
            request.diffs.clone(),
            format!("Approved change request {}", request.id),
        );
        self.store.record_audit_log(log);

        Ok(request)
    }

    pub fn reject_change(&self, id: &Uuid, approver: &str) -> Result<ChangeRequest, ConfigError> {
        let mut request = self.store.get_change_request(id)
            .ok_or_else(|| ConfigError::ChangeRequestNotFound(id.to_string()))?;

        if request.status != ChangeRequestStatus::Pending {
            return Err(ConfigError::InvalidStatusTransition(
                format!("Cannot reject request in {:?} state", request.status)
            ));
        }

        if !self.store.is_approver(&request.namespace, approver) {
            return Err(ConfigError::PermissionDenied(
                format!("{} is not an approver for namespace {}", approver, request.namespace)
            ));
        }

        request.status = ChangeRequestStatus::Rejected;
        request.approver = Some(approver.to_string());
        request.approved_at = Some(OffsetDateTime::now_utc());

        self.store.update_change_request(id, request.clone());

        let log = self.create_audit_log(
            approver,
            OperationType::Reject,
            &request.namespace,
            Some(request.id),
            request.diffs.clone(),
            format!("Rejected change request {}", request.id),
        );
        self.store.record_audit_log(log);

        Ok(request)
    }

    pub fn start_canary(&self, id: &Uuid, operator: &str) -> Result<ChangeRequest, ConfigError> {
        let mut request = self.store.get_change_request(id)
            .ok_or_else(|| ConfigError::ChangeRequestNotFound(id.to_string()))?;

        if request.status != ChangeRequestStatus::Approved {
            return Err(ConfigError::InvalidStatusTransition(
                format!("Cannot start canary for request in {:?} state", request.status)
            ));
        }

        request.status = ChangeRequestStatus::Canary;
        request.canary_started_at = Some(OffsetDateTime::now_utc());

        self.store.apply_changes_to_current(&request.namespace, &request.diffs);

        self.store.update_change_request(id, request.clone());

        let log = self.create_audit_log(
            operator,
            OperationType::Canary,
            &request.namespace,
            Some(request.id),
            request.diffs.clone(),
            format!("Started canary deployment for request {} ({}% instances)", 
                request.id, 
                request.canary_percent.unwrap_or(10)
            ),
        );
        self.store.record_audit_log(log);

        Ok(request)
    }

    pub fn deploy_full(&self, id: &Uuid, operator: &str) -> Result<ChangeRequest, ConfigError> {
        let request = self.store.get_change_request(id)
            .ok_or_else(|| ConfigError::ChangeRequestNotFound(id.to_string()))?;

        if request.status != ChangeRequestStatus::Canary {
            return Err(ConfigError::InvalidStatusTransition(
                format!("Cannot deploy full for request in {:?} state", request.status)
            ));
        }

        let version = self.store.snapshot_current_as_approved(operator);

        let mut updated = request.clone();
        updated.status = ChangeRequestStatus::Deployed;

        self.store.update_change_request(id, updated.clone());

        let log = self.create_audit_log(
            operator,
            OperationType::Deploy,
            &request.namespace,
            Some(request.id),
            request.diffs.clone(),
            format!("Deployed to all instances, version {}", version),
        );
        self.store.record_audit_log(log);

        Ok(updated)
    }

    pub fn rollback_change(&self, id: &Uuid, operator: &str) -> Result<(ChangeRequest, Vec<DiffItem>), ConfigError> {
        let request = self.store.get_change_request(id)
            .ok_or_else(|| ConfigError::ChangeRequestNotFound(id.to_string()))?;

        if request.status != ChangeRequestStatus::Canary && request.status != ChangeRequestStatus::Deployed {
            return Err(ConfigError::InvalidStatusTransition(
                format!("Cannot rollback request in {:?} state", request.status)
            ));
        }

        let last_version = self.store.get_last_approved_version()
            .ok_or(ConfigError::NoApprovedVersion)?;

        let rollback_diffs = self.store.rollback_to_version(last_version)?;

        let mut updated = request.clone();
        updated.status = ChangeRequestStatus::RolledBack;

        self.store.update_change_request(id, updated.clone());

        let log = self.create_audit_log(
            operator,
            OperationType::Rollback,
            &request.namespace,
            Some(request.id),
            rollback_diffs.clone(),
            format!("Rolled back to version {}", last_version),
        );
        self.store.record_audit_log(log);

        Ok((updated, rollback_diffs))
    }

    pub fn get_change_request(&self, id: &Uuid) -> Option<ChangeRequest> {
        self.store.get_change_request(id)
    }

    pub fn get_configs(&self, namespace: &str) -> Result<Vec<crate::models::Configuration>, ConfigError> {
        if self.store.get_namespace(namespace).is_none() {
            return Err(ConfigError::NamespaceNotFound(namespace.to_string()));
        }
        
        Ok(self.store.get_current_configs(namespace)
            .unwrap_or_default()
            .into_values()
            .collect())
    }

    pub fn get_audit_logs(&self, namespace: Option<&str>) -> Vec<AuditLog> {
        self.store.get_audit_logs(namespace)
    }
}
