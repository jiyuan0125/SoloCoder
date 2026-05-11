use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post, put},
    Json, Router,
};
use clap::Parser;
use process_route_core::error::ProcessRouteError;
use process_route_core::models::*;
use process_route_core::{InMemoryStore, ProcessRouteManager};
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::cors::{Any, CorsLayer};
use tokio::net::TcpListener;
use tracing_subscriber::EnvFilter;

#[derive(Parser, Debug, Clone)]
#[command(author, version, about, long_about = None)]
struct Config {
    #[arg(long, env = "PROCESS_ROUTE_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "PROCESS_ROUTE_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    manager: ProcessRouteManager,
}

type SharedState = Arc<AppState>;

#[derive(Debug, Serialize, Deserialize)]
struct CreateRouteRequest {
    name: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct AddVersionRequest {
    version: String,
    processes: Vec<ProcessDefinition>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CreateTaskRequest {
    name: String,
    route_id: String,
    version: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ProcessActionResponse {
    task: ProductionTask,
    warning: Option<Warning>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

fn error_to_response(err: ProcessRouteError) -> (StatusCode, Json<ErrorResponse>) {
    let status = match &err {
        ProcessRouteError::RouteNotFound(_)
        | ProcessRouteError::VersionNotFound(_)
        | ProcessRouteError::ProcessNotFound(_)
        | ProcessRouteError::TaskNotFound(_) => StatusCode::NOT_FOUND,
        ProcessRouteError::CircularDependency
        | ProcessRouteError::PredecessorsNotCompleted(_)
        | ProcessRouteError::ProcessNotWaiting
        | ProcessRouteError::ProcessNotInProgress
        | ProcessRouteError::ProcessAlreadyCompleted
        | ProcessRouteError::VersionAlreadyExists(_)
        | ProcessRouteError::TaskVersionLocked(_)
        | ProcessRouteError::InvalidOperation(_) => StatusCode::BAD_REQUEST,
        ProcessRouteError::CriticalProcessPaused(_) => StatusCode::OK,
        ProcessRouteError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };
    (status, Json(ErrorResponse { error: err.to_string() }))
}

async fn create_route(
    State(state): State<SharedState>,
    Json(req): Json<CreateRouteRequest>,
) -> impl IntoResponse {
    let route = state.manager.create_route(req.name);
    (StatusCode::CREATED, Json(route))
}

async fn list_routes(State(state): State<SharedState>) -> impl IntoResponse {
    let routes = state.manager.list_routes();
    Json(routes)
}

async fn get_route(
    State(state): State<SharedState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.manager.get_route(&id) {
        Ok(route) => (StatusCode::OK, Json(serde_json::to_value(route).unwrap())).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn delete_route(
    State(state): State<SharedState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.manager.delete_route(&id) {
        Ok(_) => StatusCode::NO_CONTENT.into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn add_version(
    State(state): State<SharedState>,
    Path(route_id): Path<String>,
    Json(req): Json<AddVersionRequest>,
) -> impl IntoResponse {
    match state.manager.add_version(&route_id, req.version, req.processes) {
        Ok(route) => (StatusCode::OK, Json(route)).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn get_critical_path(
    State(state): State<SharedState>,
    Path((route_id, version)): Path<(String, String)>,
) -> impl IntoResponse {
    match state.manager.get_critical_path_for_version(&route_id, &version) {
        Ok(info) => (StatusCode::OK, Json(info)).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn compare_versions(
    State(state): State<SharedState>,
    Path((route_id, v1, v2)): Path<(String, String, String)>,
) -> impl IntoResponse {
    match state.manager.compare_versions(&route_id, &v1, &v2) {
        Ok(comparison) => (StatusCode::OK, Json(comparison)).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn create_task(
    State(state): State<SharedState>,
    Json(req): Json<CreateTaskRequest>,
) -> impl IntoResponse {
    match state.manager.create_task(req.name, &req.route_id, req.version.as_deref()) {
        Ok(task) => (StatusCode::CREATED, Json(task)).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn list_tasks(State(state): State<SharedState>) -> impl IntoResponse {
    let tasks = state.manager.list_tasks();
    Json(tasks)
}

async fn get_task(
    State(state): State<SharedState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.manager.get_task(&id) {
        Ok(task) => (StatusCode::OK, Json(task)).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn delete_task(
    State(state): State<SharedState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match state.manager.delete_task(&id) {
        Ok(_) => StatusCode::NO_CONTENT.into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn get_available_processes(
    State(state): State<SharedState>,
    Path(task_id): Path<String>,
) -> impl IntoResponse {
    match state.manager.get_available_processes(&task_id) {
        Ok(processes) => (StatusCode::OK, Json(processes)).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn start_process(
    State(state): State<SharedState>,
    Path((task_id, process_id)): Path<(String, String)>,
) -> impl IntoResponse {
    match state.manager.start_process(&task_id, &process_id) {
        Ok(task) => (
            StatusCode::OK,
            Json(ProcessActionResponse { task, warning: None }),
        )
            .into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn complete_process(
    State(state): State<SharedState>,
    Path((task_id, process_id)): Path<(String, String)>,
) -> impl IntoResponse {
    match state.manager.complete_process(&task_id, &process_id) {
        Ok(task) => (
            StatusCode::OK,
            Json(ProcessActionResponse { task, warning: None }),
        )
            .into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn pause_process(
    State(state): State<SharedState>,
    Path((task_id, process_id)): Path<(String, String)>,
) -> impl IntoResponse {
    match state.manager.pause_process(&task_id, &process_id) {
        Ok((task, warning)) => (StatusCode::OK, Json(ProcessActionResponse { task, warning })).into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

async fn resume_process(
    State(state): State<SharedState>,
    Path((task_id, process_id)): Path<(String, String)>,
) -> impl IntoResponse {
    match state.manager.resume_process(&task_id, &process_id) {
        Ok(task) => (
            StatusCode::OK,
            Json(ProcessActionResponse { task, warning: None }),
        )
            .into_response(),
        Err(e) => {
            let (status, json) = error_to_response(e);
            (status, json).into_response()
        }
    }
}

fn create_router(state: SharedState) -> Router {
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    Router::new()
        .route("/routes", post(create_route))
        .route("/routes", get(list_routes))
        .route("/routes/:id", get(get_route))
        .route("/routes/:id", delete(delete_route))
        .route("/routes/:id/versions", post(add_version))
        .route("/routes/:id/versions/:version/critical-path", get(get_critical_path))
        .route("/routes/:id/compare/:v1/:v2", get(compare_versions))
        .route("/tasks", post(create_task))
        .route("/tasks", get(list_tasks))
        .route("/tasks/:id", get(get_task))
        .route("/tasks/:id", delete(delete_task))
        .route("/tasks/:id/available", get(get_available_processes))
        .route("/tasks/:id/processes/:process_id/start", put(start_process))
        .route("/tasks/:id/processes/:process_id/complete", put(complete_process))
        .route("/tasks/:id/processes/:process_id/pause", put(pause_process))
        .route("/tasks/:id/processes/:process_id/resume", put(resume_process))
        .with_state(state)
        .layer(cors)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let config = Config::parse();

    let store = InMemoryStore::new();
    let manager = ProcessRouteManager::new(store);

    let state = Arc::new(AppState { manager });

    let app = create_router(state);

    let addr: SocketAddr = format!("{}:{}", config.host, config.port).parse().unwrap();
    tracing::info!("工艺路线管理服务器监听于 {}", addr);

    let listener = TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
