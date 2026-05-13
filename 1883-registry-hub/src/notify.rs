use crate::heartbeat::instance_to_info;
use crate::models::{InstanceChangeNotification, ServiceInstance, Subscription, SubscriptionRequest};
use crate::store::AppState;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use reqwest::Client;
use std::time::SystemTime;
use uuid::Uuid;

pub async fn subscribe(
    State(state): State<AppState>,
    Json(request): Json<SubscriptionRequest>,
) -> impl IntoResponse {
    let subscription = Subscription {
        subscription_id: Uuid::new_v4(),
        service_name: request.service_name,
        callback_url: request.callback_url,
        created_at: SystemTime::now(),
    };

    let subscription_id = subscription.subscription_id;
    state
        .subscriptions
        .write()
        .await
        .insert(subscription_id, subscription);

    let response = serde_json::json!({
        "subscription_id": subscription_id.to_string()
    });

    (StatusCode::CREATED, Json(response))
}

pub async fn unsubscribe(
    State(state): State<AppState>,
    Path(subscription_id_str): Path<String>,
) -> impl IntoResponse {
    let subscription_id = match Uuid::parse_str(&subscription_id_str) {
        Ok(id) => id,
        Err(_) => return StatusCode::BAD_REQUEST,
    };

    let mut subscriptions = state.subscriptions.write().await;
    if subscriptions.remove(&subscription_id).is_some() {
        StatusCode::NO_CONTENT
    } else {
        StatusCode::NOT_FOUND
    }
}

pub async fn list_subscriptions(State(state): State<AppState>) -> impl IntoResponse {
    let subscriptions = state.subscriptions.read().await;
    let subs: Vec<Subscription> = subscriptions.values().cloned().collect();
    (StatusCode::OK, Json(subs))
}

pub async fn notify_subscribers(
    state: AppState,
    event_type: &str,
    service_name: &str,
    instance: &ServiceInstance,
) {
    let subscriptions = state.subscriptions.read().await;

    let matched_subscriptions: Vec<Subscription> = subscriptions
        .values()
        .filter(|sub| match &sub.service_name {
            Some(name) => name == service_name,
            None => true,
        })
        .cloned()
        .collect();

    drop(subscriptions);

    if matched_subscriptions.is_empty() {
        return;
    }

    let notification = InstanceChangeNotification {
        event_type: event_type.to_string(),
        service_name: service_name.to_string(),
        instance: instance_to_info(instance),
        timestamp: SystemTime::now(),
    };

    let client = Client::new();

    for sub in matched_subscriptions {
        let notification = notification.clone();
        let client = client.clone();
        let callback_url = sub.callback_url.clone();

        tokio::spawn(async move {
            let result = client
                .post(&callback_url)
                .json(&notification)
                .timeout(std::time::Duration::from_secs(5))
                .send()
                .await;

            match result {
                Ok(response) => {
                    if !response.status().is_success() {
                        tracing::warn!(
                            "Notification to {} failed with status: {}",
                            callback_url,
                            response.status()
                        );
                    }
                }
                Err(e) => {
                    tracing::warn!("Notification to {} failed: {}", callback_url, e);
                }
            }
        });
    }
}

pub async fn notify_batch(
    state: AppState,
    event_type: &str,
    instances: Vec<(String, ServiceInstance)>,
) {
    for (service_name, instance) in instances {
        notify_subscribers(state.clone(), event_type, &service_name, &instance).await;
    }
}
