pub mod models;
pub mod router;
pub mod package;
pub mod warehouse;
pub mod transfer_station;
pub mod order;
pub mod inventory;
pub mod exception;

pub use models::*;
pub use router::*;
pub use package::*;
pub use warehouse::*;
pub use transfer_station::*;
pub use order::*;
pub use inventory::*;
pub use exception::*;

use std::sync::Arc;
use std::sync::Mutex;

#[derive(Debug, Clone)]
pub struct LogisticsSystem {
    warehouses: Arc<Mutex<Vec<Warehouse>>>,
    transfer_stations: Arc<Mutex<Vec<TransferStation>>>,
    router: Arc<Mutex<Router>>,
    orders: Arc<Mutex<Vec<Order>>>,
    packages: Arc<Mutex<Vec<Package>>>,
    inventory_manager: Arc<Mutex<InventoryManager>>,
}

impl LogisticsSystem {
    pub fn new() -> Self {
        Self {
            warehouses: Arc::new(Mutex::new(Vec::new())),
            transfer_stations: Arc::new(Mutex::new(Vec::new())),
            router: Arc::new(Mutex::new(Router::new())),
            orders: Arc::new(Mutex::new(Vec::new())),
            packages: Arc::new(Mutex::new(Vec::new())),
            inventory_manager: Arc::new(Mutex::new(InventoryManager::new())),
        }
    }

    pub fn add_warehouse(&self, warehouse: Warehouse) {
        let mut warehouses = self.warehouses.lock().unwrap();
        warehouses.push(warehouse);
    }

    pub fn add_transfer_station(&self, station: TransferStation) {
        let mut stations = self.transfer_stations.lock().unwrap();
        stations.push(station);
    }

    pub fn add_route(&self, route: Route) {
        let mut router = self.router.lock().unwrap();
        router.add_route(route);
    }

    pub fn create_order(&self, order_request: OrderRequest) -> Result<Order, String> {
        let router = self.router.lock().unwrap();
        let warehouses = self.warehouses.lock().unwrap();
        let transfer_stations = self.transfer_stations.lock().unwrap();
        let mut inventory_manager = self.inventory_manager.lock().unwrap();
        let mut orders = self.orders.lock().unwrap();
        let mut packages = self.packages.lock().unwrap();

        let best_route = router.find_best_route(
            &order_request.product_id,
            &order_request.destination_city,
            order_request.quantity,
            &warehouses,
            &mut inventory_manager,
        )?;

        let order = Order::new(
            order_request.product_id.clone(),
            order_request.quantity,
            order_request.destination_address.clone(),
            order_request.destination_city.clone(),
            best_route.warehouse_id.clone(),
        );

        let package_splitter = PackageSplitter::new(50.0, 1.5);
        let order_packages = package_splitter.split_into_packages(
            &order.id,
            order_request.quantity,
            order_request.weight_per_item,
            order_request.length_per_item,
            order_request.width_per_item,
            order_request.height_per_item,
        );

        inventory_manager.reserve_stock(
            &best_route.warehouse_id,
            &order_request.product_id,
            order_request.quantity,
        )?;

        for mut pkg in order_packages {
            pkg.assign_route(best_route.transfer_stations.clone());
            packages.push(pkg);
        }

        orders.push(order.clone());

        Ok(order)
    }

    pub fn process_packages(&self) {
        let mut stations = self.transfer_stations.lock().unwrap();
        let mut packages = self.packages.lock().unwrap();

        for station in stations.iter_mut() {
            station.process_queued_packages();
        }

        for pkg in packages.iter_mut() {
            pkg.update_status();
        }
    }

    pub fn get_order(&self, order_id: &str) -> Option<Order> {
        let orders = self.orders.lock().unwrap();
        let packages = self.packages.lock().unwrap();
        
        let mut order = orders.iter().find(|o| o.id == order_id)?.clone();
        
        let order_packages: Vec<Package> = packages
            .iter()
            .filter(|p| p.order_id == order_id)
            .cloned()
            .collect();
        
        order.update_status(&order_packages);
        Some(order)
    }

    pub fn get_package(&self, package_id: &str) -> Option<Package> {
        let packages = self.packages.lock().unwrap();
        packages.iter().find(|p| p.id == package_id).cloned()
    }

    pub fn update_package_status(&self, package_id: &str, status: PackageStatus) -> Result<(), String> {
        let mut packages = self.packages.lock().unwrap();
        let pkg = packages
            .iter_mut()
            .find(|p| p.id == package_id)
            .ok_or_else(|| format!("Package {} not found", package_id))?;
        
        pkg.set_status(status);
        Ok(())
    }

    pub fn check_exceptions(&self) -> Vec<Exception> {
        let stations = self.transfer_stations.lock().unwrap();
        let packages = self.packages.lock().unwrap();
        
        let mut exceptions = Vec::new();
        
        for station in stations.iter() {
            if let Some(exception) = ExceptionDetector::check_transfer_station_backlog(station) {
                exceptions.push(exception);
            }
        }
        
        for pkg in packages.iter() {
            if let Some(exception) = ExceptionDetector::check_package_lost(pkg) {
                exceptions.push(exception);
            }
        }
        
        exceptions
    }

    pub fn get_all_warehouses(&self) -> Vec<Warehouse> {
        self.warehouses.lock().unwrap().clone()
    }

    pub fn get_all_transfer_stations(&self) -> Vec<TransferStation> {
        self.transfer_stations.lock().unwrap().clone()
    }

    pub fn add_inventory(&self, warehouse_id: &str, product_id: &str, quantity: u32) {
        let mut inventory_manager = self.inventory_manager.lock().unwrap();
        inventory_manager.add_stock(warehouse_id, product_id, quantity);
    }
}

impl Default for LogisticsSystem {
    fn default() -> Self {
        Self::new()
    }
}
