use std::sync::Arc;
use std::time::Duration;

use chrono::{DateTime, Utc};
use cron::Schedule;
use std::str::FromStr;
use tokio::sync::mpsc;
use tokio::sync::Mutex;
use uuid::Uuid;

use crate::models::{HealthStatus, Probe, ScheduleStrategy};
use crate::probes::execute_probe;
use crate::state::AppState;

pub struct SchedulerHandle {
    pub cancel_tx: mpsc::Sender<()>,
}

impl SchedulerHandle {
    pub fn new(cancel_tx: mpsc::Sender<()>) -> Self {
        SchedulerHandle { cancel_tx }
    }

    pub async fn cancel(&self) {
        let _ = self.cancel_tx.send(()).await;
    }
}

pub async fn start_probe_scheduler(probe: Arc<Probe>, state: AppState) -> SchedulerHandle {
    let (cancel_tx, mut cancel_rx) = mpsc::channel::<()>(1);

    let probe_id = probe.id;
    let state_clone = state.clone();

    tokio::spawn(async move {
        let mut current_interval = match probe.schedule {
            ScheduleStrategy::FixedInterval { interval_secs } => Duration::from_secs(interval_secs),
            ScheduleStrategy::ExponentialBackoff {
                initial_interval_secs,
                ..
            } => Duration::from_secs(initial_interval_secs),
            ScheduleStrategy::Cron { .. } => Duration::from_secs(1),
        };

        let initial_interval = current_interval;

        loop {
            tokio::select! {
                _ = cancel_rx.recv() => {
                    break;
                }
                _ = tokio::time::sleep(current_interval) => {
                    let result = execute_probe(probe_id, &probe.params).await;
                    state_clone.update_result(probe_id, result.clone()).await;

                    match probe.schedule {
                        ScheduleStrategy::FixedInterval { interval_secs } => {
                            current_interval = Duration::from_secs(interval_secs);
                        }
                        ScheduleStrategy::ExponentialBackoff {
                            initial_interval_secs,
                            max_interval_secs,
                        } => {
                            let initial = Duration::from_secs(initial_interval_secs);
                            let max = Duration::from_secs(max_interval_secs);

                            if result.status == HealthStatus::Healthy {
                                current_interval = initial;
                            } else {
                                let doubled = current_interval.mul_f32(2.0);
                                current_interval = if doubled > max { max } else { doubled };
                            }
                        }
                        ScheduleStrategy::Cron { ref expression } => {
                            if let Ok(schedule) = Schedule::from_str(expression) {
                                current_interval = get_next_cron_delay(&schedule);
                            } else {
                                current_interval = Duration::from_secs(60);
                            }
                        }
                    }
                }
            }
        }
    });

    SchedulerHandle::new(cancel_tx)
}

fn get_next_cron_delay(schedule: &Schedule) -> Duration {
    let now: DateTime<Utc> = Utc::now();
    if let Some(next) = schedule.upcoming(Utc).next() {
        let delay = next - now;
        if delay.num_milliseconds() > 0 {
            Duration::from_millis(delay.num_milliseconds() as u64)
        } else {
            Duration::from_secs(1)
        }
    } else {
        Duration::from_secs(60)
    }
}

pub struct SchedulerRegistry {
    handles: Mutex<std::collections::HashMap<Uuid, SchedulerHandle>>,
}

impl SchedulerRegistry {
    pub fn new() -> Self {
        SchedulerRegistry {
            handles: Mutex::new(std::collections::HashMap::new()),
        }
    }

    pub async fn register(&self, probe_id: Uuid, handle: SchedulerHandle) {
        self.handles.lock().await.insert(probe_id, handle);
    }

    pub async fn unregister(&self, probe_id: &Uuid) -> bool {
        if let Some(handle) = self.handles.lock().await.remove(probe_id) {
            handle.cancel().await;
            true
        } else {
            false
        }
    }
}
