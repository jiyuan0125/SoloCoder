use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde::Deserialize;

use crate::store::{MAX_BATCH_SIZE, MetricsStore};
use crate::types::{MetricDataPoint, MetricKey, MetricType};

#[derive(Debug, Deserialize)]
pub struct BatchMetricsRequest {
    pub metrics: Vec<MetricDataPoint>,
}

#[derive(Debug, serde::Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub struct AppError {
    pub code: StatusCode,
    pub message: String,
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        (
            self.code,
            Json(ErrorResponse {
                error: self.message,
            }),
        )
            .into_response()
    }
}

fn validate_labels(point: &MetricDataPoint) -> Result<(), AppError> {
    for (key, _) in point.labels.iter() {
        if key.is_empty() {
            return Err(AppError {
                code: StatusCode::BAD_REQUEST,
                message: "label key cannot be empty".to_string(),
            });
        }
    }
    Ok(())
}

pub async fn ingest_metrics(
    State(store): State<MetricsStore>,
    Json(batch): Json<BatchMetricsRequest>,
) -> Result<impl IntoResponse, AppError> {
    if batch.metrics.len() > MAX_BATCH_SIZE {
        return Err(AppError {
            code: StatusCode::BAD_REQUEST,
            message: format!("batch size exceeds maximum of {}", MAX_BATCH_SIZE),
        });
    }

    for point in batch.metrics.iter() {
        validate_labels(point)?;
    }

    for point in batch.metrics.into_iter() {
        let key = MetricKey::new(point.name, point.labels);
        match point.metric_type {
            MetricType::Counter => store.increment_counter(key, point.value),
            MetricType::Gauge => store.set_gauge(key, point.value),
            MetricType::Summary => store.record_summary(key, point.value),
        }
    }

    Ok(StatusCode::OK)
}

pub async fn query_metrics(
    State(store): State<MetricsStore>,
) -> impl IntoResponse {
    let snapshot = store.snapshot();
    Json(snapshot)
}
