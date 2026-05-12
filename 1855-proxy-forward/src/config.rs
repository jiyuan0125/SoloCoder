use std::sync::Arc;
use std::sync::RwLock;
use std::time::Duration;

#[derive(Debug, Clone)]
pub struct Config {
    timeout: Arc<RwLock<Duration>>,
}

impl Config {
    pub fn from_env() -> Self {
        let default_timeout = Duration::from_secs(30);
        Config {
            timeout: Arc::new(RwLock::new(default_timeout)),
        }
    }

    pub fn get_timeout(&self) -> Duration {
        *self.timeout.read().unwrap()
    }

    pub fn set_timeout(&self, duration: Duration) {
        *self.timeout.write().unwrap() = duration;
    }
}
