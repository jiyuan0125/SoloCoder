use crate::models::{
    InstanceInfo, InstanceStatus, PatchMetadataRequest, RegisterInstanceRequest,
    RegisterInstanceResponse, ServiceInstance,
};
use crate::store::AppState;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use std::collections::HashMap;
use std::time::SystemTime;
use uuid::Uuid;

pub async fn register_instance(
    State(state): State<AppState>,
    Path(service_name): Path<String>,
    Json(request): Json<RegisterInstanceRequest>,
) -> impl IntoResponse {
    let instance_id = Uuid::new_v4();
    let now = SystemTime::now();

    let instance = ServiceInstance {
        instance_id,
        service_name: service_name.clone(),
        host: request.host,
        port: request.port,
        metadata: request.metadata,
        status: InstanceStatus::Registered,
        last_heartbeat: None,
        registered_at: now,
        unhealthy_since: None,
    };

    let mut services = state.services.write().await;
    let service_instances = services
        .entry(service_name.clone())
        .or_insert_with(HashMap::new);
    service_instances.insert(instance_id, instance.clone());

    let response = RegisterInstanceResponse {
        instance_id,
        service_name,
        host: instance.host,
        port: instance.port,
        metadata: instance.metadata,
        status: instance.status.to_string(),
    };

    (StatusCode::CREATED, Json(response))
}

pub async fn deregister_instance(
    State(state): State<AppState>,
    Path((service_name, instance_id)): Path<(String, Uuid)>,
) -> impl IntoResponse {
    let mut services = state.services.write().await;

    if let Some(service_instances) = services.get_mut(&service_name) {
        if service_instances.remove(&instance_id).is_some() {
            if service_instances.is_empty() {
                services.remove(&service_name);
            }
            return StatusCode::NO_CONTENT;
        }
    }

    StatusCode::NOT_FOUND
}

pub async fn patch_metadata(
    State(state): State<AppState>,
    Path((service_name, instance_id)): Path<(String, Uuid)>,
    Json(request): Json<PatchMetadataRequest>,
) -> impl IntoResponse {
    let mut services = state.services.write().await;

    if let Some(service_instances) = services.get_mut(&service_name) {
        if let Some(instance) = service_instances.get_mut(&instance_id) {
            for (key, value) in request.updates {
                match value {
                    Some(v) => {
                        instance.metadata.insert(key, v);
                    }
                    None => {
                        instance.metadata.remove(&key);
                    }
                }
            }

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
    }

    (StatusCode::NOT_FOUND, Json(None::<InstanceInfo>))
}
