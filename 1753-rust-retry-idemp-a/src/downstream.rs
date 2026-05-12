use reqwest::Client;
use serde_json::Value;
use crate::models::ResponseData;

pub struct DownstreamService {
    client: Client,
}

impl DownstreamService {
    pub fn new() -> Self {
        Self {
            client: Client::builder()
                .timeout(std::time::Duration::from_secs(10))
                .build()
                .expect("Failed to create HTTP client"),
        }
    }

    pub async fn execute(&self, callback_url: &str, payload: &Value) -> Result<ResponseData, String> {
        tracing::info!("Calling downstream service: {}", callback_url);
        
        let response = self.client
            .post(callback_url)
            .json(payload)
            .send()
            .await
            .map_err(|e| format!("HTTP request failed: {}", e))?;

        let status_code = response.status().as_u16();
        let body: Value = response
            .json()
            .await
            .unwrap_or_else(|_| serde_json::json!({ "message": "No response body" }));

        if status_code >= 500 {
            return Err(format!("Downstream service returned error status: {}", status_code));
        }

        Ok(ResponseData {
            status_code,
            body,
        })
    }
}

pub fn new_downstream_service() -> DownstreamService {
    DownstreamService::new()
}
