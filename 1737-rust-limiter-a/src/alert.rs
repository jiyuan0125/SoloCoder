use parking_lot::RwLock;
use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct AlertEvent {
    pub timestamp: String,
    pub limit_type: String,
    pub key: String,
    pub message: String,
}

#[derive(Debug, Default)]
pub struct AlertManager {
    webhook_url: RwLock<Option<String>>,
}

impl AlertManager {
    pub fn new() -> Self {
        AlertManager {
            webhook_url: RwLock::new(None),
        }
    }

    pub fn set_webhook(&self, url: String) {
        *self.webhook_url.write() = Some(url);
    }

    pub fn get_webhook(&self) -> Option<String> {
        self.webhook_url.read().clone()
    }

    pub fn send_alert(&self, event: AlertEvent) {
        let url = self.webhook_url.read().clone();
        if let Some(webhook_url) = url {
            let event_clone = event.clone();
            tokio::spawn(async move {
                let client = reqwest::Client::new();
                match client
                    .post(&webhook_url)
                    .json(&event_clone)
                    .send()
                    .await
                {
                    Ok(_) => log::info!("Alert sent successfully to {}", webhook_url),
                    Err(e) => log::error!("Failed to send alert: {}", e),
                }
            });
        }
    }
}
