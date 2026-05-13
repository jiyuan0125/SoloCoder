use axum::http::{HeaderMap, HeaderValue};
use jsonwebtoken::{decode, decode_header, DecodingKey, Validation};
use serde::{Deserialize, Serialize};
use std::time::{SystemTime, UNIX_EPOCH};
use tracing::info;

use crate::config::AuthMethod;
use crate::store::{ApiKeyMetadata, AppState};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,
    pub exp: u64,
    #[serde(default)]
    pub roles: Vec<String>,
    #[serde(default)]
    pub permissions: Vec<String>,
    #[serde(flatten)]
    pub extra: serde_json::Map<String, serde_json::Value>,
}

#[derive(Debug, Clone)]
pub struct AuthenticatedUser {
    pub method: AuthMethod,
    pub identifier: String,
    pub service_name: Option<String>,
    pub permissions: Vec<String>,
    pub roles: Vec<String>,
    pub extra_headers: Vec<(String, String)>,
}

impl AuthenticatedUser {
    pub fn from_api_key(key: String, metadata: ApiKeyMetadata) -> Self {
        let key_clone = key.clone();
        Self {
            method: AuthMethod::ApiKey,
            identifier: key,
            service_name: Some(metadata.service_name),
            permissions: metadata.permissions,
            roles: vec![],
            extra_headers: vec![
                ("X-Auth-Method".to_string(), "api-key".to_string()),
                ("X-Auth-Identifier".to_string(), key_clone),
            ],
        }
    }

    pub fn from_jwt(claims: Claims) -> Self {
        let mut headers = vec![
            ("X-Auth-Method".to_string(), "jwt".to_string()),
            ("X-Auth-User".to_string(), claims.sub.clone()),
        ];
        if let Ok(json) = serde_json::to_string(&claims.extra) {
            headers.push(("X-Auth-Claims".to_string(), json));
        }
        Self {
            method: AuthMethod::Jwt,
            identifier: claims.sub,
            service_name: None,
            permissions: claims.permissions,
            roles: claims.roles,
            extra_headers: headers,
        }
    }

    pub fn is_admin(&self, admin_user: &str) -> bool {
        match self.method {
            AuthMethod::Jwt => self.roles.contains(&"admin".to_string()) || self.identifier == admin_user,
            AuthMethod::ApiKey => false,
        }
    }

    pub fn inject_headers(&self, headers: &mut HeaderMap) {
        if let Some(ref svc) = self.service_name {
            if let Ok(v) = HeaderValue::from_str(svc) {
                headers.insert("X-Auth-Service", v);
            }
        }
        if !self.permissions.is_empty() {
            if let Ok(v) = HeaderValue::from_str(&self.permissions.join(",")) {
                headers.insert("X-Auth-Permissions", v);
            }
        }
        if !self.roles.is_empty() {
            if let Ok(v) = HeaderValue::from_str(&self.roles.join(",")) {
                headers.insert("X-Auth-Roles", v);
            }
        }
        for (k, v) in &self.extra_headers {
            if let Ok(val) = HeaderValue::from_str(v) {
                if let Ok(name) = axum::http::HeaderName::from_bytes(k.as_bytes()) {
                    headers.insert(name, val);
                }
            }
        }
    }
}

pub enum AuthResult {
    Success(AuthenticatedUser),
    NoCredentials,
    InvalidCredentials,
    Expired,
    MethodNotAllowed,
}

pub async fn authenticate(
    state: &AppState,
    headers: &HeaderMap,
    allowed_methods: &[AuthMethod],
) -> AuthResult {
    let api_key_present = extract_api_key(headers).is_some();
    let jwt_present = extract_bearer_token(headers).is_some();

    if !api_key_present && !jwt_present {
        return AuthResult::NoCredentials;
    }

    if let Some(key) = extract_api_key(headers) {
        if allowed_methods.contains(&AuthMethod::ApiKey) {
            if let Some(meta) = state.api_keys.get(&key).await {
                info!("API Key authentication successful for service: {}", meta.service_name);
                return AuthResult::Success(AuthenticatedUser::from_api_key(key, meta));
            }
        }
    }

    if let Some(token) = extract_bearer_token(headers) {
        if allowed_methods.contains(&AuthMethod::Jwt) {
            return verify_jwt(&state.jwt_keys, &token).await;
        }
    }

    if (api_key_present && !allowed_methods.contains(&AuthMethod::ApiKey))
        || (jwt_present && !allowed_methods.contains(&AuthMethod::Jwt))
    {
        return AuthResult::MethodNotAllowed;
    }

    AuthResult::InvalidCredentials
}

fn extract_api_key(headers: &HeaderMap) -> Option<String> {
    headers
        .get("X-API-Key")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string())
}

fn extract_bearer_token(headers: &HeaderMap) -> Option<String> {
    headers
        .get("Authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| {
            if s.to_lowercase().starts_with("bearer ") {
                Some(s[7..].trim().to_string())
            } else {
                None
            }
        })
}

async fn verify_jwt(key_store: &crate::store::JwtKeyStore, token: &str) -> AuthResult {
    let now_secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs())
        .unwrap_or(0);

    let secrets = key_store.get_all_valid().await;
    let alg = match decode_header(token) {
        Ok(h) => h.alg,
        Err(_) => return AuthResult::InvalidCredentials,
    };

    for secret in secrets {
        let decoding_key = DecodingKey::from_secret(secret.as_bytes());
        let mut validation = Validation::new(alg);
        validation.validate_exp = true;
        validation.leeway = 0;

        match decode::<Claims>(token, &decoding_key, &validation) {
            Ok(data) => {
                if data.claims.exp <= now_secs {
                    info!("JWT token expired");
                    return AuthResult::Expired;
                }
                info!("JWT authentication successful for user: {}", data.claims.sub);
                return AuthResult::Success(AuthenticatedUser::from_jwt(data.claims));
            }
            Err(err) => match err.kind() {
                jsonwebtoken::errors::ErrorKind::ExpiredSignature => {
                    info!("JWT token expired: {:?}", err);
                    return AuthResult::Expired;
                }
                _ => continue,
            },
        }
    }

    info!("JWT verification failed");
    AuthResult::InvalidCredentials
}

pub fn build_www_authenticate(methods: &[AuthMethod]) -> String {
    let mut schemes = Vec::new();
    if methods.contains(&AuthMethod::Jwt) {
        schemes.push("Bearer");
    }
    if methods.contains(&AuthMethod::ApiKey) {
        schemes.push("ApiKey");
    }
    schemes.join(", ")
}
