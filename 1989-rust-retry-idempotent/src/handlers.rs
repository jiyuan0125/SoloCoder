use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use serde::Deserialize;
use uuid::Uuid;
use chrono::Utc;

use crate::models::{
    ByStatusStats, RetryDistribution, StatsResponse, Task, TaskDetail,
    TaskRequest, TaskResponse, TaskStatus,
};
use crate::scheduler::Scheduler;
use crate::store::TaskStore;

#[derive(Clone)]
pub struct AppState {
    pub store: TaskStore,
}

#[derive(Debug, Deserialize)]
pub struct QueryParams {
    pub status: Option<TaskStatus>,
    pub idempotency_key: Option<String>,
}

pub async fn submit_task(
    State(state): State<AppState>,
    Json(req): Json<TaskRequest>,
) -> impl IntoResponse {
    let idempotency_key = req.idempotency_key.clone().unwrap_or_else(|| {
        Uuid::new_v4().to_string()
    });

    if let Some(existing) = state.store.get_by_idempotency(&idempotency_key) {
        return (
            StatusCode::OK,
            Json(TaskResponse {
                task_id: existing.task_id,
                idempotency_key: existing.idempotency_key,
                status: existing.status,
            }),
        );
    }

    let url_valid = Scheduler::validate_url(&req.url);
    let mut task = Task::new(req, idempotency_key);

    if !url_valid {
        task.status = TaskStatus::Failed;
        task.last_error = Some("Invalid URL".to_string());
        task.completed_at = Some(Utc::now());
        let task_id = state.store.insert_terminal(task.clone());
        return (
            StatusCode::OK,
            Json(TaskResponse {
                task_id,
                idempotency_key: task.idempotency_key,
                status: task.status,
            }),
        );
    }

    let task_id = state.store.insert(task.clone());

    (
        StatusCode::ACCEPTED,
        Json(TaskResponse {
            task_id,
            idempotency_key: task.idempotency_key,
            status: task.status,
        }),
    )
}

pub async fn get_task(
    State(state): State<AppState>,
    Path(task_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.store.get(&task_id) {
        Some(task) => (StatusCode::OK, Json(task.to_detail())).into_response(),
        None => (StatusCode::NOT_FOUND, "Task not found").into_response(),
    }
}

pub async fn list_tasks(
    State(state): State<AppState>,
    Query(params): Query<QueryParams>,
) -> impl IntoResponse {
    let all_tasks = state.store.all_tasks();
    let filtered: Vec<TaskDetail> = all_tasks
        .into_iter()
        .filter(|t| {
            if let Some(status) = params.status {
                if t.status != status {
                    return false;
                }
            }
            if let Some(key) = &params.idempotency_key {
                if &t.idempotency_key != key {
                    return false;
                }
            }
            true
        })
        .map(|t| t.to_detail())
        .collect();
    (StatusCode::OK, Json(filtered))
}

pub async fn get_stats(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let (pending, running, retrying, succeeded, failed) = state.store.count_by_status();
    let all_tasks = state.store.all_tasks();

    let total_tasks = all_tasks.len() as u64;

    let mut total_execution_ms = 0i64;
    let mut completed_count = 0u64;
    let mut retry_dist = RetryDistribution::default();

    for task in &all_tasks {
        if task.is_terminal() {
            if let (Some(started), Some(completed)) = (task.started_at, task.completed_at) {
                let duration = completed.signed_duration_since(started);
                total_execution_ms += duration.num_milliseconds();
                completed_count += 1;
            }
        }
        let attempts = task.current_attempt;
        match attempts {
            0 => retry_dist.zero += 1,
            1 => retry_dist.one += 1,
            2 => retry_dist.two += 1,
            3 => retry_dist.three += 1,
            _ => retry_dist.four_plus += 1,
        }
    }

    let avg_execution_time_ms = if completed_count > 0 {
        Some(total_execution_ms as f64 / completed_count as f64)
    } else {
        None
    };

    let stats = StatsResponse {
        total_tasks,
        by_status: ByStatusStats {
            pending,
            running,
            retrying,
            succeeded,
            failed,
        },
        avg_execution_time_ms,
        retry_distribution: retry_dist,
    };

    (StatusCode::OK, Json(stats))
}
