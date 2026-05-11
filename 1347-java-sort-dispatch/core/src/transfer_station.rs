use crate::models::TransferStation;

pub fn create_transfer_station(id: &str, name: &str, city: &str, capacity: usize) -> TransferStation {
    TransferStation::new(id, name, city, capacity)
}
