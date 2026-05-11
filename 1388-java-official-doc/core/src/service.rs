use std::collections::HashMap;
use std::sync::Mutex;

use chrono::{Datelike, Duration, Utc};
use uuid::Uuid;

use crate::error::WorkflowError;
use crate::models::{
    Document, DocumentType, HistoryAction, StageStatus, UrgencyLevel, WorkflowHistory, WorkflowStage,
};

pub struct WorkflowService {
    documents: Mutex<HashMap<Uuid, Document>>,
    history: Mutex<Vec<WorkflowHistory>>,
    stage_statuses: Mutex<HashMap<Uuid, StageStatus>>,
    counter: Mutex<HashMap<i32, u32>>,
}

impl WorkflowService {
    pub fn new() -> Self {
        WorkflowService {
            documents: Mutex::new(HashMap::new()),
            history: Mutex::new(Vec::new()),
            stage_statuses: Mutex::new(HashMap::new()),
            counter: Mutex::new(HashMap::new()),
        }
    }

    fn generate_document_number(&self) -> String {
        let now = Utc::now();
        let year = now.year();
        let mut counter = self.counter.lock().unwrap();
        let current = counter.entry(year).or_insert(0);
        *current += 1;
        format!("{}-{:04}", year, current)
    }

    pub fn create_document(
        &self,
        title: String,
        content: String,
        doc_type: DocumentType,
        urgency: UrgencyLevel,
        author: String,
    ) -> Document {
        let now = Utc::now();
        let id = Uuid::new_v4();
        let number = self.generate_document_number();

        let doc = Document {
            id,
            number,
            title,
            content,
            doc_type,
            urgency,
            author: author.clone(),
            created_at: now,
            updated_at: now,
            current_stage: WorkflowStage::Draft,
            is_overdue: false,
        };

        self.documents.lock().unwrap().insert(id, doc.clone());

        let history_record = WorkflowHistory {
            id: Uuid::new_v4(),
            document_id: id,
            action: HistoryAction::Create,
            operator: author,
            timestamp: now,
        };
        self.history.lock().unwrap().push(history_record);

        doc
    }

    pub fn submit_document(
        &self,
        doc_id: Uuid,
        operator: String,
    ) -> Result<Document, WorkflowError> {
        let mut docs = self.documents.lock().unwrap();
        let doc = docs
            .get_mut(&doc_id)
            .ok_or(WorkflowError::DocumentNotFound)?;

        if doc.current_stage != WorkflowStage::Draft {
            return Err(WorkflowError::InvalidTransition {
                action: "submit".to_string(),
                current_stage: doc.current_stage.to_str().to_string(),
            });
        }

        let now = Utc::now();
        let from_stage = doc.current_stage;
        doc.current_stage = WorkflowStage::DepartmentReview;
        doc.updated_at = now;

        let history_record = WorkflowHistory {
            id: Uuid::new_v4(),
            document_id: doc_id,
            action: HistoryAction::Submit {
                from: from_stage,
                to: WorkflowStage::DepartmentReview,
            },
            operator,
            timestamp: now,
        };
        self.history.lock().unwrap().push(history_record);

        self.stage_statuses.lock().unwrap().insert(
            doc_id,
            StageStatus {
                document_id: doc_id,
                stage: WorkflowStage::DepartmentReview,
                entered_at: now,
                assignee: "部门审核员".to_string(),
            },
        );

        Ok(doc.clone())
    }

    pub fn department_review(
        &self,
        doc_id: Uuid,
        operator: String,
        approve: bool,
        reason: Option<String>,
    ) -> Result<Document, WorkflowError> {
        let mut docs = self.documents.lock().unwrap();
        let doc = docs
            .get_mut(&doc_id)
            .ok_or(WorkflowError::DocumentNotFound)?;

        if doc.current_stage != WorkflowStage::DepartmentReview {
            return Err(WorkflowError::InvalidTransition {
                action: "review".to_string(),
                current_stage: doc.current_stage.to_str().to_string(),
            });
        }

        let now = Utc::now();
        let from_stage = doc.current_stage;

        if approve {
            let next_stage = if doc.doc_type.needs_countersignature() {
                WorkflowStage::Countersignature
            } else {
                WorkflowStage::Issuance
            };

            doc.current_stage = next_stage;
            doc.updated_at = now;

            let history_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::ReviewApprove,
                operator,
                timestamp: now,
            };
            self.history.lock().unwrap().push(history_record);

            let assignee = if next_stage == WorkflowStage::Countersignature {
                "会签人".to_string()
            } else {
                "签发人".to_string()
            };
            self.stage_statuses.lock().unwrap().insert(
                doc_id,
                StageStatus {
                    document_id: doc_id,
                    stage: next_stage,
                    entered_at: now,
                    assignee,
                },
            );
        } else {
            if reason.is_none() || reason.as_ref().unwrap().is_empty() {
                return Err(WorkflowError::ReturnReasonRequired);
            }
            doc.current_stage = WorkflowStage::Draft;
            doc.updated_at = now;

            let history_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::Return {
                    from: from_stage,
                    reason: reason.unwrap(),
                },
                operator,
                timestamp: now,
            };
            self.history.lock().unwrap().push(history_record);

            self.stage_statuses.lock().unwrap().remove(&doc_id);
        }

        Ok(doc.clone())
    }

    pub fn countersign(
        &self,
        doc_id: Uuid,
        operator: String,
        approve: bool,
        comment: Option<String>,
        reason: Option<String>,
    ) -> Result<Document, WorkflowError> {
        let mut docs = self.documents.lock().unwrap();
        let doc = docs
            .get_mut(&doc_id)
            .ok_or(WorkflowError::DocumentNotFound)?;

        if doc.current_stage != WorkflowStage::Countersignature {
            return Err(WorkflowError::InvalidTransition {
                action: "countersign".to_string(),
                current_stage: doc.current_stage.to_str().to_string(),
            });
        }

        let now = Utc::now();
        let from_stage = doc.current_stage;

        if approve {
            doc.current_stage = WorkflowStage::Issuance;
            doc.updated_at = now;

            let history_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::CountersignApprove { comment },
                operator,
                timestamp: now,
            };
            self.history.lock().unwrap().push(history_record);

            self.stage_statuses.lock().unwrap().insert(
                doc_id,
                StageStatus {
                    document_id: doc_id,
                    stage: WorkflowStage::Issuance,
                    entered_at: now,
                    assignee: "签发人".to_string(),
                },
            );
        } else {
            if reason.is_none() || reason.as_ref().unwrap().is_empty() {
                return Err(WorkflowError::ReturnReasonRequired);
            }
            doc.current_stage = WorkflowStage::Draft;
            doc.updated_at = now;

            let history_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::Return {
                    from: from_stage,
                    reason: reason.unwrap(),
                },
                operator,
                timestamp: now,
            };
            self.history.lock().unwrap().push(history_record);

            self.stage_statuses.lock().unwrap().remove(&doc_id);
        }

        Ok(doc.clone())
    }

    pub fn issue_document(
        &self,
        doc_id: Uuid,
        operator: String,
        approve: bool,
        reason: Option<String>,
    ) -> Result<Document, WorkflowError> {
        let mut docs = self.documents.lock().unwrap();
        let doc = docs
            .get_mut(&doc_id)
            .ok_or(WorkflowError::DocumentNotFound)?;

        if doc.current_stage != WorkflowStage::Issuance {
            return Err(WorkflowError::InvalidTransition {
                action: "issue".to_string(),
                current_stage: doc.current_stage.to_str().to_string(),
            });
        }

        let now = Utc::now();
        let from_stage = doc.current_stage;

        if approve {
            doc.current_stage = WorkflowStage::Archived;
            doc.updated_at = now;

            let history_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::Issue,
                operator,
                timestamp: now,
            };
            self.history.lock().unwrap().push(history_record);

            let archive_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::Archive,
                operator: "系统".to_string(),
                timestamp: now,
            };
            self.history.lock().unwrap().push(archive_record);

            self.stage_statuses.lock().unwrap().remove(&doc_id);
        } else {
            if reason.is_none() || reason.as_ref().unwrap().is_empty() {
                return Err(WorkflowError::ReturnReasonRequired);
            }
            doc.current_stage = WorkflowStage::Draft;
            doc.updated_at = now;

            let history_record = WorkflowHistory {
                id: Uuid::new_v4(),
                document_id: doc_id,
                action: HistoryAction::Return {
                    from: from_stage,
                    reason: reason.unwrap(),
                },
                operator,
                timestamp: now,
            };
            self.history.lock().unwrap().push(history_record);

            self.stage_statuses.lock().unwrap().remove(&doc_id);
        }

        Ok(doc.clone())
    }

    pub fn update_and_resubmit(
        &self,
        doc_id: Uuid,
        operator: String,
        title: Option<String>,
        content: Option<String>,
    ) -> Result<Document, WorkflowError> {
        let mut docs = self.documents.lock().unwrap();
        let doc = docs
            .get_mut(&doc_id)
            .ok_or(WorkflowError::DocumentNotFound)?;

        if doc.current_stage != WorkflowStage::Draft {
            return Err(WorkflowError::InvalidTransition {
                action: "update and resubmit".to_string(),
                current_stage: doc.current_stage.to_str().to_string(),
            });
        }

        let now = Utc::now();
        if let Some(new_title) = title {
            doc.title = new_title;
        }
        if let Some(new_content) = content {
            doc.content = new_content;
        }
        doc.updated_at = now;

        let from_stage = WorkflowStage::Draft;
        doc.current_stage = WorkflowStage::DepartmentReview;

        let history_record = WorkflowHistory {
            id: Uuid::new_v4(),
            document_id: doc_id,
            action: HistoryAction::Resubmit,
            operator: operator.clone(),
            timestamp: now,
        };
        self.history.lock().unwrap().push(history_record);

        let submit_record = WorkflowHistory {
            id: Uuid::new_v4(),
            document_id: doc_id,
            action: HistoryAction::Submit {
                from: from_stage,
                to: WorkflowStage::DepartmentReview,
            },
            operator,
            timestamp: now,
        };
        self.history.lock().unwrap().push(submit_record);

        self.stage_statuses.lock().unwrap().insert(
            doc_id,
            StageStatus {
                document_id: doc_id,
                stage: WorkflowStage::DepartmentReview,
                entered_at: now,
                assignee: "部门审核员".to_string(),
            },
        );

        Ok(doc.clone())
    }

    pub fn get_document(&self, doc_id: Uuid) -> Option<Document> {
        self.documents.lock().unwrap().get(&doc_id).cloned()
    }

    pub fn list_documents(&self) -> Vec<Document> {
        self.documents.lock().unwrap().values().cloned().collect()
    }

    pub fn get_history(&self, doc_id: Uuid) -> Vec<WorkflowHistory> {
        let history = self.history.lock().unwrap();
        history
            .iter()
            .filter(|h| h.document_id == doc_id)
            .cloned()
            .collect()
    }

    pub fn check_overdue(&self) {
        let mut docs = self.documents.lock().unwrap();
        let stage_statuses = self.stage_statuses.lock().unwrap();
        let now = Utc::now();

        for (doc_id, doc) in docs.iter_mut() {
            if let Some(status) = stage_statuses.get(doc_id) {
                let duration = now.signed_duration_since(status.entered_at);
                let timeout_hours = doc.urgency.timeout_hours();
                let timeout_duration = Duration::hours(timeout_hours as i64);
                doc.is_overdue = duration > timeout_duration;
            } else {
                doc.is_overdue = false;
            }
        }
    }

    pub fn get_overdue_documents(&self) -> Vec<Document> {
        self.check_overdue();
        self.documents
            .lock()
            .unwrap()
            .values()
            .filter(|d| d.is_overdue)
            .cloned()
            .collect()
    }
}

impl Default for WorkflowService {
    fn default() -> Self {
        Self::new()
    }
}
