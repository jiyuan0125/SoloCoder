use axum::extract::Path;
use axum::response::Response;
use axum::body::Body;
use axum::http::StatusCode;
use axum::Json;
use axum::{Router, routing::{get, put}, Server};
use serde::Deserialize;

#[derive(Deserialize)]
struct MyRequest {
    value: String,
}

async fn simple_handler(
    Path(key): Path<String>,
) -> Response<Body> {
    tokio::time::sleep(tokio::time::Duration::from_millis(1)).await;
    Response::builder()
        .status(StatusCode::OK)
        .body(Body::from(key))
        .unwrap()
}

async fn simple_put_handler(
    Path(key): Path<String>,
    Json(body): Json<MyRequest>,
) -> Response<Body> {
    tokio::time::sleep(tokio::time::Duration::from_millis(1)).await;
    let result = format!("{}: {}", key, body.value);
    drop(result);
    Response::builder()
        .status(StatusCode::NO_CONTENT)
        .body(Body::empty())
        .unwrap()
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/:key", get(simple_handler))
        .route("/:key", put(simple_put_handler));
    
    let addr = ([0, 0, 0, 0], 3000).into();
    Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
