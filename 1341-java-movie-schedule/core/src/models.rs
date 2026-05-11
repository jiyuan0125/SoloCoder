use chrono::{DateTime, Local, NaiveTime, Duration};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum HallType {
    Normal,
    IMAX,
    VIP,
}

impl HallType {
    pub fn seat_count(&self) -> usize {
        match self {
            HallType::Normal => 100,
            HallType::IMAX => 80,
            HallType::VIP => 20,
        }
    }

    pub fn base_price(&self) -> f64 {
        match self {
            HallType::Normal => 50.0,
            HallType::IMAX => 80.0,
            HallType::VIP => 120.0,
        }
    }

    pub fn display_name(&self) -> &'static str {
        match self {
            HallType::Normal => "普通厅",
            HallType::IMAX => "IMAX厅",
            HallType::VIP => "VIP厅",
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Movie {
    pub id: Uuid,
    pub title: String,
    pub duration_minutes: i64,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Hall {
    pub id: Uuid,
    pub name: String,
    pub hall_type: HallType,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Schedule {
    pub id: Uuid,
    pub hall_id: Uuid,
    pub movie_id: Uuid,
    pub start_time: DateTime<Local>,
    pub end_time: DateTime<Local>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ScheduleCreate {
    pub hall_id: Uuid,
    pub movie_id: Uuid,
    pub start_time: DateTime<Local>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ScheduleUpdate {
    pub hall_id: Option<Uuid>,
    pub movie_id: Option<Uuid>,
    pub start_time: Option<DateTime<Local>>,
}

impl Hall {
    pub fn new(name: String, hall_type: HallType) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            hall_type,
        }
    }
}

impl Movie {
    pub fn new(title: String, duration_minutes: i64) -> Self {
        Self {
            id: Uuid::new_v4(),
            title,
            duration_minutes,
        }
    }
}

impl Schedule {
    pub fn new(
        hall_id: Uuid,
        movie_id: Uuid,
        start_time: DateTime<Local>,
        movie_duration: Duration,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            hall_id,
            movie_id,
            start_time,
            end_time: start_time + movie_duration,
        }
    }

    pub fn overlaps_with_cleaning(&self, other: &Schedule, cleaning_minutes: i64) -> bool {
        let self_start = self.start_time;
        let self_end = self.end_time + Duration::minutes(cleaning_minutes);
        let other_start = other.start_time;
        let other_end = other.end_time + Duration::minutes(cleaning_minutes);

        !(self_end <= other_start || other_end <= self_start)
    }

    pub fn is_in_prime_time(&self) -> bool {
        let prime_start = NaiveTime::from_hms_opt(18, 0, 0).unwrap();
        let prime_end = NaiveTime::from_hms_opt(22, 0, 0).unwrap();

        let mut current = self.start_time;
        while current < self.end_time {
            let time = current.time();
            if time >= prime_start && time < prime_end {
                return true;
            }
            current = current + Duration::minutes(60);
        }
        false
    }

    pub fn get_prime_hours(&self) -> Vec<chrono::NaiveDate> {
        let prime_start = NaiveTime::from_hms_opt(18, 0, 0).unwrap();
        let prime_end = NaiveTime::from_hms_opt(22, 0, 0).unwrap();
        let mut dates = std::collections::HashSet::new();

        let mut current = self.start_time;
        while current < self.end_time {
            let time = current.time();
            if time >= prime_start && time < prime_end {
                dates.insert(current.date_naive());
            }
            current = current + Duration::minutes(60);
        }

        dates.into_iter().collect()
    }
}
