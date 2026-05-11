use crate::models::Warehouse;

pub fn create_warehouse(id: &str, name: &str, city: &str) -> Warehouse {
    Warehouse::new(id, name, city)
}
