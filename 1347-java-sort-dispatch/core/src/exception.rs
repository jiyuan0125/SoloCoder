use crate::models::{Exception, ExceptionType, Package, TransferStation};
use chrono::Utc;

pub struct ExceptionDetector;

impl ExceptionDetector {
    pub fn check_transfer_station_backlog(station: &TransferStation) -> Option<Exception> {
        let oldest_queued_time = station.get_oldest_queued_time()?;
        let four_hours = chrono::Duration::hours(4);
        let now = Utc::now();

        if now.signed_duration_since(oldest_queued_time) > four_hours {
            return Some(Exception::new(
                ExceptionType::TransferStationBacklog,
                &format!(
                    "Transfer station {} has packages queued for more than 4 hours. Queue size: {}",
                    station.name,
                    station.queue.len()
                ),
                &station.id,
            ));
        }

        None
    }

    pub fn check_package_lost(package: &Package) -> Option<Exception> {
        if package.is_lost() {
            return Some(Exception::new(
                ExceptionType::PackageLost,
                &format!(
                    "Package {} has no logistics update for more than 7 days after estimated delivery",
                    package.id
                ),
                &package.id,
            ));
        }

        None
    }
}
