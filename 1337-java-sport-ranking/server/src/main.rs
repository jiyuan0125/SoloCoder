use std::sync::Arc;
use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Json, Response},
    routing::get,
    Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;
use tower_http::cors::{Any, CorsLayer};
use league_core::{League, Match, MatchResult, ValidationError};

#[derive(Parser, Debug, Clone)]
#[command(about = "足球联赛积分排名系统服务端")]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

type AppState = Arc<Mutex<League>>;

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

struct AppValidationError(ValidationError);

impl IntoResponse for AppValidationError {
    fn into_response(self) -> Response {
        (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: format!("{}", self.0),
            }),
        )
            .into_response()
    }
}

impl From<ValidationError> for AppValidationError {
    fn from(err: ValidationError) -> Self {
        AppValidationError(err)
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let args = Args::parse();

    let state: AppState = Arc::new(Mutex::new(League::new()));

    let cors = CorsLayer::new()
        .allow_methods(Any)
        .allow_origin(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/standings", get(get_standings))
        .route("/matches", get(get_matches).post(add_match))
        .route("/teams", get(get_teams).post(add_team))
        .layer(cors)
        .with_state(state);

    let addr = format!("0.0.0.0:{}", args.port);
    tracing::info!("服务启动在 {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn get_standings(State(state): State<AppState>) -> impl IntoResponse {
    let league = state.lock().await;
    Json(league.get_standings())
}

async fn get_matches(State(state): State<AppState>) -> impl IntoResponse {
    let league = state.lock().await;
    let matches: Vec<Match> = league.get_matches().to_vec();
    Json(matches)
}

async fn add_match(
    State(state): State<AppState>,
    Json(match_result): Json<MatchResult>,
) -> Result<impl IntoResponse, AppValidationError> {
    let mut league = state.lock().await;
    league.add_match(match_result).map_err(AppValidationError::from)?;
    Ok((StatusCode::CREATED, Json(serde_json::json!({"status": "ok"}))))
}

async fn get_teams(State(state): State<AppState>) -> impl IntoResponse {
    let league = state.lock().await;
    let teams: Vec<String> = league.get_teams().iter().cloned().collect();
    Json(teams)
}

async fn add_team(
    State(state): State<AppState>,
    Json(team_name): Json<String>,
) -> impl IntoResponse {
    let mut league = state.lock().await;
    league.add_team(team_name);
    (StatusCode::CREATED, Json(serde_json::json!({"status": "ok"})))
}
