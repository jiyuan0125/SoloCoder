use std::collections::HashMap;
use std::sync::Arc;

use tokio::sync::RwLock;
use uuid::Uuid;

use crate::models::{AggregatedHealth, OverallStatus, Probe, ProbeResult};
use crate::scheduler::{SchedulerRegistry, start_probe_scheduler};

const THRESHOLD: f32 = 0.8;

#[derive(Clone)]
pub struct AppState {
    pub probes: Arc<RwLock<HashMap<Uuid, Arc<Probe>>>>,
    pub results: Arc<RwLock<HashMap<Uuid, ProbeResult>>>,
    pub schedulers: Arc<SchedulerRegistry>,
}

impl AppState {
    pub fn new() -> Self {
        AppState {
            probes: Arc::new(RwLock::new(HashMap::new())),
            results: Arc::new(RwLock::new(HashMap::new())),
            schedulers: Arc::new(SchedulerRegistry::new()),
        }
    }

    pub fn start_all_schedulers(&self) {
    }

    pub async fn add_probe(&self, probe: Probe) -> Uuid {
        let id = probe.id;
        let probe_arc = Arc::new(probe);
        self.probes.write().await.insert(id, probe_arc.clone());
        let handle = start_probe_scheduler(probe_arc, self.clone()).await;
        self.schedulers.register(id, handle).await;
        id
    }

    pub async fn remove_probe(&self, id: &Uuid) -> bool {
        if self.probes.write().await.remove(id).is_some() {
            self.results.write().await.remove(id);
            self.schedulers.unregister(id).await;
            true
        } else {
            false
        }
    }

    pub async fn list_probes(&self) -> Vec<Probe> {
        self.probes
            .read()
            .await
            .values()
            .map(|p| (**p).clone())
            .collect()
    }

    pub async fn get_probe(&self, id: &Uuid) -> Option<Probe> {
        self.probes.read().await.get(id).map(|p| (**p).clone())
    }

    pub async fn update_result(&self, probe_id: Uuid, result: ProbeResult) {
        self.results.write().await.insert(probe_id, result);
    }

    pub async fn get_result(&self, id: &Uuid) -> Option<ProbeResult> {
        self.results.read().await.get(id).cloned()
    }

    pub async fn aggregate_health(&self) -> AggregatedHealth {
        let probes = self.probes.read().await;
        let results = self.results.read().await;

        let mut total_weight: u32 = 0;
        let mut healthy_weight: u32 = 0;

        for probe in probes.values() {
            total_weight += probe.weight;
            if let Some(result) = results.get(&probe.id) {
                if result.status == crate::models::HealthStatus::Healthy {
                    healthy_weight += probe.weight;
                }
            }
        }

        let health_score = if total_weight > 0 {
            healthy_weight as f32 / total_weight as f32
        } else {
            1.0
        };

        let overall_status = if health_score >= THRESHOLD {
            OverallStatus::Healthy
        } else if health_score > 0.0 {
            OverallStatus::Degraded
        } else {
            OverallStatus::Unhealthy
        };

        AggregatedHealth {
            overall_status,
            health_score,
            total_weight,
            healthy_weight,
            threshold: THRESHOLD,
        }
    }
}
