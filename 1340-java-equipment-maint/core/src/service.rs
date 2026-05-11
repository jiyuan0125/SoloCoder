use std::collections::HashMap;
use std::sync::Arc;

use chrono::{DateTime, Utc};
use parking_lot::RwLock;
use uuid::Uuid;

use crate::models::*;

#[derive(Debug, Clone, Default)]
pub struct MaintenanceService {
    equipments: Arc<RwLock<HashMap<String, Equipment>>>,
    records: Arc<RwLock<HashMap<String, MaintenanceRecord>>>,
}

impl MaintenanceService {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn create_equipment(&self, req: CreateEquipmentRequest) -> Equipment {
        let id = Uuid::new_v4().to_string();
        let equipment = Equipment {
            id: id.clone(),
            name: req.name,
            install_date: req.install_date,
            total_runtime_hours: 0,
            maintenance_rule: MaintenanceRule {
                hours_interval: req.hours_interval,
                days_interval: req.days_interval,
            },
            last_maintenance_date: None,
            last_maintenance_runtime: None,
        };
        self.equipments.write().insert(id, equipment.clone());
        equipment
    }

    pub fn get_equipment(&self, id: &str) -> Option<Equipment> {
        self.equipments.read().get(id).cloned()
    }

    pub fn list_equipments(&self) -> Vec<Equipment> {
        self.equipments.read().values().cloned().collect()
    }

    pub fn update_runtime(&self, equipment_id: &str, additional_hours: u64) -> Option<Equipment> {
        let mut equipments = self.equipments.write();
        let equipment = equipments.get_mut(equipment_id)?;
        equipment.total_runtime_hours += additional_hours;
        Some(equipment.clone())
    }

    pub fn check_maintenance_needed(&self, equipment: &Equipment, now: DateTime<Utc>) -> Option<MaintenanceReminder> {
        let rule = &equipment.maintenance_rule;

        if rule.hours_interval.is_none() && rule.days_interval.is_none() {
            return None;
        }

        let hours_since = if let Some(last_runtime) = equipment.last_maintenance_runtime {
            equipment.total_runtime_hours.saturating_sub(last_runtime)
        } else {
            equipment.total_runtime_hours
        };

        let days_since = if let Some(last_date) = equipment.last_maintenance_date {
            (now - last_date).num_days() as u32
        } else {
            (now - equipment.install_date).num_days() as u32
        };

        let hours_due = rule
            .hours_interval
            .map(|interval| hours_since >= interval)
            .unwrap_or(false);

        let days_due = rule
            .days_interval
            .map(|interval| days_since >= interval)
            .unwrap_or(false);

        if !hours_due && !days_due {
            return None;
        }

        let trigger_type = match (hours_due, days_due) {
            (true, true) => MaintenanceType::Both,
            (true, false) => MaintenanceType::Hourly,
            (false, true) => MaintenanceType::Calendar,
            (false, false) => unreachable!(),
        };

        let hours_remaining = rule
            .hours_interval
            .map(|interval| interval as i64 - hours_since as i64);

        let days_remaining = rule
            .days_interval
            .map(|interval| interval as i32 - days_since as i32);

        Some(MaintenanceReminder {
            equipment_id: equipment.id.clone(),
            equipment_name: equipment.name.clone(),
            trigger_type,
            hours_since_last: rule.hours_interval.map(|_| hours_since),
            days_since_last: rule.days_interval.map(|_| days_since),
            hours_remaining,
            days_remaining,
        })
    }

    pub fn get_all_reminders(&self, now: DateTime<Utc>) -> Vec<MaintenanceReminder> {
        self.equipments
            .read()
            .values()
            .filter_map(|eq| self.check_maintenance_needed(eq, now))
            .collect()
    }

    pub fn perform_maintenance(
        &self,
        equipment_id: &str,
        req: PerformMaintenanceRequest,
    ) -> Option<MaintenanceRecord> {
        let mut equipments = self.equipments.write();
        let equipment = equipments.get_mut(equipment_id)?;

        let record = MaintenanceRecord {
            id: Uuid::new_v4().to_string(),
            equipment_id: equipment.id.clone(),
            maintenance_type: req.maintenance_type,
            date: req.date,
            labor_hours: req.labor_hours,
            materials: req.materials,
            personnel: req.personnel,
            notes: req.notes,
            runtime_at_maintenance: equipment.total_runtime_hours,
        };

        equipment.last_maintenance_date = Some(record.date);
        equipment.last_maintenance_runtime = Some(record.runtime_at_maintenance);

        self.records.write().insert(record.id.clone(), record.clone());
        Some(record)
    }

    pub fn get_records_by_equipment(&self, equipment_id: &str) -> Vec<MaintenanceRecord> {
        self.records
            .read()
            .values()
            .filter(|r| r.equipment_id == equipment_id)
            .cloned()
            .collect()
    }

    pub fn get_records_by_personnel(&self, personnel: &str) -> Vec<MaintenanceRecord> {
        self.records
            .read()
            .values()
            .filter(|r| r.personnel == personnel)
            .cloned()
            .collect()
    }

    pub fn get_records_by_time_range(&self, start: DateTime<Utc>, end: DateTime<Utc>) -> Vec<MaintenanceRecord> {
        self.records
            .read()
            .values()
            .filter(|r| r.date >= start && r.date <= end)
            .cloned()
            .collect()
    }

    pub fn calculate_statistics(&self, records: &[MaintenanceRecord]) -> StatisticsResult {
        let total_labor_hours = records.iter().map(|r| r.labor_hours).sum();
        let total_material_cost = records.iter().map(|r| r.material_cost()).sum();
        let total_cost = records.iter().map(|r| r.total_cost()).sum();
        let record_count = records.len();

        StatisticsResult {
            total_labor_hours,
            total_material_cost,
            total_cost,
            record_count,
        }
    }

    pub fn stats_by_equipment(&self, equipment_id: &str) -> StatisticsResult {
        let records = self.get_records_by_equipment(equipment_id);
        self.calculate_statistics(&records)
    }

    pub fn stats_by_personnel(&self, personnel: &str) -> StatisticsResult {
        let records = self.get_records_by_personnel(personnel);
        self.calculate_statistics(&records)
    }

    pub fn stats_by_time_range(&self, start: DateTime<Utc>, end: DateTime<Utc>) -> StatisticsResult {
        let records = self.get_records_by_time_range(start, end);
        self.calculate_statistics(&records)
    }

    pub fn list_all_records(&self) -> Vec<MaintenanceRecord> {
        self.records.read().values().cloned().collect()
    }

    pub fn list_personnel(&self) -> Vec<String> {
        let mut names: Vec<String> = self
            .records
            .read()
            .values()
            .map(|r| r.personnel.clone())
            .collect();
        names.sort();
        names.dedup();
        names
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::TimeZone;

    #[test]
    fn test_create_equipment() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test Machine".into(),
            install_date,
            hours_interval: Some(500),
            days_interval: Some(90),
        });
        assert_eq!(eq.name, "Test Machine");
        assert_eq!(eq.total_runtime_hours, 0);
    }

    #[test]
    fn test_update_runtime() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test".into(),
            install_date,
            hours_interval: Some(500),
            days_interval: Some(90),
        });
        let updated = service.update_runtime(&eq.id, 100).unwrap();
        assert_eq!(updated.total_runtime_hours, 100);
    }

    #[test]
    fn test_hours_maintenance_trigger() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let now = Utc.with_ymd_and_hms(2024, 1, 2, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test".into(),
            install_date,
            hours_interval: Some(100),
            days_interval: None,
        });
        service.update_runtime(&eq.id, 150);
        let eq = service.get_equipment(&eq.id).unwrap();
        let reminder = service.check_maintenance_needed(&eq, now);
        assert!(reminder.is_some());
        assert_eq!(reminder.unwrap().trigger_type, MaintenanceType::Hourly);
    }

    #[test]
    fn test_calendar_maintenance_trigger() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let now = Utc.with_ymd_and_hms(2024, 4, 15, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test".into(),
            install_date,
            hours_interval: None,
            days_interval: Some(90),
        });
        let reminder = service.check_maintenance_needed(&eq, now);
        assert!(reminder.is_some());
        assert_eq!(reminder.unwrap().trigger_type, MaintenanceType::Calendar);
    }

    #[test]
    fn test_both_triggers() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let now = Utc.with_ymd_and_hms(2024, 4, 15, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test".into(),
            install_date,
            hours_interval: Some(500),
            days_interval: Some(90),
        });
        service.update_runtime(&eq.id, 600);
        let eq = service.get_equipment(&eq.id).unwrap();
        let reminder = service.check_maintenance_needed(&eq, now);
        assert!(reminder.is_some());
        assert_eq!(reminder.unwrap().trigger_type, MaintenanceType::Both);
    }

    #[test]
    fn test_perform_maintenance() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test".into(),
            install_date,
            hours_interval: Some(500),
            days_interval: Some(90),
        });
        service.update_runtime(&eq.id, 600);

        let maint_date = Utc.with_ymd_and_hms(2024, 4, 15, 0, 0, 0).unwrap();
        let record = service
            .perform_maintenance(
                &eq.id,
                PerformMaintenanceRequest {
                    maintenance_type: MaintenanceType::Both,
                    date: maint_date,
                    labor_hours: 4.0,
                    materials: vec![MaterialUsage {
                        name: "Oil".into(),
                        quantity: 5.0,
                        unit_cost: 20.0,
                    }],
                    personnel: "张三".into(),
                    notes: Some("常规保养".into()),
                },
            )
            .unwrap();

        assert_eq!(record.material_cost(), 100.0);
        assert_eq!(record.runtime_at_maintenance, 600);

        let eq = service.get_equipment(&eq.id).unwrap();
        assert_eq!(eq.last_maintenance_date, Some(maint_date));
        assert_eq!(eq.last_maintenance_runtime, Some(600));
    }

    #[test]
    fn test_statistics() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let eq = service.create_equipment(CreateEquipmentRequest {
            name: "Test".into(),
            install_date,
            hours_interval: Some(500),
            days_interval: Some(90),
        });

        let maint_date = Utc.with_ymd_and_hms(2024, 4, 15, 0, 0, 0).unwrap();
        service.perform_maintenance(
            &eq.id,
            PerformMaintenanceRequest {
                maintenance_type: MaintenanceType::Both,
                date: maint_date,
                labor_hours: 4.0,
                materials: vec![MaterialUsage {
                    name: "Oil".into(),
                    quantity: 5.0,
                    unit_cost: 20.0,
                }],
                personnel: "张三".into(),
                notes: None,
            },
        );

        let stats = service.stats_by_equipment(&eq.id);
        assert_eq!(stats.total_labor_hours, 4.0);
        assert_eq!(stats.total_material_cost, 100.0);
        assert_eq!(stats.record_count, 1);

        let stats_personnel = service.stats_by_personnel("张三");
        assert_eq!(stats_personnel.total_labor_hours, 4.0);
    }

    #[test]
    fn test_multiple_equipment_reminders() {
        let service = MaintenanceService::new();
        let install_date = Utc.with_ymd_and_hms(2024, 1, 1, 0, 0, 0).unwrap();
        let now = Utc.with_ymd_and_hms(2024, 4, 15, 0, 0, 0).unwrap();

        let eq1 = service.create_equipment(CreateEquipmentRequest {
            name: "Machine 1".into(),
            install_date,
            hours_interval: Some(500),
            days_interval: None,
        });
        let eq2 = service.create_equipment(CreateEquipmentRequest {
            name: "Machine 2".into(),
            install_date,
            hours_interval: None,
            days_interval: Some(90),
        });
        service.create_equipment(CreateEquipmentRequest {
            name: "Machine 3".into(),
            install_date,
            hours_interval: Some(1000),
            days_interval: None,
        });

        service.update_runtime(&eq1.id, 600);

        let reminders = service.get_all_reminders(now);
        assert_eq!(reminders.len(), 2);
        let names: Vec<&str> = reminders.iter().map(|r| r.equipment_name.as_str()).collect();
        assert!(names.contains(&"Machine 1"));
        assert!(names.contains(&"Machine 2"));
        assert!(!names.contains(&"Machine 3"));
    }
}
