use std::sync::Arc;
use std::thread;
use std::time::Duration;

use chrono::Utc;
use tracing::{debug, info};

use crate::models::BookingStatus;
use crate::service::BookingService;
use crate::state::AppState;

pub struct Scheduler {
    state: AppState,
    interval: Duration,
}

impl Scheduler {
    pub fn new(state: AppState, interval: Duration) -> Self {
        Self { state, interval }
    }

    pub fn start(self) -> Arc<Self> {
        let arc_self = Arc::new(self);
        let scheduler = Arc::clone(&arc_self);

        thread::spawn(move || {
            scheduler.run();
        });

        arc_self
    }

    fn run(&self) {
        loop {
            self.tick();
            thread::sleep(self.interval);
        }
    }

    fn tick(&self) {
        let now = Utc::now();

        let bookings_to_start: Vec<_> = {
            let inner = match self.state.inner.read() {
                Ok(inner) => inner,
                Err(e) => {
                    debug!("Failed to acquire read lock: {}", e);
                    return;
                }
            };

            inner
                .bookings
                .values()
                .filter(|b| {
                    b.status == BookingStatus::Confirmed && b.start_time <= now
                })
                .map(|b| b.id)
                .collect()
        };

        for booking_id in bookings_to_start {
            info!("Auto-starting booking: {}", booking_id);
            match BookingService::start_booking(&self.state, booking_id) {
                Ok(_) => info!("Successfully auto-started booking: {}", booking_id),
                Err(e) => debug!("Failed to auto-start booking {}: {}", booking_id, e),
            }
        }

        let bookings_to_complete: Vec<_> = {
            let inner = match self.state.inner.read() {
                Ok(inner) => inner,
                Err(e) => {
                    debug!("Failed to acquire read lock: {}", e);
                    return;
                }
            };

            inner
                .bookings
                .values()
                .filter(|b| {
                    b.status == BookingStatus::InProgress && b.end_time <= now
                })
                .map(|b| b.id)
                .collect()
        };

        for booking_id in bookings_to_complete {
            info!("Auto-completing booking: {}", booking_id);
            match BookingService::complete_booking(&self.state, booking_id) {
                Ok(_) => info!("Successfully auto-completed booking: {}", booking_id),
                Err(e) => debug!("Failed to auto-complete booking {}: {}", booking_id, e),
            }
        }
    }
}
