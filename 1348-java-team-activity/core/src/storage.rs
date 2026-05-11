use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::models::{Activity, Department};

#[derive(Debug, Default, Clone)]
pub struct InMemoryStorage {
    departments: Arc<RwLock<HashMap<Uuid, Department>>>,
    activities: Arc<RwLock<HashMap<Uuid, Activity>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self {
            departments: Arc::new(RwLock::new(HashMap::new())),
            activities: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn add_department(&self, dept: Department) {
        let mut departments = self.departments.write().await;
        departments.insert(dept.id, dept);
    }

    pub async fn get_department(&self, id: Uuid) -> Option<Department> {
        let departments = self.departments.read().await;
        departments.get(&id).cloned()
    }

    pub async fn update_department(&self, dept: Department) {
        let mut departments = self.departments.write().await;
        departments.insert(dept.id, dept);
    }

    pub async fn add_activity(&self, activity: Activity) {
        let mut activities = self.activities.write().await;
        activities.insert(activity.id, activity);
    }

    pub async fn get_activity(&self, id: Uuid) -> Option<Activity> {
        let activities = self.activities.read().await;
        activities.get(&id).cloned()
    }

    pub async fn update_activity(&self, activity: Activity) {
        let mut activities = self.activities.write().await;
        activities.insert(activity.id, activity);
    }

    pub async fn get_all_activities(&self) -> Vec<Activity> {
        let activities = self.activities.read().await;
        activities.values().cloned().collect()
    }

    pub async fn get_all_departments(&self) -> Vec<Department> {
        let departments = self.departments.read().await;
        departments.values().cloned().collect()
    }
}
