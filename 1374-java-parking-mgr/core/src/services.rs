use std::collections::HashMap;
use std::sync::Mutex;

use chrono::{DateTime, Duration, Utc};

use crate::error::{ParkingError, Result};
use crate::models::{MonthlyCard, ParkingRecord, ParkingSpace, SpaceType, User, VehicleType};

const TWENTY_FOUR_HOUR_CAP: f64 = 60.0;

pub struct ParkingService {
    spaces: Mutex<HashMap<String, ParkingSpace>>,
    records: Mutex<HashMap<String, ParkingRecord>>,
    active_vehicles: Mutex<HashMap<String, String>>,
    monthly_cards: Mutex<HashMap<String, MonthlyCard>>,
    users: Mutex<HashMap<String, User>>,
}

impl ParkingService {
    pub fn new() -> Self {
        Self::with_spaces(vec![])
    }

    pub fn with_spaces(spaces: Vec<ParkingSpace>) -> Self {
        let mut space_map = HashMap::new();
        for space in spaces {
            space_map.insert(space.id.clone(), space);
        }

        Self {
            spaces: Mutex::new(space_map),
            records: Mutex::new(HashMap::new()),
            active_vehicles: Mutex::new(HashMap::new()),
            monthly_cards: Mutex::new(HashMap::new()),
            users: Mutex::new(HashMap::new()),
        }
    }

    pub fn add_space(&self, space: ParkingSpace) {
        let mut spaces = self.spaces.lock().unwrap();
        spaces.insert(space.id.clone(), space);
    }

    pub fn get_all_spaces(&self) -> Vec<ParkingSpace> {
        let spaces = self.spaces.lock().unwrap();
        spaces.values().cloned().collect()
    }

    pub fn get_space(&self, space_id: &str) -> Option<ParkingSpace> {
        let spaces = self.spaces.lock().unwrap();
        spaces.get(space_id).cloned()
    }

    pub fn add_user(&self, user: User) {
        let mut users = self.users.lock().unwrap();
        users.insert(user.id.clone(), user);
    }

    pub fn get_user(&self, user_id: &str) -> Option<User> {
        let users = self.users.lock().unwrap();
        users.get(user_id).cloned()
    }

    pub fn get_all_users(&self) -> Vec<User> {
        let users = self.users.lock().unwrap();
        users.values().cloned().collect()
    }

    pub fn add_monthly_card(&self, card: MonthlyCard) -> Result<()> {
        let users = self.users.lock().unwrap();
        if !users.contains_key(&card.user_id) {
            return Err(ParkingError::UserNotFound(card.user_id.clone()));
        }
        drop(users);

        if let Some(space_id) = &card.reserved_space_id {
            let mut spaces = self.spaces.lock().unwrap();
            let space = spaces
                .get_mut(space_id)
                .ok_or_else(|| ParkingError::SpaceNotFound(space_id.clone()))?;

            if space.reserved_for_user.is_some() && space.reserved_for_user.as_ref() != Some(&card.user_id) {
                return Err(ParkingError::SpaceAlreadyReserved);
            }

            space.reserved_for_user = Some(card.user_id.clone());
        }

        let mut cards = self.monthly_cards.lock().unwrap();
        cards.insert(card.vehicle_plate.clone(), card);

        Ok(())
    }

    pub fn get_monthly_card(&self, plate: &str) -> Option<MonthlyCard> {
        let cards = self.monthly_cards.lock().unwrap();
        cards.get(plate).cloned()
    }

    pub fn get_all_monthly_cards(&self) -> Vec<MonthlyCard> {
        let cards = self.monthly_cards.lock().unwrap();
        cards.values().cloned().collect()
    }

    pub fn get_cards_needing_reminder(&self) -> Vec<MonthlyCard> {
        let cards = self.monthly_cards.lock().unwrap();
        cards
            .values()
            .filter(|c| c.needs_reminder())
            .cloned()
            .collect()
    }

    pub fn enter_parking(
        &self,
        plate_number: &str,
        vehicle_type: VehicleType,
        user_id: Option<String>,
    ) -> Result<ParkingRecord> {
        let active_vehicles = self.active_vehicles.lock().unwrap();
        if active_vehicles.contains_key(plate_number) {
            return Err(ParkingError::VehicleAlreadyParked(plate_number.to_string()));
        }
        drop(active_vehicles);

        let (is_monthly, reserved_space_id) = {
            let cards = self.monthly_cards.lock().unwrap();
            if let Some(card) = cards.get(plate_number) {
                if card.is_active() {
                    (true, card.reserved_space_id.clone())
                } else {
                    let now = Utc::now();
                    let hours_since_expiry = (now - card.end_time).num_hours();
                    if hours_since_expiry <= 24 {
                        (true, card.reserved_space_id.clone())
                    } else {
                        (false, None)
                    }
                }
            } else {
                (false, None)
            }
        };

        let space_id = self.allocate_space(vehicle_type, user_id.as_ref(), reserved_space_id.as_deref())?;

        let record = ParkingRecord::new(plate_number.to_string(), vehicle_type, space_id.clone(), is_monthly);

        let mut spaces = self.spaces.lock().unwrap();
        let space = spaces.get_mut(&space_id).ok_or_else(|| ParkingError::SpaceNotFound(space_id.clone()))?;
        space.is_occupied = true;

        let mut active_vehicles = self.active_vehicles.lock().unwrap();
        active_vehicles.insert(plate_number.to_string(), record.id.clone());

        let mut records = self.records.lock().unwrap();
        records.insert(record.id.clone(), record.clone());

        Ok(record)
    }

    fn allocate_space(
        &self,
        vehicle_type: VehicleType,
        user_id: Option<&String>,
        preferred_space_id: Option<&str>,
    ) -> Result<String> {
        let spaces = self.spaces.lock().unwrap();

        if let Some(preferred) = preferred_space_id {
            if let Some(space) = spaces.get(preferred) {
                if space.is_available_for(vehicle_type, user_id) {
                    return Ok(preferred.to_string());
                }
            }
        }

        let priority_order: &[SpaceType] = match vehicle_type {
            VehicleType::Electric => &[SpaceType::Charging, SpaceType::Normal],
            VehicleType::Disabled => &[SpaceType::Accessible, SpaceType::Normal],
            VehicleType::Regular => &[SpaceType::Normal],
        };

        for space_type in priority_order {
            for space in spaces.values() {
                if space.space_type == *space_type && space.is_available_for(vehicle_type, user_id) {
                    return Ok(space.id.clone());
                }
            }
        }

        Err(ParkingError::NoSpaceAvailable)
    }

    pub fn exit_parking(&self, plate_number: &str) -> Result<ParkingRecord> {
        let active_vehicles = self.active_vehicles.lock().unwrap();
        let record_id = active_vehicles
            .get(plate_number)
            .ok_or_else(|| ParkingError::VehicleNotFound(plate_number.to_string()))?
            .clone();
        drop(active_vehicles);

        let mut records = self.records.lock().unwrap();
        let record = records
            .get_mut(&record_id)
            .ok_or_else(|| ParkingError::ParkingRecordNotFound(record_id.clone()))?;

        let exit_time = Utc::now();
        record.exit_time = Some(exit_time);

        let fee = self.calculate_fee(record, exit_time);
        record.fee = Some(fee);

        let space_id = record.space_id.clone();
        let mut spaces = self.spaces.lock().unwrap();
        if let Some(space) = spaces.get_mut(&space_id) {
            space.is_occupied = false;
        }

        let mut active_vehicles = self.active_vehicles.lock().unwrap();
        active_vehicles.remove(plate_number);

        Ok(record.clone())
    }

    fn calculate_fee(&self, record: &ParkingRecord, exit_time: DateTime<Utc>) -> f64 {
        if record.is_monthly {
            if let Some(card) = self.get_monthly_card(&record.plate_number) {
                if card.is_active() {
                    return 0.0;
                }

                let hours_since_expiry = (exit_time - card.end_time).num_hours();
                if hours_since_expiry <= 24 {
                    return 0.0;
                }

                let charge_start = card.end_time + Duration::hours(24);
                if exit_time <= charge_start {
                    return 0.0;
                }
                let billable_duration = exit_time - charge_start;
                return self.calculate_temporary_fee(billable_duration);
            }
        }

        let duration = exit_time - record.entry_time;
        self.calculate_temporary_fee(duration)
    }

    fn calculate_temporary_fee(&self, duration: Duration) -> f64 {
        let total_seconds = duration.num_seconds();
        if total_seconds <= 0 {
            return 0.0;
        }

        let total_hours_float = total_seconds as f64 / 3600.0;
        let full_days = (total_hours_float / 24.0) as i64;
        let remaining_hours_float = total_hours_float - (full_days as f64 * 24.0);

        let daily_fee = self.calculate_single_day_fee(remaining_hours_float);
        (full_days as f64 * TWENTY_FOUR_HOUR_CAP) + daily_fee
    }

    fn calculate_single_day_fee(&self, hours_float: f64) -> f64 {
        if hours_float <= 0.5 {
            return 0.0;
        }

        let mut fee = 0.0;

        if hours_float > 0.5 {
            fee += 4.0;

            if hours_float > 1.0 {
                fee += 6.0;

                if hours_float > 2.0 {
                    let remaining = hours_float - 2.0;
                    let whole_hours = remaining.ceil() as i64;
                    fee += whole_hours as f64 * 8.0;
                }
            }
        }

        if fee > TWENTY_FOUR_HOUR_CAP {
            TWENTY_FOUR_HOUR_CAP
        } else {
            fee
        }
    }

    pub fn get_active_record(&self, plate_number: &str) -> Option<ParkingRecord> {
        let active_vehicles = self.active_vehicles.lock().unwrap();
        let record_id = active_vehicles.get(plate_number)?;
        let records = self.records.lock().unwrap();
        records.get(record_id).cloned()
    }

    pub fn get_all_records(&self) -> Vec<ParkingRecord> {
        let records = self.records.lock().unwrap();
        records.values().cloned().collect()
    }

    pub fn get_record(&self, record_id: &str) -> Option<ParkingRecord> {
        let records = self.records.lock().unwrap();
        records.get(record_id).cloned()
    }
}

impl Default for ParkingService {
    fn default() -> Self {
        Self::new()
    }
}
