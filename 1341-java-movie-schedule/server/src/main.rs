use std::env;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::Mutex;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::get,
    Json, Router,
};
use serde_json::json;
use uuid::Uuid;

use movie_schedule_core::{
    Scheduler, Schedule, ScheduleCreate, ScheduleUpdate, ScheduleError,
};

type AppState = Arc<Mutex<Scheduler>>;

#[tokio::main]
async fn main() {
    let port = env::args()
        .nth(1)
        .or_else(|| env::var("PORT").ok())
        .unwrap_or_else(|| "8100".to_string());
    
    let addr: SocketAddr = format!("127.0.0.1:{}", port)
        .parse()
        .expect("无效的端口号");

    let scheduler = Scheduler::new();
    let state = Arc::new(Mutex::new(scheduler));

    let app = Router::new()
        .route("/", get(root))
        .route("/api/halls", get(list_halls))
        .route("/api/movies", get(list_movies))
        .route("/api/schedules", get(list_schedules))
        .route("/api/schedules", axum::routing::post(create_schedule))
        .route("/api/schedules/:id", get(get_schedule))
        .route("/api/schedules/:id", axum::routing::put(update_schedule))
        .route("/api/schedules/:id", axum::routing::delete(delete_schedule))
        .with_state(state);

    println!("影院排片管理系统服务端已启动");
    println!("监听地址: {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn root() -> &'static str {
    "影院排片管理系统 API"
}

async fn list_halls(State(state): State<AppState>) -> impl IntoResponse {
    let scheduler = state.lock().await;
    let halls = scheduler.get_halls();
    Json(halls.iter().cloned().cloned().collect::<Vec<_>>()).into_response()
}

async fn list_movies(State(state): State<AppState>) -> impl IntoResponse {
    let scheduler = state.lock().await;
    let movies = scheduler.get_movies();
    Json(movies.iter().cloned().cloned().collect::<Vec<_>>()).into_response()
}

async fn list_schedules(State(state): State<AppState>) -> impl IntoResponse {
    let scheduler = state.lock().await;
    let schedules: Vec<Schedule> = scheduler.get_schedules().into_iter().cloned().collect();
    Json(schedules).into_response()
}

async fn get_schedule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let scheduler = state.lock().await;
    match scheduler.get_schedule(id) {
        Some(schedule) => Json(schedule.clone()).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(json!({ "error": "排片不存在" })),
        ).into_response(),
    }
}

async fn create_schedule(
    State(state): State<AppState>,
    Json(create): Json<ScheduleCreate>,
) -> impl IntoResponse {
    let mut scheduler = state.lock().await;
    match scheduler.create_schedule(create) {
        Ok(schedule) => (
            StatusCode::CREATED,
            Json(schedule),
        ).into_response(),
        Err(e) => match e {
            ScheduleError::Conflict(conflict) => {
                let error_msg = format!("时间冲突：与排片 {} 重叠", conflict);
                (
                    StatusCode::CONFLICT,
                    Json(json!({ "error": error_msg, "conflict": conflict })),
                ).into_response()
            }
            ScheduleError::PrimeTimeLimitExceeded => (
                StatusCode::BAD_REQUEST,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
            ScheduleError::HallNotFound => (
                StatusCode::BAD_REQUEST,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
            ScheduleError::MovieNotFound => (
                StatusCode::BAD_REQUEST,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
            _ => (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
        },
    }
}

async fn update_schedule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(update): Json<ScheduleUpdate>,
) -> impl IntoResponse {
    let mut scheduler = state.lock().await;
    match scheduler.update_schedule(id, update) {
        Ok(schedule) => Json(schedule).into_response(),
        Err(e) => match e {
            ScheduleError::ScheduleNotFound => (
                StatusCode::NOT_FOUND,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
            ScheduleError::Conflict(conflict) => {
                let error_msg = format!("时间冲突：与排片 {} 重叠", conflict);
                (
                    StatusCode::CONFLICT,
                    Json(json!({ "error": error_msg, "conflict": conflict })),
                ).into_response()
            }
            ScheduleError::PrimeTimeLimitExceeded => (
                StatusCode::BAD_REQUEST,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
            ScheduleError::HallNotFound => (
                StatusCode::BAD_REQUEST,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
            ScheduleError::MovieNotFound => (
                StatusCode::BAD_REQUEST,
                Json(json!({ "error": e.to_string() })),
            ).into_response(),
        },
    }
}

async fn delete_schedule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let mut scheduler = state.lock().await;
    match scheduler.delete_schedule(id) {
        Ok(_) => StatusCode::NO_CONTENT.into_response(),
        Err(_) => (
            StatusCode::NOT_FOUND,
            Json(json!({ "error": "排片不存在" })),
        ).into_response(),
    }
}
