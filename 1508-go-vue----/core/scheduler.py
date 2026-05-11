from __future__ import annotations

import math
from typing import List, Optional

from .models import (
    DeliveryTask,
    Location,
    TaskStatus,
    TemperatureConflict,
    Vehicle,
    VehicleStatus,
)
from .store import DataStore


class Scheduler:
    def __init__(self, store: DataStore) -> None:
        self.store = store

    @staticmethod
    def _calculate_distance(loc1: Location, loc2: Location) -> float:
        lat_diff = loc1.latitude - loc2.latitude
        lon_diff = loc1.longitude - loc2.longitude
        return math.sqrt(lat_diff**2 + lon_diff**2) * 111.0

    def _can_accept_task(self, vehicle: Vehicle, task: DeliveryTask) -> bool:
        if vehicle.status == VehicleStatus.MAINTENANCE:
            return False

        if vehicle.current_load + task.weight > vehicle.max_load:
            return False

        for assigned_task_id in vehicle.assigned_tasks:
            assigned_task = self.store.get_task(assigned_task_id)
            if not assigned_task:
                continue
            if assigned_task.status == TaskStatus.COMPLETED:
                continue

            temp_overlap = (
                task.required_max_temp >= assigned_task.required_min_temp
                and task.required_min_temp <= assigned_task.required_max_temp
            )
            if not temp_overlap:
                return False

        return True

    def _score_vehicle(self, vehicle: Vehicle, task: DeliveryTask) -> float:
        origin = self.store.get_location(task.origin_id)
        if not origin:
            return float("inf")

        vehicle_loc = self.store.get_location(vehicle.current_location_id)
        if not vehicle_loc:
            vehicle_loc = origin

        distance = self._calculate_distance(vehicle_loc, origin)

        temp_score = abs(vehicle.current_temperature - (task.required_min_temp + task.required_max_temp) / 2)

        load_score = (vehicle.current_load + task.weight) / vehicle.max_load

        return distance * 10.0 + temp_score * 5.0 + load_score * 2.0

    def assign_task(self, task: DeliveryTask) -> Optional[Vehicle]:
        eligible_vehicles: List[Vehicle] = []

        for vehicle in self.store.list_vehicles():
            if self._can_accept_task(vehicle, task):
                eligible_vehicles.append(vehicle)

        if not eligible_vehicles:
            return None

        eligible_vehicles.sort(key=lambda v: self._score_vehicle(v, task))
        selected_vehicle = eligible_vehicles[0]

        task.assigned_vehicle_id = selected_vehicle.id
        task.status = TaskStatus.ASSIGNED
        selected_vehicle.current_load += task.weight
        selected_vehicle.assigned_tasks.append(task.id)
        selected_vehicle.status = VehicleStatus.IN_TRANSIT

        self.store.update_task(task)
        self.store.update_vehicle(selected_vehicle)

        return selected_vehicle

    def start_task(self, task_id: str) -> DeliveryTask:
        task = self.store.get_task(task_id)
        if not task:
            raise ValueError(f"Task {task_id} not found")

        if task.status not in (TaskStatus.ASSIGNED, TaskStatus.TEMPERATURE_ABNORMAL):
            raise ValueError(f"Task {task_id} is not in a startable state")

        from datetime import datetime

        task.status = TaskStatus.IN_TRANSIT
        task.started_at = datetime.utcnow()
        self.store.update_task(task)
        return task
