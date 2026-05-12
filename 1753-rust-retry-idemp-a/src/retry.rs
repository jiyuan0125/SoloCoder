use std::time::Duration;
use chrono::Utc;
use crate::models::{RequestStatus, RetryHistory};

pub struct RetryPolicy {
    initial_delay: Duration,
    max_delay: Duration,
    max_retries: u32,
}

impl RetryPolicy {
    pub fn new(initial_delay: Duration, max_delay: Duration, max_retries: u32) -> Self {
        Self {
            initial_delay,
            max_delay,
            max_retries,
        }
    }
    
    pub fn default_policy() -> Self {
        Self {
            initial_delay: Duration::from_secs(1),
            max_delay: Duration::from_secs(30),
            max_retries: 5,
        }
    }
    
    pub fn get_delay(&self, attempt: u32) -> Duration {
        let base = self.initial_delay.as_secs();
        let delay = base * 2u64.saturating_pow(attempt);
        let delay_duration = Duration::from_secs(delay);
        std::cmp::min(delay_duration, self.max_delay)
    }
    
    pub fn should_retry(&self, attempt: u32) -> bool {
        attempt < self.max_retries
    }
    
    pub fn max_retries(&self) -> u32 {
        self.max_retries
    }
}

pub struct StateMachine;

impl StateMachine {
    pub fn can_transition(from: RequestStatus, to: RequestStatus) -> bool {
        match (from, to) {
            (RequestStatus::Initial, RequestStatus::Retrying) => true,
            (RequestStatus::Initial, RequestStatus::Success) => true,
            (RequestStatus::Initial, RequestStatus::Failed) => true,
            (RequestStatus::Retrying, RequestStatus::Success) => true,
            (RequestStatus::Retrying, RequestStatus::Failed) => true,
            _ => false,
        }
    }
}

pub fn create_retry_history(
    attempt: u32,
    duration_ms: u64,
    success: bool,
    error_message: Option<String>,
) -> RetryHistory {
    RetryHistory {
        attempt,
        timestamp: Utc::now(),
        duration_ms,
        success,
        error_message,
    }
}
