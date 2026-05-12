use std::env;

use actix_web::{web, App, HttpServer};
use log::info;

mod admin;
mod app_state;
mod forward;
mod health_check;
mod models;

use crate::app_state::AppState;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .expect("PORT 必须是有效的端口号");

    let app_state = AppState::new();
    let health_state = app_state.clone();

    tokio::spawn(async move {
        health_check::run_health_checks(health_state).await;
    });

    info!("反向代理服务启动于端口 {}", port);

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            .route("/admin/add", web::post().to(admin::add_backend))
            .route("/admin/delete", web::post().to(admin::delete_backend))
            .route("/admin/status", web::get().to(admin::list_status))
            .default_service(web::route().to(forward::forward_request))
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
