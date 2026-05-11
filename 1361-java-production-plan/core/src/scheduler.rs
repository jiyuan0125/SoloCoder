use chrono::{DateTime, Duration, Utc};
use std::collections::HashSet;

use crate::errors::SchedulerError;
use crate::models::{MaintenanceWindow, Order, OrderPriority, ScheduleResult, ScheduledTask, Warning};
use crate::store::InMemoryStore;

pub struct Scheduler {
    store: InMemoryStore,
}

impl Scheduler {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn add_urgent_order(&self, order: Order) -> Result<ScheduleResult, SchedulerError> {
        if order.assigned_line_id.is_some() {
            return Err(SchedulerError::OrderAlreadyAssigned);
        }

        let all_lines = self.store.get_all_lines();
        if all_lines.is_empty() {
            return Err(SchedulerError::NoAvailableLines);
        }

        let mut best_result: Option<ScheduleResult> = None;
        let mut best_warning_count = usize::MAX;

        for line in &all_lines {
            let result = self.simulate_insert_urgent(&order, &line.id)?;
            let warning_count = result.warnings.len();
            
            if warning_count < best_warning_count {
                best_warning_count = warning_count;
                best_result = Some(result);
            }
        }

        if let Some(mut result) = best_result {
            let line_id = result.tasks.first()
                .map(|t| t.line_id.clone())
                .ok_or(SchedulerError::InternalError("No line assigned".to_string()))?;
            
            let mut final_order = order.clone();
            final_order.assigned_line_id = Some(line_id.clone());
            self.store.add_order(final_order);
            self.reschedule_line(&line_id)?;
            result.tasks = self.store.get_scheduled_tasks(&line_id);
            Ok(result)
        } else {
            Err(SchedulerError::NoAvailableLines)
        }
    }

    fn simulate_insert_urgent(&self, order: &Order, line_id: &str) -> Result<ScheduleResult, SchedulerError> {
        let mut orders = self.get_unscheduled_orders_for_line(line_id);
        orders.insert(0, order.clone());
        
        let line_windows = self.store.get_maintenance_windows_for_line(line_id);
        self.schedule_orders_for_line(&orders, line_id, &line_windows)
    }

    pub fn add_normal_order(&self, order: Order) -> Result<ScheduleResult, SchedulerError> {
        if order.assigned_line_id.is_some() {
            return Err(SchedulerError::OrderAlreadyAssigned);
        }

        let all_lines = self.store.get_all_lines();
        if all_lines.is_empty() {
            return Err(SchedulerError::NoAvailableLines);
        }

        let mut best_line: Option<String> = None;
        let mut earliest_end: Option<DateTime<Utc>> = None;

        for line in &all_lines {
            let line_orders = self.get_unscheduled_orders_for_line(&line.id);
            let mut test_orders = line_orders.clone();
            test_orders.push(order.clone());
            
            let line_windows = self.store.get_maintenance_windows_for_line(&line.id);
            let result = self.schedule_orders_for_line(&test_orders, &line.id, &line_windows)?;
            
            if let Some(last_task) = result.tasks.last() {
                if earliest_end.is_none() || last_task.end_time < earliest_end.unwrap() {
                    earliest_end = Some(last_task.end_time);
                    best_line = Some(line.id.clone());
                }
            }
        }

        if let Some(line_id) = best_line {
            let mut final_order = order.clone();
            final_order.assigned_line_id = Some(line_id.clone());
            self.store.add_order(final_order);
            self.reschedule_line(&line_id)?;
            
            let tasks = self.store.get_scheduled_tasks(&line_id);
            Ok(ScheduleResult {
                tasks,
                warnings: Vec::new(),
            })
        } else {
            Err(SchedulerError::NoAvailableLines)
        }
    }

    pub fn reschedule_all(&self) -> Result<ScheduleResult, SchedulerError> {
        let lines = self.store.get_all_lines();
        let mut all_tasks = Vec::new();
        let mut all_warnings = Vec::new();

        for line in &lines {
            let result = self.reschedule_line(&line.id)?;
            all_tasks.extend(result.tasks);
            all_warnings.extend(result.warnings);
        }

        Ok(ScheduleResult {
            tasks: all_tasks,
            warnings: all_warnings,
        })
    }

    pub fn reschedule_line(&self, line_id: &str) -> Result<ScheduleResult, SchedulerError> {
        let orders = self.get_unscheduled_orders_for_line(line_id);
        let line_windows = self.store.get_maintenance_windows_for_line(line_id);
        
        let result = self.schedule_orders_for_line(&orders, line_id, &line_windows)?;
        self.store.set_scheduled_tasks(line_id.to_string(), result.tasks.clone());
        
        Ok(result)
    }

    fn get_unscheduled_orders_for_line(&self, line_id: &str) -> Vec<Order> {
        let mut orders: Vec<Order> = self.store.get_all_orders()
            .into_iter()
            .filter(|o| o.assigned_line_id.as_deref() == Some(line_id))
            .collect();
        
        orders.sort_by(|a, b| {
            b.priority.cmp(&a.priority)
                .then_with(|| a.due_date.cmp(&b.due_date))
        });
        
        orders
    }

    fn schedule_orders_for_line(
        &self,
        orders: &[Order],
        line_id: &str,
        maintenance_windows: &[MaintenanceWindow],
    ) -> Result<ScheduleResult, SchedulerError> {
        let mut tasks: Vec<ScheduledTask> = Vec::new();
        let mut warnings: Vec<Warning> = Vec::new();
        
        let mut current_time = Utc::now();
        let mut sorted_windows: Vec<MaintenanceWindow> = maintenance_windows.to_vec();
        sorted_windows.sort_by(|a, b| a.start_time.cmp(&b.start_time));

        let urgent_inserted = orders.iter().position(|o| o.priority == OrderPriority::Urgent);
        let normal_orders_after_urgent: HashSet<String> = if let Some(urgent_idx) = urgent_inserted {
            orders[urgent_idx + 1..]
                .iter()
                .filter(|o| o.priority == OrderPriority::Normal)
                .map(|o| o.id.clone())
                .collect()
        } else {
            HashSet::new()
        };

        for order in orders {
            let duration = Duration::minutes(order.duration_minutes as i64);
            let (start, end) = self.find_next_available_slot(
                current_time,
                duration,
                &sorted_windows,
                &tasks,
            );

            let task = ScheduledTask {
                order_id: order.id.clone(),
                line_id: line_id.to_string(),
                start_time: start,
                end_time: end,
            };

            if normal_orders_after_urgent.contains(&order.id) && end > order.due_date {
                warnings.push(Warning {
                    order_id: order.id.clone(),
                    message: format!(
                        "Order {} will be overdue: scheduled end {} > due date {}",
                        order.id,
                        end.format("%Y-%m-%d %H:%M"),
                        order.due_date.format("%Y-%m-%d %H:%M")
                    ),
                });
            }

            tasks.push(task);
            current_time = end;
        }

        Ok(ScheduleResult { tasks, warnings })
    }

    fn find_next_available_slot(
        &self,
        earliest_start: DateTime<Utc>,
        duration: Duration,
        maintenance_windows: &[MaintenanceWindow],
        existing_tasks: &[ScheduledTask],
    ) -> (DateTime<Utc>, DateTime<Utc>) {
        let mut current_start = earliest_start;
        
        loop {
            let end = current_start + duration;
            let mut conflict = false;

            for window in maintenance_windows {
                if self.time_overlaps(current_start, end, window.start_time, window.end_time) {
                    current_start = window.end_time;
                    conflict = true;
                    break;
                }
            }

            if !conflict {
                for task in existing_tasks {
                    if self.time_overlaps(current_start, end, task.start_time, task.end_time) {
                        current_start = task.end_time;
                        conflict = true;
                        break;
                    }
                }
            }

            if !conflict {
                return (current_start, current_start + duration);
            }
        }
    }

    fn time_overlaps(
        &self,
        s1: DateTime<Utc>,
        e1: DateTime<Utc>,
        s2: DateTime<Utc>,
        e2: DateTime<Utc>,
    ) -> bool {
        s1 < e2 && s2 < e1
    }

    pub fn get_schedule(&self) -> Vec<ScheduledTask> {
        self.store.get_all_scheduled_tasks()
    }

    pub fn get_schedule_for_line(&self, line_id: &str) -> Vec<ScheduledTask> {
        self.store.get_scheduled_tasks(line_id)
    }
}
