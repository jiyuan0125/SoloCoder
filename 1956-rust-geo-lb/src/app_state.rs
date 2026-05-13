use crate::models::{BackendState, ZoneConfig, ZoneState};
use ipnet::IpNet;
use std::collections::HashMap;
use std::net::IpAddr;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::info;

#[derive(Debug)]
pub struct AppState {
    pub zones: RwLock<HashMap<String, Arc<ZoneState>>>,
    pub next_backend_index: AtomicUsize,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            zones: RwLock::new(HashMap::new()),
            next_backend_index: AtomicUsize::new(0),
        }
    }

    pub async fn add_zone(&self, config: ZoneConfig) -> Result<(), String> {
        let name = config.name.clone();

        let mut cidrs = Vec::new();
        for cidr_str in &config.cidrs {
            match cidr_str.parse::<IpNet>() {
                Ok(cidr) => cidrs.push(cidr),
                Err(e) => return Err(format!("Invalid CIDR '{}': {}", cidr_str, e)),
            }
        }

        let backends: Vec<BackendState> = config
            .backends
            .iter()
            .map(|addr| BackendState {
                address: addr.clone(),
                healthy: true,
                consecutive_failures: 0,
                last_checked: None,
            })
            .collect();

        let zone = Arc::new(ZoneState {
            name: name.clone(),
            cidrs,
            backends,
            health_interval_sec: config.health_interval_sec,
            request_count: std::sync::atomic::AtomicU64::new(0),
            total_latency_ms: std::sync::atomic::AtomicU64::new(0),
        });

        let mut zones = self.zones.write().await;
        zones.insert(name.clone(), zone);

        info!("Zone '{}' added with {} backends", name, config.backends.len());
        Ok(())
    }

    pub async fn find_zone_for_ip(&self, ip: IpAddr) -> Option<Arc<ZoneState>> {
        let zones = self.zones.read().await;

        for zone in zones.values() {
            if zone.contains_ip(ip) {
                return Some(zone.clone());
            }
        }

        None
    }

    pub async fn get_next_available_zones(&self, preferred_zone_name: &str) -> Vec<Arc<ZoneState>> {
        let zones = self.zones.read().await;

        let mut all_zones: Vec<Arc<ZoneState>> = zones.values().cloned().collect();

        all_zones.sort_by(|a, b| {
            let a_pref = if a.name == preferred_zone_name { 0 } else { 1 };
            let b_pref = if b.name == preferred_zone_name { 0 } else { 1 };
            a_pref.cmp(&b_pref)
        });

        all_zones
    }

    pub async fn get_all_zones(&self) -> Vec<Arc<ZoneState>> {
        let zones = self.zones.read().await;
        zones.values().cloned().collect()
    }

    pub fn get_next_backend_index(&self, count: usize) -> usize {
        if count == 0 {
            return 0;
        }
        self.next_backend_index.fetch_add(1, Ordering::SeqCst) % count
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}
