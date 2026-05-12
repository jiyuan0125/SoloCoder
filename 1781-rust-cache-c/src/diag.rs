use chrono::{DateTime, Duration, Utc};
use std::collections::VecDeque;

#[allow(dead_code)]
#[derive(Debug, Clone)]
pub struct EvictionFailureLog {
    pub timestamp: DateTime<Utc>,
    pub message: String,
}

pub struct DiagLogger {
    eviction_failures: VecDeque<EvictionFailureLog>,
}

impl DiagLogger {
    pub fn new() -> Self {
        DiagLogger {
            eviction_failures: VecDeque::new(),
        }
    }

    pub fn log_eviction_failure(&mut self, message: &str) {
        let entry = EvictionFailureLog {
            timestamp: Utc::now(),
            message: message.to_string(),
        };
        self.eviction_failures.push_back(entry);
        tracing::warn!("Eviction failure logged: {}", message);
    }

    pub fn cleanup_old_logs(&mut self, days: i64) -> usize {
        let cutoff = Utc::now() - Duration::days(days);
        let mut removed = 0;

        while let Some(front) = self.eviction_failures.front() {
            if front.timestamp < cutoff {
                self.eviction_failures.pop_front();
                removed += 1;
            } else {
                break;
            }
        }

        if removed > 0 {
            tracing::info!("Cleaned {} old eviction failure logs", removed);
        }

        removed
    }

}
