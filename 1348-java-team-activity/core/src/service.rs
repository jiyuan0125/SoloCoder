use chrono::Utc;
use std::sync::Arc;
use uuid::Uuid;

use crate::error::ActivityError;
use crate::models::{
    Activity, ActivityStatus, CreateActivityRequest, Department, Registration, RegistrationStatus,
    SettlingRequest, UpdateLimitRequest,
};
use crate::storage::InMemoryStorage;

pub struct ActivityService {
    storage: Arc<InMemoryStorage>,
}

impl ActivityService {
    pub fn new(storage: Arc<InMemoryStorage>) -> Self {
        Self { storage }
    }

    pub async fn create_department(
        &self,
        name: String,
        annual_budget: f64,
    ) -> Department {
        let dept = Department::new(name, annual_budget);
        self.storage.add_department(dept.clone()).await;
        dept
    }

    pub async fn get_department(&self, id: Uuid) -> Option<Department> {
        self.storage.get_department(id).await
    }

    pub async fn get_all_departments(&self) -> Vec<Department> {
        self.storage.get_all_departments().await
    }

    pub async fn create_activity(
        &self,
        req: CreateActivityRequest,
    ) -> Result<Activity, ActivityError> {
        let dept = self
            .storage
            .get_department(req.department_id)
            .await
            .ok_or(ActivityError::DepartmentNotFound)?;

        let activity = Activity::new(
            dept.id,
            req.name,
            req.description,
            req.cost_per_person,
            req.max_participants,
            req.start_time,
            req.end_time,
        );

        self.storage.add_activity(activity.clone()).await;
        Ok(activity)
    }

    pub async fn publish_activity(&self, activity_id: Uuid) -> Result<Activity, ActivityError> {
        let mut activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        if activity.status != ActivityStatus::Draft {
            return Err(ActivityError::InvalidState);
        }

        let dept = self
            .storage
            .get_department(activity.department_id)
            .await
            .ok_or(ActivityError::DepartmentNotFound)?;

        let estimated_cost = activity.estimated_total_cost();
        if estimated_cost > dept.remaining_budget {
            return Err(ActivityError::InsufficientBudget);
        }

        activity.status = ActivityStatus::Published;
        self.storage.update_activity(activity.clone()).await;

        Ok(activity)
    }

    pub async fn cancel_activity(&self, activity_id: Uuid) -> Result<Activity, ActivityError> {
        let activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        let now = Utc::now();
        if now >= activity.start_time {
            return Err(ActivityError::CannotCancelStartedActivity);
        }

        let mut activity = activity;
        if matches!(
            activity.status,
            ActivityStatus::Completed | ActivityStatus::Cancelled
        ) {
            return Err(ActivityError::InvalidState);
        }

        activity.status = ActivityStatus::Cancelled;
        activity.waitlist.clear();
        self.storage.update_activity(activity.clone()).await;

        Ok(activity)
    }

    pub async fn update_registration_limit(
        &self,
        activity_id: Uuid,
        req: UpdateLimitRequest,
    ) -> Result<Activity, ActivityError> {
        let mut activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        if req.new_limit > activity.max_participants {
            return Err(ActivityError::CannotIncreaseLimit);
        }

        if req.new_limit == activity.max_participants {
            return Ok(activity);
        }

        activity.max_participants = req.new_limit;

        let registered_count = activity.registered_count();
        if registered_count > req.new_limit {
            let excess = registered_count - req.new_limit;
            let mut to_waitlist: Vec<String> = Vec::new();

            for reg in activity.registrations.iter_mut() {
                if reg.status == RegistrationStatus::Registered
                    && to_waitlist.len() < excess
                {
                    reg.status = RegistrationStatus::Waitlisted;
                    to_waitlist.push(reg.user_id.clone());
                }
            }

            activity.waitlist.extend(to_waitlist);
        }

        self.storage.update_activity(activity.clone()).await;
        Ok(activity)
    }

    pub async fn register_for_activity(
        &self,
        activity_id: Uuid,
        user_id: String,
    ) -> Result<RegistrationStatus, ActivityError> {
        let activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        let now = Utc::now();

        if !activity.is_registration_open(now) {
            return Err(ActivityError::RegistrationClosed);
        }

        if activity.has_registration(&user_id) {
            return Err(ActivityError::UserAlreadyRegistered);
        }

        let mut activity = activity;
        let status = if activity.is_full() {
            if !activity.has_waitlist_entry(&user_id) {
                activity.waitlist.push(user_id.clone());
            }
            RegistrationStatus::Waitlisted
        } else {
            activity.registrations.push(Registration {
                user_id: user_id.clone(),
                status: RegistrationStatus::Registered,
                registered_at: now,
            });
            RegistrationStatus::Registered
        };

        self.storage.update_activity(activity).await;
        Ok(status)
    }

    pub async fn cancel_registration(
        &self,
        activity_id: Uuid,
        user_id: String,
    ) -> Result<(), ActivityError> {
        let activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        let now = Utc::now();
        let before_24h = activity.is_before_24_hours(now);

        let mut activity = activity;

        let was_on_waitlist = activity.waitlist.iter().position(|u| u == &user_id);
        if let Some(pos) = was_on_waitlist {
            activity.waitlist.remove(pos);
            self.storage.update_activity(activity).await;
            return Ok(());
        }

        let registration = activity
            .get_registration(&user_id)
            .ok_or(ActivityError::UserNotRegistered)?;

        if registration.status != RegistrationStatus::Registered {
            return Err(ActivityError::UserNotRegistered);
        }

        let new_status = if before_24h {
            RegistrationStatus::Cancelled
        } else {
            RegistrationStatus::TemporaryExit
        };

        for reg in activity.registrations.iter_mut() {
            if reg.user_id == user_id {
                reg.status = new_status;
                break;
            }
        }

        self.promote_from_waitlist(&mut activity);
        self.storage.update_activity(activity).await;
        Ok(())
    }

    fn promote_from_waitlist(&self, activity: &mut Activity) {
        while !activity.waitlist.is_empty() && !activity.is_full() {
            let user_id = activity.waitlist.remove(0);
            for reg in activity.registrations.iter_mut() {
                if reg.user_id == user_id {
                    reg.status = RegistrationStatus::Registered;
                }
            }
            if !activity.has_registration(&user_id) {
                activity.registrations.push(Registration {
                    user_id,
                    status: RegistrationStatus::Registered,
                    registered_at: Utc::now(),
                });
            }
        }
    }

    pub async fn close_registration(&self, activity_id: Uuid) -> Result<Activity, ActivityError> {
        let mut activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        if activity.status == ActivityStatus::Published {
            activity.status = ActivityStatus::RegistrationClosed;
            activity.waitlist.clear();
            self.storage.update_activity(activity.clone()).await;
        }

        Ok(activity)
    }

    pub async fn settle_activity(
        &self,
        activity_id: Uuid,
        req: SettlingRequest,
    ) -> Result<Activity, ActivityError> {
        let mut activity = self
            .storage
            .get_activity(activity_id)
            .await
            .ok_or(ActivityError::ActivityNotFound)?;

        if activity.status == ActivityStatus::Cancelled {
            return Err(ActivityError::InvalidState);
        }

        if activity.status == ActivityStatus::Completed {
            return Ok(activity);
        }

        let registered_users: Vec<String> = activity
            .registrations
            .iter()
            .filter(|r| r.status == RegistrationStatus::Registered)
            .map(|r| r.user_id.clone())
            .collect();

        if req.actual_participants.len() != registered_users.len() {
            return Err(ActivityError::ParticipantCountMismatch);
        }

        for participant in &req.actual_participants {
            if !registered_users.contains(participant) {
                return Err(ActivityError::ParticipantCountMismatch);
            }
        }

        for user in &registered_users {
            if !req.actual_participants.contains(user) {
                return Err(ActivityError::ParticipantCountMismatch);
            }
        }

        let mut dept = self
            .storage
            .get_department(activity.department_id)
            .await
            .ok_or(ActivityError::DepartmentNotFound)?;

        let actual_cost = activity.cost_per_person * req.actual_participants.len() as f64;
        dept.remaining_budget -= actual_cost;

        activity.status = ActivityStatus::Completed;
        activity.actual_participants = Some(req.actual_participants);

        self.storage.update_department(dept).await;
        self.storage.update_activity(activity.clone()).await;

        Ok(activity)
    }

    pub async fn get_activity(&self, id: Uuid) -> Option<Activity> {
        self.storage.get_activity(id).await
    }

    pub async fn get_all_activities(&self) -> Vec<Activity> {
        self.storage.get_all_activities().await
    }
}
