use axum::{Router, routing::{get, post, put}};
use crate::handlers::*;
use crate::state::AppState;

pub fn create_routes(state: AppState) -> Router {
    Router::new()
        .route("/departments", get(list_departments).post(create_department))
        .route("/departments/:id", get(get_department))
        .route("/employees", get(list_employees).post(create_employee))
        .route("/employees/:id", get(get_employee))
        .route("/employees/:id/certificates", get(get_employee_certificates))
        .route("/certificates", get(list_certificates).post(create_certificate))
        .route("/certificates/:id", get(get_certificate))
        .route("/certificates/:id/renew", put(renew_certificate))
        .route("/certificates/:id/revoke", put(revoke_certificate))
        .route("/training", get(list_training_records).post(create_training_record))
        .route("/training/:id", get(get_training_record))
        .route("/stats/department-rates", get(get_department_rates))
        .route("/stats/expiry-rates", get(get_expiry_rates))
        .route("/alerts/expiring", get(get_expiring_alerts))
        .with_state(state)
}
