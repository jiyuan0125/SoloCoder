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
    ConfigChangeRequest, ConfigVersion, DeployRecord, DeployStatus, DeployStrategy,
    ImpactResponse, ValidationResult,
};
use crate::storage::ConfigStore;
use crate::validator::validate;

#[derive(Debug, Deserialize)]
pub struct DeployQuery {
    strategy: String,
    percentage: Option<u32>,
}

#[derive(Debug, Deserialize)]
pub struct RollbackQuery {
    version: Option<Uuid>,
}

pub async fn create_config(
    State(store): State<ConfigStore>,
    Json(request): Json<ConfigChangeRequest>,
) -> impl IntoResponse {
    let validation_result = validate(&request.value, &request.schema, "");
    
    if !validation_result.valid {
        return (
            StatusCode::UNPROCESSABLE_ENTITY,
            Json(serde_json::json!({
                "valid": false,
                "errors": validation_result.errors
            })),
        );
    }
    
    let version = ConfigVersion {
        version: Uuid::new_v4(),
        key: request.key.clone(),
        value: request.value,
        schema: request.schema,
        description: request.description,
        author: request.author,
        created_at: Utc::now(),
    };
    
    store.add_version(version.clone());
    
    (
        StatusCode::CREATED,
        Json(serde_json::json!({
            "valid": true,
            "version": version.version,
            "key": version.key,
            "value": version.value,
            "created_at": version.created_at
        })),
    )
}

pub async fn get_config_impact(
    State(store): State<ConfigStore>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    if let Some(mapping) = store.get_service_mapping(&key) {
        let total_instances: u32 = mapping.services.iter().map(|s| s.instance_count).sum();
        
        return (
            StatusCode::OK,
            Json(ImpactResponse {
                key,
                services: mapping.services,
                total_instances,
            }),
        );
    }
    
    (
        StatusCode::NOT_FOUND,
        Json(ImpactResponse {
            key,
            services: vec![],
            total_instances: 0,
        }),
    )
}

pub async fn deploy_config(
    State(store): State<ConfigStore>,
    Path(key): Path<String>,
    Query(query): Query<DeployQuery>,
) -> impl IntoResponse {
    let strategy = match query.strategy.to_lowercase().as_str() {
        "canary" => DeployStrategy::Canary,
        "full" => DeployStrategy::Full,
        _ => {
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({
                    "error": "Invalid strategy. Must be 'canary' or 'full'"
                })),
            );
        }
    };
    
    if strategy == DeployStrategy::Canary && query.percentage.is_none() {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({
                "error": "Canary deploy requires 'percentage' parameter"
            })),
        );
    }
    
    let latest_version = match store.get_latest_version(&key) {
        Some(v) => v,
        None => {
            return (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({
                    "error": format!("No config versions found for key '{}'", key)
                })),
            );
        }
    };
    
    let deployment = DeployRecord {
        id: Uuid::new_v4(),
        key: key.clone(),
        version: latest_version.version,
        strategy: strategy.clone(),
        percentage: query.percentage,
        status: DeployStatus::InProgress,
        created_at: Utc::now(),
        completed_at: None,
    };
    
    store.add_deploy_record(deployment.clone());
    
    match strategy {
        DeployStrategy::Canary => {
            store.set_canary_version(&key, latest_version.version);
            store.update_deploy_status(deployment.id, DeployStatus::Success);
            
            (
                StatusCode::OK,
                Json(serde_json::json!({
                    "deployment_id": deployment.id,
                    "key": key,
                    "strategy": "canary",
                    "percentage": query.percentage,
                    "version": latest_version.version,
                    "status": "success",
                    "message": format!("Canary deployment initiated for {}% instances", query.percentage.unwrap_or(0))
                })),
            )
        }
        DeployStrategy::Full => {
            store.set_full_version(&key, latest_version.version);
            store.clear_canary_version(&key);
            store.update_deploy_status(deployment.id, DeployStatus::Success);
            
            (
                StatusCode::OK,
                Json(serde_json::json!({
                    "deployment_id": deployment.id,
                    "key": key,
                    "strategy": "full",
                    "version": latest_version.version,
                    "status": "success",
                    "message": "Full deployment completed successfully"
                })),
            )
        }
    }
}

pub async fn get_config_versions(
    State(store): State<ConfigStore>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    let versions = store.get_versions(&key);
    
    if versions.is_empty() {
        return (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({
                "error": format!("No versions found for key '{}'", key)
            })),
        );
    }
    
    (
        StatusCode::OK,
        Json(serde_json::json!({
            "key": key,
            "count": versions.len(),
            "versions": versions
        })),
    )
}

pub async fn rollback_config(
    State(store): State<ConfigStore>,
    Path(key): Path<String>,
    Query(query): Query<RollbackQuery>,
) -> impl IntoResponse {
    let versions = store.get_versions(&key);
    
    if versions.len() < 2 {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({
                "error": "Not enough versions to rollback"
            })),
        );
    }
    
    let target_version = match query.version {
        Some(v) => {
            match versions.iter().find(|ver| ver.version == v) {
                Some(ver) => ver.clone(),
                None => {
                    return (
                        StatusCode::NOT_FOUND,
                        Json(serde_json::json!({
                            "error": format!("Version {} not found for key '{}'", v, key)
                        })),
                    );
                }
            }
        }
        None => {
            versions.get(versions.len() - 2).unwrap().clone()
        }
    };
    
    let rollback_version = ConfigVersion {
        version: Uuid::new_v4(),
        key: key.clone(),
        value: target_version.value.clone(),
        schema: target_version.schema.clone(),
        description: Some(format!("Rollback to version {}", target_version.version)),
        author: None,
        created_at: Utc::now(),
    };
    
    store.add_version(rollback_version.clone());
    store.set_full_version(&key, rollback_version.version);
    store.clear_canary_version(&key);
    
    let deployment = DeployRecord {
        id: Uuid::new_v4(),
        key: key.clone(),
        version: rollback_version.version,
        strategy: DeployStrategy::Full,
        percentage: None,
        status: DeployStatus::RolledBack,
        created_at: Utc::now(),
        completed_at: Some(Utc::now()),
    };
    
    store.add_deploy_record(deployment);
    
    (
        StatusCode::OK,
        Json(serde_json::json!({
            "new_version": rollback_version.version,
            "rolled_back_to": target_version.version,
            "key": key,
            "value": rollback_version.value,
            "status": "success",
            "message": "Rollback completed successfully"
        })),
    )
}

pub async fn get_deploy_status(
    State(store): State<ConfigStore>,
    Path(key): Path<String>,
) -> impl IntoResponse {
    let full_version = store.get_full_version(&key);
    let canary_version = store.get_canary_version(&key);
    let deployments = store.get_deploy_records(&key);
    
    (
        StatusCode::OK,
        Json(serde_json::json!({
            "key": key,
            "full_version": full_version.map(|v| serde_json::json!({
                "version": v.version,
                "value": v.value,
                "created_at": v.created_at
            })),
            "canary_version": canary_version.map(|v| serde_json::json!({
                "version": v.version,
                "value": v.value,
                "created_at": v.created_at
            })),
            "deployments": deployments
        })),
    )
}

pub async fn health_check() -> impl IntoResponse {
    (
        StatusCode::OK,
        Json(serde_json::json!({
            "status": "ok",
            "service": "config-validator"
        })),
    )
}
