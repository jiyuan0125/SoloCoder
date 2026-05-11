use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::errors::{ComplaintError, Result};
use crate::models::*;
use crate::sla::calculate_deadline;
use crate::store::InMemoryStore;

#[derive(Clone)]
pub struct ComplaintService {
    store: InMemoryStore,
}

impl ComplaintService {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn create_complaint(&self, req: CreateComplaintRequest) -> Result<Complaint> {
        if let Some(handler_id) = &req.handler_id {
            if self.store.get_handler(handler_id)?.is_none() {
                return Err(ComplaintError::HandlerNotFound(handler_id.clone()));
            }
        }

        let now = Utc::now();
        let deadline = calculate_deadline(now, req.severity.response_hours());
        
        let complaint = Complaint::new(
            req.title,
            req.description,
            req.customer_name,
            req.customer_contact,
            req.severity,
            req.handler_id,
            now,
            deadline,
        );

        self.store.save_complaint(complaint.clone())?;
        tracing::info!(complaint_id = %complaint.id, "创建投诉成功");
        
        Ok(complaint)
    }

    pub fn get_complaint(&self, id: Uuid) -> Result<Complaint> {
        self.store.get_complaint(id)?.ok_or_else(|| ComplaintError::NotFound(id.to_string()))
    }

    pub fn list_complaints(&self) -> Result<Vec<Complaint>> {
        self.store.list_complaints()
    }

    pub fn respond_to_complaint(&self, id: Uuid, req: RespondToComplaintRequest) -> Result<Complaint> {
        let mut complaint = self.get_complaint(id)?;
        
        if !matches!(complaint.status, Status::Pending | Status::Reopened | Status::Escalated) {
            return Err(ComplaintError::InvalidState);
        }

        if self.store.get_handler(&req.handler_id)?.is_none() {
            return Err(ComplaintError::HandlerNotFound(req.handler_id.clone()));
        }

        let now = Utc::now();
        complaint.status = Status::Processing;
        complaint.handler_id = Some(req.handler_id.clone());
        if complaint.response_time.is_none() {
            complaint.response_time = Some(now);
        }
        
        complaint.add_history(
            "响应投诉",
            Some(format!("处理人: {}，响应内容: {}", req.handler_id, req.response)),
            req.handler_id,
        );

        self.store.save_complaint(complaint.clone())?;
        tracing::info!(complaint_id = %id, "响应投诉成功");
        
        Ok(complaint)
    }

    pub fn propose_solution(&self, id: Uuid, req: ProposeSolutionRequest) -> Result<Complaint> {
        let mut complaint = self.get_complaint(id)?;
        
        if !matches!(complaint.status, Status::Processing) {
            return Err(ComplaintError::InvalidState);
        }

        if self.store.get_handler(&req.handler_id)?.is_none() {
            return Err(ComplaintError::HandlerNotFound(req.handler_id.clone()));
        }

        complaint.status = Status::AwaitingFeedback;
        complaint.add_history(
            "提交处理方案",
            Some(format!("方案: {}", req.solution)),
            req.handler_id,
        );

        self.store.save_complaint(complaint.clone())?;
        tracing::info!(complaint_id = %id, "提交处理方案");
        
        Ok(complaint)
    }

    pub fn customer_feedback(&self, id: Uuid, req: CustomerFeedbackRequest) -> Result<Complaint> {
        let mut complaint = self.get_complaint(id)?;
        
        if complaint.status != Status::AwaitingFeedback {
            return Err(ComplaintError::InvalidState);
        }

        if req.accepted {
            let now = Utc::now();
            complaint.status = Status::Closed;
            complaint.closed_at = Some(now);
            complaint.add_history(
                "客户接受方案",
                req.comments.clone(),
                "客户".to_string(),
            );
            tracing::info!(complaint_id = %id, "客户接受方案，投诉结案");
        } else {
            complaint.status = Status::Reopened;
            complaint.reopen_count += 1;
            let now = Utc::now();
            complaint.deadline = calculate_deadline(now, complaint.severity.response_hours());
            complaint.add_history(
                "客户不接受方案，重新打开投诉",
                req.comments.clone(),
                "客户".to_string(),
            );
            tracing::info!(complaint_id = %id, "客户不接受方案，投诉重新打开");
        }

        self.store.save_complaint(complaint.clone())?;
        Ok(complaint)
    }

    pub fn follow_up_feedback(&self, id: Uuid, req: FollowUpFeedbackRequest) -> Result<Complaint> {
        let mut complaint = self.get_complaint(id)?;
        
        if complaint.status != Status::Closed {
            return Err(ComplaintError::InvalidState);
        }

        let follow_up = FollowUp {
            id: Uuid::new_v4(),
            complaint_id: id,
            satisfaction: req.satisfied,
            comments: req.comments.clone(),
            created_at: Utc::now(),
        };
        complaint.follow_ups.push(follow_up);

        if !req.satisfied {
            complaint.status = Status::Reopened;
            complaint.closed_at = None;
            complaint.reopen_count += 1;
            
            let current_handler_level = complaint
                .handler_id
                .as_ref()
                .and_then(|id| self.store.get_handler(id).ok().flatten())
                .map(|h| h.level)
                .unwrap_or(1);

            if let Some(higher_handler) = self.store.get_higher_level_handler(current_handler_level)? {
                complaint.handler_id = Some(higher_handler.id.clone());
                let new_severity = match complaint.severity {
                    Severity::General => Severity::Serious,
                    Severity::Serious => Severity::Urgent,
                    Severity::Urgent => Severity::Urgent,
                };
                complaint.severity = new_severity;
            }

            let now = Utc::now();
            complaint.deadline = calculate_deadline(now, complaint.severity.response_hours());
            
            complaint.add_history(
                "回访不满意，重新打开并升级",
                req.comments,
                "系统".to_string(),
            );
            tracing::info!(complaint_id = %id, "回访不满意，投诉重新打开并升级");
        } else {
            complaint.add_history(
                "回访满意",
                req.comments,
                "系统".to_string(),
            );
            tracing::info!(complaint_id = %id, "回访满意");
        }

        self.store.save_complaint(complaint.clone())?;
        Ok(complaint)
    }

    pub fn check_and_escalate(&self, now: DateTime<Utc>) -> Result<Vec<Complaint>> {
        let complaints = self.store.list_complaints()?;
        let mut escalated = Vec::new();

        for mut complaint in complaints {
            if complaint.status == Status::Closed {
                continue;
            }

            if !complaint.is_overdue(now) {
                continue;
            }

            let handler_level = complaint
                .handler_id
                .as_ref()
                .and_then(|id| self.store.get_handler(id).ok().flatten())
                .map(|h| h.level)
                .unwrap_or(1);

            if let Some(new_severity) = complaint.severity.upgrade() {
                let old_severity = complaint.severity;
                complaint.severity = new_severity;
                complaint.deadline = calculate_deadline(now, new_severity.response_hours());
                
                let record = EscalationRecord {
                    id: Uuid::new_v4(),
                    from_severity: Some(old_severity),
                    to_severity: Some(new_severity),
                    reason: "超时未响应".to_string(),
                    escalated_at: now,
                    new_handler_id: None,
                };
                complaint.escalations.push(record);
                complaint.add_history(
                    "自动升级",
                    Some(format!("从 {} 升级到 {}", old_severity, new_severity)),
                    "系统".to_string(),
                );
                
                tracing::info!(complaint_id = %complaint.id, ?old_severity, ?new_severity, "投诉自动升级");
            } else {
                if let Some(higher_handler) = self.store.get_higher_level_handler(handler_level)? {
                    let old_handler_id = complaint.handler_id.clone();
                    complaint.handler_id = Some(higher_handler.id.clone());
                    complaint.status = Status::Escalated;
                    complaint.deadline = calculate_deadline(now, complaint.severity.response_hours());

                    let record = EscalationRecord {
                        id: Uuid::new_v4(),
                        from_severity: None,
                        to_severity: None,
                        reason: "紧急投诉超时，升级给上级主管".to_string(),
                        escalated_at: now,
                        new_handler_id: Some(higher_handler.id.clone()),
                    };
                    complaint.escalations.push(record);
                    complaint.add_history(
                        "升级给上级主管",
                        Some(format!("从处理人 {:?} 升级到 {}", old_handler_id, higher_handler.name)),
                        "系统".to_string(),
                    );
                    
                    tracing::info!(complaint_id = %complaint.id, new_handler = %higher_handler.id, "紧急投诉超时，升级给上级主管");
                }
            }

            self.store.save_complaint(complaint.clone())?;
            escalated.push(complaint);
        }

        Ok(escalated)
    }

    pub fn list_handlers(&self) -> Result<Vec<Handler>> {
        self.store.list_handlers()
    }

    pub fn add_handler(&self, handler: Handler) -> Result<()> {
        self.store.add_handler(handler)
    }
}
