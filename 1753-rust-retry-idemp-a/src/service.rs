use std::sync::Arc;
use uuid::Uuid;
use chrono::Utc;
use crate::idempotency::IdempotencyStore;
use crate::models::{
    RequestRecord, RequestStatus, ResponseData, 
    ApiResponse, QueryResponse, Statistics, RetryHistory
};
use crate::retry::{RetryPolicy, create_retry_history};
use crate::downstream::DownstreamService;

pub struct RetryIdempotentService {
    idempotency_store: IdempotencyStore,
    retry_policy: RetryPolicy,
    downstream: DownstreamService,
}

impl RetryIdempotentService {
    pub fn new(
        idempotency_store: IdempotencyStore,
        retry_policy: RetryPolicy,
        downstream: DownstreamService,
    ) -> Self {
        Self {
            idempotency_store,
            retry_policy,
            downstream,
        }
    }

    pub async fn process_request(
        &self,
        idempotency_key: &str,
        payload: &serde_json::Value,
    ) -> ApiResponse {
        if let Some(record) = self.idempotency_store.get(idempotency_key) {
            if matches!(record.status, RequestStatus::Success | RequestStatus::Failed) {
                return ApiResponse {
                    request_id: record.request_id,
                    idempotency_key: idempotency_key.to_string(),
                    status: record.status,
                    from_cache: true,
                    response: record.response,
                    retry_count: record.total_retries,
                };
            }
        }

        let record = RequestRecord {
            idempotency_key: idempotency_key.to_string(),
            request_id: Uuid::new_v4(),
            status: RequestStatus::Initial,
            created_at: Utc::now(),
            completed_at: None,
            retry_history: Vec::new(),
            total_retries: 0,
            response: None,
            expires_at: Utc::now() + chrono::Duration::hours(24),
        };
        self.idempotency_store.insert(record.clone());

        let result = self.execute_with_retry(idempotency_key, payload).await;

        match result {
            Ok(response) => {
                self.idempotency_store.update_response(
                    idempotency_key,
                    response.clone(),
                    RequestStatus::Success,
                );
                
                let final_record = self.idempotency_store.get(idempotency_key).unwrap();
                ApiResponse {
                    request_id: final_record.request_id,
                    idempotency_key: idempotency_key.to_string(),
                    status: RequestStatus::Success,
                    from_cache: false,
                    response: Some(response),
                    retry_count: final_record.total_retries,
                }
            }
            Err(e) => {
                self.idempotency_store.update_status(idempotency_key, RequestStatus::Failed);
                
                let final_record = self.idempotency_store.get(idempotency_key).unwrap();
                ApiResponse {
                    request_id: final_record.request_id,
                    idempotency_key: idempotency_key.to_string(),
                    status: RequestStatus::Failed,
                    from_cache: false,
                    response: Some(ResponseData {
                        status_code: 500,
                        body: serde_json::json!({ "error": e }),
                    }),
                    retry_count: final_record.total_retries,
                }
            }
        }
    }

    async fn execute_with_retry(
        &self,
        idempotency_key: &str,
        payload: &serde_json::Value,
    ) -> Result<ResponseData, String> {
        let mut attempt: u32 = 0;
        let mut last_error: Option<String>;

        loop {
            if attempt > 0 {
                self.idempotency_store.update_status(idempotency_key, RequestStatus::Retrying);
            }

            let start = std::time::Instant::now();
            let result = self.downstream.execute(payload).await;
            let duration = start.elapsed();

            match result {
                Ok(response) => {
                    let history = create_retry_history(
                        attempt,
                        duration.as_millis() as u64,
                        true,
                        None,
                    );
                    self.add_retry_history(idempotency_key, history);
                    return Ok(response);
                }
                Err(e) => {
                    last_error = Some(e.clone());
                    let history = create_retry_history(
                        attempt,
                        duration.as_millis() as u64,
                        false,
                        Some(e),
                    );
                    self.add_retry_history(idempotency_key, history);

                    if !self.retry_policy.should_retry(attempt) {
                        return Err(format!("All retries failed: {}", last_error.unwrap_or_else(|| "unknown error".to_string())));
                    }

                    attempt += 1;
                    self.idempotency_store.increment_retry(idempotency_key);
                    
                    let delay = self.retry_policy.get_delay(attempt);
                    tokio::time::sleep(delay).await;
                }
            }
        }
    }

    fn add_retry_history(&self, idempotency_key: &str, history: RetryHistory) {
        self.idempotency_store.add_retry_history(idempotency_key, history);
    }

    pub fn query_request(&self, idempotency_key: &str) -> Option<QueryResponse> {
        self.idempotency_store.get(idempotency_key).map(|record| QueryResponse {
            request: record,
        })
    }

    pub fn get_statistics(&self) -> Statistics {
        let records = self.idempotency_store.get_all();
        
        let total_requests = records.len() as u64;
        let success_requests = records.iter()
            .filter(|r| matches!(r.status, RequestStatus::Success))
            .count() as u64;
        let total_with_retry = records.iter()
            .filter(|r| r.total_retries > 0)
            .count() as u64;
        let retry_successful = records.iter()
            .filter(|r| r.total_retries > 0 && matches!(r.status, RequestStatus::Success))
            .count() as u64;
        
        let total_retries_sum: u64 = records.iter()
            .map(|r| r.total_retries as u64)
            .sum();
        
        let average_retries = if total_requests > 0 {
            total_retries_sum as f64 / total_requests as f64
        } else {
            0.0
        };
        
        let retry_success_rate = if total_with_retry > 0 {
            retry_successful as f64 / total_with_retry as f64
        } else {
            0.0
        };
        
        Statistics {
            total_requests,
            total_with_retry,
            success_requests,
            retry_success_rate,
            average_retries,
        }
    }
}

pub fn new_service(
    idempotency_store: IdempotencyStore,
    retry_policy: RetryPolicy,
    downstream: DownstreamService,
) -> Arc<RetryIdempotentService> {
    Arc::new(RetryIdempotentService::new(idempotency_store, retry_policy, downstream))
}
