use std::sync::Arc;
use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::error::{Result, SealBorrowError};
use crate::models::*;
use crate::store::InMemoryStore;

#[derive(Clone)]
pub struct SealBorrowService {
    store: Arc<InMemoryStore>,
    reminder_interval_days: i64,
}

impl SealBorrowService {
    pub fn new(store: Arc<InMemoryStore>) -> Self {
        Self {
            store,
            reminder_interval_days: 3,
        }
    }

    pub fn with_reminder_interval(mut self, days: i64) -> Self {
        self.reminder_interval_days = days;
        self
    }

    pub fn reminder_interval_days(&self) -> i64 {
        self.reminder_interval_days
    }

    pub async fn create_seal(&self, req: CreateSealRequest) -> Result<Seal> {
        if self.store.get_employee(req.custodian_id).await.is_none() {
            return Err(SealBorrowError::EmployeeNotFound);
        }
        Ok(self.store.create_seal(req).await)
    }

    pub async fn get_seal(&self, id: Uuid) -> Result<Seal> {
        self.store.get_seal(id).await.ok_or(SealBorrowError::SealNotFound)
    }

    pub async fn list_seals(&self) -> Vec<Seal> {
        self.store.list_seals().await
    }

    pub async fn update_seal_status(&self, req: UpdateSealStatusRequest) -> Result<Seal> {
        self.store.update_seal_status(req.seal_id, req.status)
            .await
            .ok_or(SealBorrowError::SealNotFound)
    }

    pub async fn create_employee(&self, req: CreateEmployeeRequest) -> Employee {
        self.store.create_employee(req).await
    }

    pub async fn get_employee(&self, id: Uuid) -> Result<Employee> {
        self.store.get_employee(id).await.ok_or(SealBorrowError::EmployeeNotFound)
    }

    pub async fn list_employees(&self) -> Vec<Employee> {
        self.store.list_employees().await
    }

    pub async fn create_borrow_request(&self, req: CreateBorrowRequest) -> Result<BorrowRequest> {
        let seal = self.store.get_seal(req.seal_id).await.ok_or(SealBorrowError::SealNotFound)?;
        
        if self.store.get_employee(req.borrower_id).await.is_none() {
            return Err(SealBorrowError::EmployeeNotFound);
        }

        if seal.status == SealStatus::Maintenance {
            return Err(SealBorrowError::SealInMaintenance);
        }

        let borrower_requests = self.store.list_borrow_requests_by_borrower(req.borrower_id).await;
        let has_active_borrow = borrower_requests.iter().any(|r| {
            r.seal_id == req.seal_id 
                && r.status == BorrowRequestStatus::Approved 
                && r.actual_return_date.is_none()
        });
        if has_active_borrow {
            return Err(SealBorrowError::AlreadyBorrowingThisSeal);
        }

        let pending_requests = self.store.list_borrow_requests_by_seal(req.seal_id).await;
        let has_pending = pending_requests.iter().any(|r| r.status == BorrowRequestStatus::Pending);
        let has_approved_not_returned = pending_requests.iter().any(|r| {
            r.status == BorrowRequestStatus::Approved && r.actual_return_date.is_none()
        });
        
        if seal.status == SealStatus::Borrowed || has_pending || has_approved_not_returned {
            return Err(SealBorrowError::SealAlreadyBorrowed);
        }

        Ok(self.store.create_borrow_request(req, false, None).await)
    }

    pub async fn get_borrow_request(&self, id: Uuid) -> Result<BorrowRequest> {
        self.store.get_borrow_request(id).await.ok_or(SealBorrowError::BorrowRequestNotFound)
    }

    pub async fn list_borrow_requests(&self) -> Vec<BorrowRequest> {
        self.store.list_borrow_requests().await
    }

    pub async fn list_borrow_requests_by_seal(&self, seal_id: Uuid) -> Vec<BorrowRequest> {
        self.store.list_borrow_requests_by_seal(seal_id).await
    }

    pub async fn list_borrow_requests_by_borrower(&self, borrower_id: Uuid) -> Vec<BorrowRequest> {
        self.store.list_borrow_requests_by_borrower(borrower_id).await
    }

    pub async fn approve_request(&self, req: ApproveRequest) -> Result<BorrowRequest> {
        let mut request = self.store.get_borrow_request(req.request_id).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)?;

        if request.status != BorrowRequestStatus::Pending {
            return Err(SealBorrowError::InvalidStatusForOperation);
        }

        if req.approver_id == request.borrower_id {
            return Err(SealBorrowError::ApproverCannotBeBorrower);
        }

        let seal = self.store.get_seal(request.seal_id).await
            .ok_or(SealBorrowError::SealNotFound)?;

        if req.approver_id != seal.custodian_id {
            return Err(SealBorrowError::OnlyCustodianCanApprove);
        }

        request.status = BorrowRequestStatus::Approved;
        request.approver_id = Some(req.approver_id);
        request.approved_at = Some(Utc::now());

        self.store.update_seal_status(request.seal_id, SealStatus::Borrowed).await;

        self.store.update_borrow_request(request.clone()).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)
    }

    pub async fn reject_request(&self, req: RejectRequest) -> Result<BorrowRequest> {
        let mut request = self.store.get_borrow_request(req.request_id).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)?;

        if request.status != BorrowRequestStatus::Pending {
            return Err(SealBorrowError::InvalidStatusForOperation);
        }

        if req.approver_id == request.borrower_id {
            return Err(SealBorrowError::ApproverCannotBeBorrower);
        }

        let seal = self.store.get_seal(request.seal_id).await
            .ok_or(SealBorrowError::SealNotFound)?;

        if req.approver_id != seal.custodian_id {
            return Err(SealBorrowError::OnlyCustodianCanApprove);
        }

        request.status = BorrowRequestStatus::Rejected;
        request.approver_id = Some(req.approver_id);
        request.reject_reason = Some(req.reason);

        self.store.update_borrow_request(request.clone()).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)
    }

    pub async fn renew_request(&self, req: RenewRequest) -> Result<BorrowRequest> {
        let original_request = self.store.get_borrow_request(req.request_id).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)?;

        if req.borrower_id != original_request.borrower_id {
            return Err(SealBorrowError::OnlyBorrowerCanRenew);
        }

        if original_request.status != BorrowRequestStatus::Approved || original_request.actual_return_date.is_some() {
            return Err(SealBorrowError::InvalidStatusForOperation);
        }

        if req.new_expected_return_date <= original_request.expected_return_date {
            return Err(SealBorrowError::NewReturnDateMustBeAfterOriginal);
        }

        let renewal_request = CreateBorrowRequest {
            seal_id: original_request.seal_id,
            borrower_id: req.borrower_id,
            reason: req.reason,
            expected_return_date: req.new_expected_return_date,
        };

        Ok(self.store.create_borrow_request(renewal_request, true, Some(original_request.id)).await)
    }

    pub async fn process_renewal_approval(&self, renewal_request_id: Uuid, approved: bool, approver_id: Uuid, reject_reason: Option<String>) -> Result<BorrowRequest> {
        let renewal_request = self.store.get_borrow_request(renewal_request_id).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)?;

        if !renewal_request.is_renewal || renewal_request.original_request_id.is_none() {
            return Err(SealBorrowError::InvalidStatusForOperation);
        }

        let original_request_id = renewal_request.original_request_id.unwrap();
        let original_request = self.store.get_borrow_request(original_request_id).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)?;

        if renewal_request.status != BorrowRequestStatus::Pending {
            return Err(SealBorrowError::InvalidStatusForOperation);
        }

        if approver_id == renewal_request.borrower_id {
            return Err(SealBorrowError::ApproverCannotBeBorrower);
        }

        let seal = self.store.get_seal(renewal_request.seal_id).await
            .ok_or(SealBorrowError::SealNotFound)?;

        if approver_id != seal.custodian_id {
            return Err(SealBorrowError::OnlyCustodianCanApprove);
        }

        if approved {
            let mut updated_original = original_request.clone();
            updated_original.expected_return_date = renewal_request.expected_return_date;
            self.store.update_borrow_request(updated_original).await;

            let mut updated_renewal = renewal_request.clone();
            updated_renewal.status = BorrowRequestStatus::Approved;
            updated_renewal.approver_id = Some(approver_id);
            updated_renewal.approved_at = Some(Utc::now());
            self.store.update_borrow_request(updated_renewal.clone()).await;

            Ok(updated_renewal)
        } else {
            let mut updated_renewal = renewal_request.clone();
            updated_renewal.status = BorrowRequestStatus::Rejected;
            updated_renewal.approver_id = Some(approver_id);
            updated_renewal.reject_reason = reject_reason;
            self.store.update_borrow_request(updated_renewal.clone()).await;

            Ok(updated_renewal)
        }
    }

    pub async fn return_seal(&self, req: ReturnRequest) -> Result<BorrowRequest> {
        let mut request = self.store.get_borrow_request(req.request_id).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)?;

        if request.status != BorrowRequestStatus::Approved || request.actual_return_date.is_some() {
            return Err(SealBorrowError::InvalidStatusForOperation);
        }

        request.actual_return_date = Some(Utc::now());
        self.store.update_seal_status(request.seal_id, SealStatus::InStock).await;

        self.store.update_borrow_request(request.clone()).await
            .ok_or(SealBorrowError::BorrowRequestNotFound)
    }

    pub async fn check_and_create_reminders(&self, now: DateTime<Utc>) -> Vec<ReminderRecord> {
        let overdue_requests = self.store.get_overdue_requests(now).await;
        let mut new_reminders = Vec::new();

        for request in overdue_requests {
            let should_send = match self.store.get_latest_reminder(request.id).await {
                Some(latest) => {
                    let days_since = (now - latest.reminder_date).num_days();
                    days_since >= self.reminder_interval_days
                }
                None => true,
            };

            if should_send {
                let reminder = self.store.create_reminder(request.id, now).await;
                new_reminders.push(reminder);
            }
        }

        new_reminders
    }

    pub async fn list_reminders(&self) -> Vec<ReminderRecord> {
        self.store.list_reminders().await
    }

    pub async fn list_reminders_by_request(&self, borrow_request_id: Uuid) -> Vec<ReminderRecord> {
        self.store.list_reminders_by_request(borrow_request_id).await
    }
}
