use chrono::{DateTime, NaiveTime, Utc};

pub fn is_time_in_range(time: NaiveTime, start: NaiveTime, end: NaiveTime) -> bool {
    if start <= end {
        time >= start && time <= end
    } else {
        time >= start || time <= end
    }
}

pub fn datetime_to_naive_time(dt: DateTime<Utc>) -> NaiveTime {
    dt.naive_utc().time()
}

pub fn duration_minutes(a: DateTime<Utc>, b: DateTime<Utc>) -> i64 {
    (b - a).num_minutes()
}

pub fn duration_seconds(a: DateTime<Utc>, b: DateTime<Utc>) -> i64 {
    (b - a).num_seconds()
}
