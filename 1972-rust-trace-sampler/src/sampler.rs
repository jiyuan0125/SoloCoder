use crate::models::{SamplerConfig, SamplingStrategy, Span};
use crate::stats::StatsCollector;
use rand::Rng;
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::SystemTime;

pub enum SampleReason {
    PriorityError,
    PrioritySlow,
    StrategySampled,
}

pub enum SampleDecision {
    Keep(SampleReason),
    Drop(String),
}

pub struct SamplerEngine {
    adaptive_state: AdaptiveState,
}

struct AdaptiveState {
    last_update: AtomicU64,
    current_rate: parking_lot::RwLock<f64>,
    window_samples: AtomicU64,
}

impl Default for AdaptiveState {
    fn default() -> Self {
        AdaptiveState {
            last_update: AtomicU64::new(now_secs()),
            current_rate: parking_lot::RwLock::new(0.01),
            window_samples: AtomicU64::new(0),
        }
    }
}

fn now_secs() -> u64 {
    SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

impl SamplerEngine {
    pub fn new() -> Self {
        SamplerEngine {
            adaptive_state: AdaptiveState::default(),
        }
    }

    pub fn should_sample(
        &self,
        span: &Span,
        config: &SamplerConfig,
        stats: &StatsCollector,
    ) -> SampleDecision {
        if config.priority.enabled {
            if let Some(reason) = self.check_priority(span, config, stats) {
                return SampleDecision::Keep(reason);
            }
        }

        let rate = self.get_effective_rate(config, stats, span);
        let sampled = rand::thread_rng().gen::<f64>() < rate;

        if sampled {
            stats.record_strategy_sample();
            SampleDecision::Keep(SampleReason::StrategySampled)
        } else {
            SampleDecision::Drop(format!("strategy_rate={:.4}", rate))
        }
    }

    fn check_priority(
        &self,
        span: &Span,
        config: &SamplerConfig,
        stats: &StatsCollector,
    ) -> Option<SampleReason> {
        let is_error = match (span.http_status_code, span.error) {
            (Some(code), _) => code >= 500,
            (_, Some(true)) => true,
            _ => false,
        };

        if is_error {
            stats.record_priority_error();
            return Some(SampleReason::PriorityError);
        }

        if let Some(duration) = span.duration_ms {
            if duration >= config.priority.slow_request_p99_threshold_ms {
                stats.record_priority_slow();
                return Some(SampleReason::PrioritySlow);
            }
        }

        None
    }

    fn get_effective_rate(
        &self,
        config: &SamplerConfig,
        stats: &StatsCollector,
        span: &Span,
    ) -> f64 {
        match config.strategy {
            SamplingStrategy::FixedRate => {
                config.fixed_rate.rate.clamp(0.0, 1.0)
            }
            SamplingStrategy::Adaptive => {
                self.get_adaptive_rate(&config.adaptive, stats)
            }
            SamplingStrategy::ErrorBased => {
                self.get_error_based_rate(&config.error_based, stats, span)
            }
        }
    }

    fn get_adaptive_rate(
        &self,
        config: &crate::models::AdaptiveConfig,
        _stats: &StatsCollector,
    ) -> f64 {
        let now = now_secs();
        let last = self.adaptive_state.last_update.load(Ordering::Relaxed);

        if now - last >= 1 {
            let samples = self.adaptive_state.window_samples.swap(0, Ordering::Relaxed);
            let current_qps = samples.max(1);
            let mut rate = *self.adaptive_state.current_rate.read();

            let ratio = config.target_qps as f64 / current_qps as f64;
            if ratio > 1.05 {
                rate *= 1.0 + config.adjustment_factor;
            } else if ratio < 0.95 {
                rate *= 1.0 - config.adjustment_factor;
            }

            rate = rate.clamp(config.min_rate, config.max_rate);
            *self.adaptive_state.current_rate.write() = rate;
            self.adaptive_state.last_update.store(now, Ordering::Relaxed);
        }

        self.adaptive_state.window_samples.fetch_add(1, Ordering::Relaxed);
        *self.adaptive_state.current_rate.read()
    }

    fn get_error_based_rate(
        &self,
        config: &crate::models::ErrorBasedConfig,
        stats: &StatsCollector,
        span: &Span,
    ) -> f64 {
        let service_name = span.service_name.as_deref().unwrap_or("unknown");
        let service_stats = stats.get_service_stats(service_name);
        let error_rate = service_stats.error_rate;

        let mut rate = config.base_rate;
        if error_rate > config.error_rate_threshold {
            let excess = (error_rate - config.error_rate_threshold) / config.error_rate_threshold;
            let multiplier = 1.0 + excess * (config.max_multiplier - 1.0);
            rate = (config.base_rate * multiplier).clamp(config.base_rate, 1.0);
        }

        rate.clamp(0.0, 1.0)
    }

    pub fn get_current_rates(&self, config: &SamplerConfig) -> std::collections::HashMap<String, f64> {
        let mut rates = std::collections::HashMap::new();
        rates.insert("fixed_rate".to_string(), config.fixed_rate.rate);
        rates.insert("adaptive_current".to_string(), *self.adaptive_state.current_rate.read());
        rates.insert("adaptive_target_qps".to_string(), config.adaptive.target_qps as f64);
        rates.insert("error_based_base".to_string(), config.error_based.base_rate);
        rates
    }
}

impl Default for SamplerEngine {
    fn default() -> Self {
        Self::new()
    }
}
