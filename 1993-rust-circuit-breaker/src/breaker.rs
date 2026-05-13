use std::collections::{HashMap, VecDeque};
use std::sync::{Arc, RwLock, Mutex};
use chrono::{DateTime, Utc};
use crate::models::{CircuitConfig, CircuitState, CircuitStatus, RequestRecord, WindowStats, StateChangeNotification};

type NotifyCallback = Arc<dyn Fn(&str, StateChangeNotification) + Send + Sync + 'static>;

pub struct CircuitBreaker {
    pub service_name: String,
    state: RwLock<CircuitState>,
    config: RwLock<CircuitConfig>,
    request_window: Mutex<VecDeque<bool>>,
    recent_requests: Mutex<VecDeque<RequestRecord>>,
    opened_at: RwLock<Option<DateTime<Utc>>>,
    last_state_change: RwLock<DateTime<Utc>>,
    half_open_in_progress: Mutex<bool>,
    notify_callback: Option<NotifyCallback>,
}

impl CircuitBreaker {
    pub fn new(service_name: String, config: CircuitConfig) -> Self {
        Self {
            service_name,
            state: RwLock::new(CircuitState::Closed),
            config: RwLock::new(config),
            request_window: Mutex::new(VecDeque::new()),
            recent_requests: Mutex::new(VecDeque::new()),
            opened_at: RwLock::new(None),
            last_state_change: RwLock::new(Utc::now()),
            half_open_in_progress: Mutex::new(false),
            notify_callback: None,
        }
    }

    pub fn with_callback(
        service_name: String,
        config: CircuitConfig,
        callback: NotifyCallback,
    ) -> Self {
        Self {
            service_name,
            state: RwLock::new(CircuitState::Closed),
            config: RwLock::new(config),
            request_window: Mutex::new(VecDeque::new()),
            recent_requests: Mutex::new(VecDeque::new()),
            opened_at: RwLock::new(None),
            last_state_change: RwLock::new(Utc::now()),
            half_open_in_progress: Mutex::new(false),
            notify_callback: Some(callback),
        }
    }

    pub fn get_state(&self) -> CircuitState {
        *self.state.read().unwrap()
    }

    pub fn get_config(&self) -> CircuitConfig {
        self.config.read().unwrap().clone()
    }

    pub fn update_config(&self, updates: &CircuitConfigUpdate) {
        let mut config = self.config.write().unwrap();
        if let Some(threshold) = updates.failure_threshold {
            config.failure_threshold = threshold;
        }
        if let Some(window_size) = updates.window_size {
            config.window_size = window_size;
        }
        if let Some(cooling_period) = updates.cooling_period_secs {
            config.cooling_period_secs = cooling_period;
        }
        if let Some(notify_url) = &updates.notify_url {
            config.notify_url = notify_url.clone();
        }
        if let Some(target_url) = &updates.target_url {
            config.target_url = target_url.clone();
        }
    }

    pub fn get_status(&self) -> CircuitStatus {
        let state = self.get_state();
        let config = self.get_config();
        let window = self.request_window.lock().unwrap();
        let recent = self.recent_requests.lock().unwrap();
        let opened_at = *self.opened_at.read().unwrap();
        let last_state_change = *self.last_state_change.read().unwrap();

        let (successes, failures, total) = window.iter().fold(
            (0usize, 0usize, 0usize),
            |(s, f, t), &success| {
                (
                    s + if success { 1 } else { 0 },
                    f + if success { 0 } else { 1 },
                    t + 1,
                )
            },
        );

        let failure_rate = if total > 0 {
            Some(failures as f64 / total as f64)
        } else {
            None
        };

        let mut recent_reqs: Vec<RequestRecord> = recent.iter().cloned().collect();
        recent_reqs.reverse();

        CircuitStatus {
            service_name: self.service_name.clone(),
            state,
            config,
            current_window: WindowStats {
                successes,
                failures,
                total,
                failure_rate,
            },
            recent_requests: recent_reqs,
            opened_at,
            last_state_change,
        }
    }

    fn change_state(&self, new_state: CircuitState, reason: String) {
        let mut state = self.state.write().unwrap();
        let old_state = *state;

        if old_state == new_state {
            return;
        }

        *state = new_state;
        drop(state);

        let mut last_change = self.last_state_change.write().unwrap();
        *last_change = Utc::now();
        drop(last_change);

        if new_state == CircuitState::Open {
            let mut opened = self.opened_at.write().unwrap();
            *opened = Some(Utc::now());
        } else if new_state == CircuitState::Closed {
            let mut opened = self.opened_at.write().unwrap();
            *opened = None;
        }

        if let Some(callback) = &self.notify_callback {
            let notification = StateChangeNotification {
                service_name: self.service_name.clone(),
                old_state,
                new_state,
                reason,
                timestamp: Utc::now(),
            };
            let notify_url = self.get_config().notify_url;
            let url_str = notify_url.unwrap_or_default();
            callback(&url_str, notification);
        }
    }

    fn record_result(&self, success: bool) {
        let config = self.get_config();
        let window_size = config.window_size;

        let mut window = self.request_window.lock().unwrap();
        window.push_back(success);
        while window.len() > window_size {
            window.pop_front();
        }
        drop(window);

        let mut recent = self.recent_requests.lock().unwrap();
        recent.push_back(RequestRecord {
            timestamp: Utc::now(),
            success,
        });
        while recent.len() > 100 {
            recent.pop_front();
        }
    }

    fn reset_counter(&self) {
        let mut window = self.request_window.lock().unwrap();
        window.clear();
    }

    fn check_failure_threshold(&self) -> bool {
        let config = self.get_config();
        let window = self.request_window.lock().unwrap();

        if window.len() < config.window_size {
            return false;
        }

        let failures = window.iter().filter(|&&s| !s).count();
        let failure_rate = failures as f64 / window.len() as f64;
        failure_rate >= config.failure_threshold
    }

    fn check_cooling_period_elapsed(&self) -> bool {
        let config = self.get_config();
        let opened_at = *self.opened_at.read().unwrap();

        if let Some(opened) = opened_at {
            let elapsed = Utc::now().signed_duration_since(opened);
            elapsed.num_seconds() >= config.cooling_period_secs as i64
        } else {
            false
        }
    }

    pub fn try_acquire(&self) -> bool {
        let state = self.get_state();

        match state {
            CircuitState::Closed => true,
            CircuitState::Open => {
                if self.check_cooling_period_elapsed() {
                    self.change_state(
                        CircuitState::HalfOpen,
                        "Cooling period elapsed, entering half-open state".to_string(),
                    );
                    let mut in_progress = self.half_open_in_progress.lock().unwrap();
                    if !*in_progress {
                        *in_progress = true;
                        true
                    } else {
                        false
                    }
                } else {
                    false
                }
            }
            CircuitState::HalfOpen => {
                let mut in_progress = self.half_open_in_progress.lock().unwrap();
                if !*in_progress {
                    *in_progress = true;
                    true
                } else {
                    false
                }
            }
        }
    }

    pub fn on_success(&self) {
        let state = self.get_state();

        match state {
            CircuitState::Closed => {
                self.record_result(true);
            }
            CircuitState::HalfOpen => {
                self.record_result(true);
                self.reset_counter();
                let mut in_progress = self.half_open_in_progress.lock().unwrap();
                *in_progress = false;
                self.change_state(
                    CircuitState::Closed,
                    "Half-open probe succeeded, closing circuit".to_string(),
                );
            }
            CircuitState::Open => {}
        }
    }

    pub fn on_failure(&self) {
        let state = self.get_state();

        match state {
            CircuitState::Closed => {
                self.record_result(false);
                if self.check_failure_threshold() {
                    self.change_state(
                        CircuitState::Open,
                        "Failure threshold exceeded".to_string(),
                    );
                }
            }
            CircuitState::HalfOpen => {
                self.record_result(false);
                let mut in_progress = self.half_open_in_progress.lock().unwrap();
                *in_progress = false;
                self.change_state(
                    CircuitState::Open,
                    "Half-open probe failed, reopening circuit".to_string(),
                );
            }
            CircuitState::Open => {}
        }
    }

    pub fn force_open(&self) {
        self.change_state(
            CircuitState::Open,
            "Manually forced open".to_string(),
        );
    }

    pub fn force_close(&self) {
        self.reset_counter();
        self.change_state(
            CircuitState::Closed,
            "Manually forced close".to_string(),
        );
    }

    pub fn force_close_for_deletion(&self) {
        let mut state = self.state.write().unwrap();
        *state = CircuitState::Closed;
    }
}

pub struct CircuitConfigUpdate {
    pub failure_threshold: Option<f64>,
    pub window_size: Option<usize>,
    pub cooling_period_secs: Option<u64>,
    pub notify_url: Option<Option<String>>,
    pub target_url: Option<Option<String>>,
}

pub struct CircuitBreakerRegistry {
    breakers: Mutex<HashMap<String, Arc<CircuitBreaker>>>,
    notifier: crate::notifier::Notifier,
}

impl CircuitBreakerRegistry {
    pub fn new() -> Self {
        Self {
            breakers: Mutex::new(HashMap::new()),
            notifier: crate::notifier::Notifier::new(),
        }
    }

    pub fn create(
        &self,
        service_name: String,
        config: CircuitConfig,
    ) -> Option<Arc<CircuitBreaker>> {
        let mut breakers = self.breakers.lock().unwrap();

        if breakers.contains_key(&service_name) {
            return None;
        }

        let notifier = self.notifier.clone();
        let breaker = Arc::new(CircuitBreaker::with_callback(
            service_name.clone(),
            config,
            Arc::new(move |url: &str, notification| {
                let notify_url = if url.is_empty() { None } else { Some(url.to_string()) };
                notifier.send(notification, notify_url);
            }),
        ));

        breakers.insert(service_name, breaker.clone());
        Some(breaker)
    }

    pub fn get(&self, service_name: &str) -> Option<Arc<CircuitBreaker>> {
        let breakers = self.breakers.lock().unwrap();
        breakers.get(service_name).cloned()
    }

    pub fn list(&self) -> Vec<Arc<CircuitBreaker>> {
        let breakers = self.breakers.lock().unwrap();
        breakers.values().cloned().collect()
    }

    pub fn delete(&self, service_name: &str) -> bool {
        let mut breakers = self.breakers.lock().unwrap();

        if let Some(breaker) = breakers.get(service_name) {
            breaker.force_close_for_deletion();
            breakers.remove(service_name);
            true
        } else {
            false
        }
    }
}
