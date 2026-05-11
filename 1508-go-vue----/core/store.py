from __future__ import annotations

from datetime import datetime, timedelta
from typing import Dict, List, Optional

from .models import (
    DeliveryRecord,
    DeliveryTask,
    Location,
    LocationReport,
    TaskStatus,
    TemperatureReport,
    Vehicle,
    VehicleStatus,
)


class DataStore:
    def __init__(self) -> None:
        self.locations: Dict[str, Location] = {}
        self.vehicles: Dict[str, Vehicle] = {}
        self.tasks: Dict[str, DeliveryTask] = {}
        self.delivery_records: Dict[str, DeliveryRecord] = {}

    def add_location(self, location: Location) -> Location:
        self.locations[location.id] = location
        return location

    def get_location(self, location_id: str) -> Optional[Location]:
        return self.locations.get(location_id)

    def list_locations(self) -> List[Location]:
        return list(self.locations.values())

    def add_vehicle(self, vehicle: Vehicle) -> Vehicle:
        self.vehicles[vehicle.id] = vehicle
        return vehicle

    def get_vehicle(self, vehicle_id: str) -> Optional[Vehicle]:
        return self.vehicles.get(vehicle_id)

    def list_vehicles(self) -> List[Vehicle]:
        return list(self.vehicles.values())

    def update_vehicle(self, vehicle: Vehicle) -> Vehicle:
        self.vehicles[vehicle.id] = vehicle
        return vehicle

    def add_task(self, task: DeliveryTask) -> DeliveryTask:
        self.tasks[task.id] = task
        return task

    def get_task(self, task_id: str) -> Optional[DeliveryTask]:
        return self.tasks.get(task_id)

    def list_tasks(self, status: Optional[TaskStatus] = None) -> List[DeliveryTask]:
        tasks = list(self.tasks.values())
        if status:
            tasks = [t for t in tasks if t.status == status]
        return tasks

    def update_task(self, task: DeliveryTask) -> DeliveryTask:
        self.tasks[task.id] = task
        return task

    def add_temperature_report(
        self, task_id: str, vehicle_id: str, temperature: float
    ) -> TemperatureReport:
        task = self.tasks[task_id]
        report = TemperatureReport(
            task_id=task_id,
            vehicle_id=vehicle_id,
            temperature=temperature,
            is_normal=(
                task.required_min_temp <= temperature <= task.required_max_temp
            ),
        )
        task.temperature_reports.append(report)

        vehicle = self.vehicles[vehicle_id]
        vehicle.last_report_time = report.reported_at
        vehicle.current_temperature = temperature

        if not report.is_normal and task.status == TaskStatus.IN_TRANSIT:
            task.status = TaskStatus.TEMPERATURE_ABNORMAL

        self.tasks[task_id] = task
        self.vehicles[vehicle_id] = vehicle
        return report

    def add_location_report(
        self, task_id: str, vehicle_id: str, latitude: float, longitude: float
    ) -> LocationReport:
        task = self.tasks[task_id]
        report = LocationReport(
            task_id=task_id,
            vehicle_id=vehicle_id,
            latitude=latitude,
            longitude=longitude,
        )
        task.location_reports.append(report)
        self.tasks[task_id] = task
        return report

    def check_communication_status(self) -> List[DeliveryTask]:
        now = datetime.utcnow()
        cutoff = now - timedelta(hours=24)
        affected_tasks: List[DeliveryTask] = []

        for task in self.tasks.values():
            if task.status not in (
                TaskStatus.IN_TRANSIT,
                TaskStatus.TEMPERATURE_ABNORMAL,
            ):
                continue

            has_recent_report = any(
                r.reported_at >= cutoff for r in task.temperature_reports
            )
            if not has_recent_report:
                task.status = TaskStatus.COMMUNICATION_LOST
                self.tasks[task.id] = task
                affected_tasks.append(task)

        return affected_tasks

    def complete_task(self, task_id: str) -> DeliveryTask:
        task = self.tasks[task_id]
        task.status = TaskStatus.COMPLETED
        task.completed_at = datetime.utcnow()
        self.tasks[task_id] = task

        vehicle = self.vehicles[task.assigned_vehicle_id]
        vehicle.current_load -= task.weight
        vehicle.assigned_tasks.remove(task_id)
        if not vehicle.assigned_tasks:
            vehicle.status = VehicleStatus.IDLE
        self.vehicles[vehicle.id] = vehicle

        self._create_delivery_record(task)
        return task

    def _create_delivery_record(self, task: DeliveryTask) -> DeliveryRecord:
        vehicle = self.vehicles[task.assigned_vehicle_id]
        origin = self.locations[task.origin_id]
        destination = self.locations[task.destination_id]

        record = DeliveryRecord(
            task_id=task.id,
            order_id=task.order_id,
            vehicle_id=vehicle.id,
            vehicle_number=vehicle.vehicle_number,
            origin_name=origin.name,
            destination_name=destination.name,
            weight=task.weight,
            required_min_temp=task.required_min_temp,
            required_max_temp=task.required_max_temp,
            created_at=task.created_at,
            started_at=task.started_at,
            completed_at=task.completed_at,
            final_status=task.status,
        )
        self.delivery_records[task.id] = record
        return record

    def get_delivery_records(
        self, start_date: datetime, end_date: datetime
    ) -> List[DeliveryRecord]:
        return [
            record
            for record in self.delivery_records.values()
            if start_date <= record.created_at <= end_date
        ]
