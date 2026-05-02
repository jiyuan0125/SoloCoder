use crate::dag::BuildDAG;
use std::collections::{HashMap, HashSet};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use std::process::Command;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum TargetStatus {
    Built,
    Cached,
    Failed,
    Skipped,
}

impl std::fmt::Display for TargetStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            TargetStatus::Built => write!(f, "built"),
            TargetStatus::Cached => write!(f, "cached"),
            TargetStatus::Failed => write!(f, "failed"),
            TargetStatus::Skipped => write!(f, "skipped"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct TargetResult {
    pub name: String,
    pub status: TargetStatus,
    pub duration_ms: u64,
}

pub struct ParallelExecutor {
    parallelism: usize,
}

impl ParallelExecutor {
    pub fn new(parallelism: usize) -> Self {
        Self {
            parallelism: parallelism.max(1),
        }
    }

    pub fn default() -> Self {
        Self::new(4)
    }

    pub fn execute<F>(
        &self,
        dag: &BuildDAG,
        should_rebuild: F,
    ) -> (Vec<TargetResult>, Duration)
    where
        F: Fn(&str) -> bool + Send + Sync + Clone + 'static,
    {
        let total_start = Instant::now();
        
        let topological_order = dag.topological_order();
        let results: Arc<Mutex<HashMap<String, TargetResult>>> = Arc::new(Mutex::new(HashMap::new()));
        let completed: Arc<Mutex<HashSet<String>>> = Arc::new(Mutex::new(HashSet::new()));
        let in_flight: Arc<Mutex<HashSet<String>>> = Arc::new(Mutex::new(HashSet::new()));
        let failed: Arc<Mutex<HashSet<String>>> = Arc::new(Mutex::new(HashSet::new()));

        let (tx, rx) = std::sync::mpsc::channel();

        loop {
            {
                let completed_lock = completed.lock().unwrap();
                if completed_lock.len() == topological_order.len() {
                    break;
                }
            }

            let ready_nodes = {
                let completed_lock = completed.lock().unwrap();
                let in_flight_lock = in_flight.lock().unwrap();
                let failed_lock = failed.lock().unwrap();
                
                let mut ready = dag.get_ready_nodes(&completed_lock);
                ready.retain(|name| !in_flight_lock.contains(name));
                ready.retain(|name| {
                    if let Some(node) = dag.nodes.get(name) {
                        !node.dependencies.iter().any(|dep| failed_lock.contains(dep))
                    } else {
                        false
                    }
                });
                ready
            };

            {
                let in_flight_lock = in_flight.lock().unwrap();
                let slots_available = self.parallelism.saturating_sub(in_flight_lock.len());
                
                if slots_available > 0 && !ready_nodes.is_empty() {
                    let nodes_to_spawn: Vec<_> = ready_nodes.into_iter().take(slots_available).collect();
                    
                    for node_name in nodes_to_spawn {
                        {
                            let mut in_flight_lock = in_flight.lock().unwrap();
                            in_flight_lock.insert(node_name.clone());
                        }

                        let node = dag.nodes.get(&node_name).unwrap().clone();
                        let should_rebuild_clone = should_rebuild.clone();
                        let tx_clone = tx.clone();
                        let results_clone = results.clone();
                        let completed_clone = completed.clone();
                        let in_flight_clone = in_flight.clone();
                        let failed_clone = failed.clone();

                        rayon::spawn(move || {
                            let start = Instant::now();
                            let needs_rebuild = should_rebuild_clone(&node.name);
                            
                            let (status, success) = if needs_rebuild {
                                let output = Command::new("sh")
                                    .arg("-c")
                                    .arg(&node.config.command)
                                    .output();

                                match output {
                                    Ok(o) if o.status.success() => (TargetStatus::Built, true),
                                    _ => (TargetStatus::Failed, false),
                                }
                            } else {
                                (TargetStatus::Cached, true)
                            };

                            let duration_ms = start.elapsed().as_millis() as u64;

                            let result = TargetResult {
                                name: node.name.clone(),
                                status: status.clone(),
                                duration_ms,
                            };

                            {
                                let mut results_lock = results_clone.lock().unwrap();
                                results_lock.insert(node.name.clone(), result);
                            }

                            {
                                let mut completed_lock = completed_clone.lock().unwrap();
                                completed_lock.insert(node.name.clone());
                            }

                            {
                                let mut in_flight_lock = in_flight_clone.lock().unwrap();
                                in_flight_lock.remove(&node.name);
                            }

                            if !success {
                                let mut failed_lock = failed_clone.lock().unwrap();
                                failed_lock.insert(node.name.clone());
                            }

                            let _ = tx_clone.send(());
                        });
                    }
                }
            }

            let _ = rx.recv_timeout(Duration::from_millis(100));
        }

        let results_map = results.lock().unwrap().clone();
        let mut results_vec: Vec<TargetResult> = Vec::new();
        
        for name in &topological_order {
            if let Some(result) = results_map.get(name) {
                results_vec.push(result.clone());
            } else {
                let deps_failed = {
                    let failed_lock = failed.lock().unwrap();
                    dag.nodes.get(name).map_or(false, |node| {
                        node.dependencies.iter().any(|dep| failed_lock.contains(dep))
                    })
                };

                results_vec.push(TargetResult {
                    name: name.clone(),
                    status: if deps_failed { TargetStatus::Skipped } else { TargetStatus::Failed },
                    duration_ms: 0,
                });
            }
        }

        let total_duration = total_start.elapsed();
        (results_vec, total_duration)
    }
}
