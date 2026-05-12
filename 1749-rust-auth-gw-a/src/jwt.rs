use chrono::{Duration, Utc};
use jsonwebtoken::{
    decode, encode, DecodingKey, EncodingKey, Header, TokenData, Validation,
};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub sub: String,
    pub user_id: String,
    pub username: String,
    pub role: String,
    pub exp: i64,
    pub iat: i64,
    pub jti: String,
}

#[derive(Clone)]
pub struct JwtService {
    encoding_key: EncodingKey,
    decoding_key: DecodingKey,
    access_token_expiry: Duration,
    refresh_threshold: Duration,
}

impl JwtService {
    pub fn new(secret: &str) -> Self {
        Self {
            encoding_key: EncodingKey::from_secret(secret.as_bytes()),
            decoding_key: DecodingKey::from_secret(secret.as_bytes()),
            access_token_expiry: Duration::minutes(60),
            refresh_threshold: Duration::minutes(5),
        }
    }

    pub fn generate_token(
        &self,
        user_id: &str,
        username: &str,
        role: &str,
    ) -> Result<String, jsonwebtoken::errors::Error> {
        let now = Utc::now();
        let claims = Claims {
            sub: user_id.to_string(),
            user_id: user_id.to_string(),
            username: username.to_string(),
            role: role.to_string(),
            iat: now.timestamp(),
            exp: (now + self.access_token_expiry).timestamp(),
            jti: Uuid::new_v4().to_string(),
        };

        encode(&Header::default(), &claims, &self.encoding_key)
    }

    pub fn validate_token(&self, token: &str) -> Result<TokenData<Claims>, jsonwebtoken::errors::Error> {
        let validation = Validation::default();
        decode::<Claims>(token, &self.decoding_key, &validation)
    }

    pub fn should_refresh(&self, claims: &Claims) -> bool {
        let now = Utc::now().timestamp();
        (claims.exp - now) <= self.refresh_threshold.num_seconds()
    }

    pub fn refresh_token(&self, claims: &Claims) -> Result<String, jsonwebtoken::errors::Error> {
        self.generate_token(&claims.user_id, &claims.username, &claims.role)
    }

    pub fn get_jti(&self, claims: &Claims) -> String {
        claims.jti.clone()
    }

    pub fn get_expiry(&self, claims: &Claims) -> i64 {
        claims.exp
    }
}

impl Default for JwtService {
    fn default() -> Self {
        let default_secret = "default-jwt-secret-change-in-production-please";
        Self::new(default_secret)
    }
}
