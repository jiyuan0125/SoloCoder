use actix_web::dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform};
use actix_web::http::StatusCode;
use actix_web::{Error, HttpResponse};
use futures::future::LocalBoxFuture;
use std::future::{ready, Ready};
use std::sync::Arc;

use crate::alert::{AlertEvent, AlertManager};
use crate::sliding_window::SlidingWindowLimiter;
use crate::token_bucket::TokenBucketLimiter;

pub struct RateLimiterMiddleware {
    sliding_window: Arc<SlidingWindowLimiter>,
    token_bucket: Arc<TokenBucketLimiter>,
    alert_manager: Arc<AlertManager>,
}

impl RateLimiterMiddleware {
    pub fn new(
        sliding_window: Arc<SlidingWindowLimiter>,
        token_bucket: Arc<TokenBucketLimiter>,
        alert_manager: Arc<AlertManager>,
    ) -> Self {
        RateLimiterMiddleware {
            sliding_window,
            token_bucket,
            alert_manager,
        }
    }
}

impl<S, B> Transform<S, ServiceRequest> for RateLimiterMiddleware
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type InitError = ();
    type Transform = RateLimiterService<S>;
    type Future = Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        ready(Ok(RateLimiterService {
            service,
            sliding_window: self.sliding_window.clone(),
            token_bucket: self.token_bucket.clone(),
            alert_manager: self.alert_manager.clone(),
        }))
    }
}

pub struct RateLimiterService<S> {
    service: S,
    sliding_window: Arc<SlidingWindowLimiter>,
    token_bucket: Arc<TokenBucketLimiter>,
    alert_manager: Arc<AlertManager>,
}

impl<S, B> Service<ServiceRequest> for RateLimiterService<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;

    forward_ready!(service);

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let path = req.path().to_string();
        let ip = get_client_ip(&req);

        let sliding_allowed = self.sliding_window.try_acquire(&ip);
        let token_allowed = self.token_bucket.try_acquire(&path, 1);

        if !sliding_allowed {
            let alert = AlertEvent {
                timestamp: chrono::Utc::now().to_rfc3339(),
                limit_type: "sliding_window".to_string(),
                key: ip.clone(),
                message: format!("IP {} exceeded sliding window rate limit", ip),
            };
            self.alert_manager.send_alert(alert);
            return Box::pin(ready(Err(actix_web::error::Error::from(
                actix_web::error::InternalError::from_response(
                    "",
                    HttpResponse::new(StatusCode::TOO_MANY_REQUESTS),
                ),
            ))));
        }

        if !token_allowed {
            let alert = AlertEvent {
                timestamp: chrono::Utc::now().to_rfc3339(),
                limit_type: "token_bucket".to_string(),
                key: path.clone(),
                message: format!("Path {} exceeded token bucket rate limit", path),
            };
            self.alert_manager.send_alert(alert);
            return Box::pin(ready(Err(actix_web::error::Error::from(
                actix_web::error::InternalError::from_response(
                    "",
                    HttpResponse::new(StatusCode::TOO_MANY_REQUESTS),
                ),
            ))));
        }

        let fut = self.service.call(req);
        Box::pin(async move {
            let res = fut.await?;
            Ok(res)
        })
    }
}

fn get_client_ip(req: &ServiceRequest) -> String {
    if let Some(forwarded) = req.headers().get("X-Forwarded-For") {
        if let Ok(ip_str) = forwarded.to_str() {
            if let Some(first_ip) = ip_str.split(',').next() {
                return first_ip.trim().to_string();
            }
        }
    }

    if let Some(real_ip) = req.headers().get("X-Real-IP") {
        if let Ok(ip_str) = real_ip.to_str() {
            return ip_str.to_string();
        }
    }

    req.peer_addr()
        .map(|addr| addr.ip().to_string())
        .unwrap_or_else(|| "unknown".to_string())
}
