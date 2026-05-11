use crate::AppState;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json, Router,
    routing::{get, post},
};
use insurance_core::{PolicyApplication, Policy, RuleEngine, ValidationError, ValidationResult};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

pub fn create_app(state: AppState) -> Router {
    Router::new()
        .route("/products", get(list_products))
        .route("/policies", get(list_policies).post(create_policy))
        .route("/policies/:policy_number", get(get_policy))
        .route("/validate", post(validate_application))
        .with_state(state)
}

async fn list_products(State(state): State<AppState>) -> Json<Vec<insurance_core::Product>> {
    Json(state.store.list_products())
}

async fn list_policies(State(state): State<AppState>) -> Json<Vec<Policy>> {
    Json(state.store.list_policies())
}

async fn get_policy(
    State(state): State<AppState>,
    Path(policy_number): Path<String>,
) -> impl IntoResponse {
    match state.store.get_policy(&policy_number) {
        Some(policy) => (StatusCode::OK, Json(policy)).into_response(),
        None => (StatusCode::NOT_FOUND, "保单不存在").into_response(),
    }
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

#[derive(Debug, Serialize)]
struct CreatePolicyResponse {
    success: bool,
    policy: Option<Policy>,
    validation_result: ValidationResult,
}

async fn create_policy(
    State(state): State<AppState>,
    Json(application): Json<PolicyApplication>,
) -> Json<CreatePolicyResponse> {
    let product = match state.store.get_product(&application.product_code) {
        Some(p) => p,
        None => {
            return Json(CreatePolicyResponse {
                success: false,
                policy: None,
                validation_result: ValidationResult {
                    valid: false,
                    errors: vec![ValidationError::ProductNotFound(application.product_code)],
                    premium: None,
                },
            });
        }
    };

    let has_existing = state
        .store
        .has_existing_policy(&application.insured.id_card, &application.product_code);

    let validation_result =
        RuleEngine::validate_application(&application, &product, has_existing);

    if !validation_result.valid {
        return Json(CreatePolicyResponse {
            success: false,
            policy: None,
            validation_result,
        });
    }

    let policy = Policy {
        policy_number: format!("POL-{}", Uuid::new_v4()),
        insured: application.insured.clone(),
        product,
        application_date: application.application_date,
        beneficiaries: application.beneficiaries,
        insured_amount: application.insured_amount,
        premium: validation_result.premium.unwrap(),
    };

    state.store.add_policy(policy.clone());

    Json(CreatePolicyResponse {
        success: true,
        policy: Some(policy),
        validation_result,
    })
}

#[derive(Debug, Deserialize)]
struct ValidateRequest {
    application: PolicyApplication,
}

async fn validate_application(
    State(state): State<AppState>,
    Json(request): Json<ValidateRequest>,
) -> impl IntoResponse {
    let product = match state.store.get_product(&request.application.product_code) {
        Some(p) => p,
        None => {
            return Json(ValidationResult {
                valid: false,
                errors: vec![ValidationError::ProductNotFound(
                    request.application.product_code,
                )],
                premium: None,
            });
        }
    };

    let has_existing = state.store.has_existing_policy(
        &request.application.insured.id_card,
        &request.application.product_code,
    );

    let result = RuleEngine::validate_application(&request.application, &product, has_existing);

    Json(result)
}
