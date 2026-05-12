use actix_web::dev::{Service, ServiceRequest, ServiceResponse, Transform};
use actix_web::{Error, HttpMessage};
use futures_util::future::{ok, LocalBoxFuture, Ready};
use std::rc::Rc;
use std::task::{Context, Poll};

use crate::auth::{validate_token, AuthState, Claims, TokenType};
use crate::errors::{unauthorized, ErrorResponse};

pub struct AuthMiddleware;

impl<S, B> Transform<S, ServiceRequest> for AuthMiddleware
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error> + 'static,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type Transform = AuthMiddlewareService<S>;
    type InitError = ();
    type Future = Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        ok(AuthMiddlewareService {
            service: Rc::new(service),
        })
    }
}

pub struct AuthMiddlewareService<S> {
    service: Rc<S>,
}

impl<S, B> Service<ServiceRequest> for AuthMiddlewareService<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error> + 'static,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;

    fn poll_ready(&self, ctx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.service.poll_ready(ctx)
    }

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let service = self.service.clone();

        Box::pin(async move {
            let auth_header = req.headers().get("Authorization");

            let token = match auth_header {
                Some(h) => {
                    let header_val = h.to_str().unwrap_or("");
                    if header_val.starts_with("Bearer ") {
                        Some(header_val[7..].to_string())
                    } else {
                        None
                    }
                }
                None => None,
            };

            let claims: Option<Claims> = match token {
                Some(t) => {
                    let auth_data_opt = req.app_data::<actix_web::web::Data<AuthState>>();
                    match auth_data_opt {
                        Some(auth_data) => {
                            match validate_token(&t, auth_data, TokenType::Access) {
                                Ok(c) => {
                                    log::info!(
                                        "AUTH_SUCCESS: user_id={}, method={}, path={}",
                                        c.sub,
                                        req.method(),
                                        req.path()
                                    );
                                    Some(c)
                                }
                                Err(e) => {
                                    log::warn!(
                                        "AUTH_FAILED: method={}, path={}, reason={:?}",
                                        req.method(),
                                        req.path(),
                                        e
                                    );
                                    let err_resp = ErrorResponse::from(e);
                                    return Err(actix_web::error::InternalError::from_response(
                                        "",
                                        actix_web::HttpResponse::Unauthorized().json(err_resp),
                                    )
                                    .into());
                                }
                            }
                        }
                        None => {
                            return Err(actix_web::error::InternalError::from_response(
                                "",
                                actix_web::HttpResponse::InternalServerError().json(
                                    crate::errors::ErrorResponse {
                                        success: false,
                                        error: crate::errors::ErrorDetail {
                                            code: "CONFIG_ERROR".to_string(),
                                            message: "Auth state not configured".to_string(),
                                        },
                                    },
                                ),
                            )
                            .into());
                        }
                    }
                }
                None => {
                    log::warn!(
                        "AUTH_MISSING: method={}, path={}",
                        req.method(),
                        req.path()
                    );
                    return Err(actix_web::error::InternalError::from_response(
                        "",
                        unauthorized("MISSING_TOKEN", "No authorization token provided"),
                    )
                    .into());
                }
            };

            if let Some(c) = claims {
                req.extensions_mut().insert(c);
            }

            service.call(req).await
        })
    }
}
