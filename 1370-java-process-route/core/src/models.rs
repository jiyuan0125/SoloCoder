use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ProcessStatus {
    Waiting,
    InProgress,
    Paused,
    Completed,
}

impl std::fmt::Display for ProcessStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ProcessStatus::Waiting => write!(f, "等待中"),
            ProcessStatus::InProgress => write!(f, "进行中"),
            ProcessStatus::Paused => write!(f, "暂停"),
            ProcessStatus::Completed => write!(f, "已完成"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessDefinition {
    pub id: String,
    pub name: String,
    pub standard_time_minutes: u32,
    pub predecessors: Vec<String>,
}

impl ProcessDefinition {
    pub fn new(name: impl Into<String>, standard_time_minutes: u32, predecessors: Vec<String>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name: name.into(),
            standard_time_minutes,
            predecessors,
        }
    }

    pub fn with_id(id: impl Into<String>, name: impl Into<String>, standard_time_minutes: u32, predecessors: Vec<String>) -> Self {
        Self {
            id: id.into(),
            name: name.into(),
            standard_time_minutes,
            predecessors,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessRouteVersion {
    pub version: String,
    pub processes: Vec<ProcessDefinition>,
    pub created_at: DateTime<Utc>,
}

impl ProcessRouteVersion {
    pub fn new(version: impl Into<String>, processes: Vec<ProcessDefinition>) -> Self {
        Self {
            version: version.into(),
            processes,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessRoute {
    pub id: String,
    pub name: String,
    pub versions: Vec<ProcessRouteVersion>,
    pub created_at: DateTime<Utc>,
}

impl ProcessRoute {
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name: name.into(),
            versions: Vec::new(),
            created_at: Utc::now(),
        }
    }

    pub fn add_version(&mut self, version: ProcessRouteVersion) {
        self.versions.push(version);
    }

    pub fn get_version(&self, version: &str) -> Option<&ProcessRouteVersion> {
        self.versions.iter().find(|v| v.version == version)
    }

    pub fn get_latest_version(&self) -> Option<&ProcessRouteVersion> {
        self.versions.last()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessInstance {
    pub definition_id: String,
    pub name: String,
    pub standard_time_minutes: u32,
    pub predecessors: Vec<String>,
    pub status: ProcessStatus,
    pub start_time: Option<DateTime<Utc>>,
    pub end_time: Option<DateTime<Utc>>,
    pub is_critical: bool,
}

impl ProcessInstance {
    pub fn from_definition(def: &ProcessDefinition, is_critical: bool) -> Self {
        Self {
            definition_id: def.id.clone(),
            name: def.name.clone(),
            standard_time_minutes: def.standard_time_minutes,
            predecessors: def.predecessors.clone(),
            status: ProcessStatus::Waiting,
            start_time: None,
            end_time: None,
            is_critical,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProductionTask {
    pub id: String,
    pub name: String,
    pub route_id: String,
    pub route_name: String,
    pub version: String,
    pub processes: Vec<ProcessInstance>,
    pub critical_path: Vec<String>,
    pub total_time_minutes: u32,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
}

impl ProductionTask {
    pub fn new(
        name: impl Into<String>,
        route_id: impl Into<String>,
        route_name: impl Into<String>,
        version: impl Into<String>,
        processes: Vec<ProcessInstance>,
        critical_path: Vec<String>,
        total_time_minutes: u32,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            name: name.into(),
            route_id: route_id.into(),
            route_name: route_name.into(),
            version: version.into(),
            processes,
            critical_path,
            total_time_minutes,
            started_at: None,
            completed_at: None,
        }
    }

    pub fn get_process(&self, definition_id: &str) -> Option<&ProcessInstance> {
        self.processes.iter().find(|p| p.definition_id == definition_id)
    }

    pub fn get_process_mut(&mut self, definition_id: &str) -> Option<&mut ProcessInstance> {
        self.processes.iter_mut().find(|p| p.definition_id == definition_id)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CriticalPathInfo {
    pub path: Vec<String>,
    pub total_time_minutes: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VersionComparison {
    pub version1: String,
    pub version2: String,
    pub total_time1: u32,
    pub total_time2: u32,
    pub time_difference: i32,
    pub critical_path1: Vec<String>,
    pub critical_path2: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Warning {
    pub task_id: String,
    pub task_name: String,
    pub process_id: String,
    pub process_name: String,
    pub message: String,
}
