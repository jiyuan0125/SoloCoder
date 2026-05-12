use serde_json::Value;
use crate::models::ResponseData;

pub struct DownstreamService;

impl DownstreamService {
    pub async fn execute(&self, payload: &Value) -> Result<ResponseData, String> {
        let result = serde_json::json!({
            "message": "Operation completed successfully",
            "payload": payload,
            "timestamp": chrono::Utc::now().to_rfc3339()
        });
        
        Ok(ResponseData {
            status_code: 200,
            body: result,
        })
    }
}

pub fn new_downstream_service() -> DownstreamService {
    DownstreamService
}
