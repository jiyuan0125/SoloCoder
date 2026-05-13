use crate::heartbeat::instance_to_info;
use crate::models::{InstanceInfo, InstanceStatus, PaginatedResponse};
use crate::store::AppState;
use axum::{
    extract::{Path, Query, State, RawQuery},
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
}

fn default_page() -> usize {
    1
}

fn default_page_size() -> usize {
    20
}

fn parse_metadata_filters(raw_query: Option<&str>) -> HashMap<String, String> {
    let mut filters = HashMap::new();
    
    if let Some(query) = raw_query {
        for param in query.split('&') {
            if param.is_empty() {
                continue;
            }
            
            let decoded = percent_decode(param);
            
            if let Some(idx) = decoded.find('=') {
                let key = &decoded[..idx];
                let value = &decoded[idx + 1..];
                
                if key.starts_with("metadata[") && key.ends_with(']') {
                    let metadata_key = &key[9..key.len() - 1];
                    if !metadata_key.is_empty() {
                        filters.insert(metadata_key.to_string(), value.to_string());
                    }
                }
            }
        }
    }
    
    filters
}

fn percent_decode(s: &str) -> String {
    let mut result = String::with_capacity(s.len());
    let mut chars = s.chars();
    
    while let Some(c) = chars.next() {
        if c == '+' {
            result.push(' ');
        } else if c == '%' {
            let hex1 = chars.next().unwrap_or('0');
            let hex2 = chars.next().unwrap_or('0');
            let hex = format!("{}{}", hex1, hex2);
            if let Ok(byte) = u8::from_str_radix(&hex, 16) {
                result.push(byte as char);
            }
        } else {
            result.push(c);
        }
    }
    
    result
}

pub async fn discover_service(
    State(state): State<AppState>,
    Path(service_name): Path<String>,
    Query(query): Query<DiscoveryQuery>,
    raw_query: RawQuery,
) -> impl IntoResponse {
    let metadata_filters = parse_metadata_filters(raw_query.0.as_deref());
    
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
            metadata_filters.iter().all(|(key, value_substr)| {
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
