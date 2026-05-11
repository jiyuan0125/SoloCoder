use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use thesis_core::{AdvisorId, StudentId, ThesisSystem, *};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "THESIS_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(short, long, env = "THESIS_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    system: ThesisSystem,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "thesis_server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let system = ThesisSystem::new();
    let state = Arc::new(AppState { system });

    let app = Router::new()
        .route("/status", get(get_status))
        .route("/phase/advance", post(advance_phase))
        .route("/students", post(create_student))
        .route("/students", get(list_students))
        .route("/students/:id", get(get_student))
        .route("/students/:id/submit", post(submit_preferences))
        .route("/students/:id/modify", post(modify_preferences))
        .route("/advisors", post(create_advisor))
        .route("/advisors", get(list_advisors))
        .route("/advisors/:id", get(get_advisor))
        .route("/advisors/:id/pool", get(get_advisor_pool))
        .route("/advisors/:id/select", post(advisor_select))
        .route("/appeals", post(create_appeal))
        .route("/appeals", get(list_appeals))
        .route("/assign", post(manual_assign))
        .route("/changes", get(get_change_logs))
        .route("/changes/:student_id", get(get_student_changes))
        .route("/results", get(get_match_results))
        .with_state(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse()?;
    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

async fn get_status(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let status = state.system.get_status().await;
    (StatusCode::OK, Json(ApiResponse::success(status)))
}

async fn advance_phase(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.system.advance_phase().await {
        Ok(phase) => (StatusCode::OK, Json(ApiResponse::success(phase))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<SystemPhase>::error(e.to_string())),
        ),
    }
}

async fn create_student(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateStudentRequest>,
) -> impl IntoResponse {
    match state.system.create_student(req.name, req.student_no).await {
        Ok(student) => (StatusCode::CREATED, Json(ApiResponse::success(student))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Student>::error(e.to_string())),
        ),
    }
}

async fn list_students(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.system.list_students().await {
        Ok(students) => (StatusCode::OK, Json(ApiResponse::success(students))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<Student>>::error(e.to_string())),
        ),
    }
}

async fn get_student(
    State(state): State<Arc<AppState>>,
    Path(id): Path<StudentId>,
) -> impl IntoResponse {
    match state.system.get_student(id).await {
        Ok(student) => (StatusCode::OK, Json(ApiResponse::success(student))),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Student>::error(e.to_string())),
        ),
    }
}

async fn submit_preferences(
    State(state): State<Arc<AppState>>,
    Path(id): Path<StudentId>,
    Json(req): Json<SubmitPreferencesRequest>,
) -> impl IntoResponse {
    match state.system.submit_preferences(id, req.preferences).await {
        Ok(student) => (StatusCode::OK, Json(ApiResponse::success(student))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Student>::error(e.to_string())),
        ),
    }
}

async fn modify_preferences(
    State(state): State<Arc<AppState>>,
    Path(id): Path<StudentId>,
    Json(req): Json<ModifyPreferencesRequest>,
) -> impl IntoResponse {
    match state.system.modify_preferences(id, req.preferences).await {
        Ok(student) => (StatusCode::OK, Json(ApiResponse::success(student))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Student>::error(e.to_string())),
        ),
    }
}

async fn create_advisor(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateAdvisorRequest>,
) -> impl IntoResponse {
    match state.system.create_advisor(req.name, req.capacity).await {
        Ok(advisor) => (StatusCode::CREATED, Json(ApiResponse::success(advisor))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Advisor>::error(e.to_string())),
        ),
    }
}

async fn list_advisors(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.system.list_advisors().await {
        Ok(advisors) => (StatusCode::OK, Json(ApiResponse::success(advisors))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<Advisor>>::error(e.to_string())),
        ),
    }
}

async fn get_advisor(
    State(state): State<Arc<AppState>>,
    Path(id): Path<AdvisorId>,
) -> impl IntoResponse {
    match state.system.get_advisor(id).await {
        Ok(advisor) => (StatusCode::OK, Json(ApiResponse::success(advisor))),
        Err(e) => (
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<Advisor>::error(e.to_string())),
        ),
    }
}

async fn get_advisor_pool(
    State(state): State<Arc<AppState>>,
    Path(id): Path<AdvisorId>,
) -> impl IntoResponse {
    match state.system.get_advisor_pool(id).await {
        Ok(students) => (StatusCode::OK, Json(ApiResponse::success(students))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<Student>>::error(e.to_string())),
        ),
    }
}

async fn advisor_select(
    State(state): State<Arc<AppState>>,
    Path(advisor_id): Path<AdvisorId>,
    Json(req): Json<AdvisorSelectRequest>,
) -> impl IntoResponse {
    match state
        .system
        .advisor_select(advisor_id, req.student_id, req.accept)
        .await
    {
        Ok(_) => (
            StatusCode::OK,
            Json(ApiResponse::success("Selection processed".to_string())),
        ),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<String>::error(e.to_string())),
        ),
    }
}

async fn create_appeal(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateAppealRequest>,
) -> impl IntoResponse {
    match state.system.create_appeal(req.student_id, req.reason).await {
        Ok(appeal) => (StatusCode::CREATED, Json(ApiResponse::success(appeal))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<AppealRecord>::error(e.to_string())),
        ),
    }
}

async fn list_appeals(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.system.list_appeals().await {
        Ok(appeals) => (StatusCode::OK, Json(ApiResponse::success(appeals))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<AppealRecord>>::error(e.to_string())),
        ),
    }
}

async fn manual_assign(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ManualAssignRequest>,
) -> impl IntoResponse {
    match state
        .system
        .manual_assign(req.student_id, req.advisor_id, req.reason, req.operator)
        .await
    {
        Ok(_) => (
            StatusCode::OK,
            Json(ApiResponse::success("Manual assignment done".to_string())),
        ),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<String>::error(e.to_string())),
        ),
    }
}

async fn get_change_logs(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.system.get_change_logs(None).await {
        Ok(logs) => (StatusCode::OK, Json(ApiResponse::success(logs))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<ChangeLog>>::error(e.to_string())),
        ),
    }
}

async fn get_student_changes(
    State(state): State<Arc<AppState>>,
    Path(student_id): Path<StudentId>,
) -> impl IntoResponse {
    match state.system.get_change_logs(Some(student_id)).await {
        Ok(logs) => (StatusCode::OK, Json(ApiResponse::success(logs))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<ChangeLog>>::error(e.to_string())),
        ),
    }
}

async fn get_match_results(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    match state.system.get_match_results().await {
        Ok(results) => (StatusCode::OK, Json(ApiResponse::success(results))),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<Vec<MatchResult>>::error(e.to_string())),
        ),
    }
}
