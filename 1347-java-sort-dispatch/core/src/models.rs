use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Address {
    pub city: String,
    pub street: String,
    pub postal_code: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Warehouse {
    pub id: String,
    pub name: String,
    pub city: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferStation {
    pub id: String,
    pub name: String,
    pub city: String,
    pub capacity: usize,
    pub current_load: usize,
    pub queue: Vec<QueuedPackage>,
    pub last_processed_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QueuedPackage {
    pub package_id: String,
    pub enqueued_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Route {
    pub warehouse_id: String,
    pub destination_city: String,
    pub transfer_stations: Vec<String>,
    pub priority: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BestRoute {
    pub warehouse_id: String,
    pub transfer_stations: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: String,
    pub product_id: String,
    pub quantity: u32,
    pub destination_address: Address,
    pub destination_city: String,
    pub warehouse_id: String,
    pub created_at: DateTime<Utc>,
    pub status: OrderStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum OrderStatus {
    Created,
    Processing,
    Shipped,
    PartiallyDelivered,
    Delivered,
    Exception,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Package {
    pub id: String,
    pub order_id: String,
    pub weight_kg: f64,
    pub length_m: f64,
    pub width_m: f64,
    pub height_m: f64,
    pub quantity: u32,
    pub route: Vec<String>,
    pub current_station_index: usize,
    pub status: PackageStatus,
    pub created_at: DateTime<Utc>,
    pub last_update: DateTime<Utc>,
    pub estimated_delivery: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PackageStatus {
    Created,
    InWarehouse,
    InTransit,
    AtTransferStation,
    OutForDelivery,
    Delivered,
    Lost,
    Exception,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderRequest {
    pub product_id: String,
    pub quantity: u32,
    pub destination_address: String,
    pub destination_city: String,
    pub weight_per_item: f64,
    pub length_per_item: f64,
    pub width_per_item: f64,
    pub height_per_item: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Exception {
    pub id: String,
    pub exception_type: ExceptionType,
    pub description: String,
    pub created_at: DateTime<Utc>,
    pub related_entity_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum ExceptionType {
    TransferStationBacklog,
    PackageLost,
    InventoryShortage,
}

impl Warehouse {
    pub fn new(id: &str, name: &str, city: &str) -> Self {
        Self {
            id: id.to_string(),
            name: name.to_string(),
            city: city.to_string(),
        }
    }
}

impl TransferStation {
    pub fn new(id: &str, name: &str, city: &str, capacity: usize) -> Self {
        Self {
            id: id.to_string(),
            name: name.to_string(),
            city: city.to_string(),
            capacity,
            current_load: 0,
            queue: Vec::new(),
            last_processed_time: Utc::now(),
        }
    }

    pub fn can_accept(&self) -> bool {
        self.current_load < self.capacity
    }

    pub fn enqueue_package(&mut self, package_id: &str) {
        self.queue.push(QueuedPackage {
            package_id: package_id.to_string(),
            enqueued_at: Utc::now(),
        });
    }

    pub fn process_queued_packages(&mut self) {
        while !self.queue.is_empty() && self.can_accept() {
            let queued = self.queue.remove(0);
            self.current_load += 1;
        }
        self.last_processed_time = Utc::now();
    }

    pub fn get_oldest_queued_time(&self) -> Option<DateTime<Utc>> {
        self.queue.first().map(|q| q.enqueued_at)
    }
}

impl Route {
    pub fn new(
        warehouse_id: &str,
        destination_city: &str,
        transfer_stations: Vec<&str>,
        priority: u32,
    ) -> Self {
        Self {
            warehouse_id: warehouse_id.to_string(),
            destination_city: destination_city.to_string(),
            transfer_stations: transfer_stations.iter().map(|s| s.to_string()).collect(),
            priority,
        }
    }
}

impl Order {
    pub fn new(
        product_id: String,
        quantity: u32,
        destination_address: String,
        destination_city: String,
        warehouse_id: String,
    ) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            product_id,
            quantity,
            destination_address: Address {
                city: destination_city.clone(),
                street: destination_address,
                postal_code: String::new(),
            },
            destination_city,
            warehouse_id,
            created_at: Utc::now(),
            status: OrderStatus::Created,
        }
    }

    pub fn update_status(&mut self, packages: &[Package]) {
        let total_packages = packages.len();
        let delivered_count = packages
            .iter()
            .filter(|p| p.status == PackageStatus::Delivered)
            .count();
        let lost_count = packages
            .iter()
            .filter(|p| p.status == PackageStatus::Lost)
            .count();

        if lost_count > 0 {
            self.status = OrderStatus::Exception;
        } else if delivered_count == total_packages {
            self.status = OrderStatus::Delivered;
        } else if delivered_count > 0 {
            self.status = OrderStatus::PartiallyDelivered;
        } else {
            let any_in_transit = packages.iter().any(|p| {
                matches!(p.status, PackageStatus::InTransit | PackageStatus::AtTransferStation)
            });
            if any_in_transit {
                self.status = OrderStatus::Shipped;
            }
        }
    }
}

impl Package {
    pub fn new(
        order_id: &str,
        weight_kg: f64,
        length_m: f64,
        width_m: f64,
        height_m: f64,
        quantity: u32,
    ) -> Self {
        let now = Utc::now();
        let estimated_delivery = now + chrono::Duration::days(3);

        Self {
            id: Uuid::new_v4().to_string(),
            order_id: order_id.to_string(),
            weight_kg,
            length_m,
            width_m,
            height_m,
            quantity,
            route: Vec::new(),
            current_station_index: 0,
            status: PackageStatus::Created,
            created_at: now,
            last_update: now,
            estimated_delivery,
        }
    }

    pub fn assign_route(&mut self, route: Vec<String>) {
        self.route = route;
        self.status = PackageStatus::InWarehouse;
    }

    pub fn set_status(&mut self, status: PackageStatus) {
        self.status = status;
        self.last_update = Utc::now();
    }

    pub fn update_status(&mut self) {
        match self.status {
            PackageStatus::InWarehouse => {
                if !self.route.is_empty() {
                    self.status = PackageStatus::InTransit;
                    self.last_update = Utc::now();
                }
            }
            PackageStatus::InTransit => {
                if self.current_station_index < self.route.len() {
                    self.status = PackageStatus::AtTransferStation;
                    self.last_update = Utc::now();
                }
            }
            PackageStatus::AtTransferStation => {
                self.current_station_index += 1;
                if self.current_station_index >= self.route.len() {
                    self.status = PackageStatus::OutForDelivery;
                } else {
                    self.status = PackageStatus::InTransit;
                }
                self.last_update = Utc::now();
            }
            _ => {}
        }
    }

    pub fn is_lost(&self) -> bool {
        if self.status == PackageStatus::Delivered || self.status == PackageStatus::Lost {
            return false;
        }

        let seven_days = chrono::Duration::days(7);
        let now = Utc::now();

        now.signed_duration_since(self.estimated_delivery) > seven_days
            && now.signed_duration_since(self.last_update) > seven_days
    }
}

impl Exception {
    pub fn new(exception_type: ExceptionType, description: &str, related_entity_id: &str) -> Self {
        Self {
            id: Uuid::new_v4().to_string(),
            exception_type,
            description: description.to_string(),
            created_at: Utc::now(),
            related_entity_id: related_entity_id.to_string(),
        }
    }
}
