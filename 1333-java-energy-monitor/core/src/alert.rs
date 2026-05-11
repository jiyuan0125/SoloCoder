use crate::models::{Alert, PointId};
use chrono::NaiveDate;
use crate::storage::InMemoryStorage;
use std::sync::Arc;

pub const ALERT_THRESHOLD_PERCENT: f64 = 50.0;

pub fn check_alert_for_point(
    storage: &InMemoryStorage,
    point_id: &PointId,
    today: &NaiveDate,
) -> Option<Alert> {
    if storage.should_skip_alert(today) {
        return None;
    }

    let current_record = storage.get_record(point_id, today)?;
    let current_usage = current_record.usage.total();

    if current_usage <= 0.0 {
        return None;
    }

    let (_prev_date, prev_record) = storage.get_latest_usage_before(point_id, today)?;
    let prev_usage = prev_record.usage.total();

    if prev_usage <= 0.0 {
        return None;
    }

    let increase = current_usage - prev_usage;
    let increase_percent = (increase / prev_usage) * 100.0;

    if increase_percent > ALERT_THRESHOLD_PERCENT {
        Some(Alert {
            point_id: point_id.clone(),
            date: *today,
            current_usage,
            previous_usage: prev_usage,
            increase_percent,
        })
    } else {
        None
    }
}

pub fn check_all_alerts(
    storage: Arc<InMemoryStorage>,
    today: NaiveDate,
) -> Vec<Alert> {
    let points = storage.get_all_points();
    let mut alerts = Vec::new();

    for point in points {
        if let Some(alert) = check_alert_for_point(&storage, &point, &today) {
            alerts.push(alert);
        }
    }

    alerts
}
