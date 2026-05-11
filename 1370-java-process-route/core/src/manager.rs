use crate::dag::ProcessDag;
use crate::error::{ProcessRouteError, Result};
use crate::models::{
    CriticalPathInfo, ProcessDefinition, ProcessInstance, ProcessRoute, ProcessRouteVersion,
    ProcessStatus, ProductionTask, VersionComparison, Warning,
};
use crate::store::InMemoryStore;
use chrono::Utc;

#[derive(Clone)]
pub struct ProcessRouteManager {
    store: InMemoryStore,
}

impl ProcessRouteManager {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn create_route(&self, name: impl Into<String>) -> ProcessRoute {
        let route = ProcessRoute::new(name);
        self.store.save_route(route.clone());
        route
    }

    pub fn list_routes(&self) -> Vec<ProcessRoute> {
        self.store.list_routes()
    }

    pub fn get_route(&self, id: &str) -> Result<ProcessRoute> {
        self.store
            .get_route(id)
            .ok_or_else(|| ProcessRouteError::RouteNotFound(id.to_string()))
    }

    pub fn delete_route(&self, id: &str) -> Result<()> {
        if self.store.delete_route(id) {
            Ok(())
        } else {
            Err(ProcessRouteError::RouteNotFound(id.to_string()))
        }
    }

    pub fn add_version(
        &self,
        route_id: &str,
        version: impl Into<String>,
        processes: Vec<ProcessDefinition>,
    ) -> Result<ProcessRoute> {
        let version_str = version.into();

        let dag = ProcessDag::new(&processes)?;
        let _ = dag.calculate_critical_path()?;

        let mut route = self.get_route(route_id)?;

        if route.versions.iter().any(|v| v.version == version_str) {
            return Err(ProcessRouteError::VersionAlreadyExists(version_str));
        }

        route.add_version(ProcessRouteVersion::new(version_str, processes));
        self.store.save_route(route.clone());
        Ok(route)
    }

    pub fn get_version(&self, route_id: &str, version: &str) -> Result<ProcessRouteVersion> {
        let route = self.get_route(route_id)?;
        route
            .get_version(version)
            .cloned()
            .ok_or_else(|| ProcessRouteError::VersionNotFound(version.to_string()))
    }

    pub fn get_critical_path_for_version(
        &self,
        route_id: &str,
        version: &str,
    ) -> Result<CriticalPathInfo> {
        let version_data = self.get_version(route_id, version)?;
        let dag = ProcessDag::new(&version_data.processes)?;
        dag.calculate_critical_path()
    }

    pub fn compare_versions(
        &self,
        route_id: &str,
        version1: &str,
        version2: &str,
    ) -> Result<VersionComparison> {
        let cp1 = self.get_critical_path_for_version(route_id, version1)?;
        let cp2 = self.get_critical_path_for_version(route_id, version2)?;

        Ok(VersionComparison {
            version1: version1.to_string(),
            version2: version2.to_string(),
            total_time1: cp1.total_time_minutes,
            total_time2: cp2.total_time_minutes,
            time_difference: cp2.total_time_minutes as i32 - cp1.total_time_minutes as i32,
            critical_path1: cp1.path,
            critical_path2: cp2.path,
        })
    }

    pub fn create_task(
        &self,
        name: impl Into<String>,
        route_id: &str,
        version: Option<&str>,
    ) -> Result<ProductionTask> {
        let route = self.get_route(route_id)?;

        let version_to_use = match version {
            Some(v) => v.to_string(),
            None => route
                .get_latest_version()
                .map(|v| v.version.clone())
                .ok_or_else(|| {
                    ProcessRouteError::InvalidOperation("工艺路线没有可用版本".to_string())
                })?,
        };

        let version_data = self.get_version(route_id, &version_to_use)?;
        let dag = ProcessDag::new(&version_data.processes)?;
        let critical_info = dag.calculate_critical_path()?;
        let critical_ids: std::collections::HashSet<_> = critical_info.path.iter().collect();

        let processes: Vec<ProcessInstance> = version_data
            .processes
            .iter()
            .map(|p| {
                ProcessInstance::from_definition(p, critical_ids.contains(&p.id))
            })
            .collect();

        let task = ProductionTask::new(
            name,
            route_id,
            &route.name,
            version_to_use,
            processes,
            critical_info.path,
            critical_info.total_time_minutes,
        );

        self.store.save_task(task.clone());
        Ok(task)
    }

    pub fn list_tasks(&self) -> Vec<ProductionTask> {
        self.store.list_tasks()
    }

    pub fn get_task(&self, id: &str) -> Result<ProductionTask> {
        self.store
            .get_task(id)
            .ok_or_else(|| ProcessRouteError::TaskNotFound(id.to_string()))
    }

    pub fn delete_task(&self, id: &str) -> Result<()> {
        if self.store.delete_task(id) {
            Ok(())
        } else {
            Err(ProcessRouteError::TaskNotFound(id.to_string()))
        }
    }

    pub fn start_process(&self, task_id: &str, process_id: &str) -> Result<ProductionTask> {
        let task = self.get_task(task_id)?;

        {
            let process = task
                .get_process(process_id)
                .ok_or_else(|| ProcessRouteError::ProcessNotFound(process_id.to_string()))?;

            if process.status != ProcessStatus::Waiting {
                return Err(ProcessRouteError::ProcessNotWaiting);
            }

            let uncompleted_preds: Vec<String> = process
                .predecessors
                .iter()
                .filter(|pred_id| {
                    task.get_process(pred_id)
                        .map(|p| p.status != ProcessStatus::Completed)
                        .unwrap_or(true)
                })
                .cloned()
                .collect();

            if !uncompleted_preds.is_empty() {
                return Err(ProcessRouteError::PredecessorsNotCompleted(uncompleted_preds));
            }
        }

        self.store.update_task(task_id, |t| {
            if t.started_at.is_none() {
                t.started_at = Some(Utc::now());
            }

            if let Some(p) = t.get_process_mut(process_id) {
                p.status = ProcessStatus::InProgress;
                p.start_time = Some(Utc::now());
            }
        });

        self.get_task(task_id)
    }

    pub fn complete_process(&self, task_id: &str, process_id: &str) -> Result<ProductionTask> {
        let task = self.get_task(task_id)?;

        {
            let process = task
                .get_process(process_id)
                .ok_or_else(|| ProcessRouteError::ProcessNotFound(process_id.to_string()))?;

            if process.status == ProcessStatus::Completed {
                return Err(ProcessRouteError::ProcessAlreadyCompleted);
            }

            if process.status != ProcessStatus::InProgress {
                return Err(ProcessRouteError::ProcessNotInProgress);
            }
        }

        self.store.update_task(task_id, |t| {
            if let Some(p) = t.get_process_mut(process_id) {
                p.status = ProcessStatus::Completed;
                p.end_time = Some(Utc::now());
            }

            let all_completed = t
                .processes
                .iter()
                .all(|p| p.status == ProcessStatus::Completed);
            if all_completed {
                t.completed_at = Some(Utc::now());
            }
        });

        self.get_task(task_id)
    }

    pub fn pause_process(&self, task_id: &str, process_id: &str) -> Result<(ProductionTask, Option<Warning>)> {
        let task = self.get_task(task_id)?;

        {
            let process = task
                .get_process(process_id)
                .ok_or_else(|| ProcessRouteError::ProcessNotFound(process_id.to_string()))?;

            if process.status == ProcessStatus::Completed {
                return Err(ProcessRouteError::ProcessAlreadyCompleted);
            }

            if process.status != ProcessStatus::InProgress {
                return Err(ProcessRouteError::ProcessNotInProgress);
            }
        }

        let mut warning = None;
        let is_critical = task
            .get_process(process_id)
            .map(|p| p.is_critical)
            .unwrap_or(false);

        if is_critical {
            let process_name = task
                .get_process(process_id)
                .map(|p| p.name.clone())
                .unwrap_or_default();
            warning = Some(Warning {
                task_id: task.id.clone(),
                task_name: task.name.clone(),
                process_id: process_id.to_string(),
                process_name,
                message: format!(
                    "关键工序已暂停，可能影响整体工期。关键路径: {:?}",
                    task.critical_path
                ),
            });
        }

        self.store.update_task(task_id, |t| {
            if let Some(p) = t.get_process_mut(process_id) {
                p.status = ProcessStatus::Paused;
            }
        });

        Ok((self.get_task(task_id)?, warning))
    }

    pub fn resume_process(&self, task_id: &str, process_id: &str) -> Result<ProductionTask> {
        let task = self.get_task(task_id)?;

        {
            let process = task
                .get_process(process_id)
                .ok_or_else(|| ProcessRouteError::ProcessNotFound(process_id.to_string()))?;

            if process.status == ProcessStatus::Completed {
                return Err(ProcessRouteError::ProcessAlreadyCompleted);
            }

            if process.status != ProcessStatus::Paused {
                return Err(ProcessRouteError::InvalidOperation(
                    "工序不是暂停状态".to_string(),
                ));
            }
        }

        self.store.update_task(task_id, |t| {
            if let Some(p) = t.get_process_mut(process_id) {
                p.status = ProcessStatus::InProgress;
            }
        });

        self.get_task(task_id)
    }

    pub fn get_available_processes(&self, task_id: &str) -> Result<Vec<ProcessInstance>> {
        let task = self.get_task(task_id)?;

        let available: Vec<ProcessInstance> = task
            .processes
            .iter()
            .filter(|p| p.status == ProcessStatus::Waiting)
            .filter(|p| {
                p.predecessors.iter().all(|pred_id| {
                    task.get_process(pred_id)
                        .map(|pred| pred.status == ProcessStatus::Completed)
                        .unwrap_or(false)
                })
            })
            .cloned()
            .collect();

        Ok(available)
    }

    pub fn calculate_task_actual_duration(&self, task_id: &str) -> Result<Option<u32>> {
        let task = self.get_task(task_id)?;

        match (task.started_at, task.completed_at) {
            (Some(start), Some(end)) => {
                let duration = end.signed_duration_since(start);
                Ok(Some(duration.num_minutes().max(0) as u32))
            }
            _ => Ok(None),
        }
    }
}
