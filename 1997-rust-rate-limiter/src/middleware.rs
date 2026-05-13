use crate::store::RateLimiterEngine;
use std::future::Future;
use std::net::SocketAddr;
use std::pin::Pin;
use std::sync::Arc;
use std::task::{Context, Poll};
use tower::Service;

use axum::body::Body;
use axum::http::{Request, Response, StatusCode};
use axum::response::IntoResponse;
use http::HeaderValue;

#[derive(Clone)]
pub struct RateLimitMiddleware<S> {
    inner: S,
    engine: Arc<RateLimiterEngine>,
}

impl<S> RateLimitMiddleware<S> {
    pub fn new(inner: S, engine: Arc<RateLimiterEngine>) -> Self {
        RateLimitMiddleware { inner, engine }
    }
}

impl<S> Service<Request<Body>> for RateLimitMiddleware<S>
where
    S: Service<Request<Body>, Response = Response<Body>> + Clone + Send + 'static,
    S::Future: Send + 'static,
    S::Error: std::fmt::Display + Send,
{
    type Response = S::Response;
    type Error = S::Error;
    type Future = Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>> + Send>>;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, req: Request<Body>) -> Self::Future {
        let engine = self.engine.clone();
        let path = req.uri().path().to_string();
        let ip = extract_client_ip(&req);

        let result = engine.check_request(&path, ip);

        let clone = self.inner.clone();
        let mut inner = std::mem::replace(&mut self.inner, clone);

        Box::pin(async move {
            if !result.allowed {
                let mut response = (
                    StatusCode::TOO_MANY_REQUESTS,
                    "Too Many Requests",
                )
                    .into_response();

                let remaining = result.remaining.to_string();
                if let Ok(header_value) = HeaderValue::from_str(&remaining) {
                    response
                        .headers_mut()
                        .insert("X-RateLimit-Remaining", header_value);
                }

                if let Some(retry_after) = result.retry_after {
                    let retry_secs = retry_after.as_secs().max(1).to_string();
                    if let Ok(header_value) = HeaderValue::from_str(&retry_secs) {
                        response.headers_mut().insert("Retry-After", header_value);
                    }
                }

                return Ok(response);
            }

            inner.call(req).await
        })
    }
}

fn extract_client_ip(req: &Request<Body>) -> Option<std::net::IpAddr> {
    if let Some(forwarded) = req.headers().get("X-Forwarded-For") {
        if let Ok(s) = forwarded.to_str() {
            if let Some(first) = s.split(',').next() {
                if let Ok(ip) = first.trim().parse::<std::net::IpAddr>() {
                    return Some(ip);
                }
            }
        }
    }

    if let Some(real_ip) = req.headers().get("X-Real-IP") {
        if let Ok(s) = real_ip.to_str() {
            if let Ok(ip) = s.parse::<std::net::IpAddr>() {
                return Some(ip);
            }
        }
    }

    if let Some(addr) = req.extensions().get::<SocketAddr>() {
        return Some(addr.ip());
    }

    None
}

pub fn make_rate_limit_layer(
    engine: Arc<RateLimiterEngine>,
) -> impl tower::Layer<
    axum::routing::Route,
    Service = RateLimitMiddleware<axum::routing::Route>,
> + Clone {
    tower::layer::layer_fn(move |service| {
        RateLimitMiddleware::new(service, engine.clone())
    })
}
