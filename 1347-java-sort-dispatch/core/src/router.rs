use crate::inventory::InventoryManager;
use crate::models::{BestRoute, Route, Warehouse};
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct Router {
    routes: HashMap<String, Vec<Route>>,
}

impl Router {
    pub fn new() -> Self {
        Self {
            routes: HashMap::new(),
        }
    }

    pub fn add_route(&mut self, route: Route) {
        let key = route.destination_city.clone();
        self.routes
            .entry(key)
            .or_default()
            .push(route);
    }

    pub fn find_best_route(
        &self,
        product_id: &str,
        destination_city: &str,
        quantity: u32,
        warehouses: &[Warehouse],
        inventory_manager: &mut InventoryManager,
    ) -> Result<BestRoute, String> {
        let routes = self
            .routes
            .get(destination_city)
            .ok_or_else(|| format!("No routes found for city {}", destination_city))?;

        let mut sorted_routes = routes.clone();
        sorted_routes.sort_by_key(|r| r.priority);

        for route in sorted_routes {
            if inventory_manager.has_stock(&route.warehouse_id, product_id, quantity) {
                return Ok(BestRoute {
                    warehouse_id: route.warehouse_id.clone(),
                    transfer_stations: route.transfer_stations.clone(),
                });
            }
        }

        Err(format!(
            "No warehouse has sufficient stock for product {} (quantity {})",
            product_id, quantity
        ))
    }
}

impl Default for Router {
    fn default() -> Self {
        Self::new()
    }
}
