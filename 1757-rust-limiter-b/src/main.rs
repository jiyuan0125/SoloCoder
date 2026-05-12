use std::env;

use actix_web::{web, App, HttpServer};

use pay_limiter::{handlers, limiter::RateLimiter, middleware::RateLimitMiddleware};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(8080);

    let limiter = RateLimiter::new();
    let limiter_data = web::Data::new(limiter.clone());

    HttpServer::new(move || {
        App::new()
            .app_data(limiter_data.clone())
            .wrap(RateLimitMiddleware::new(limiter.clone()))
            .route("/", web::get().to(handlers::index))
            .route("/rules", web::get().to(handlers::get_rules))
            .route("/rules", web::put().to(handlers::upsert_rule))
            .route("/rules/{id}", web::get().to(handlers::get_rule))
            .route("/rules/{id}", web::delete().to(handlers::delete_rule))
            .route("/stats", web::get().to(handlers::get_stats))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
