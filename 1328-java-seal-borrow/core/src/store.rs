use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;
use chrono::{DateTime, Utc};

use crate::models::*;

#[derive(Clone, Default)]
pub struct InMemoryStore {
    seals: Arc<RwLock<HashMap<Uuid, Seal>>>,
    employees: Arc<RwLock<HashMap<Uuid, Employee>>>,
    borrow_requests: Arc<RwLock<HashMap<Uuid, BorrowRequest>>>,
    reminders: Arc<RwLock<HashMap<Uuid, ReminderRecord>>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub async fn create_seal(&self, req: CreateSealRequest) -> Seal {
        let seal = Seal {
            id: Uuid::new_v4(),
            name: req.name,
            seal_type: req.seal_type,
            custodian_id: req.custodian_id,
            status: SealStatus::InStock,
            created_at: Utc::now(),
        };
        self.seals.write().await.insert(seal.id, seal.clone());
        seal
    }

    pub async fn get_seal(&self, id: Uuid) -> Option<Seal> {
        self.seals.read().await.get(&id).cloned()
    }

    pub async fn list_seals(&self) -> Vec<Seal> {
        self.seals.read().await.values().cloned().collect()
    }

    pub async fn update_seal_status(&self, seal_id: Uuid, status: SealStatus) -> Option<Seal> {
        let mut seals = self.seals.write().await;
        if let Some(seal) = seals.get_mut(&seal_id) {
            seal.status = status;
            Some(seal.clone())
        } else {
            None
        }
    }

    pub async fn create_employee(&self, req: CreateEmployeeRequest) -> Employee {
        let employee = Employee {
            id: Uuid::new_v4(),
            name: req.name,
            email: req.email,
            created_at: Utc::now(),
        };
        self.employees.write().await.insert(employee.id, employee.clone());
        employee
    }

    pub async fn get_employee(&self, id: Uuid) -> Option<Employee> {
        self.employees.read().await.get(&id).cloned()
    }

    pub async fn list_employees(&self) -> Vec<Employee> {
        self.employees.read().await.values().cloned().collect()
    }

    pub async fn create_borrow_request(&self, req: CreateBorrowRequest, is_renewal: bool, original_request_id: Option<Uuid>) -> BorrowRequest {
        let request = BorrowRequest {
            id: Uuid::new_v4(),
            seal_id: req.seal_id,
            borrower_id: req.borrower_id,
            reason: req.reason,
            expected_return_date: req.expected_return_date,
            actual_return_date: None,
            status: BorrowRequestStatus::Pending,
            approver_id: None,
            reject_reason: None,
            is_renewal,
            original_request_id,
            created_at: Utc::now(),
            approved_at: None,
        };
        self.borrow_requests.write().await.insert(request.id, request.clone());
        request
    }

    pub async fn get_borrow_request(&self, id: Uuid) -> Option<BorrowRequest> {
        self.borrow_requests.read().await.get(&id).cloned()
    }

    pub async fn list_borrow_requests(&self) -> Vec<BorrowRequest> {
        self.borrow_requests.read().await.values().cloned().collect()
    }

    pub async fn list_borrow_requests_by_seal(&self, seal_id: Uuid) -> Vec<BorrowRequest> {
        self.borrow_requests.read().await.values()
            .filter(|r| r.seal_id == seal_id)
            .cloned()
            .collect()
    }

    pub async fn list_borrow_requests_by_borrower(&self, borrower_id: Uuid) -> Vec<BorrowRequest> {
        self.borrow_requests.read().await.values()
            .filter(|r| r.borrower_id == borrower_id)
            .cloned()
            .collect()
    }

    pub async fn update_borrow_request(&self, request: BorrowRequest) -> Option<BorrowRequest> {
        let mut requests = self.borrow_requests.write().await;
        if requests.contains_key(&request.id) {
            requests.insert(request.id, request.clone());
            Some(request)
        } else {
            None
        }
    }

    pub async fn get_overdue_requests(&self, now: DateTime<Utc>) -> Vec<BorrowRequest> {
        self.borrow_requests.read().await.values()
            .filter(|r| {
                r.status == BorrowRequestStatus::Approved 
                    && r.actual_return_date.is_none() 
                    && r.expected_return_date < now
            })
            .cloned()
            .collect()
    }

    pub async fn create_reminder(&self, borrow_request_id: Uuid, reminder_date: DateTime<Utc>) -> ReminderRecord {
        let reminder = ReminderRecord {
            id: Uuid::new_v4(),
            borrow_request_id,
            reminder_date,
            created_at: Utc::now(),
        };
        self.reminders.write().await.insert(reminder.id, reminder.clone());
        reminder
    }

    pub async fn list_reminders(&self) -> Vec<ReminderRecord> {
        self.reminders.read().await.values().cloned().collect()
    }

    pub async fn list_reminders_by_request(&self, borrow_request_id: Uuid) -> Vec<ReminderRecord> {
        self.reminders.read().await.values()
            .filter(|r| r.borrow_request_id == borrow_request_id)
            .cloned()
            .collect()
    }

    pub async fn get_latest_reminder(&self, borrow_request_id: Uuid) -> Option<ReminderRecord> {
        let mut reminders: Vec<ReminderRecord> = self.reminders.read().await.values()
            .filter(|r| r.borrow_request_id == borrow_request_id)
            .cloned()
            .collect();
        reminders.sort_by(|a, b| b.reminder_date.cmp(&a.reminder_date));
        reminders.first().cloned()
    }
}
