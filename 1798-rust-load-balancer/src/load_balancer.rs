use std::collections::HashMap;
use std::sync::atomic::Ordering;
use std::sync::Arc;

use chrono::{DateTime, Utc};
use parking_lot::RwLock;
use rand::seq::SliceRandom;
use rand::thread_rng;
use tracing::info;

use crate::types::{
    AllocationRecord, Backend, BackendStats, BackendStatus, LoadBalancingStrategy, SelectionReason,
};

const MAX_ALLOCATION_RECORDS: usize = 50;
const WARMUP_DURATION_SECONDS: i64 = 10;
const WARMUP_TRAFFIC_PERCENT: f64 = 0.05;

#[derive(Clone)]
pub struct LoadBalancer {
    backends: Arc<RwLock<HashMap<String, Arc<Backend>>>>,
    allocation_records: Arc<RwLock<Vec<AllocationRecord>>>,
    strategy: LoadBalancingStrategy,
}

impl LoadBalancer {
    pub fn new(strategy: LoadBalancingStrategy) -> Self {
        Self {
            backends: Arc::new(RwLock::new(HashMap::new())),
            allocation_records: Arc::new(RwLock::new(Vec::new())),
            strategy,
        }
    }

    pub fn add_backend(&self, address: String, weight: u32) {
        let now = Utc::now();
        let warmup_end = now + chrono::Duration::seconds(WARMUP_DURATION_SECONDS);
        let addr_clone = address.clone();
        
        let backend = Arc::new(Backend {
            address,
            weight,
            current_connections: 0.into(),
            total_requests: 0.into(),
            total_failures: 0.into(),
            added_at: now,
            warmup_end,
            status: RwLock::new(BackendStatus::Healthy),
            consecutive_failures: 0.into(),
            consecutive_successes: 0.into(),
        });

        let mut backends = self.backends.write();
        backends.insert(addr_clone.clone(), backend);
        info!("Backend added: {}", addr_clone);
    }

    pub fn remove_backend(&self, address: &str) -> bool {
        let mut backends = self.backends.write();
        if let Some(backend) = backends.get(address) {
            let current_connections = backend.current_connections.load(Ordering::Relaxed);
            if current_connections == 0 {
                backends.remove(address);
                info!("Backend removed: {}", address);
                true
            } else {
                let mut status = backend.status.write();
                *status = BackendStatus::Draining;
                info!("Backend marked as draining: {}", address);
                false
            }
        } else {
            false
        }
    }

    pub fn clean_up_draining(&self) {
        let mut backends = self.backends.write();
        let to_remove: Vec<String> = backends
            .iter()
            .filter(|(_, backend)| {
                let status = *backend.status.read();
                status == BackendStatus::Draining
                    && backend.current_connections.load(Ordering::Relaxed) == 0
            })
            .map(|(addr, _)| addr.clone())
            .collect();
        
        for addr in to_remove {
            backends.remove(&addr);
            info!("Draining backend removed: {}", addr);
        }
    }

    pub fn get_backends(&self) -> Vec<Arc<Backend>> {
        let backends = self.backends.read();
        backends.values().cloned().collect()
    }

    pub fn get_stats(&self) -> Vec<BackendStats> {
        let backends = self.backends.read();
        let now = Utc::now();
        
        backends
            .values()
            .map(|b| BackendStats {
                address: b.address.clone(),
                weight: b.weight,
                current_connections: b.current_connections.load(Ordering::Relaxed),
                total_requests: b.total_requests.load(Ordering::Relaxed),
                total_failures: b.total_failures.load(Ordering::Relaxed),
                added_at: b.added_at,
                is_warming_up: now < b.warmup_end,
                status: *b.status.read(),
            })
            .collect()
    }

    pub fn get_allocation_records(&self) -> Vec<AllocationRecord> {
        let records = self.allocation_records.read();
        records.clone()
    }

    pub fn select_backend(&self) -> Option<(Arc<Backend>, SelectionReason)> {
        let backends = self.backends.read();
        let now = Utc::now();
        
        let available_backends: Vec<Arc<Backend>> = backends
            .values()
            .filter(|b| {
                let status = *b.status.read();
                status == BackendStatus::Healthy
            })
            .cloned()
            .collect();

        if available_backends.is_empty() {
            return None;
        }

        let (warming_up, warmed_up): (Vec<Arc<Backend>>, Vec<Arc<Backend>>) = available_backends
            .into_iter()
            .partition(|b| now < b.warmup_end);

        match self.strategy {
            LoadBalancingStrategy::LeastConnections => {
                self.select_least_connections(warmed_up, warming_up, now)
            }
            LoadBalancingStrategy::Random => {
                self.select_random(warmed_up, warming_up, now)
            }
        }
    }

    fn select_least_connections(
        &self,
        warmed_up: Vec<Arc<Backend>>,
        warming_up: Vec<Arc<Backend>>,
        _now: DateTime<Utc>,
    ) -> Option<(Arc<Backend>, SelectionReason)> {
        let mut rng = thread_rng();
        
        if !warming_up.is_empty() {
            let total_count = warmed_up.len() + warming_up.len();
            let warmup_max_traffic = (total_count as f64 * WARMUP_TRAFFIC_PERCENT) as usize;
            let warmup_count = warming_up.len().min(warmup_max_traffic.max(1));
            
            let select_warmup = rand::random::<f64>() < (warmup_count as f64 / total_count as f64);
            
            if select_warmup {
                let selected = warming_up.choose(&mut rng)?.clone();
                return Some((selected, SelectionReason::WarmupLimit));
            }
        }

        if warmed_up.is_empty() {
            return warming_up.choose(&mut rng).cloned().map(|b| (b, SelectionReason::WarmupLimit));
        }

        let min_connections = warmed_up
            .iter()
            .map(|b| b.current_connections.load(Ordering::Relaxed))
            .min()?;
        
        let candidates: Vec<Arc<Backend>> = warmed_up
            .into_iter()
            .filter(|b| b.current_connections.load(Ordering::Relaxed) == min_connections)
            .collect();
        
        let selected = if candidates.len() == 1 {
            candidates[0].clone()
        } else {
            candidates.choose(&mut rng)?.clone()
        };

        Some((selected, SelectionReason::LeastConnections))
    }

    fn select_random(
        &self,
        warmed_up: Vec<Arc<Backend>>,
        warming_up: Vec<Arc<Backend>>,
        _now: DateTime<Utc>,
    ) -> Option<(Arc<Backend>, SelectionReason)> {
        let mut rng = thread_rng();
        
        if !warming_up.is_empty() {
            let total_count = warmed_up.len() + warming_up.len();
            let warmup_max_traffic = (total_count as f64 * WARMUP_TRAFFIC_PERCENT) as usize;
            let warmup_count = warming_up.len().min(warmup_max_traffic.max(1));
            
            let select_warmup = rand::random::<f64>() < (warmup_count as f64 / total_count as f64);
            
            if select_warmup {
                let selected = warming_up.choose(&mut rng)?.clone();
                return Some((selected, SelectionReason::WarmupLimit));
            }
        }

        if warmed_up.is_empty() {
            return warming_up.choose(&mut rng).cloned().map(|b| (b, SelectionReason::WarmupLimit));
        }

        let selected = warmed_up.choose(&mut rng)?.clone();
        Some((selected, SelectionReason::Random))
    }

    pub fn acquire_connection(&self, backend: &Backend) {
        backend.current_connections.fetch_add(1, Ordering::Relaxed);
        backend.total_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn release_connection(&self, backend: &Backend) {
        backend.current_connections.fetch_sub(1, Ordering::Relaxed);
    }

    pub fn record_failure(&self, backend: &Backend) {
        backend.total_failures.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_allocation(
        &self,
        backend_address: String,
        strategy: LoadBalancingStrategy,
        reason: SelectionReason,
    ) {
        let record = AllocationRecord {
            timestamp: Utc::now(),
            backend_address,
            strategy,
            reason,
        };

        let mut records = self.allocation_records.write();
        records.push(record);
        if records.len() > MAX_ALLOCATION_RECORDS {
            records.remove(0);
        }
    }

    pub fn get_strategy(&self) -> LoadBalancingStrategy {
        self.strategy
    }

    pub fn update_backend_status(&self, address: &str, status: BackendStatus) {
        let backends = self.backends.read();
        if let Some(backend) = backends.get(address) {
            let mut current_status = backend.status.write();
            if *current_status != BackendStatus::Draining {
                *current_status = status;
            }
        }
    }

    pub fn get_backend(&self, address: &str) -> Option<Arc<Backend>> {
        let backends = self.backends.read();
        backends.get(address).cloned()
    }
}
