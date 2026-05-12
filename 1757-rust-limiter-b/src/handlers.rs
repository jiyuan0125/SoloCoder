use actix_web::{web, HttpResponse, Responder};
use serde::Serialize;

use crate::limiter::{LimitCounter, LimitRule, RateLimiter};

#[derive(Serialize)]
struct StatsResponse {
    total_allowed: u64,
    total_rejected: u64,
    counters: Vec<LimitCounter>,
}

pub async fn get_rules(limiter: web::Data<RateLimiter>) -> impl Responder {
    let rules = limiter.list_rules();
    HttpResponse::Ok().json(rules)
}

pub async fn get_rule(path: web::Path<String>, limiter: web::Data<RateLimiter>) -> impl Responder {
    match limiter.get_rule(&path.into_inner()) {
        Some(rule) => HttpResponse::Ok().json(rule),
        None => HttpResponse::NotFound().json(serde_json::json!({"error":"rule not found"})),
    }
}

pub async fn upsert_rule(
    rule: web::Json<LimitRule>,
    limiter: web::Data<RateLimiter>,
) -> impl Responder {
    limiter.upsert_rule(rule.into_inner());
    HttpResponse::Ok().json(serde_json::json!({"ok":true}))
}

pub async fn delete_rule(
    path: web::Path<String>,
    limiter: web::Data<RateLimiter>,
) -> impl Responder {
    match limiter.delete_rule(&path.into_inner()) {
        Some(_) => HttpResponse::Ok().json(serde_json::json!({"ok":true})),
        None => HttpResponse::NotFound().json(serde_json::json!({"error":"rule not found"})),
    }
}

pub async fn get_stats(limiter: web::Data<RateLimiter>) -> impl Responder {
    let counters = limiter.list_counters();
    let total_allowed = counters.iter().map(|c| c.allowed).sum();
    let total_rejected = counters.iter().map(|c| c.rejected).sum();
    HttpResponse::Ok().json(StatsResponse {
        total_allowed,
        total_rejected,
        counters,
    })
}

pub async fn index() -> impl Responder {
    HttpResponse::Ok().json(serde_json::json!({"message":"ok"}))
}
