use crate::heartbeat;
use crate::models::{
    InstanceInfo, InstanceStatus, PatchMetadataRequest, RegisterInstanceRequest,
    RegisterInstanceResponse, ServiceInstance,
};
use crate::notify;
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

    {
        let mut services = state.services.write().await;
        let service_instances = services
            .entry(service_name.clone())
            .or_insert_with(HashMap::new);
        service_instances.insert(instance_id, instance.clone());
    }

    notify::notify_subscribers(state.clone(), "registered", &service_name, &instance).await;

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
    Path((service_name, instance_id_str)): Path<(String, String)>,
) -> impl IntoResponse {
    let instance_id = match Uuid::parse_str(&instance_id_str) {
        Ok(id) => id,
        Err(_) => return StatusCode::BAD_REQUEST,
    };

    let instance_to_notify = {
        let mut services = state.services.write().await;

        if let Some(service_instances) = services.get_mut(&service_name) {
            if let Some(instance) = service_instances.remove(&instance_id) {
                if service_instances.is_empty() {
                    services.remove(&service_name);
                }
                Some(instance)
            } else {
                None
            }
        } else {
            None
        }
    };

    if let Some(instance) = instance_to_notify {
        notify::notify_subscribers(state.clone(), "deregistered", &service_name, &instance).await;
        StatusCode::NO_CONTENT
    } else {
        StatusCode::NOT_FOUND
    }
}

pub async fn patch_metadata(
    State(state): State<AppState>,
    Path((service_name, instance_id_str)): Path<(String, String)>,
    Json(request): Json<PatchMetadataRequest>,
) -> impl IntoResponse {
    let instance_id = match Uuid::parse_str(&instance_id_str) {
        Ok(id) => id,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(None::<InstanceInfo>)),
    };

    let result = {
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

                let info = heartbeat::instance_to_info(instance);
                Some((instance.clone(), info))
            } else {
                None
            }
        } else {
            None
        }
    };

    match result {
        Some((instance, info)) => {
            notify::notify_subscribers(state.clone(), "metadata_updated", &service_name, &instance).await;
            (StatusCode::OK, Json(Some(info)))
        }
        None => (StatusCode::NOT_FOUND, Json(None::<InstanceInfo>)),
    }
}

pub async fn handle_heartbeat(
    State(state): State<AppState>,
    Path((service_name, instance_id_str)): Path<(String, String)>,
) -> impl IntoResponse {
    let instance_id = match Uuid::parse_str(&instance_id_str) {
        Ok(id) => id,
        Err(_) => return (StatusCode::BAD_REQUEST, Json(None::<InstanceInfo>)),
    };

    let now = SystemTime::now();
    
    let result = {
        let mut services = state.services.write().await;

        if let Some(service_instances) = services.get_mut(&service_name) {
            if let Some(instance) = service_instances.get_mut(&instance_id) {
                let old_status = instance.status;
                instance.last_heartbeat = Some(now);
                instance.unhealthy_since = None;
                instance.status = InstanceStatus::Healthy;

                let info = heartbeat::instance_to_info(instance);
                Some((instance.clone(), info, old_status))
            } else {
                None
            }
        } else {
            None
        }
    };

    match result {
        Some((instance, info, old_status)) => {
            if old_status != InstanceStatus::Healthy {
                notify::notify_subscribers(state.clone(), "healthy", &service_name, &instance).await;
            }
            (StatusCode::OK, Json(Some(info)))
        }
        None => (StatusCode::NOT_FOUND, Json(None::<InstanceInfo>)),
    }
}
