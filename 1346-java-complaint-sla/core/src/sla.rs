use chrono::{DateTime, Datelike, Duration, Timelike, Utc, Weekday};

const WORK_START_HOUR: u32 = 9;
const WORK_END_HOUR: u32 = 18;

fn is_working_day(weekday: Weekday) -> bool {
    matches!(weekday, Weekday::Mon | Weekday::Tue | Weekday::Wed | Weekday::Thu | Weekday::Fri)
}

fn is_working_time(dt: DateTime<Utc>) -> bool {
    if !is_working_day(dt.weekday()) {
        return false;
    }
    let hour = dt.hour();
    hour >= WORK_START_HOUR && hour < WORK_END_HOUR
}

fn next_working_day_start(dt: DateTime<Utc>) -> DateTime<Utc> {
    let mut current = dt.date_naive().and_hms_opt(WORK_START_HOUR, 0, 0).unwrap();
    loop {
        current = current + Duration::days(1);
        if is_working_day(current.weekday()) {
            break;
        }
    }
    current.and_utc()
}

pub fn calculate_deadline(start: DateTime<Utc>, working_hours: u32) -> DateTime<Utc> {
    let mut remaining = Duration::hours(working_hours as i64);
    let mut current = start;

    while remaining.num_seconds() > 0 {
        if !is_working_time(current) {
            current = next_working_day_start(current);
            continue;
        }

        let today_end = current
            .date_naive()
            .and_hms_opt(WORK_END_HOUR, 0, 0)
            .unwrap()
            .and_utc();
        
        let available_today = today_end - current;
        
        if available_today >= remaining {
            current = current + remaining;
            remaining = Duration::zero();
        } else {
            remaining = remaining - available_today;
            current = next_working_day_start(today_end);
        }
    }

    current
}

pub fn calculate_working_hours_between(start: DateTime<Utc>, end: DateTime<Utc>) -> Duration {
    if end <= start {
        return Duration::zero();
    }

    let mut total = Duration::zero();
    let mut current = start;

    while current < end {
        if !is_working_time(current) {
            current = next_working_day_start(current);
            continue;
        }

        let today_end = current
            .date_naive()
            .and_hms_opt(WORK_END_HOUR, 0, 0)
            .unwrap()
            .and_utc();

        let period_end = std::cmp::min(today_end, end);
        total = total + (period_end - current);
        current = period_end;

        if current == today_end {
            current = next_working_day_start(current);
        }
    }

    total
}

pub fn is_deadline_passed(deadline: DateTime<Utc>, now: DateTime<Utc>) -> bool {
    now > deadline
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::TimeZone;

    #[test]
    fn test_is_working_time() {
        let monday_10am = Utc.with_ymd_and_hms(2024, 5, 13, 10, 0, 0).unwrap();
        assert!(is_working_time(monday_10am));

        let monday_8am = Utc.with_ymd_and_hms(2024, 5, 13, 8, 0, 0).unwrap();
        assert!(!is_working_time(monday_8am));

        let monday_7pm = Utc.with_ymd_and_hms(2024, 5, 13, 19, 0, 0).unwrap();
        assert!(!is_working_time(monday_7pm));

        let saturday_10am = Utc.with_ymd_and_hms(2024, 5, 11, 10, 0, 0).unwrap();
        assert!(!is_working_time(saturday_10am));
    }

    #[test]
    fn test_calculate_deadline_same_day() {
        let start = Utc.with_ymd_and_hms(2024, 5, 13, 10, 0, 0).unwrap();
        let deadline = calculate_deadline(start, 2);
        let expected = Utc.with_ymd_and_hms(2024, 5, 13, 12, 0, 0).unwrap();
        assert_eq!(deadline, expected);
    }

    #[test]
    fn test_calculate_deadline_cross_day() {
        let start = Utc.with_ymd_and_hms(2024, 5, 13, 17, 0, 0).unwrap();
        let deadline = calculate_deadline(start, 3);
        let expected = Utc.with_ymd_and_hms(2024, 5, 14, 11, 0, 0).unwrap();
        assert_eq!(deadline, expected);
    }

    #[test]
    fn test_calculate_deadline_friday_to_monday() {
        let start = Utc.with_ymd_and_hms(2024, 5, 10, 17, 0, 0).unwrap();
        let deadline = calculate_deadline(start, 24);
        let expected = Utc.with_ymd_and_hms(2024, 5, 15, 14, 0, 0).unwrap();
        assert_eq!(deadline, expected);
    }

    #[test]
    fn test_calculate_deadline_friday_late() {
        let start = Utc.with_ymd_and_hms(2024, 5, 10, 17, 0, 0).unwrap();
        let deadline = calculate_deadline(start, 2);
        let expected = Utc.with_ymd_and_hms(2024, 5, 13, 10, 0, 0).unwrap();
        assert_eq!(deadline, expected);
    }

    #[test]
    fn test_calculate_working_hours_between() {
        let start = Utc.with_ymd_and_hms(2024, 5, 13, 10, 0, 0).unwrap();
        let end = Utc.with_ymd_and_hms(2024, 5, 13, 12, 0, 0).unwrap();
        let hours = calculate_working_hours_between(start, end);
        assert_eq!(hours.num_hours(), 2);
    }

    #[test]
    fn test_calculate_working_hours_weekend() {
        let start = Utc.with_ymd_and_hms(2024, 5, 10, 17, 0, 0).unwrap();
        let end = Utc.with_ymd_and_hms(2024, 5, 13, 10, 0, 0).unwrap();
        let hours = calculate_working_hours_between(start, end);
        assert_eq!(hours.num_hours(), 2);
    }
}
