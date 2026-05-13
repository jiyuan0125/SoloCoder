use crate::heartbeat::instance_to_info;
use crate::models::{InstanceInfo, InstanceStatus, PaginatedResponse};
use crate::store::AppState;
use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use serde::Deserialize;
use std::collections::HashMap;

#[derive(Debug, Clone, Deserialize)]
pub struct DiscoveryQuery {
    #[serde(default = "default_page")]
    pub page: usize,
    #[serde(default = "default_page_size")]
    pub page_size: usize,
    #[serde(default)]
    pub metadata: HashMap<String, String>,
}

fn default_page() -> usize {
    1
}

fn default_page_size() -> usize {
    20
}

pub async fn discover_service(
    State(state): State<AppState>,
    Path(service_name): Path<String>,
    Query(query): Query<DiscoveryQuery>,
) -> impl IntoResponse {
    let services = state.services.read().await;

    let instances = match services.get(&service_name) {
        Some(instances) => instances,
        None => {
            let response = PaginatedResponse {
                items: Vec::new(),
                total: 0,
                page: query.page,
                page_size: query.page_size,
                total_pages: 0,
            };
            return (StatusCode::OK, Json(response));
        }
    };

    let mut healthy_instances: Vec<InstanceInfo> = instances
        .values()
        .filter(|i| i.status == InstanceStatus::Healthy)
        .filter(|i| {
            query.metadata.iter().all(|(key, value_substr)| {
                i.metadata
                    .get(key)
                    .map(|v| v.contains(value_substr))
                    .unwrap_or(false)
            })
        })
        .map(instance_to_info)
        .collect();

    healthy_instances.sort_by(|a, b| a.instance_id.cmp(&b.instance_id));

    let total = healthy_instances.len();
    let page = if query.page < 1 { 1 } else { query.page };
    let page_size = if query.page_size < 1 {
        20
    } else {
        query.page_size
    };
    let total_pages = if total == 0 {
        0
    } else {
        (total + page_size - 1) / page_size
    };

    let start = (page - 1) * page_size;
    let items: Vec<InstanceInfo> = healthy_instances
        .into_iter()
        .skip(start)
        .take(page_size)
        .collect();

    let response = PaginatedResponse {
        items,
        total,
        page,
        page_size,
        total_pages,
    };

    (StatusCode::OK, Json(response))
}
