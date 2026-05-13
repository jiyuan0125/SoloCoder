use tokio::sync::mpsc;
use crate::models::StateChangeNotification;

#[derive(Clone)]
pub struct Notifier {
    sender: mpsc::Sender<NotificationTask>,
}

struct NotificationTask {
    notification: StateChangeNotification,
    notify_url: Option<String>,
}

impl Notifier {
    pub fn new() -> Self {
        let (sender, mut receiver) = mpsc::channel::<NotificationTask>(100);

        tokio::spawn(async move {
            while let Some(task) = receiver.recv().await {
                Self::send_with_retry(task).await;
            }
        });

        Notifier { sender }
    }

    pub fn send(&self, notification: StateChangeNotification, notify_url: Option<String>) {
        let task = NotificationTask {
            notification,
            notify_url,
        };
        
        if let Err(e) = self.sender.try_send(task) {
            log::error!("Failed to queue notification: {}", e);
        }
    }

    async fn send_with_retry(task: NotificationTask) {
        let url = match task.notify_url {
            Some(url) => url,
            None => {
                log::debug!("No notify URL configured for service: {}", task.notification.service_name);
                return;
            }
        };

        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(5))
            .build()
            .expect("Failed to create HTTP client");

        let max_retries = 2;
        let mut last_error = None;

        for attempt in 0..max_retries {
            log::debug!(
                "Sending state change notification (attempt {}/{}): {} -> {} for service {}",
                attempt + 1,
                max_retries,
                task.notification.old_state.as_str(),
                task.notification.new_state.as_str(),
                task.notification.service_name
            );

            match client.post(&url).json(&task.notification).send().await {
                Ok(response) => {
                    if response.status().is_success() {
                        log::info!(
                            "Notification sent successfully for service: {}",
                            task.notification.service_name
                        );
                        return;
                    } else {
                        let status = response.status();
                        last_error = Some(format!("HTTP status: {}", status));
                    }
                }
                Err(e) => {
                    last_error = Some(e.to_string());
                }
            }

            if attempt < max_retries - 1 {
                tokio::time::sleep(std::time::Duration::from_secs(1)).await;
            }
        }

        log::error!(
            "Failed to send notification after {} retries for service {}: {:?}",
            max_retries,
            task.notification.service_name,
            last_error
        );
    }
}
