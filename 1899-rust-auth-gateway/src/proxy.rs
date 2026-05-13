use axum::body::Body;
use axum::response::Response;
use axum::http::{HeaderMap, Method, Uri};
use reqwest::Client;
use tracing::error;

use crate::auth::AuthenticatedUser;

pub fn build_proxy_client() -> Client {
    Client::builder()
        .build()
        .expect("Failed to build proxy client")
}

pub async fn proxy_request(
    method: Method,
    uri: &Uri,
    mut headers: HeaderMap,
    body_bytes: Option<Vec<u8>>,
    backend_url: &str,
    user: Option<&AuthenticatedUser>,
) -> Result<Response, String> {
    let client = build_proxy_client();
    let path = uri.path();
    let query = uri.query().map(|q| format!("?{}", q)).unwrap_or_default();

    let target_url = if backend_url.ends_with('/') {
        let trimmed = &backend_url[..backend_url.len() - 1];
        format!("{}{}{}", trimmed, path, query)
    } else {
        format!("{}{}{}", backend_url, path, query)
    };

    headers.remove(axum::http::header::HOST);

    if let Some(u) = user {
        u.inject_headers(&mut headers);
    }

    let mut req_builder = client.request(method.clone(), &target_url);
    for (name, value) in headers.iter() {
        let name_str = name.as_str();
        let value_bytes = value.as_bytes();
        req_builder = req_builder.header(name_str, value_bytes);
    }

    let resp = match body_bytes {
        Some(bytes) => {
            if !bytes.is_empty() {
                req_builder.body(bytes).send().await
            } else {
                req_builder.send().await
            }
        }
        None => req_builder.send().await,
    };

    match resp {
        Ok(backend_resp) => {
            let status = backend_resp.status();
            let resp_headers = backend_resp.headers().clone();
            let resp_body = match backend_resp.bytes().await {
                Ok(b) => b,
                Err(e) => {
                    error!("Failed to read backend response body: {}", e);
                    return Ok(Response::builder()
                        .status(502)
                        .body(Body::from(format!("Bad Gateway: {}", e)))
                        .unwrap());
                }
            };
            let mut builder = Response::builder().status(status);
            for (name, value) in resp_headers.iter() {
                let name_str = name.as_str();
                let value_bytes = value.as_bytes();
                builder = builder.header(name_str, value_bytes);
            }
            Ok(builder.body(Body::from(resp_body)).unwrap_or_else(|e| {
                error!("Failed to build response: {}", e);
                Response::builder()
                    .status(502)
                    .body(Body::from("Bad Gateway"))
                    .unwrap()
            }))
        }
        Err(e) => {
            error!("Proxy error: {}", e);
            Ok(Response::builder()
                .status(502)
                .body(Body::from(format!("Bad Gateway: {}", e)))
                .unwrap())
        }
    }
}
