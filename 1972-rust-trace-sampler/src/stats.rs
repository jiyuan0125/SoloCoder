use crate::models::{SamplerConfig, SamplerStats, ServiceStats};
use crate::sampler::SamplerEngine;
use std::collections::HashMap;
use std::sync::atomic::{AtomicU64, Ordering};
use parking_lot::RwLock;

struct ServiceData {
    total: AtomicU64,
    kept: AtomicU64,
    errors: AtomicU64,
}

impl Default for ServiceData {
    fn default() -> Self {
        ServiceData {
            total: AtomicU64::new(0),
            kept: AtomicU64::new(0),
            errors: AtomicU64::new(0),
        }
    }
}

pub struct StatsCollector {
    total_received: AtomicU64,
    total_kept: AtomicU64,
    total_dropped: AtomicU64,
    priority_errors: AtomicU64,
    priority_slow: AtomicU64,
    strategy_samples: AtomicU64,
    services: RwLock<HashMap<String, ServiceData>>,
}

impl StatsCollector {
    pub fn new() -> Self {
        StatsCollector {
            total_received: AtomicU64::new(0),
            total_kept: AtomicU64::new(0),
            total_dropped: AtomicU64::new(0),
            priority_errors: AtomicU64::new(0),
            priority_slow: AtomicU64::new(0),
            strategy_samples: AtomicU64::new(0),
            services: RwLock::new(HashMap::new()),
        }
    }

    pub fn record_received(&self) {
        self.total_received.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_kept(&self) {
        self.total_kept.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_dropped(&self) {
        self.total_dropped.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_priority_error(&self) {
        self.priority_errors.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_priority_slow(&self) {
        self.priority_slow.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_strategy_sample(&self) {
        self.strategy_samples.fetch_add(1, Ordering::Relaxed);
    }

    pub fn record_service(&self, service: &str, kept: bool, is_error: bool) {
        let services = self.services.read();
        if let Some(data) = services.get(service) {
            data.total.fetch_add(1, Ordering::Relaxed);
            if kept {
                data.kept.fetch_add(1, Ordering::Relaxed);
            }
            if is_error {
                data.errors.fetch_add(1, Ordering::Relaxed);
            }
            return;
        }
        drop(services);

        let mut services = self.services.write();
        let data = services.entry(service.to_string()).or_default();
        data.total.fetch_add(1, Ordering::Relaxed);
        if kept {
            data.kept.fetch_add(1, Ordering::Relaxed);
        }
        if is_error {
            data.errors.fetch_add(1, Ordering::Relaxed);
        }
    }

    pub fn get_service_stats(&self, service: &str) -> ServiceStats {
        let services = self.services.read();
        match services.get(service) {
            Some(data) => {
                let total = data.total.load(Ordering::Relaxed);
                let errors = data.errors.load(Ordering::Relaxed);
                ServiceStats {
                    total,
                    kept: data.kept.load(Ordering::Relaxed),
                    error_count: errors,
                    error_rate: if total > 0 { errors as f64 / total as f64 } else { 0.0 },
                }
            }
            None => ServiceStats::default(),
        }
    }

    pub fn get_stats(&self, engine: &SamplerEngine, config: &SamplerConfig) -> SamplerStats {
        let total_received = self.total_received.load(Ordering::Relaxed);
        let total_kept = self.total_kept.load(Ordering::Relaxed);
        let total_dropped = self.total_dropped.load(Ordering::Relaxed);

        let effective_rate = if total_received > 0 {
            total_kept as f64 / total_received as f64
        } else {
            0.0
        };

        let priority_errors = self.priority_errors.load(Ordering::Relaxed);
        let priority_slow = self.priority_slow.load(Ordering::Relaxed);

        let per_service: HashMap<String, ServiceStats> = {
            let services = self.services.read();
            services
                .iter()
                .map(|(name, data)| {
                    let total = data.total.load(Ordering::Relaxed);
                    let errors = data.errors.load(Ordering::Relaxed);
                    (
                        name.clone(),
                        ServiceStats {
                            total,
                            kept: data.kept.load(Ordering::Relaxed),
                            error_count: errors,
                            error_rate: if total > 0 { errors as f64 / total as f64 } else { 0.0 },
                        },
                    )
                })
                .collect()
        };

        SamplerStats {
            total_received,
            total_kept,
            total_dropped,
            effective_rate,
            current_strategy: format!("{:?}", config.strategy),
            strategy_rates: engine.get_current_rates(config),
            priority_sampled: crate::models::PriorityStats {
                error_requests: priority_errors,
                slow_requests: priority_slow,
                total_priority: priority_errors + priority_slow,
            },
            per_service,
        }
    }
}

impl Default for StatsCollector {
    fn default() -> Self {
        Self::new()
    }
}
