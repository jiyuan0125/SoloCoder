use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum MetricType {
    Counter,
    Gauge,
    Summary,
}

impl fmt::Display for MetricType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            MetricType::Counter => write!(f, "counter"),
            MetricType::Gauge => write!(f, "gauge"),
            MetricType::Summary => write!(f, "summary"),
        }
    }
}

pub type Labels = BTreeMap<String, String>;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct MetricKey {
    pub name: String,
    pub labels: Labels,
}

impl MetricKey {
    pub fn new(name: String, labels: Labels) -> Self {
        Self { name, labels }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricDataPoint {
    pub name: String,
    #[serde(rename = "type")]
    pub metric_type: MetricType,
    pub value: f64,
    #[serde(default)]
    pub labels: Labels,
}

#[derive(Debug, Clone, Serialize)]
#[serde(tag = "type")]
pub enum MetricValue {
    Counter { value: f64 },
    Gauge { value: f64 },
    Summary {
        count: usize,
        min: f64,
        max: f64,
        avg: f64,
        p50: Option<f64>,
        p90: Option<f64>,
        p99: Option<f64>,
        #[serde(skip_serializing_if = "Option::is_none")]
        error: Option<&'static str>,
    },
}

#[derive(Debug, Clone, Serialize)]
pub struct MetricSnapshot {
    pub name: String,
    #[serde(rename = "type")]
    pub metric_type: MetricType,
    pub value: MetricValue,
    pub labels: Labels,
}
