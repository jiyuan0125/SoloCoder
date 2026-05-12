mod api;
mod health_check;
mod load_balancer;
mod types;

use std::env;

use api::create_routes;
use health_check::run_health_checks;
use load_balancer::LoadBalancer;
use types::LoadBalancingStrategy;
use tracing_subscriber;
use warp::Filter;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(8080);

    let strategy = match env::var("LB_STRATEGY")
        .unwrap_or_else(|_| "least_connections".to_string())
        .to_lowercase()
        .as_str()
    {
        "random" => LoadBalancingStrategy::Random,
        _ => LoadBalancingStrategy::LeastConnections,
    };

    let lb = LoadBalancer::new(strategy);

    println!("Starting load balancer on port {}", port);
    println!("Load balancing strategy: {:?}", strategy);

    let lb_clone = lb.clone();
    tokio::spawn(async move {
        run_health_checks(lb_clone).await;
    });

    let routes = create_routes(lb).with(warp::log("load_balancer"));

    warp::serve(routes).run(([0, 0, 0, 0], port)).await;
}
