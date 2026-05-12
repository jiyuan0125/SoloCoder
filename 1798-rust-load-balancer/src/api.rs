use serde::Serialize;
use warp::{Filter, Rejection, Reply};

use crate::load_balancer::LoadBalancer;
use crate::types::{AddBackendRequest, AllocationRecord, BackendStats};

#[derive(Serialize)]
#[allow(dead_code)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    message: Option<String>,
}

#[derive(Serialize)]
struct StatsResponse {
    backends: Vec<BackendStats>,
}

#[derive(Serialize)]
struct AllocationHistoryResponse {
    records: Vec<AllocationRecord>,
}

#[derive(Serialize)]
struct SimpleResponse {
    message: String,
}

fn with_lb(
    lb: LoadBalancer,
) -> impl Filter<Extract = (LoadBalancer,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || lb.clone())
}

fn json_body() -> impl Filter<Extract = (AddBackendRequest,), Error = Rejection> + Clone {
    warp::body::content_length_limit(1024 * 16).and(warp::body::json())
}

pub fn create_routes(lb: LoadBalancer) -> impl Filter<Extract = impl Reply, Error = Rejection> + Clone {
    let health = warp::path("health")
        .and(warp::get())
        .map(|| warp::reply::json(&serde_json::json!({"status": "ok"})));

    let get_stats = warp::path("stats")
        .and(warp::get())
        .and(with_lb(lb.clone()))
        .map(|lb: LoadBalancer| {
            let stats = lb.get_stats();
            let response = StatsResponse { backends: stats };
            warp::reply::json(&response)
        });

    let get_allocation_records = warp::path("allocations")
        .and(warp::get())
        .and(with_lb(lb.clone()))
        .map(|lb: LoadBalancer| {
            let records = lb.get_allocation_records();
            let response = AllocationHistoryResponse { records };
            warp::reply::json(&response)
        });

    let add_backend = warp::path("backends")
        .and(warp::post())
        .and(json_body())
        .and(with_lb(lb.clone()))
        .map(|body: AddBackendRequest, lb: LoadBalancer| {
            let weight = body.weight.unwrap_or(1);
            lb.add_backend(body.address.clone(), weight);
            warp::reply::json(&SimpleResponse {
                message: format!("Backend {} added successfully", body.address),
            })
        });

    let remove_backend = warp::path!("backends" / String)
        .and(warp::delete())
        .and(with_lb(lb.clone()))
        .map(|address: String, lb: LoadBalancer| {
            if lb.remove_backend(&address) {
                warp::reply::json(&SimpleResponse {
                    message: format!("Backend {} removed", address),
                })
            } else {
                warp::reply::json(&SimpleResponse {
                    message: format!("Backend {} marked as draining, will be removed when connections are released", address),
                })
            }
        });

    health
        .or(get_stats)
        .or(get_allocation_records)
        .or(add_backend)
        .or(remove_backend)
}
