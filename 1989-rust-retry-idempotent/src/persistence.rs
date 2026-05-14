use std::fs::{File, create_dir_all};
use std::io::{Read, Write};
use std::path::Path;
use serde::{Deserialize, Serialize};
use tracing::{info, warn};

use crate::models::{Task, TaskStatus};

const DEFAULT_STORAGE_DIR: &str = "./data";
const TASKS_FILE: &str = "tasks.json";

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PersistedState {
    pub tasks: Vec<Task>,
}

pub struct PersistenceManager {
    storage_dir: String,
}

impl PersistenceManager {
    pub fn new(storage_dir: Option<String>) -> Self {
        let dir = storage_dir.unwrap_or_else(|| DEFAULT_STORAGE_DIR.to_string());
        if let Err(e) = create_dir_all(&dir) {
            warn!("Failed to create storage directory {}: {}", dir, e);
        }
        Self { storage_dir: dir }
    }

    fn tasks_file_path(&self) -> String {
        Path::new(&self.storage_dir)
            .join(TASKS_FILE)
            .to_string_lossy()
            .to_string()
    }

    pub fn save(&self, tasks: &[Task]) -> Result<(), String> {
        let state = PersistedState {
            tasks: tasks.to_vec(),
        };
        
        let json = serde_json::to_string_pretty(&state)
            .map_err(|e| format!("Failed to serialize tasks: {}", e))?;
        
        let file_path = self.tasks_file_path();
        let mut file = File::create(&file_path)
            .map_err(|e| format!("Failed to create tasks file {}: {}", file_path, e))?;
        
        file.write_all(json.as_bytes())
            .map_err(|e| format!("Failed to write tasks file: {}", e))?;
        
        info!("Saved {} tasks to {}", tasks.len(), file_path);
        Ok(())
    }

    pub fn load(&self) -> Result<Vec<Task>, String> {
        let file_path = self.tasks_file_path();
        
        if !Path::new(&file_path).exists() {
            info!("No persisted state found at {}", file_path);
            return Ok(Vec::new());
        }

        let mut file = File::open(&file_path)
            .map_err(|e| format!("Failed to open tasks file {}: {}", file_path, e))?;
        
        let mut contents = String::new();
        file.read_to_string(&mut contents)
            .map_err(|e| format!("Failed to read tasks file: {}", e))?;
        
        if contents.trim().is_empty() {
            return Ok(Vec::new());
        }

        let state: PersistedState = serde_json::from_str(&contents)
            .map_err(|e| format!("Failed to deserialize tasks: {}", e))?;
        
        info!("Loaded {} tasks from {}", state.tasks.len(), file_path);
        
        let tasks = state.tasks
            .into_iter()
            .map(|mut task| {
                if !task.is_terminal() {
                    task.status = TaskStatus::Pending;
                }
                task
            })
            .collect();
        
        Ok(tasks)
    }
}
