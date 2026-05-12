use crate::summary::SummaryWindow;
use crate::types::{MetricKey, MetricSnapshot, MetricType, MetricValue};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};

pub const DEFAULT_SUMMARY_WINDOW_SIZE: usize = 1000;
pub const MAX_BATCH_SIZE: usize = 200;

#[derive(Debug)]
pub enum MetricEntry {
    Counter { value: f64 },
    Gauge { value: f64 },
    Summary { window: SummaryWindow },
}

impl MetricEntry {
    fn new_counter() -> Self {
        MetricEntry::Counter { value: 0.0 }
    }

    #[allow(dead_code)]
    fn new_gauge() -> Self {
        MetricEntry::Gauge { value: 0.0 }
    }

    fn new_summary(window_size: usize) -> Self {
        MetricEntry::Summary {
            window: SummaryWindow::new(window_size),
        }
    }

    fn metric_type(&self) -> MetricType {
        match self {
            MetricEntry::Counter { .. } => MetricType::Counter,
            MetricEntry::Gauge { .. } => MetricType::Gauge,
            MetricEntry::Summary { .. } => MetricType::Summary,
        }
    }

    fn snapshot_value(&self) -> MetricValue {
        match self {
            MetricEntry::Counter { value } => MetricValue::Counter { value: *value },
            MetricEntry::Gauge { value } => MetricValue::Gauge { value: *value },
            MetricEntry::Summary { window } => window.snapshot_value(),
        }
    }
}

#[derive(Clone, Debug)]
pub struct MetricsStore {
    inner: Arc<RwLock<HashMap<MetricKey, MetricEntry>>>,
    summary_window_size: usize,
}

impl MetricsStore {
    pub fn new() -> Self {
        Self::with_summary_window_size(DEFAULT_SUMMARY_WINDOW_SIZE)
    }

    pub fn with_summary_window_size(window_size: usize) -> Self {
        Self {
            inner: Arc::new(RwLock::new(HashMap::new())),
            summary_window_size: window_size,
        }
    }

    pub fn increment_counter(&self, key: MetricKey, delta: f64) {
        let mut map = self.inner.write().unwrap();
        map.entry(key)
            .and_modify(|entry| {
                if let MetricEntry::Counter { value } = entry {
                    *value += delta;
                }
            })
            .or_insert_with(|| {
                let mut entry = MetricEntry::new_counter();
                if let MetricEntry::Counter { value } = &mut entry {
                    *value = delta;
                }
                entry
            });
    }

    pub fn set_gauge(&self, key: MetricKey, value: f64) {
        let mut map = self.inner.write().unwrap();
        map.insert(key, MetricEntry::Gauge { value });
    }

    pub fn record_summary(&self, key: MetricKey, value: f64) {
        let mut map = self.inner.write().unwrap();
        let entry = map.entry(key).or_insert_with(|| {
            MetricEntry::new_summary(self.summary_window_size)
        });
        if let MetricEntry::Summary { window } = entry {
            window.push(value);
        }
    }

    pub fn snapshot(&self) -> Vec<MetricSnapshot> {
        let map = self.inner.read().unwrap();
        map.iter()
            .map(|(key, entry)| MetricSnapshot {
                name: key.name.clone(),
                metric_type: entry.metric_type(),
                value: entry.snapshot_value(),
                labels: key.labels.clone(),
            })
            .collect()
    }
}
