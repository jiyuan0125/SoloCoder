use std::future::{ready, Ready};

use actix_web::{
    body::EitherBody,
    dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform},
    Error, HttpResponse,
};
use futures_util::future::LocalBoxFuture;

use crate::limiter::{LimitDimension, RateLimiter};

pub struct RateLimitMiddleware {
    limiter: RateLimiter,
}

impl RateLimitMiddleware {
    pub fn new(limiter: RateLimiter) -> Self {
        RateLimitMiddleware { limiter }
    }
}

impl<S, B> Transform<S, ServiceRequest> for RateLimitMiddleware
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<EitherBody<B>>;
    type Error = Error;
    type Transform = RateLimitService<S>;
    type InitError = ();
    type Future = Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        ready(Ok(RateLimitService {
            service,
            limiter: self.limiter.clone(),
        }))
    }
}

pub struct RateLimitService<S> {
    service: S,
    limiter: RateLimiter,
}

impl<S, B> Service<ServiceRequest> for RateLimitService<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<EitherBody<B>>;
    type Error = Error;
    type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;

    forward_ready!(service);

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let path = req.path().to_string();
        let ip = req
            .peer_addr()
            .map(|a| a.ip().to_string())
            .unwrap_or_else(|| "unknown".to_string());
        let user = req
            .headers()
            .get("X-User-Id")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("anonymous")
            .to_string();

        let limiter = self.limiter.clone();
        let rules = limiter.list_rules();

        let mut allowed = true;
        for rule in rules {
            if !rule.enabled {
                continue;
            }
            let key = match rule.dimension {
                LimitDimension::Ip => format!("ip:{}", ip),
                LimitDimension::Path => format!("path:{}", path),
                LimitDimension::IpPath => format!("ippath:{}:{}", ip, path),
                LimitDimension::User => format!("user:{}", user),
            };
            if !limiter.check(&key) {
                allowed = false;
                break;
            }
        }

        let fut = self.service.call(req);
        Box::pin(async move {
            if !allowed {
                let (req, _) = fut.await?.into_parts();
                let resp = HttpResponse::TooManyRequests()
                    .content_type("application/json")
                    .body(r#"{"error":"Too Many Requests"}"#);
                Ok(ServiceResponse::new(req, resp).map_into_right_body())
            } else {
                Ok(fut.await?.map_into_left_body())
            }
        })
    }
}
