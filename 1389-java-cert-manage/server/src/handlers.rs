use axum::{
    extract::{Path, State, Json},
    http::StatusCode,
    response::{IntoResponse, Response},
};
use serde::{Serialize, Deserialize};
use uuid::Uuid;
use cert_core::{CertificateRenewalResult, CertError};
use crate::state::AppState;

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct CreateDepartmentRequest {
    pub name: String,
}

#[derive(Debug, Deserialize)]
pub struct CreateEmployeeRequest {
    pub name: String,
    pub department_id: Uuid,
}

#[derive(Debug, Deserialize)]
pub struct CreateCertificateRequest {
    pub employee_id: Uuid,
    pub certificate_number: String,
    pub issuer: String,
    pub issue_date: String,
    pub expiry_date: String,
    pub certificate_type: String,
    pub required_hours_for_renewal: u32,
}

#[derive(Debug, Deserialize)]
pub struct RenewCertificateRequest {
    pub new_expiry_date: String,
}

#[derive(Debug, Deserialize)]
pub struct CreateTrainingRecordRequest {
    pub name: String,
    pub date: String,
    pub hours: u32,
    pub attendee_ids: Vec<Uuid>,
}

fn parse_date(date_str: &str) -> Result<chrono::NaiveDate, String> {
    chrono::NaiveDate::parse_from_str(date_str, "%Y-%m-%d")
        .map_err(|_| format!("无效的日期格式: {}", date_str))
}

fn json_response<T: Serialize>(status: StatusCode, success: bool, data: Option<T>, message: Option<String>) -> Response {
    let body = Json(ApiResponse { success, data, message });
    (status, body).into_response()
}

fn success_response<T: Serialize>(data: T) -> Response {
    json_response(StatusCode::OK, true, Some(data), None)
}

fn error_response(status: StatusCode, message: String) -> Response {
    json_response::<()>(status, false, None, Some(message))
}

fn handle_error(err: CertError) -> Response {
    match err {
        CertError::DepartmentNotFound => error_response(StatusCode::NOT_FOUND, "部门不存在".to_string()),
        CertError::EmployeeNotFound => error_response(StatusCode::NOT_FOUND, "员工不存在".to_string()),
        CertError::CertificateNotFound => error_response(StatusCode::NOT_FOUND, "证书不存在".to_string()),
        CertError::TrainingRecordNotFound => error_response(StatusCode::NOT_FOUND, "培训记录不存在".to_string()),
        CertError::DepartmentNameExists => error_response(StatusCode::CONFLICT, "部门名称已存在".to_string()),
        CertError::CertificateNumberExists => error_response(StatusCode::CONFLICT, "证书编号已存在".to_string()),
    }
}

pub async fn list_departments(State(state): State<AppState>) -> Response {
    let depts = state.service.list_departments();
    success_response(depts)
}

pub async fn create_department(
    State(state): State<AppState>,
    Json(req): Json<CreateDepartmentRequest>,
) -> Response {
    match state.service.create_department(&req.name) {
        Ok(dept) => success_response(dept),
        Err(e) => handle_error(e),
    }
}

pub async fn get_department(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    match state.service.get_department(&id) {
        Some(dept) => success_response(dept),
        None => error_response(StatusCode::NOT_FOUND, "部门不存在".to_string()),
    }
}

pub async fn list_employees(State(state): State<AppState>) -> Response {
    let employees = state.service.list_employees();
    success_response(employees)
}

pub async fn create_employee(
    State(state): State<AppState>,
    Json(req): Json<CreateEmployeeRequest>,
) -> Response {
    match state.service.create_employee(&req.name, req.department_id) {
        Ok(emp) => success_response(emp),
        Err(e) => handle_error(e),
    }
}

pub async fn get_employee(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    match state.service.get_employee(&id) {
        Some(emp) => success_response(emp),
        None => error_response(StatusCode::NOT_FOUND, "员工不存在".to_string()),
    }
}

pub async fn get_employee_certificates(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    let certs = state.service.get_employee_certificates(&id);
    success_response(certs)
}

pub async fn list_certificates(State(state): State<AppState>) -> Response {
    let certs = state.service.list_certificates();
    success_response(certs)
}

pub async fn create_certificate(
    State(state): State<AppState>,
    Json(req): Json<CreateCertificateRequest>,
) -> Response {
    let issue_date = match parse_date(&req.issue_date) {
        Ok(d) => d,
        Err(e) => return error_response(StatusCode::BAD_REQUEST, e),
    };
    let expiry_date = match parse_date(&req.expiry_date) {
        Ok(d) => d,
        Err(e) => return error_response(StatusCode::BAD_REQUEST, e),
    };
    
    match state.service.create_certificate(
        req.employee_id,
        &req.certificate_number,
        &req.issuer,
        issue_date,
        expiry_date,
        &req.certificate_type,
        req.required_hours_for_renewal,
    ) {
        Ok(cert) => success_response(cert),
        Err(e) => handle_error(e),
    }
}

pub async fn get_certificate(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    match state.service.get_certificate(&id) {
        Some(cert) => success_response(cert),
        None => error_response(StatusCode::NOT_FOUND, "证书不存在".to_string()),
    }
}

pub async fn renew_certificate(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<RenewCertificateRequest>,
) -> Response {
    let new_expiry = match parse_date(&req.new_expiry_date) {
        Ok(d) => d,
        Err(e) => return error_response(StatusCode::BAD_REQUEST, e),
    };
    
    match state.service.renew_certificate(id, new_expiry) {
        Ok(result) => {
            let message = match result {
                CertificateRenewalResult::RenewedSuccessfully => "续证成功".to_string(),
                CertificateRenewalResult::RenewedWithInsufficientHours { actual_hours, required_hours } => {
                    format!("续证成功但学时不足: 实际{}学时，需{}学时", actual_hours, required_hours)
                },
                CertificateRenewalResult::CannotRenewRevoked => return error_response(StatusCode::BAD_REQUEST, "已注销的证书不能续证".to_string()),
            };
            json_response(StatusCode::OK, true, Some(serde_json::json!({
                "result": format!("{:?}", result),
                "message": message
            })), Some(message))
        },
        Err(e) => handle_error(e),
    }
}

pub async fn revoke_certificate(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    match state.service.revoke_certificate(id) {
        Ok(cert) => success_response(cert),
        Err(e) => handle_error(e),
    }
}

pub async fn list_training_records(State(state): State<AppState>) -> Response {
    let records = state.service.list_training_records();
    success_response(records)
}

pub async fn create_training_record(
    State(state): State<AppState>,
    Json(req): Json<CreateTrainingRecordRequest>,
) -> Response {
    let date = match parse_date(&req.date) {
        Ok(d) => d,
        Err(e) => return error_response(StatusCode::BAD_REQUEST, e),
    };
    
    match state.service.create_training_record(
        &req.name,
        date,
        req.hours,
        req.attendee_ids,
    ) {
        Ok(record) => success_response(record),
        Err(e) => handle_error(e),
    }
}

pub async fn get_training_record(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Response {
    match state.service.get_training_record(&id) {
        Some(record) => success_response(record),
        None => error_response(StatusCode::NOT_FOUND, "培训记录不存在".to_string()),
    }
}

pub async fn get_department_rates(State(state): State<AppState>) -> Response {
    let rates = state.service.get_department_certification_rates();
    success_response(rates)
}

pub async fn get_expiry_rates(State(state): State<AppState>) -> Response {
    let rates = state.service.get_certificate_type_expiry_rates();
    success_response(rates)
}

pub async fn get_expiring_alerts(State(state): State<AppState>) -> Response {
    let alerts = state.service.get_expiring_certificates_alerts();
    success_response(alerts)
}
