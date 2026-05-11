use crate::models::{Order, OrderRequest, Package};

pub fn create_order(request: OrderRequest, warehouse_id: String) -> Order {
    Order::new(
        request.product_id,
        request.quantity,
        request.destination_address,
        request.destination_city,
        warehouse_id,
    )
}

pub fn calculate_order_status(packages: &[Package]) -> String {
    let total = packages.len();
    let delivered = packages.iter().filter(|p| p.status == crate::models::PackageStatus::Delivered).count();
    let lost = packages.iter().filter(|p| p.status == crate::models::PackageStatus::Lost).count();

    if lost > 0 {
        "Exception".to_string()
    } else if delivered == total {
        "Delivered".to_string()
    } else if delivered > 0 {
        "Partially Delivered".to_string()
    } else if packages.iter().any(|p| matches!(p.status, crate::models::PackageStatus::InTransit | crate::models::PackageStatus::AtTransferStation)) {
        "Shipped".to_string()
    } else {
        "Processing".to_string()
    }
}
