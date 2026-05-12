use actix_web::{web, HttpResponse, HttpMessage};
use serde::{Deserialize, Serialize};

use crate::auth::{
    create_token_pair, revoke_refresh_token, validate_token, AuthState, Claims, TokenType,
};
use crate::errors::{internal_error, unauthorized, ErrorResponse};

#[derive(Debug, Deserialize)]
pub struct LoginRequest {
    username: String,
    password: String,
}

#[derive(Debug, Serialize)]
pub struct LoginResponse {
    pub success: bool,
    pub data: Option<TokenData>,
    pub error: Option<ErrorDetail>,
}

#[derive(Debug, Serialize)]
pub struct TokenData {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_in: usize,
    pub token_type: String,
}

#[derive(Debug, Serialize)]
pub struct ErrorDetail {
    pub code: String,
    pub message: String,
}

#[derive(Debug, Deserialize)]
pub struct RefreshRequest {
    refresh_token: String,
}

#[derive(Debug, Serialize)]
pub struct ProtectedResponse {
    pub success: bool,
    pub data: ProtectedData,
}

#[derive(Debug, Serialize)]
pub struct ProtectedData {
    pub user_id: String,
    pub message: String,
    pub timestamp: i64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ExamplePayload {
    pub id: u64,
    pub name: String,
    pub description: String,
}

fn validate_credentials(username: &str, password: &str) -> Option<String> {
    if username == "admin" && password == "password123" {
        Some("user-001".to_string())
    } else if username == "demo" && password == "demo123" {
        Some("user-002".to_string())
    } else {
        None
    }
}

#[actix_web::post("/login")]
pub async fn login(
    auth_state: web::Data<AuthState>,
    credentials: web::Json<LoginRequest>,
) -> HttpResponse {
    log::info!(
        "LOGIN_ATTEMPT: username={}",
        credentials.username
    );

    let user_id = match validate_credentials(&credentials.username, &credentials.password) {
        Some(id) => id,
        None => {
            log::warn!(
                "LOGIN_FAILED: invalid credentials for username={}",
                credentials.username
            );
            return unauthorized("INVALID_CREDENTIALS", "Invalid username or password");
        }
    };

    let token_pair = match create_token_pair(&user_id, &auth_state) {
        Ok(tp) => tp,
        Err(e) => {
            log::error!("LOGIN_ERROR: failed to create tokens: {:?}", e);
            return internal_error("Failed to create authentication tokens");
        }
    };

    log::info!("LOGIN_SUCCESS: user_id={}", user_id);

    HttpResponse::Ok().json(LoginResponse {
        success: true,
        data: Some(TokenData {
            access_token: token_pair.access_token,
            refresh_token: token_pair.refresh_token,
            expires_in: token_pair.expires_in,
            token_type: "Bearer".to_string(),
        }),
        error: None,
    })
}

#[actix_web::post("/refresh")]
pub async fn refresh(
    auth_state: web::Data<AuthState>,
    refresh_req: web::Json<RefreshRequest>,
) -> HttpResponse {
    let claims = match validate_token(&refresh_req.refresh_token, &auth_state, TokenType::Refresh) {
        Ok(c) => c,
        Err(e) => {
            log::warn!("REFRESH_FAILED: invalid refresh token: {:?}", e);
            let err_resp = ErrorResponse::from(e);
            return HttpResponse::Unauthorized().json(err_resp);
        }
    };

    let old_jti = claims.jti.clone();
    let user_id = claims.sub;

    revoke_refresh_token(&auth_state, &old_jti);

    let token_pair = match create_token_pair(&user_id, &auth_state) {
        Ok(tp) => tp,
        Err(e) => {
            log::error!("REFRESH_ERROR: failed to create new tokens: {:?}", e);
            return internal_error("Failed to refresh tokens");
        }
    };

    log::info!("REFRESH_SUCCESS: user_id={}", user_id);

    HttpResponse::Ok().json(LoginResponse {
        success: true,
        data: Some(TokenData {
            access_token: token_pair.access_token,
            refresh_token: token_pair.refresh_token,
            expires_in: token_pair.expires_in,
            token_type: "Bearer".to_string(),
        }),
        error: None,
    })
}

#[actix_web::get("/protected")]
pub async fn protected(req: actix_web::HttpRequest) -> HttpResponse {
    let claims = req.extensions().get::<Claims>().cloned().unwrap();
    HttpResponse::Ok().json(ProtectedResponse {
        success: true,
        data: ProtectedData {
            user_id: claims.sub.clone(),
            message: "You have accessed a protected resource!".to_string(),
            timestamp: chrono::Utc::now().timestamp(),
        },
    })
}

#[actix_web::post("/data")]
pub async fn protected_data(
    req: actix_web::HttpRequest,
    payload: web::Json<ExamplePayload>,
) -> HttpResponse {
    let claims = req.extensions().get::<Claims>().cloned().unwrap();
    log::info!(
        "PROTECTED_DATA: user_id={}, received data for id={}",
        claims.sub,
        payload.id
    );

    HttpResponse::Ok().json(serde_json::json!({
        "success": true,
        "data": {
            "received": payload.into_inner(),
            "processed_by": claims.sub,
            "processed_at": chrono::Utc::now().timestamp()
        }
    }))
}
