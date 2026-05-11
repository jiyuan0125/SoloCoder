use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use chrono::Utc;
use clap::Parser;
use freight_core::*;
use serde::Serialize;
use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::{Arc, Mutex};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "FREIGHT_SERVER_HOST", default_value = "127.0.0.1")]
    host: String,

    #[arg(long, env = "FREIGHT_SERVER_PORT", default_value_t = 3000)]
    port: u16,
}

type AppState = Arc<Mutex<HashMap<Uuid, Order>>>;

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

fn json_error<T: ToString>(status: StatusCode, message: T) -> impl IntoResponse {
    (
        status,
        Json(ErrorResponse {
            error: message.to_string(),
        }),
    )
}

async fn create_order(
    State(state): State<AppState>,
    Json(req): Json<CreateOrderRequest>,
) -> impl IntoResponse {
    if req.packages.is_empty() {
        return json_error(StatusCode::BAD_REQUEST, "至少需要一个包裹").into_response();
    }

    for pkg in &req.packages {
        if let Err(e) = validate_package(pkg) {
            return json_error(StatusCode::BAD_REQUEST, e.to_string()).into_response();
        }
    }

    let mut items = Vec::new();
    let mut total = 0.0;

    for (idx, pkg_req) in req.packages.iter().enumerate() {
        let package = Package {
            id: format!("pkg-{}-{}", Uuid::new_v4(), idx + 1),
            length_cm: pkg_req.length_cm,
            width_cm: pkg_req.width_cm,
            height_cm: pkg_req.height_cm,
            actual_weight_kg: pkg_req.actual_weight_kg,
            destination: pkg_req.destination,
            cargo_type: pkg_req.cargo_type,
            declared_value: pkg_req.declared_value,
        };

        let breakdown = calculate_freight_breakdown(&package);
        total += breakdown.total_freight;

        items.push(OrderItem { package, breakdown });
    }

    let order = Order {
        id: Uuid::new_v4(),
        items,
        total_freight: total,
        status: OrderStatus::Created,
        created_at: Utc::now(),
        cancelled_at: None,
    };

    let order_id = order.id;
    let response = OrderResponse {
        id: order.id,
        items: order
            .items
            .iter()
            .map(|item| OrderItemResponse {
                package_id: item.package.id.clone(),
                destination: item.package.destination,
                cargo_type: item.package.cargo_type,
                breakdown: item.breakdown.clone(),
            })
            .collect(),
        total_freight: order.total_freight,
        status: order.status,
        created_at: order.created_at,
    };

    {
        let mut store = state.lock().unwrap();
        store.insert(order_id, order);
    }

    (StatusCode::CREATED, Json(response)).into_response()
}

async fn get_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> impl IntoResponse {
    let store = state.lock().unwrap();
    match store.get(&order_id) {
        Some(order) => {
            let response = OrderResponse {
                id: order.id,
                items: order
                    .items
                    .iter()
                    .map(|item| OrderItemResponse {
                        package_id: item.package.id.clone(),
                        destination: item.package.destination,
                        cargo_type: item.package.cargo_type,
                        breakdown: item.breakdown.clone(),
                    })
                    .collect(),
                total_freight: order.total_freight,
                status: order.status,
                created_at: order.created_at,
            };
            (StatusCode::OK, Json(response)).into_response()
        }
        None => json_error(StatusCode::NOT_FOUND, "订单不存在").into_response(),
    }
}

async fn list_orders(State(state): State<AppState>) -> impl IntoResponse {
    let store = state.lock().unwrap();
    let orders: Vec<OrderResponse> = store
        .values()
        .map(|order| OrderResponse {
            id: order.id,
            items: order
                .items
                .iter()
                .map(|item| OrderItemResponse {
                    package_id: item.package.id.clone(),
                    destination: item.package.destination,
                    cargo_type: item.package.cargo_type,
                    breakdown: item.breakdown.clone(),
                })
                .collect(),
            total_freight: order.total_freight,
            status: order.status,
            created_at: order.created_at,
        })
        .collect();
    Json(orders).into_response()
}

async fn cancel_order(
    State(state): State<AppState>,
    Path(order_id): Path<Uuid>,
) -> impl IntoResponse {
    let mut store = state.lock().unwrap();

    let order = match store.get_mut(&order_id) {
        Some(o) => o,
        None => return json_error(StatusCode::NOT_FOUND, CancelError::OrderNotFound.to_string()).into_response(),
    };

    if order.status != OrderStatus::Created {
        return json_error(
            StatusCode::BAD_REQUEST,
            CancelError::NotCancellable(order.status.name().to_string()).to_string(),
        ).into_response();
    }

    if !can_cancel_order(order) {
        return json_error(StatusCode::BAD_REQUEST, CancelError::CancelPeriodExpired.to_string()).into_response();
    }

    order.status = OrderStatus::Cancelled;
    order.cancelled_at = Some(Utc::now());

    let response = OrderResponse {
        id: order.id,
        items: order
            .items
            .iter()
            .map(|item| OrderItemResponse {
                package_id: item.package.id.clone(),
                destination: item.package.destination,
                cargo_type: item.package.cargo_type,
                breakdown: item.breakdown.clone(),
            })
            .collect(),
        total_freight: order.total_freight,
        status: order.status,
        created_at: order.created_at,
    };

    (StatusCode::OK, Json(response)).into_response()
}

async fn health() -> &'static str {
    "OK"
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let state: AppState = Arc::new(Mutex::new(HashMap::new()));

    let app = Router::new()
        .route("/health", get(health))
        .route("/api/orders", post(create_order).get(list_orders))
        .route("/api/orders/:id", get(get_order))
        .route("/api/orders/:id/cancel", post(cancel_order))
        .with_state(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();

    println!("货运物流报价服务已启动: http://{}", addr);
    println!("健康检查: http://{}/health", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
