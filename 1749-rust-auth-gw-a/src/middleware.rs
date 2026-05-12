use axum::{
    extract::{Request, State},
    http::{HeaderMap, HeaderName, HeaderValue, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
};
use crate::access_log::{AccessLogEntry, AccessLogger};
use crate::blacklist::TokenBlacklist;
use crate::jwt::{Claims, JwtService};
use chrono::Utc;

#[derive(Clone)]
pub struct AuthState {
    pub jwt_service: JwtService,
    pub blacklist: TokenBlacklist,
    pub access_logger: AccessLogger,
}

fn extract_bearer_token(headers: &HeaderMap) -> Option<String> {
    let auth_header = headers.get("authorization")?;
    let auth_str = auth_header.to_str().ok()?;
    if !auth_str.starts_with("Bearer ") {
        return None;
    }
    Some(auth_str["Bearer ".len()..].to_string())
}

fn extract_remote_addr(headers: &HeaderMap) -> Option<String> {
    headers
        .get("x-forwarded-for")
        .or_else(|| headers.get("x-real-ip"))
        .and_then(|h| h.to_str().ok())
        .map(|s| s.to_string())
}

fn extract_user_agent(headers: &HeaderMap) -> Option<String> {
    headers
        .get("user-agent")
        .and_then(|h| h.to_str().ok())
        .map(|s| s.to_string())
}

fn build_user_headers(claims: &Claims) -> HeaderMap {
    let mut headers = HeaderMap::new();
    if let Ok(val) = HeaderValue::from_str(&claims.user_id) {
        headers.insert(
            HeaderName::from_static("x-auth-user-id"),
            val,
        );
    }
    if let Ok(val) = HeaderValue::from_str(&claims.username) {
        headers.insert(
            HeaderName::from_static("x-auth-username"),
            val,
        );
    }
    if let Ok(val) = HeaderValue::from_str(&claims.role) {
        headers.insert(
            HeaderName::from_static("x-auth-role"),
            val,
        );
    }
    headers
}

fn unauthorized_response(message: &'static str) -> Response {
    let body = serde_json::json!({
        "error": "Unauthorized",
        "message": message
    });
    (
        StatusCode::UNAUTHORIZED,
        [(axum::http::header::CONTENT_TYPE, "application/json")],
        axum::Json(body),
    ).into_response()
}

pub async fn auth_middleware(
    State(state): State<AuthState>,
    mut req: Request,
    next: Next,
) -> Response {
    let method = req.method().to_string();
    let path = req.uri().path().to_string();
    let headers = req.headers();
    let remote_addr = extract_remote_addr(headers);
    let user_agent = extract_user_agent(headers);
    let token = match extract_bearer_token(headers) {
        Some(t) => t,
        None => {
            let resp = unauthorized_response("缺少 Authorization Token");
            log_access(
                &state.access_logger,
                "".to_string(),
                "".to_string(),
                method,
                path,
                remote_addr,
                user_agent,
                StatusCode::UNAUTHORIZED.as_u16(),
            );
            return resp;
        }
    };

    let token_data = match state.jwt_service.validate_token(&token) {
        Ok(data) => data,
        Err(_) => {
            let resp = unauthorized_response("Token 无效或已过期");
            log_access(
                &state.access_logger,
                "".to_string(),
                "".to_string(),
                method,
                path,
                remote_addr,
                user_agent,
                StatusCode::UNAUTHORIZED.as_u16(),
            );
            return resp;
        }
    };

    let jti = state.jwt_service.get_jti(&token_data.claims);
    if state.blacklist.contains(&jti) {
        let resp = unauthorized_response("Token 已被注销");
        log_access(
            &state.access_logger,
            token_data.claims.user_id.clone(),
            token_data.claims.username.clone(),
            method,
            path,
            remote_addr.clone(),
            user_agent.clone(),
            StatusCode::UNAUTHORIZED.as_u16(),
        );
        return resp;
    }

    let claims = token_data.claims.clone();
    let user_headers = build_user_headers(&claims);
    req.headers_mut().extend(user_headers);

    let should_refresh = state.jwt_service.should_refresh(&claims);
    let new_token = if should_refresh {
        state.jwt_service.refresh_token(&claims).ok()
    } else {
        None
    };

    let mut response = next.run(req).await;

    if let Some(new_tok) = new_token {
        if let Ok(val) = HeaderValue::from_str(&new_tok) {
            response.headers_mut().insert(
                HeaderName::from_static("x-new-auth-token"),
                val,
            );
        }
    }

    log_access(
        &state.access_logger,
        claims.user_id,
        claims.username,
        method,
        path,
        remote_addr,
        user_agent,
        response.status().as_u16(),
    );

    response
}

fn log_access(
    logger: &AccessLogger,
    user_id: String,
    username: String,
    method: String,
    path: String,
    remote_addr: Option<String>,
    user_agent: Option<String>,
    status: u16,
) {
    if user_id.is_empty() {
        return;
    }
    let entry = AccessLogEntry {
        user_id,
        username,
        method,
        path,
        remote_addr,
        user_agent,
        timestamp: Utc::now(),
        status,
    };
    logger.log(entry);
}
