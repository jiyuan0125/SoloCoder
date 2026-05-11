use chrono::{DateTime, Duration, Utc};

pub struct FeeCalculator;

impl FeeCalculator {
    pub fn calculate(deposited_at: &DateTime<Utc>, picked_at: &DateTime<Utc>) -> (u64, f64) {
        let duration = *picked_at - *deposited_at;
        let total_seconds = duration.num_seconds();
        
        let total_hours = if total_seconds <= 0 {
            0
        } else {
            (total_seconds + 3599) / 3600
        };

        let total_hours = total_hours as u64;
        let fee = Self::calculate_fee(total_hours);

        (total_hours, fee)
    }

    fn calculate_fee(total_hours: u64) -> f64 {
        if total_hours <= 24 {
            0.0
        } else {
            let hours_24_48 = std::cmp::min(total_hours - 24, 24);
            let hours_48_72 = if total_hours > 48 {
                std::cmp::min(total_hours - 48, 24)
            } else {
                0
            };
            let hours_over_72 = if total_hours > 72 {
                total_hours - 72
            } else {
                0
            };

            (hours_24_48 as f64 * 0.5)
                + (hours_48_72 as f64 * 1.0)
                + (hours_over_72 as f64 * 2.0)
        }
    }
}

pub fn is_locked(locked_until: &Option<DateTime<Utc>>) -> bool {
    match locked_until {
        Some(until) => *until > Utc::now(),
        None => false,
    }
}

pub fn lock_duration_minutes() -> i64 {
    30
}

pub fn max_attempts() -> u32 {
    3
}

pub fn create_lock_time() -> DateTime<Utc> {
    Utc::now() + Duration::minutes(lock_duration_minutes())
}
