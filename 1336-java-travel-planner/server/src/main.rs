use axum::{
    extract::{Json, Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::get,
    Router,
};
use clap::Parser;
use std::net::SocketAddr;
use std::sync::{Arc, Mutex};
use tower_http::cors::{Any, CorsLayer};
use travel_planner_core::*;
use serde::{Deserialize, Serialize};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

type AppState = Arc<Mutex<TravelPlannerService>>;

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let service = TravelPlannerService::new();
    let state: AppState = Arc::new(Mutex::new(service));

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/health", get(health))
        .route("/attractions", get(list_attractions).post(create_attraction))
        .route("/attractions/:id", get(get_attraction))
        .route("/transportations", get(list_transportations).post(create_transportation))
        .route("/itineraries", get(list_itineraries).post(create_itinerary))
        .route("/itineraries/:id", get(get_itinerary).post(add_nodes_to_itinerary))
        .layer(cors)
        .with_state(state);

    let addr = SocketAddr::from(([127, 0, 0, 1], args.port));
    println!("Server running on http://{}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn health() -> &'static str {
    "OK"
}

async fn list_attractions(State(state): State<AppState>) -> Json<Vec<Attraction>> {
    let service = state.lock().unwrap();
    Json(service.get_all_attractions())
}

async fn create_attraction(
    State(state): State<AppState>,
    Json(payload): Json<NewAttraction>,
) -> Json<Attraction> {
    let mut service = state.lock().unwrap();
    Json(service.add_attraction(payload))
}

async fn get_attraction(
    State(state): State<AppState>,
    Path(id): Path<uuid::Uuid>,
) -> Response {
    let service = state.lock().unwrap();
    match service.get_attraction(id) {
        Some(attraction) => Json(attraction).into_response(),
        None => StatusCode::NOT_FOUND.into_response(),
    }
}

async fn list_transportations(State(state): State<AppState>) -> Json<Vec<Transportation>> {
    let service = state.lock().unwrap();
    Json(service.get_all_transportations())
}

async fn create_transportation(
    State(state): State<AppState>,
    Json(payload): Json<NewTransportation>,
) -> Response {
    let mut service = state.lock().unwrap();
    match service.add_transportation(payload) {
        Ok(transportation) => (StatusCode::CREATED, Json(transportation)).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ErrorResponse::from(e))).into_response(),
    }
}

async fn list_itineraries(State(state): State<AppState>) -> Json<Vec<Itinerary>> {
    let service = state.lock().unwrap();
    Json(service.get_all_itineraries())
}

async fn create_itinerary(
    State(state): State<AppState>,
    Json(payload): Json<NewItinerary>,
) -> (StatusCode, Json<Itinerary>) {
    let mut service = state.lock().unwrap();
    (StatusCode::CREATED, Json(service.create_itinerary(payload)))
}

async fn get_itinerary(
    State(state): State<AppState>,
    Path(id): Path<uuid::Uuid>,
) -> Response {
    let service = state.lock().unwrap();
    match service.get_itinerary(id) {
        Some(itinerary) => Json(itinerary).into_response(),
        None => StatusCode::NOT_FOUND.into_response(),
    }
}

async fn add_nodes_to_itinerary(
    State(state): State<AppState>,
    Path(id): Path<uuid::Uuid>,
    Json(payload): Json<AddItineraryNodesRequest>,
) -> Response {
    let mut service = state.lock().unwrap();
    match service.add_nodes_to_itinerary(id, payload.nodes) {
        Ok(itinerary) => Json(itinerary).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ErrorResponse::from(e))).into_response(),
    }
}

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
    details: Vec<String>,
}

impl From<TravelPlannerError> for ErrorResponse {
    fn from(err: TravelPlannerError) -> Self {
        let details = match &err {
            TravelPlannerError::MultipleErrors(errors) => {
                errors.iter().map(|e| e.to_string()).collect()
            }
            _ => vec![err.to_string()],
        };

        ErrorResponse {
            error: err.to_string(),
            details,
        }
    }
}
