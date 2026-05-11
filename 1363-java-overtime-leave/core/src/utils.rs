use chrono::{DateTime, Datelike, Duration, Utc};

use crate::models::OvertimeType;

pub const DAYS_PER_MONTH: f64 = 21.75;
pub const HOURS_PER_DAY: f64 = 8.0;
pub const SUBMISSION_WINDOW_HOURS: i64 = 24;

pub fn calculate_hourly_wage(monthly_salary: f64) -> f64 {
    monthly_salary / DAYS_PER_MONTH / HOURS_PER_DAY
}

pub fn calculate_overtime_pay(
    hourly_wage: f64,
    overtime_type: OvertimeType,
    duration_hours: f64,
) -> f64 {
    let multiplier = match overtime_type {
        OvertimeType::Workday => 1.5,
        OvertimeType::Weekend => 2.0,
        OvertimeType::Holiday => 3.0,
    };
    hourly_wage * duration_hours * multiplier
}

pub fn calculate_duration(start: DateTime<Utc>, end: DateTime<Utc>) -> Result<f64, crate::errors::SystemError> {
    if start >= end {
        return Err(crate::errors::SystemError::InvalidOvertimePeriod);
    }
    let duration = end.signed_duration_since(start);
    let total_minutes = duration.num_minutes() as f64;
    let half_hours = (total_minutes / 30.0).ceil();
    Ok(half_hours * 0.5)
}

pub fn is_overtime_submission_expired(
    end_time: DateTime<Utc>,
    submitted_at: DateTime<Utc>,
) -> bool {
    let deadline = end_time + Duration::hours(SUBMISSION_WINDOW_HOURS);
    submitted_at > deadline
}

pub fn determine_overtime_type(
    start_time: DateTime<Utc>,
    _end_time: DateTime<Utc>,
    is_holiday: bool,
) -> OvertimeType {
    if is_holiday {
        return OvertimeType::Holiday;
    }
    
    let weekday = start_time.weekday();
    if matches!(weekday, chrono::Weekday::Sat | chrono::Weekday::Sun) {
        return OvertimeType::Weekend;
    }
    
    OvertimeType::Workday
}

pub fn is_leave_past(end_time: DateTime<Utc>) -> bool {
    end_time < Utc::now()
}
