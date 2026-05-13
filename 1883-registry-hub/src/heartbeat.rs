use crate::models::{InstanceInfo, InstanceStatus, ServiceInstance};
use crate::store::AppState;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use std::collections::HashMap;
use std::time::{Duration, SystemTime};
use uuid::Uuid;

pub const HEARTBEAT_TIMEOUT_SECS: u64 = 30;
pub const UNHEALTHY_CLEANUP_SECS: u64 = 60;

pub async fn handle_heartbeat(
    State(state): State<AppState>,
    Path((service_name, instance_id)): Path<(String, Uuid)>,
) -> impl IntoResponse {
    let now = SystemTime::now();
    let mut services = state.services.write().await;

    let service_instances = services
        .entry(service_name.clone())
        .or_insert_with(HashMap::new);

    if let Some(instance) = service_instances.get_mut(&instance_id) {
        instance.last_heartbeat = Some(now);
        instance.unhealthy_since = None;
        instance.status = InstanceStatus::Healthy;

        let info = InstanceInfo {
            instance_id: instance.instance_id,
            service_name: instance.service_name.clone(),
            host: instance.host.clone(),
            port: instance.port,
            metadata: instance.metadata.clone(),
            status: instance.status.to_string(),
        };

        return (StatusCode::OK, Json(Some(info)));
    }

    (StatusCode::NOT_FOUND, Json(None::<InstanceInfo>))
}

pub fn instance_to_info(instance: &ServiceInstance) -> InstanceInfo {
    InstanceInfo {
        instance_id: instance.instance_id,
        service_name: instance.service_name.clone(),
        host: instance.host.clone(),
        port: instance.port,
        metadata: instance.metadata.clone(),
        status: instance.status.to_string(),
    }
}

pub struct HeartbeatCheckResult {
    pub became_unhealthy: Vec<(String, ServiceInstance)>,
    pub cleaned_up: Vec<(String, ServiceInstance)>,
    pub became_healthy: Vec<(String, ServiceInstance)>,
}

pub async fn check_heartbeat_timeouts(state: AppState) -> HeartbeatCheckResult {
    let now = SystemTime::now();
    let timeout = Duration::from_secs(HEARTBEAT_TIMEOUT_SECS);
    let cleanup_timeout = Duration::from_secs(UNHEALTHY_CLEANUP_SECS);

    let mut result = HeartbeatCheckResult {
        became_unhealthy: Vec::new(),
        cleaned_up: Vec::new(),
        became_healthy: Vec::new(),
    };

    let mut services = state.services.write().await;
    let mut services_to_remove: Vec<String> = Vec::new();

    for (service_name, instances) in services.iter_mut() {
        let mut instances_to_remove: Vec<Uuid> = Vec::new();

        for (instance_id, instance) in instances.iter_mut() {
            match instance.status {
                InstanceStatus::Registered => {
                    if let Some(last_heartbeat) = instance.last_heartbeat {
                        if now.duration_since(last_heartbeat).ok().map(|d| d < timeout).unwrap_or(false) {
                            instance.status = InstanceStatus::Healthy;
                            instance.unhealthy_since = None;
                            result.became_healthy.push((service_name.clone(), instance.clone()));
                        }
                    }

                    let age = now.duration_since(instance.registered_at).ok();
                    if age.map(|d| d >= timeout).unwrap_or(false) && instance.last_heartbeat.is_none() {
                        instance.status = InstanceStatus::Unhealthy;
                        instance.unhealthy_since = Some(now);
                        result.became_unhealthy.push((service_name.clone(), instance.clone()));
                    }
                }
                InstanceStatus::Healthy => {
                    if let Some(last_heartbeat) = instance.last_heartbeat {
                        if now.duration_since(last_heartbeat).ok().map(|d| d >= timeout).unwrap_or(false) {
                            instance.status = InstanceStatus::Unhealthy;
                            instance.unhealthy_since = Some(now);
                            result.became_unhealthy.push((service_name.clone(), instance.clone()));
                        }
                    }
                }
                InstanceStatus::Unhealthy => {
                    if let Some(last_heartbeat) = instance.last_heartbeat {
                        if now.duration_since(last_heartbeat).ok().map(|d| d < timeout).unwrap_or(false) {
                            instance.status = InstanceStatus::Healthy;
                            instance.unhealthy_since = None;
                            result.became_healthy.push((service_name.clone(), instance.clone()));
                            continue;
                        }
                    }

                    if let Some(unhealthy_since) = instance.unhealthy_since {
                        if now.duration_since(unhealthy_since).ok().map(|d| d >= cleanup_timeout).unwrap_or(false) {
                            instances_to_remove.push(*instance_id);
                            result.cleaned_up.push((service_name.clone(), instance.clone()));
                        }
                    }
                }
            }
        }

        for instance_id in &instances_to_remove {
            instances.remove(instance_id);
        }

        if instances.is_empty() {
            services_to_remove.push(service_name.clone());
        }
    }

    for service_name in &services_to_remove {
        services.remove(service_name);
    }

    result
}
