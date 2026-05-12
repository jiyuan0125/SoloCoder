use chrono::{Duration, Utc};
use jsonwebtoken::{DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub sub: String,
    pub jti: String,
    pub exp: usize,
    pub iat: usize,
    pub token_type: TokenType,
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq)]
pub enum TokenType {
    Access,
    Refresh,
}

#[derive(Clone)]
pub struct AuthState {
    pub encoding_key: EncodingKey,
    pub decoding_key: DecodingKey,
    pub refresh_tokens: Arc<Mutex<HashMap<String, String>>>,
    pub jwt_secret: String,
}

impl AuthState {
    pub fn new(secret: String) -> Self {
        AuthState {
            encoding_key: EncodingKey::from_secret(secret.as_bytes()),
            decoding_key: DecodingKey::from_secret(secret.as_bytes()),
            refresh_tokens: Arc::new(Mutex::new(HashMap::new())),
            jwt_secret: secret,
        }
    }
}

#[derive(Debug, Serialize)]
pub struct TokenPair {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_in: usize,
}

const ACCESS_TOKEN_DURATION_MINUTES: i64 = 15;
const REFRESH_TOKEN_DURATION_DAYS: i64 = 7;

pub fn create_token_pair(
    user_id: &str,
    auth_state: &AuthState,
) -> Result<TokenPair, jsonwebtoken::errors::Error> {
    let access_jti = uuid::Uuid::new_v4().to_string();
    let refresh_jti = uuid::Uuid::new_v4().to_string();

    let now = Utc::now();

    let access_claims = Claims {
        sub: user_id.to_string(),
        jti: access_jti.clone(),
        exp: (now + Duration::minutes(ACCESS_TOKEN_DURATION_MINUTES)).timestamp() as usize,
        iat: now.timestamp() as usize,
        token_type: TokenType::Access,
    };

    let refresh_claims = Claims {
        sub: user_id.to_string(),
        jti: refresh_jti.clone(),
        exp: (now + Duration::days(REFRESH_TOKEN_DURATION_DAYS)).timestamp() as usize,
        iat: now.timestamp() as usize,
        token_type: TokenType::Refresh,
    };

    let access_token = jsonwebtoken::encode(
        &Header::default(),
        &access_claims,
        &auth_state.encoding_key,
    )?;

    let refresh_token = jsonwebtoken::encode(
        &Header::default(),
        &refresh_claims,
        &auth_state.encoding_key,
    )?;

    if let Ok(mut rt) = auth_state.refresh_tokens.lock() {
        rt.insert(refresh_jti, user_id.to_string());
    }

    Ok(TokenPair {
        access_token,
        refresh_token,
        expires_in: ACCESS_TOKEN_DURATION_MINUTES as usize * 60,
    })
}

pub fn validate_token(
    token: &str,
    auth_state: &AuthState,
    expected_type: TokenType,
) -> Result<Claims, AuthError> {
    let claims = jsonwebtoken::decode::<Claims>(
        token,
        &auth_state.decoding_key,
        &Validation::default(),
    )
    .map_err(|e| match e.kind() {
        jsonwebtoken::errors::ErrorKind::ExpiredSignature => AuthError::TokenExpired,
        _ => AuthError::InvalidToken,
    })?;

    if claims.claims.token_type != expected_type {
        return Err(AuthError::InvalidToken);
    }

    if expected_type == TokenType::Refresh {
        if let Ok(rt) = auth_state.refresh_tokens.lock() {
            if !rt.contains_key(&claims.claims.jti) {
                return Err(AuthError::InvalidToken);
            }
        }
    }

    Ok(claims.claims)
}

pub fn revoke_refresh_token(auth_state: &AuthState, jti: &str) {
    if let Ok(mut rt) = auth_state.refresh_tokens.lock() {
        rt.remove(jti);
    }
}

#[derive(Debug)]
pub enum AuthError {
    TokenExpired,
    InvalidToken,
    MissingToken,
    InvalidCredentials,
}
