use chrono::NaiveDate;
use std::collections::HashMap;
use uuid::Uuid;

use crate::errors::SchedulerError;
use crate::models::*;

#[derive(Debug, Clone, Default)]
pub struct Scheduler {
    patients: HashMap<Uuid, Patient>,
    machines: HashMap<Uuid, Machine>,
    schedules: HashMap<Uuid, ScheduleEntry>,
    maintenances: HashMap<Uuid, Maintenance>,
}

impl Scheduler {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add_patient(&mut self, patient: Patient) -> &Patient {
        let id = patient.id;
        self.patients.insert(id, patient);
        self.patients.get(&id).unwrap()
    }

    pub fn get_patient(&self, id: &Uuid) -> Result<&Patient, SchedulerError> {
        self.patients
            .get(id)
            .ok_or_else(|| SchedulerError::PatientNotFound(id.to_string()))
    }

    pub fn list_patients(&self) -> Vec<&Patient> {
        self.patients.values().collect()
    }

    pub fn add_machine(&mut self, machine: Machine) -> &Machine {
        let id = machine.id;
        self.machines.insert(id, machine);
        self.machines.get(&id).unwrap()
    }

    pub fn get_machine(&self, id: &Uuid) -> Result<&Machine, SchedulerError> {
        self.machines
            .get(id)
            .ok_or_else(|| SchedulerError::MachineNotFound(id.to_string()))
    }

    pub fn list_machines(&self) -> Vec<&Machine> {
        self.machines.values().collect()
    }

    pub fn list_schedules(&self) -> Vec<&ScheduleEntry> {
        self.schedules.values().collect()
    }

    pub fn get_schedule(&self, id: &Uuid) -> Result<&ScheduleEntry, SchedulerError> {
        self.schedules
            .get(id)
            .ok_or_else(|| SchedulerError::ScheduleNotFound(id.to_string()))
    }

    pub fn list_maintenances(&self) -> Vec<&Maintenance> {
        self.maintenances.values().collect()
    }

    fn is_machine_available(
        &self,
        machine_id: &Uuid,
        date: NaiveDate,
        time_slot: TimeSlot,
        exclude_schedule_id: Option<&Uuid>,
    ) -> bool {
        let occupied = self.schedules.values().any(|s| {
            if exclude_schedule_id.map_or(false, |id| id == &s.id) {
                return false;
            }
            s.machine_id == *machine_id && s.date == date && s.time_slot == time_slot
        });

        if occupied {
            return false;
        }

        let under_maintenance = self.maintenances.values().any(|m| {
            m.machine_id == *machine_id && m.date == date && m.time_slot == time_slot
        });

        !under_maintenance
    }

    fn is_patient_available(
        &self,
        patient_id: &Uuid,
        date: NaiveDate,
        exclude_schedule_id: Option<&Uuid>,
    ) -> bool {
        !self.schedules.values().any(|s| {
            if exclude_schedule_id.map_or(false, |id| id == &s.id) {
                return false;
            }
            s.patient_id == *patient_id && s.date == date
        })
    }

    pub fn create_schedule(
        &mut self,
        request: CreateScheduleRequest,
    ) -> Result<&ScheduleEntry, SchedulerError> {
        let patient = self.get_patient(&request.patient_id)?;
        let machine = self.get_machine(&request.machine_id)?;

        if !machine.can_treat(&patient.disease) {
            if machine.is_hepatitis_b_only {
                return Err(SchedulerError::NonHepatitisBCannotUseHepatitisBMachine);
            } else {
                return Err(SchedulerError::HepatitisBMachineRequired);
            }
        }

        if !self.is_machine_available(&request.machine_id, request.date, request.time_slot, None) {
            return Err(SchedulerError::MachineAlreadyOccupied);
        }

        if !self.is_patient_available(&request.patient_id, request.date, None) {
            return Err(SchedulerError::PatientAlreadyScheduledOnDay);
        }

        let entry = ScheduleEntry::new(
            request.patient_id,
            request.machine_id,
            request.date,
            request.time_slot,
        );

        let id = entry.id;
        self.schedules.insert(id, entry);
        Ok(self.schedules.get(&id).unwrap())
    }

    pub fn bulk_create_schedules(
        &mut self,
        request: BulkCreateScheduleRequest,
    ) -> Result<Vec<&ScheduleEntry>, SchedulerError> {
        for (i, req) in request.schedules.iter().enumerate() {
            let patient = self.get_patient(&req.patient_id)?;
            let machine = self.get_machine(&req.machine_id)?;

            if !machine.can_treat(&patient.disease) {
                return Err(if machine.is_hepatitis_b_only {
                    SchedulerError::NonHepatitisBCannotUseHepatitisBMachine
                } else {
                    SchedulerError::HepatitisBMachineRequired
                });
            }

            if !self.is_machine_available(&req.machine_id, req.date, req.time_slot, None) {
                return Err(SchedulerError::MachineAlreadyOccupied);
            }

            if !self.is_patient_available(&req.patient_id, req.date, None) {
                return Err(SchedulerError::PatientAlreadyScheduledOnDay);
            }

            for (j, other) in request.schedules.iter().enumerate() {
                if i == j {
                    continue;
                }

                if req.machine_id == other.machine_id
                    && req.date == other.date
                    && req.time_slot == other.time_slot
                {
                    return Err(SchedulerError::BulkScheduleConflict);
                }

                if req.patient_id == other.patient_id && req.date == other.date {
                    return Err(SchedulerError::BulkScheduleConflict);
                }
            }
        }

        let mut ids = Vec::new();
        for req in request.schedules {
            let entry = ScheduleEntry::new(req.patient_id, req.machine_id, req.date, req.time_slot);
            ids.push(entry.id);
            self.schedules.insert(entry.id, entry);
        }

        Ok(ids.iter().map(|id| self.schedules.get(id).unwrap()).collect())
    }

    pub fn cancel_schedule(&mut self, schedule_id: &Uuid) -> Result<ScheduleEntry, SchedulerError> {
        self.schedules
            .remove(schedule_id)
            .ok_or_else(|| SchedulerError::ScheduleNotFound(schedule_id.to_string()))
    }

    pub fn reschedule(
        &mut self,
        request: RescheduleRequest,
    ) -> Result<&ScheduleEntry, SchedulerError> {
        let original = self
            .schedules
            .get(&request.schedule_id)
            .cloned()
            .ok_or_else(|| SchedulerError::ScheduleNotFound(request.schedule_id.to_string()))?;

        let patient = self.get_patient(&original.patient_id)?;

        let machine_id = if let Some(id) = request.new_machine_id {
            let machine = self.get_machine(&id)?;
            if !machine.can_treat(&patient.disease) {
                return Err(if machine.is_hepatitis_b_only {
                    SchedulerError::NonHepatitisBCannotUseHepatitisBMachine
                } else {
                    SchedulerError::HepatitisBMachineRequired
                });
            }
            id
        } else {
            original.machine_id
        };

        if !self.is_machine_available(
            &machine_id,
            request.new_date,
            request.new_time_slot,
            Some(&request.schedule_id),
        ) {
            return Err(SchedulerError::RescheduleTargetNotAvailable);
        }

        if !self.is_patient_available(
            &original.patient_id,
            request.new_date,
            Some(&request.schedule_id),
        ) {
            return Err(SchedulerError::PatientAlreadyScheduledOnDay);
        }

        self.schedules.remove(&request.schedule_id);

        let new_entry = ScheduleEntry::new(
            original.patient_id,
            machine_id,
            request.new_date,
            request.new_time_slot,
        );

        let id = new_entry.id;
        self.schedules.insert(id, new_entry);
        Ok(self.schedules.get(&id).unwrap())
    }

    pub fn create_maintenance(
        &mut self,
        request: CreateMaintenanceRequest,
    ) -> Result<(Maintenance, Vec<RescheduleResult>), SchedulerError> {
        let _machine = self.get_machine(&request.machine_id)?;

        if !self.is_machine_available(&request.machine_id, request.date, request.time_slot, None) {
            if !request.is_emergency {
                return Err(SchedulerError::MachineAlreadyOccupied);
            }
        }

        let affected_schedules: Vec<ScheduleEntry> = self
            .schedules
            .values()
            .filter(|s| {
                s.machine_id == request.machine_id
                    && s.date == request.date
                    && s.time_slot == request.time_slot
            })
            .cloned()
            .collect();

        let maintenance = Maintenance::new(
            request.machine_id,
            request.date,
            request.time_slot,
            request.is_emergency,
            request.description,
        );

        let maint_id = maintenance.id;
        self.maintenances.insert(maint_id, maintenance.clone());

        let mut results = Vec::new();

        for schedule in affected_schedules {
            let patient = self.get_patient(&schedule.patient_id).unwrap();

            let alternative_machine = self.machines.values().find(|m| {
                m.can_treat(&patient.disease)
                    && self.is_machine_available(&m.id, schedule.date, schedule.time_slot, None)
            });

            if let Some(alt_machine) = alternative_machine {
                self.schedules.remove(&schedule.id);
                let new_entry = ScheduleEntry::new(
                    schedule.patient_id,
                    alt_machine.id,
                    schedule.date,
                    schedule.time_slot,
                );
                let new_id = new_entry.id;
                self.schedules.insert(new_id, new_entry);

                results.push(RescheduleResult {
                    original_schedule_id: schedule.id,
                    new_schedule: Some(self.schedules.get(&new_id).unwrap().clone()),
                    canceled: false,
                    message: format!("已调换到机器 {}", alt_machine.name),
                });
            } else {
                self.schedules.remove(&schedule.id);
                results.push(RescheduleResult {
                    original_schedule_id: schedule.id,
                    new_schedule: None,
                    canceled: true,
                    message: "无可用替代机器，已取消".to_string(),
                });
            }
        }

        Ok((maintenance, results))
    }

    pub fn cancel_maintenance(&mut self, maintenance_id: &Uuid) -> Result<Maintenance, SchedulerError> {
        self.maintenances
            .remove(maintenance_id)
            .ok_or_else(|| SchedulerError::MaintenanceNotFound(maintenance_id.to_string()))
    }

    pub fn get_schedules_by_date(&self, date: NaiveDate) -> Vec<&ScheduleEntry> {
        self.schedules
            .values()
            .filter(|s| s.date == date)
            .collect()
    }

    pub fn get_maintenances_by_date(&self, date: NaiveDate) -> Vec<&Maintenance> {
        self.maintenances
            .values()
            .filter(|m| m.date == date)
            .collect()
    }
}
