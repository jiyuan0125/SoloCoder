mod auth;
mod errors;
mod handlers;
mod middleware;

use actix_web::{App, HttpServer};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let addr = format!("0.0.0.0:{}", port);

    let jwt_secret = std::env::var("JWT_SECRET").unwrap_or_else(|_| "super-secret-key-change-in-production".to_string());
    let auth_data = auth::AuthState::new(jwt_secret);

    log::info!("Starting auth gateway on {}", addr);

    HttpServer::new(move || {
        App::new()
            .app_data(actix_web::web::Data::new(auth_data.clone()))
            .service(handlers::login)
            .service(handlers::refresh)
            .service(
                actix_web::web::scope("/api")
                    .wrap(middleware::AuthMiddleware)
                    .service(handlers::protected)
                    .service(handlers::protected_data)
            )
    })
    .bind(&addr)?
    .run()
    .await
}
