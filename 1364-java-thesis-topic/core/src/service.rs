use crate::{api::*, error::*, models::*, *};
use chrono::{DateTime, Utc};
use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};
use tokio::sync::RwLock;
use uuid::Uuid;

pub struct ThesisSystem {
    phase: RwLock<SystemPhase>,
    students: RwLock<HashMap<StudentId, Student>>,
    advisors: RwLock<HashMap<AdvisorId, Advisor>>,
    selection_records: RwLock<Vec<SelectionRecord>>,
    appeals: RwLock<HashMap<Uuid, AppealRecord>>,
    change_logs: RwLock<Vec<ChangeLog>>,
    student_locks: RwLock<HashMap<StudentId, AtomicUsize>>,
}

impl Default for ThesisSystem {
    fn default() -> Self {
        Self::new()
    }
}

impl ThesisSystem {
    pub fn new() -> Self {
        Self {
            phase: RwLock::new(SystemPhase::StudentSubmission),
            students: RwLock::new(HashMap::new()),
            advisors: RwLock::new(HashMap::new()),
            selection_records: RwLock::new(Vec::new()),
            appeals: RwLock::new(HashMap::new()),
            change_logs: RwLock::new(Vec::new()),
            student_locks: RwLock::new(HashMap::new()),
        }
    }

    async fn ensure_phase(&self, allowed: &[SystemPhase]) -> Result<()> {
        let phase = *self.phase.read().await;
        if allowed.contains(&phase) {
            Ok(())
        } else {
            Err(ThesisError::PhaseError(format!(
                "Operation not allowed in phase {:?}",
                phase
            )))
        }
    }

    async fn lock_student(&self, student_id: StudentId) -> Result<bool> {
        let locks = self.student_locks.read().await;
        if let Some(lock) = locks.get(&student_id) {
            Ok(lock.fetch_add(1, Ordering::SeqCst) == 0)
        } else {
            drop(locks);
            let mut locks = self.student_locks.write().await;
            locks.entry(student_id).or_insert_with(|| AtomicUsize::new(1));
            Ok(true)
        }
    }

    async fn unlock_student(&self, student_id: StudentId) {
        let locks = self.student_locks.read().await;
        if let Some(lock) = locks.get(&student_id) {
            lock.fetch_sub(1, Ordering::SeqCst);
        }
    }

    async fn is_student_locked(&self, student_id: StudentId) -> bool {
        let locks = self.student_locks.read().await;
        locks
            .get(&student_id)
            .map(|l| l.load(Ordering::SeqCst) > 0)
            .unwrap_or(false)
    }

    async fn log_change(
        &self,
        student_id: StudentId,
        change_type: ChangeType,
        operator: String,
        reason: Option<String>,
    ) {
        let log = ChangeLog {
            id: Uuid::new_v4(),
            student_id,
            change_type,
            operator,
            reason,
            timestamp: Utc::now(),
        };
        self.change_logs.write().await.push(log);
    }

    pub async fn get_phase(&self) -> SystemPhase {
        *self.phase.read().await
    }

    pub async fn advance_phase(&self) -> Result<SystemPhase> {
        let mut phase = self.phase.write().await;
        *phase = match *phase {
            SystemPhase::StudentSubmission => SystemPhase::AdvisorSelection,
            SystemPhase::AdvisorSelection => {
                self.finalize_auto().await?;
                SystemPhase::Appeal
            }
            SystemPhase::Appeal => SystemPhase::Finalized,
            SystemPhase::Finalized => return Err(ThesisError::InvalidOperation("Already finalized".into())),
        };
        Ok(*phase)
    }

    async fn finalize_auto(&self) -> Result<()> {
        let mut students = self.students.write().await;
        let mut advisors = self.advisors.write().await;
        let phase = self.phase.read().await;

        if *phase != SystemPhase::AdvisorSelection {
            return Ok(());
        }

        for student in students.values_mut() {
            match student.status {
                StudentStatus::AcceptedByAdvisor(adv_id) => {
                    if let Some(adv) = advisors.get_mut(&adv_id) {
                        if !adv.accepted.contains(&student.id) {
                            adv.accepted.push(student.id);
                        }
                    }
                    let old_status = student.status;
                    student.status = StudentStatus::Assigned(adv_id);
                    self.log_change(
                        student.id,
                        ChangeType::StatusChange {
                            from: old_status,
                            to: student.status,
                        },
                        "system-auto".into(),
                        None,
                    )
                    .await;
                }
                StudentStatus::WaitingAdvisor(_) | StudentStatus::RejectedByCurrent => {
                    let old_status = student.status;
                    student.status = StudentStatus::ToBeAssigned;
                    self.log_change(
                        student.id,
                        ChangeType::StatusChange {
                            from: old_status,
                            to: student.status,
                        },
                        "system-auto".into(),
                        None,
                    )
                    .await;
                }
                _ => {}
            }
        }
        Ok(())
    }

    pub async fn create_student(&self, name: String, student_no: String) -> Result<Student> {
        let student = Student::new(name, student_no);
        let id = student.id;
        self.students.write().await.insert(id, student.clone());
        Ok(student)
    }

    pub async fn create_advisor(&self, name: String, capacity: u32) -> Result<Advisor> {
        let advisor = Advisor::new(name, capacity);
        let id = advisor.id;
        self.advisors.write().await.insert(id, advisor.clone());
        Ok(advisor)
    }

    pub async fn list_students(&self) -> Result<Vec<Student>> {
        let students = self.students.read().await;
        Ok(students.values().cloned().collect())
    }

    pub async fn list_advisors(&self) -> Result<Vec<Advisor>> {
        let advisors = self.advisors.read().await;
        Ok(advisors.values().cloned().collect())
    }

    pub async fn get_student(&self, id: StudentId) -> Result<Student> {
        self.students
            .read()
            .await
            .get(&id)
            .cloned()
            .ok_or_else(|| ThesisError::StudentNotFound(id.to_string()))
    }

    pub async fn get_advisor(&self, id: AdvisorId) -> Result<Advisor> {
        self.advisors
            .read()
            .await
            .get(&id)
            .cloned()
            .ok_or_else(|| ThesisError::AdvisorNotFound(id.to_string()))
    }

    pub async fn submit_preferences(
        &self,
        student_id: StudentId,
        preferences: [Option<AdvisorId>; 3],
    ) -> Result<Student> {
        self.ensure_phase(&[SystemPhase::StudentSubmission]).await?;

        let mut students = self.students.write().await;
        let student = students
            .get_mut(&student_id)
            .ok_or_else(|| ThesisError::StudentNotFound(student_id.to_string()))?;

        if matches!(student.status, StudentStatus::Submitted | StudentStatus::WaitingAdvisor(_)) {
            return Err(ThesisError::AlreadySubmitted);
        }

        self.validate_preferences(&preferences).await?;

        student.preferences = preferences;
        let old_status = student.status;
        student.status = StudentStatus::WaitingAdvisor(PriorityLevel::First);
        student.can_modify_once = true;

        self.log_change(
            student.id,
            ChangeType::StatusChange {
                from: old_status,
                to: student.status,
            },
            "student".into(),
            Some("submitted preferences".into()),
        )
        .await;

        Ok(student.clone())
    }

    async fn validate_preferences(&self, prefs: &[Option<AdvisorId>; 3]) -> Result<()> {
        let advisors = self.advisors.read().await;
        let mut seen = HashMap::new();
        let mut count = 0;

        for (i, p) in prefs.iter().enumerate() {
            if let Some(id) = p {
                if !advisors.contains_key(id) {
                    return Err(ThesisError::AdvisorNotFound(id.to_string()));
                }
                *seen.entry(id).or_insert(0) += 1;
                count += 1;

                if i > 0 && prefs[i - 1].is_none() {
                    return Err(ThesisError::InvalidPreferences);
                }
            }
        }

        if count == 0 || seen.values().any(|&c| c > 1) {
            return Err(ThesisError::InvalidPreferences);
        }

        Ok(())
    }

    pub async fn modify_preferences(
        &self,
        student_id: StudentId,
        preferences: [Option<AdvisorId>; 3],
    ) -> Result<Student> {
        self.ensure_phase(&[SystemPhase::StudentSubmission]).await?;

        let mut students = self.students.write().await;
        let student = students
            .get_mut(&student_id)
            .ok_or_else(|| ThesisError::StudentNotFound(student_id.to_string()))?;

        if !student.can_modify_once {
            return Err(ThesisError::CannotModify);
        }

        self.validate_preferences(&preferences).await?;

        student.preferences = preferences;
        student.can_modify_once = false;

        self.log_change(
            student.id,
            ChangeType::StatusChange {
                from: StudentStatus::WaitingAdvisor(PriorityLevel::First),
                to: StudentStatus::WaitingAdvisor(PriorityLevel::First),
            },
            "student".into(),
            Some("modified preferences".into()),
        )
        .await;

        Ok(student.clone())
    }

    pub async fn get_advisor_pool(&self, advisor_id: AdvisorId) -> Result<Vec<Student>> {
        self.ensure_phase(&[
            SystemPhase::AdvisorSelection,
            SystemPhase::Appeal,
            SystemPhase::Finalized,
        ])
        .await?;

        self.advisors
            .read()
            .await
            .get(&advisor_id)
            .ok_or_else(|| ThesisError::AdvisorNotFound(advisor_id.to_string()))?;

        let students = self.students.read().await;
        let mut result = Vec::new();

        for student in students.values() {
            let eligible = match student.status {
                StudentStatus::WaitingAdvisor(p) => {
                    student.preferences[p.index()] == Some(advisor_id)
                        && !self.is_student_locked(student.id).await
                }
                _ => false,
            };

            if eligible {
                result.push(student.clone());
            }
        }

        Ok(result)
    }

    pub async fn advisor_select(
        &self,
        advisor_id: AdvisorId,
        student_id: StudentId,
        accept: bool,
    ) -> Result<()> {
        self.ensure_phase(&[SystemPhase::AdvisorSelection]).await?;

        if !self.lock_student(student_id).await? {
            return Err(ThesisError::ConcurrentConflict(format!(
                "Student {} is being processed by another advisor",
                student_id
            )));
        }

        let result = self
            .do_advisor_select(advisor_id, student_id, accept)
            .await;

        self.unlock_student(student_id).await;
        result
    }

    async fn do_advisor_select(
        &self,
        advisor_id: AdvisorId,
        student_id: StudentId,
        accept: bool,
    ) -> Result<()> {
        let mut advisors = self.advisors.write().await;
        let advisor = advisors
            .get_mut(&advisor_id)
            .ok_or_else(|| ThesisError::AdvisorNotFound(advisor_id.to_string()))?;

        if accept && !advisor.has_capacity() {
            return Err(ThesisError::NoCapacity);
        }

        let mut students = self.students.write().await;
        let student = students
            .get_mut(&student_id)
            .ok_or_else(|| ThesisError::StudentNotFound(student_id.to_string()))?;

        let current_priority = match student.status {
            StudentStatus::WaitingAdvisor(p) => p,
            _ => return Err(ThesisError::NotEligible),
        };

        if student.preferences[current_priority.index()] != Some(advisor_id) {
            return Err(ThesisError::NotEligible);
        }

        let record = SelectionRecord {
            id: Uuid::new_v4(),
            advisor_id,
            student_id,
            priority: current_priority,
            accepted: accept,
            timestamp: Utc::now(),
        };
        self.selection_records.write().await.push(record);

        if accept {
            if !advisor.has_capacity() {
                return Err(ThesisError::NoCapacity);
            }

            let old_status = student.status;
            student.status = StudentStatus::AcceptedByAdvisor(advisor_id);

            self.log_change(
                student.id,
                ChangeType::StatusChange {
                    from: old_status,
                    to: student.status,
                },
                format!("advisor-{}", advisor_id),
                Some("accepted by advisor".into()),
            )
            .await;
        } else {
            let old_status = student.status;
            student.status = StudentStatus::RejectedByCurrent;

            self.log_change(
                student.id,
                ChangeType::StatusChange {
                    from: old_status,
                    to: student.status,
                },
                format!("advisor-{}", advisor_id),
                Some("rejected by advisor".into()),
            )
            .await;

            self.activate_next_preference(student_id).await?;
        }

        Ok(())
    }

    async fn activate_next_preference(&self, student_id: StudentId) -> Result<()> {
        let mut students = self.students.write().await;
        let student = students
            .get_mut(&student_id)
            .ok_or_else(|| ThesisError::StudentNotFound(student_id.to_string()))?;

        let current_priority = match student.status {
            StudentStatus::RejectedByCurrent => {
                let mut p = PriorityLevel::First;
                for pref in &student.preferences {
                    if pref.is_some() {
                        p = match p {
                            PriorityLevel::First => PriorityLevel::Second,
                            PriorityLevel::Second => PriorityLevel::Third,
                            PriorityLevel::Third => break,
                        };
                    }
                }
                PriorityLevel::First
            }
            _ => return Ok(()),
        };

        let mut next_p = None;
        for (idx, pref) in student.preferences.iter().enumerate() {
            if pref.is_some() && idx > current_priority.index() {
                next_p = match idx {
                    1 => Some(PriorityLevel::Second),
                    2 => Some(PriorityLevel::Third),
                    _ => None,
                };
                break;
            }
        }

        let old_status = student.status;
        if let Some(np) = next_p {
            student.status = StudentStatus::WaitingAdvisor(np);
        } else {
            student.status = StudentStatus::ToBeAssigned;
        }

        self.log_change(
            student.id,
            ChangeType::StatusChange {
                from: old_status,
                to: student.status,
            },
            "system-auto".into(),
            Some("activated next preference or moved to to-be-assigned".into()),
        )
        .await;

        Ok(())
    }

    pub async fn create_appeal(
        &self,
        student_id: StudentId,
        reason: String,
    ) -> Result<AppealRecord> {
        self.ensure_phase(&[SystemPhase::Appeal]).await?;

        self.get_student(student_id).await?;

        let appeal = AppealRecord {
            id: Uuid::new_v4(),
            student_id,
            reason,
            resolved: false,
            created_at: Utc::now(),
            resolved_at: None,
        };

        self.appeals
            .write()
            .await
            .insert(appeal.id, appeal.clone());
        Ok(appeal)
    }

    pub async fn list_appeals(&self) -> Result<Vec<AppealRecord>> {
        Ok(self.appeals.read().await.values().cloned().collect())
    }

    pub async fn manual_assign(
        &self,
        student_id: StudentId,
        advisor_id: AdvisorId,
        reason: String,
        operator: String,
    ) -> Result<()> {
        self.ensure_phase(&[SystemPhase::Appeal, SystemPhase::Finalized])
            .await?;

        if !self.lock_student(student_id).await? {
            return Err(ThesisError::ConcurrentConflict(
                "Student is locked".into(),
            ));
        }

        let result = self
            .do_manual_assign(student_id, advisor_id, reason, operator)
            .await;

        self.unlock_student(student_id).await;
        result
    }

    async fn do_manual_assign(
        &self,
        student_id: StudentId,
        advisor_id: AdvisorId,
        reason: String,
        operator: String,
    ) -> Result<()> {
        {
            let advisors = self.advisors.read().await;
            let advisor = advisors
                .get(&advisor_id)
                .ok_or_else(|| ThesisError::AdvisorNotFound(advisor_id.to_string()))?;

            if !advisor.has_capacity() && !advisor.accepted.contains(&student_id) {
                return Err(ThesisError::NoCapacity);
            }
        }

        let old_assigned_advisor: Option<AdvisorId>;
        let old_status: StudentStatus;
        {
            let students = self.students.read().await;
            let student = students
                .get(&student_id)
                .ok_or_else(|| ThesisError::StudentNotFound(student_id.to_string()))?;

            old_status = student.status;
            old_assigned_advisor = match student.status {
                StudentStatus::Assigned(id) => {
                    if id == advisor_id {
                        return Ok(());
                    }
                    Some(id)
                }
                _ => None,
            };
        }

        if let Some(old_adv_id) = old_assigned_advisor {
            let mut advisors = self.advisors.write().await;
            if let Some(old_adv) = advisors.get_mut(&old_adv_id) {
                old_adv.accepted.retain(|&id| id != student_id);
            }
            self.log_change(
                student_id,
                ChangeType::Unassignment {
                    advisor_id: old_adv_id,
                },
                operator.clone(),
                Some(reason.clone()),
            )
            .await;
        }

        {
            let mut students = self.students.write().await;
            let student = students.get_mut(&student_id).unwrap();
            student.status = StudentStatus::Assigned(advisor_id);
        }

        {
            let mut advisors = self.advisors.write().await;
            if let Some(advisor) = advisors.get_mut(&advisor_id) {
                if !advisor.accepted.contains(&student_id) {
                    advisor.accepted.push(student_id);
                }
            }
        }

        self.log_change(
            student_id,
            ChangeType::Assignment { advisor_id },
            operator,
            Some(reason),
        )
        .await;

        self.log_change(
            student_id,
            ChangeType::StatusChange {
                from: old_status,
                to: StudentStatus::Assigned(advisor_id),
            },
            "system".into(),
            None,
        )
        .await;

        Ok(())
    }

    pub async fn get_change_logs(&self, student_id: Option<StudentId>) -> Result<Vec<ChangeLog>> {
        let logs = self.change_logs.read().await;
        Ok(match student_id {
            Some(id) => logs
                .iter()
                .filter(|l| l.student_id == id)
                .cloned()
                .collect(),
            None => logs.clone(),
        })
    }

    pub async fn get_match_results(&self) -> Result<Vec<MatchResult>> {
        let students = self.students.read().await;
        let advisors = self.advisors.read().await;

        let mut results = Vec::new();
        for student in students.values() {
            let advisor = match student.status {
                StudentStatus::Assigned(id) => advisors.get(&id).cloned(),
                StudentStatus::AcceptedByAdvisor(id) => advisors.get(&id).cloned(),
                _ => None,
            };
            results.push(MatchResult {
                student: student.clone(),
                advisor,
            });
        }
        Ok(results)
    }

    pub async fn get_status(&self) -> SystemStatus {
        SystemStatus {
            phase: self.get_phase().await,
            student_count: self.students.read().await.len(),
            advisor_count: self.advisors.read().await.len(),
        }
    }
}
