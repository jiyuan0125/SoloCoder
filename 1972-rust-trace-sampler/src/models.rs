use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Span {
    pub trace_id: Option<String>,
    pub span_id: Option<String>,
    pub service_name: Option<String>,
    pub operation_name: Option<String>,
    pub duration_ms: Option<u64>,
    pub http_status_code: Option<u16>,
    pub error: Option<bool>,
    pub tags: Option<HashMap<String, serde_json::Value>>,
    #[serde(flatten)]
    pub extra: HashMap<String, serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum SamplingStrategy {
    FixedRate,
    Adaptive,
    ErrorBased,
}

impl Default for SamplingStrategy {
    fn default() -> Self {
        SamplingStrategy::FixedRate
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SamplerConfig {
    pub strategy: SamplingStrategy,
    pub fixed_rate: FixedRateConfig,
    pub adaptive: AdaptiveConfig,
    pub error_based: ErrorBasedConfig,
    pub priority: PriorityConfig,
}

impl Default for SamplerConfig {
    fn default() -> Self {
        SamplerConfig {
            strategy: SamplingStrategy::FixedRate,
            fixed_rate: FixedRateConfig::default(),
            adaptive: AdaptiveConfig::default(),
            error_based: ErrorBasedConfig::default(),
            priority: PriorityConfig::default(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FixedRateConfig {
    pub rate: f64,
}

impl Default for FixedRateConfig {
    fn default() -> Self {
        FixedRateConfig { rate: 0.01 }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AdaptiveConfig {
    pub target_qps: u64,
    pub min_rate: f64,
    pub max_rate: f64,
    pub adjustment_factor: f64,
}

impl Default for AdaptiveConfig {
    fn default() -> Self {
        AdaptiveConfig {
            target_qps: 100,
            min_rate: 0.001,
            max_rate: 1.0,
            adjustment_factor: 0.1,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorBasedConfig {
    pub base_rate: f64,
    pub error_rate_threshold: f64,
    pub max_multiplier: f64,
    pub window_seconds: u64,
}

impl Default for ErrorBasedConfig {
    fn default() -> Self {
        ErrorBasedConfig {
            base_rate: 0.01,
            error_rate_threshold: 0.05,
            max_multiplier: 10.0,
            window_seconds: 60,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PriorityConfig {
    pub enabled: bool,
    pub slow_request_p99_threshold_ms: u64,
}

impl Default for PriorityConfig {
    fn default() -> Self {
        PriorityConfig {
            enabled: true,
            slow_request_p99_threshold_ms: 500,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SamplerStats {
    pub total_received: u64,
    pub total_kept: u64,
    pub total_dropped: u64,
    pub effective_rate: f64,
    pub current_strategy: String,
    pub strategy_rates: HashMap<String, f64>,
    pub priority_sampled: PriorityStats,
    pub per_service: HashMap<String, ServiceStats>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct PriorityStats {
    pub error_requests: u64,
    pub slow_requests: u64,
    pub total_priority: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ServiceStats {
    pub total: u64,
    pub kept: u64,
    pub error_count: u64,
    pub error_rate: f64,
}
