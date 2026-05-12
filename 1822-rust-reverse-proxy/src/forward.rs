use actix_web::{web, HttpRequest, HttpResponse, Responder};
use awc::Client;
use log::error;
use rand::Rng;
use std::collections::HashMap;
use url::Url;

use crate::app_state::AppState;

fn select_backend(
    backends: &[(String, u32, bool, bool, u64)],
) -> Option<String> {
    let available: Vec<&(String, u32, bool, bool, u64)> = backends
        .iter()
        .filter(|(_, weight, healthy, marked, _)| {
            *healthy && !*marked && *weight > 0
        })
        .collect();

    if available.is_empty() {
        return None;
    }

    let total_weight: u32 = available.iter().map(|(_, w, _, _, _)| *w).sum();
    if total_weight == 0 {
        return None;
    }

    let mut rng = rand::thread_rng();
    let mut roll = rng.gen_range(0..total_weight);

    for (addr, weight, _, _, _) in available.iter() {
        if roll < *weight {
            return Some(addr.clone());
        }
        roll -= *weight;
    }

    None
}

fn build_forward_url(
    backend_addr: &str,
    path: &str,
    query_string: &str,
    captures: HashMap<String, String>,
) -> String {
    let mut url = if backend_addr.starts_with("http://") || backend_addr.starts_with("https://") {
        backend_addr.to_string()
    } else {
        format!("http://{}", backend_addr)
    };

    if !path.is_empty() {
        url.push_str(path);
    }

    let mut params: Vec<(String, String)> = Vec::new();

    for (k, v) in captures {
        params.push((k, v));
    }

    if !query_string.is_empty() {
        if let Ok(parsed_url) = Url::parse(&format!("http://dummy?{}", query_string)) {
            for (k, v) in parsed_url.query_pairs() {
                params.push((k.into_owned(), v.into_owned()));
            }
        }
    }

    if !params.is_empty() {
        url.push('?');
        let query: Vec<String> = params
            .iter()
            .map(|(k, v)| format!("{}={}", k, v))
            .collect();
        url.push_str(&query.join("&"));
    }

    url
}

pub async fn forward_request(
    req: HttpRequest,
    body: web::Bytes,
    state: web::Data<AppState>,
) -> impl Responder {
    let path = req.path().to_string();
    let query = req.query_string().to_string();

    let (pattern, backends) = match state.find_matching_route(&path).await {
        Some((p, b)) => (p, b),
        None => {
            return HttpResponse::NotFound().body("未找到匹配的路由");
        }
    };

    let captures = match state.get_captures(&pattern, &path).await {
        Some(c) => c,
        None => HashMap::new(),
    };

    let backend_addr = match select_backend(&backends) {
        Some(addr) => addr,
        None => {
            return HttpResponse::ServiceUnavailable().body("无可用后端");
        }
    };

    let forward_url = build_forward_url(&backend_addr, &path, &query, captures);

    state.increment_request(&pattern, &backend_addr).await;

    let client = Client::default();
    let method = req.method().clone();

    let result: Result<HttpResponse, ()> = async move {
        let mut request = client.request(method.clone(), &forward_url);

        for (key, value) in req.headers() {
            if key.as_str().to_lowercase() != "host" {
                request = request.insert_header((key.as_str(), value.as_bytes()));
            }
        }

        let response = request.send_body(body).await;

        match response {
            Ok(mut res) => {
                let mut client_resp = HttpResponse::build(res.status());
                for (key, value) in res.headers() {
                    let lower = key.as_str().to_lowercase();
                    if lower != "connection" && lower != "content-length" {
                        client_resp.insert_header((key.as_str(), value.as_bytes()));
                    }
                }

                let body_bytes = match res.body().await {
                    Ok(b) => b,
                    Err(e) => {
                        error!("读取响应体失败: {}", e);
                        return Ok(HttpResponse::InternalServerError().body("读取响应体失败"));
                    }
                };

                Ok(client_resp.body(body_bytes))
            }
            Err(e) => {
                error!("转发请求失败: {}", e);
                Ok(HttpResponse::BadGateway().body("后端服务不可用"))
            }
        }
    }
    .await;

    state.decrement_pending(&pattern, &backend_addr).await;

    result.unwrap_or_else(|_| HttpResponse::InternalServerError().finish())
}
