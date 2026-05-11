use chrono::{Duration, Utc};
use uuid::Uuid;

use crate::error::ArchiveError;
use crate::models::*;
use crate::repository::Repository;

pub struct ArchiveService<R: Repository> {
    repo: R,
}

impl<R: Repository> ArchiveService<R> {
    pub fn new(repo: R) -> Self {
        Self { repo }
    }

    pub fn create_user(&self, user: User) {
        self.repo.save_user(user);
    }

    pub fn get_user(&self, id: &str) -> Result<User, ArchiveError> {
        self.repo
            .get_user(id)
            .ok_or_else(|| ArchiveError::UserNotFound(id.to_string()))
    }

    pub fn get_all_users(&self) -> Vec<User> {
        self.repo.get_all_users()
    }

    pub fn create_archive(&self, archive: Archive) {
        self.repo.save_archive(archive);
    }

    pub fn get_archive(&self, id: &str) -> Result<Archive, ArchiveError> {
        self.repo
            .get_archive(id)
            .ok_or_else(|| ArchiveError::ArchiveNotFound(id.to_string()))
    }

    pub fn get_all_archives(&self) -> Vec<Archive> {
        self.repo.get_all_archives()
    }

    pub fn request_borrow(
        &self,
        user_id: &str,
        archive_id: &str,
    ) -> Result<Borrow, ArchiveError> {
        let user = self.get_user(user_id)?;
        let archive = self.get_archive(archive_id)?;

        if user.is_suspended {
            return Err(ArchiveError::UserSuspended);
        }

        if archive.is_borrowed {
            return Err(ArchiveError::ArchiveAlreadyBorrowed);
        }

        let active_borrows = self.repo.get_active_borrows_by_user(user_id);
        if active_borrows.len() >= user.max_borrow_count as usize {
            return Err(ArchiveError::UserHasOverdueItems);
        }

        let has_overdue = active_borrows.iter().any(|b| {
            matches!(b.status, BorrowStatus::Overdue)
                || (matches!(b.status, BorrowStatus::Active) && b.due_date < Utc::now())
        });
        if has_overdue {
            return Err(ArchiveError::UserHasOverdueItems);
        }

        let mut borrow = Borrow::new(archive_id, user_id, archive.classification);

        if archive.classification.requires_second_approval() {
            let approval_process = self.create_approval_process(&borrow, &user);
            borrow.approval_process_id = Some(approval_process.id.clone());
            self.repo.save_approval_process(approval_process);
        } else {
            borrow.status = BorrowStatus::Active;
        }

        self.repo.save_borrow(borrow.clone());

        if borrow.status == BorrowStatus::Active {
            let mut archive = archive;
            archive.is_borrowed = true;
            archive.current_borrow_id = Some(borrow.id.clone());
            self.repo.update_archive(archive).map_err(|_| {
                ArchiveError::Internal("更新档案状态失败".to_string())
            })?;
        }

        Ok(borrow)
    }

    fn create_approval_process(
        &self,
        borrow: &Borrow,
        requester: &User,
    ) -> ApprovalProcess {
        let steps = vec![
            ApprovalStep {
                approver_id: "approver-1".to_string(),
                step: 1,
                status: ApprovalStatus::Pending,
                comment: None,
                approved_at: None,
            },
            ApprovalStep {
                approver_id: "approver-2".to_string(),
                step: 2,
                status: ApprovalStatus::Pending,
                comment: None,
                approved_at: None,
            },
        ];

        ApprovalProcess {
            id: Uuid::new_v4().to_string(),
            borrow_id: borrow.id.clone(),
            archive_id: borrow.archive_id.clone(),
            requester_id: requester.id.clone(),
            steps,
            created_at: Utc::now(),
        }
    }

    pub fn approve_borrow(
        &self,
        approval_process_id: &str,
        approver_id: &str,
        step: u8,
        comment: Option<String>,
    ) -> Result<Borrow, ArchiveError> {
        let mut process = self
            .repo
            .get_approval_process(approval_process_id)
            .ok_or_else(|| ArchiveError::ApprovalNotFound)?;

        let step_index = step as usize - 1;
        {
            let approval_step = process
                .steps
                .get(step_index)
                .ok_or_else(|| ArchiveError::ApprovalNotFound)?;

            if approval_step.status != ApprovalStatus::Pending {
                return Err(ArchiveError::ApprovalAlreadyCompleted);
            }

            if approval_step.approver_id != approver_id {
                return Err(ArchiveError::ApproverMismatch);
            }

            if step > 1 {
                let prev_step = &process.steps[step_index - 1];
                if prev_step.status != ApprovalStatus::Approved {
                    return Err(ArchiveError::ApprovalOrderError);
                }
            }
        }

        let approval_step = process
            .steps
            .get_mut(step_index)
            .ok_or_else(|| ArchiveError::ApprovalNotFound)?;

        approval_step.status = ApprovalStatus::Approved;
        approval_step.comment = comment;
        approval_step.approved_at = Some(Utc::now());

        self.repo
            .update_approval_process(process.clone())
            .map_err(|_| ArchiveError::Internal("更新审批流程失败".to_string()))?;

        let all_approved = process
            .steps
            .iter()
            .all(|s| s.status == ApprovalStatus::Approved);

        let mut borrow = self
            .repo
            .get_borrow(&process.borrow_id)
            .ok_or_else(|| ArchiveError::BorrowNotFound(process.borrow_id.clone()))?;

        if all_approved {
            borrow.status = BorrowStatus::Active;
            self.repo
                .update_borrow(borrow.clone())
                .map_err(|_| ArchiveError::Internal("更新借阅状态失败".to_string()))?;

            let mut archive = self.get_archive(&borrow.archive_id)?;
            archive.is_borrowed = true;
            archive.current_borrow_id = Some(borrow.id.clone());
            self.repo.update_archive(archive).map_err(|_| {
                ArchiveError::Internal("更新档案状态失败".to_string())
            })?;
        }

        Ok(borrow)
    }

    pub fn reject_borrow(
        &self,
        approval_process_id: &str,
        approver_id: &str,
        step: u8,
        comment: Option<String>,
    ) -> Result<Borrow, ArchiveError> {
        let mut process = self
            .repo
            .get_approval_process(approval_process_id)
            .ok_or_else(|| ArchiveError::ApprovalNotFound)?;

        let step_index = step as usize - 1;
        let approval_step = process
            .steps
            .get_mut(step_index)
            .ok_or_else(|| ArchiveError::ApprovalNotFound)?;

        if approval_step.status != ApprovalStatus::Pending {
            return Err(ArchiveError::ApprovalAlreadyCompleted);
        }

        if approval_step.approver_id != approver_id {
            return Err(ArchiveError::ApproverMismatch);
        }

        approval_step.status = ApprovalStatus::Rejected;
        approval_step.comment = comment;
        approval_step.approved_at = Some(Utc::now());

        self.repo
            .update_approval_process(process.clone())
            .map_err(|_| ArchiveError::Internal("更新审批流程失败".to_string()))?;

        let mut borrow = self
            .repo
            .get_borrow(&process.borrow_id)
            .ok_or_else(|| ArchiveError::BorrowNotFound(process.borrow_id.clone()))?;

        borrow.status = BorrowStatus::Cancelled;
        self.repo
            .update_borrow(borrow.clone())
            .map_err(|_| ArchiveError::Internal("更新借阅状态失败".to_string()))?;

        Ok(borrow)
    }

    pub fn get_approval_process(&self, id: &str) -> Result<ApprovalProcess, ArchiveError> {
        self.repo
            .get_approval_process(id)
            .ok_or_else(|| ArchiveError::ApprovalNotFound)
    }

    pub fn get_borrow(&self, id: &str) -> Result<Borrow, ArchiveError> {
        self.repo
            .get_borrow(id)
            .ok_or_else(|| ArchiveError::BorrowNotFound(id.to_string()))
    }

    pub fn get_all_borrows(&self) -> Vec<Borrow> {
        self.repo.get_all_borrows()
    }

    pub fn get_user_borrows(&self, user_id: &str) -> Result<Vec<Borrow>, ArchiveError> {
        let _ = self.get_user(user_id)?;
        Ok(self.repo.get_active_borrows_by_user(user_id))
    }

    pub fn renew_borrow(&self, borrow_id: &str) -> Result<Borrow, ArchiveError> {
        let mut borrow = self.get_borrow(borrow_id)?;
        let archive = self.get_archive(&borrow.archive_id)?;

        if !archive.classification.can_renew() {
            return Err(ArchiveError::ConfidentialRenewalNotAllowed);
        }

        if borrow.renewal_count >= 1 {
            return Err(ArchiveError::RenewalLimitExceeded);
        }

        borrow.due_date += Duration::days(archive.classification.max_borrow_days());
        borrow.renewal_count += 1;
        borrow.status = BorrowStatus::Renewed;

        self.repo
            .update_borrow(borrow.clone())
            .map_err(|_| ArchiveError::Internal("更新借阅状态失败".to_string()))?;

        Ok(borrow)
    }

    pub fn return_archive(
        &self,
        borrow_id: &str,
        check_result: ReturnCheckResult,
    ) -> Result<Borrow, ArchiveError> {
        let mut borrow = self.get_borrow(borrow_id)?;

        if !matches!(
            borrow.status,
            BorrowStatus::Active | BorrowStatus::Renewed | BorrowStatus::Overdue
        ) {
            return Err(ArchiveError::ArchiveNotAvailable);
        }

        if check_result.is_damaged {
            let damage_report = DamageReport {
                id: Uuid::new_v4().to_string(),
                borrow_id: borrow.id.clone(),
                description: check_result.damage_description.unwrap_or_default(),
                compensation_amount: check_result.compensation_amount,
                reported_by: check_result.checked_by,
                reported_at: check_result.checked_at,
            };
            borrow.damage_report = Some(damage_report);
        }

        borrow.return_date = Some(check_result.checked_at);
        borrow.status = BorrowStatus::Returned;

        self.repo
            .update_borrow(borrow.clone())
            .map_err(|_| ArchiveError::Internal("更新借阅状态失败".to_string()))?;

        let mut archive = self.get_archive(&borrow.archive_id)?;
        archive.is_borrowed = false;
        archive.current_borrow_id = None;
        self.repo
            .update_archive(archive.clone())
            .map_err(|_| ArchiveError::Internal("更新档案状态失败".to_string()))?;

        self.notify_next_reservation(&archive.id);

        Ok(borrow)
    }

    fn notify_next_reservation(&self, archive_id: &str) {
        if let Some(mut reservation) = self.repo.get_next_pending_reservation(archive_id) {
            reservation.status = ReservationStatus::Notified;
            reservation.notified_at = Some(Utc::now());
            let _ = self.repo.update_reservation(reservation);
        }
    }

    pub fn reserve_archive(
        &self,
        user_id: &str,
        archive_id: &str,
    ) -> Result<Reservation, ArchiveError> {
        let user = self.get_user(user_id)?;
        let archive = self.get_archive(archive_id)?;

        if user.is_suspended {
            return Err(ArchiveError::UserSuspended);
        }

        let reservations = self.repo.get_reservations_by_archive(archive_id);
        let existing_reservation = reservations.iter().find(|r| {
            r.user_id == user_id
                && matches!(
                    r.status,
                    ReservationStatus::Pending | ReservationStatus::Notified
                )
        });

        if existing_reservation.is_some() {
            return Err(ArchiveError::UserAlreadyReserved);
        }

        if archive.is_borrowed {
            if let Some(current_borrow) = archive.current_borrow_id.as_ref() {
                if let Ok(borrow) = self.get_borrow(current_borrow) {
                    if borrow.user_id == user_id {
                        return Err(ArchiveError::UserAlreadyBorrowed);
                    }
                }
            }
        }

        let active_pending_reservations: Vec<_> = reservations
            .into_iter()
            .filter(|r| {
                matches!(
                    r.status,
                    ReservationStatus::Pending | ReservationStatus::Notified
                )
            })
            .collect();

        let queue_position = active_pending_reservations.len() as u32 + 1;
        let reservation = Reservation::new(archive_id, user_id, queue_position);

        self.repo.save_reservation(reservation.clone());

        if !archive.is_borrowed && queue_position == 1 {
            let mut reservation = reservation.clone();
            reservation.status = ReservationStatus::Notified;
            reservation.notified_at = Some(Utc::now());
            self.repo
                .update_reservation(reservation.clone())
                .map_err(|_| {
                    ArchiveError::Internal("更新预约状态失败".to_string())
                })?;
        }

        Ok(reservation)
    }

    pub fn get_reservation(&self, id: &str) -> Result<Reservation, ArchiveError> {
        self.repo
            .get_reservation(id)
            .ok_or_else(|| ArchiveError::ReservationNotFound(id.to_string()))
    }

    pub fn get_all_reservations(&self) -> Vec<Reservation> {
        self.repo.get_all_reservations()
    }

    pub fn get_archive_reservations(&self, archive_id: &str) -> Vec<Reservation> {
        self.repo.get_reservations_by_archive(archive_id)
    }

    pub fn claim_reservation(&self, reservation_id: &str) -> Result<Borrow, ArchiveError> {
        let mut reservation = self.get_reservation(reservation_id)?;

        if reservation.status != ReservationStatus::Notified {
            return Err(ArchiveError::ReservationExpired);
        }

        if reservation.is_claim_deadline_passed() {
            reservation.status = ReservationStatus::Expired;
            self.repo
                .update_reservation(reservation)
                .map_err(|_| ArchiveError::Internal("更新预约状态失败".to_string()))?;
            return Err(ArchiveError::ReservationNotClaimed);
        }

        let archive = self.get_archive(&reservation.archive_id)?;
        if archive.is_borrowed {
            return Err(ArchiveError::ArchiveAlreadyBorrowed);
        }

        reservation.status = ReservationStatus::Claimed;
        reservation.claimed_at = Some(Utc::now());
        self.repo
            .update_reservation(reservation.clone())
            .map_err(|_| ArchiveError::Internal("更新预约状态失败".to_string()))?;

        let borrow = self.request_borrow(&reservation.user_id, &reservation.archive_id)?;

        let reservations = self.repo.get_reservations_by_archive(&reservation.archive_id);
        for (idx, mut r) in reservations.into_iter().enumerate() {
            if matches!(
                r.status,
                ReservationStatus::Pending | ReservationStatus::Notified
            ) && r.id != reservation.id
            {
                r.queue_position = idx as u32 + 1;
                let _ = self.repo.update_reservation(r);
            }
        }

        Ok(borrow)
    }

    pub fn cancel_reservation(&self, reservation_id: &str) -> Result<Reservation, ArchiveError> {
        let mut reservation = self.get_reservation(reservation_id)?;

        if !matches!(
            reservation.status,
            ReservationStatus::Pending | ReservationStatus::Notified
        ) {
            return Err(ArchiveError::ReservationExpired);
        }

        reservation.status = ReservationStatus::Cancelled;
        self.repo
            .update_reservation(reservation.clone())
            .map_err(|_| ArchiveError::Internal("更新预约状态失败".to_string()))?;

        Ok(reservation)
    }

    pub fn check_expired_reservations(&self) -> Vec<Reservation> {
        let mut expired = Vec::new();
        let all_reservations = self.repo.get_all_reservations();

        for mut reservation in all_reservations {
            if reservation.status == ReservationStatus::Notified
                && reservation.is_claim_deadline_passed()
            {
                reservation.status = ReservationStatus::Expired;
                if self.repo.update_reservation(reservation.clone()).is_ok() {
                    expired.push(reservation);
                }
            }
        }

        expired
    }

    pub fn get_expiring_soon(&self, days: i64) -> Vec<Borrow> {
        let threshold = Utc::now() + Duration::days(days);
        self.repo
            .get_all_borrows()
            .into_iter()
            .filter(|b| {
                matches!(
                    b.status,
                    BorrowStatus::Active | BorrowStatus::Renewed
                ) && b.due_date <= threshold
                    && b.due_date > Utc::now()
            })
            .collect()
    }

    pub fn process_overdue(&self) -> Vec<OverdueInfo> {
        let mut overdue_infos = Vec::new();
        let now = Utc::now();
        let overdue_borrows = self.repo.get_overdue_borrows();

        for mut borrow in overdue_borrows {
            let days_overdue = (now - borrow.due_date).num_days();

            let stage = if days_overdue <= 3 {
                OverdueStage::Warning
            } else if days_overdue <= 7 {
                OverdueStage::DepartmentNotify
            } else {
                OverdueStage::Suspension
            };

            if borrow.status != BorrowStatus::Overdue {
                borrow.status = BorrowStatus::Overdue;
                let _ = self.repo.update_borrow(borrow.clone());
            }

            if stage == OverdueStage::Suspension {
                if let Ok(mut user) = self.get_user(&borrow.user_id) {
                    if !user.is_suspended {
                        user.is_suspended = true;
                        let _ = self.repo.update_user(user);
                    }
                }
            }

            overdue_infos.push(OverdueInfo {
                borrow_id: borrow.id.clone(),
                days_overdue,
                stage,
                notified_at: Some(now),
            });
        }

        overdue_infos
    }

    pub fn get_overdue_stage(&self, days_overdue: i64) -> OverdueStage {
        if days_overdue <= 3 {
            OverdueStage::Warning
        } else if days_overdue <= 7 {
            OverdueStage::DepartmentNotify
        } else {
            OverdueStage::Suspension
        }
    }

    pub fn get_due_date_reminders(&self) -> Vec<Borrow> {
        self.get_expiring_soon(3)
    }
}
