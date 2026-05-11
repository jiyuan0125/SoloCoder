import math
from datetime import datetime, timedelta
from typing import List, Optional, Tuple
from .models import (
    Vehicle, 
    VehicleStatus, 
    DeliveryTask, 
    DeliveryTaskStatus, 
    Origin,
    Location
)
from .store import store


def calculate_distance(loc1: Location, loc2: Location) -> float:
    lat1, lon1 = loc1.latitude, loc1.longitude
    lat2, lon2 = loc2.latitude, loc2.longitude
    
    R = 6371.0
    
    lat1_rad = math.radians(lat1)
    lon1_rad = math.radians(lon1)
    lat2_rad = math.radians(lat2)
    lon2_rad = math.radians(lon2)
    
    dlat = lat2_rad - lat1_rad
    dlon = lon2_rad - lon1_rad
    
    a = math.sin(dlat / 2) ** 2 + math.cos(lat1_rad) * math.cos(lat2_rad) * math.sin(dlon / 2) ** 2
    c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))
    
    distance = R * c
    return distance


def can_vehicle_accept_task(vehicle: Vehicle, task: DeliveryTask) -> bool:
    if vehicle.status != VehicleStatus.IDLE and vehicle.status != VehicleStatus.IN_TRANSIT:
        return False
    
    if vehicle.current_load + task.weight > vehicle.max_load:
        return False
    
    for existing_task_id in vehicle.current_task_ids:
        existing_task = store.get_task(existing_task_id)
        if existing_task and existing_task.status != DeliveryTaskStatus.COMPLETED:
            if (task.min_temperature > existing_task.max_temperature or 
                task.max_temperature < existing_task.min_temperature):
                return False
    
    if not (task.min_temperature <= vehicle.current_temperature <= task.max_temperature):
        pass
    
    return True


def get_vehicle_temperature_range(vehicle: Vehicle) -> Tuple[Optional[float], Optional[float]]:
    min_temp = None
    max_temp = None
    
    for task_id in vehicle.current_task_ids:
        task = store.get_task(task_id)
        if task and task.status != DeliveryTaskStatus.COMPLETED:
            if min_temp is None or task.min_temperature < min_temp:
                min_temp = task.min_temperature
            if max_temp is None or task.max_temperature > max_temp:
                max_temp = task.max_temperature
    
    return min_temp, max_temp


def find_best_vehicle(task: DeliveryTask) -> Optional[Vehicle]:
    origin = store.get_origin(task.origin_id)
    if not origin:
        return None
    
    available_vehicles = []
    
    for vehicle in store.list_vehicles():
        if can_vehicle_accept_task(vehicle, task):
            available_vehicles.append(vehicle)
    
    if not available_vehicles:
        return None
    
    scored_vehicles = []
    for vehicle in available_vehicles:
        if vehicle.location:
            distance = calculate_distance(vehicle.location, origin.location)
        else:
            distance = float('inf')
        
        score = distance
        
        temp_match = (
            task.min_temperature <= vehicle.current_temperature <= task.max_temperature
        )
        if not temp_match:
            score += 1000
        
        scored_vehicles.append((score, vehicle))
    
    scored_vehicles.sort(key=lambda x: x[0])
    return scored_vehicles[0][1]


def assign_task_to_vehicle(task: DeliveryTask, vehicle: Vehicle) -> bool:
    if not can_vehicle_accept_task(vehicle, task):
        return False
    
    vehicle.current_load += task.weight
    vehicle.current_task_ids.append(task.id)
    if vehicle.status == VehicleStatus.IDLE:
        vehicle.status = VehicleStatus.IN_TRANSIT
    
    task.vehicle_id = vehicle.id
    task.status = DeliveryTaskStatus.ASSIGNED
    task.assigned_at = datetime.now()
    
    store.update_vehicle(vehicle)
    store.update_task(task)
    
    return True


def complete_task(task: DeliveryTask) -> bool:
    if not task.vehicle_id:
        return False
    
    vehicle = store.get_vehicle(task.vehicle_id)
    if not vehicle:
        return False
    
    vehicle.current_load -= task.weight
    if task.id in vehicle.current_task_ids:
        vehicle.current_task_ids.remove(task.id)
    
    if len(vehicle.current_task_ids) == 0 and vehicle.current_load == 0:
        vehicle.status = VehicleStatus.IDLE
    
    task.status = DeliveryTaskStatus.COMPLETED
    task.completed_at = datetime.now()
    
    store.update_vehicle(vehicle)
    store.update_task(task)
    
    return True


def check_temperature_abnormal(task: DeliveryTask) -> bool:
    if not task.temperature_reports:
        return False
    
    latest_report = task.temperature_reports[-1]
    return not (task.min_temperature <= latest_report.temperature <= task.max_temperature)


def check_communication_lost(task: DeliveryTask) -> bool:
    if task.status == DeliveryTaskStatus.COMPLETED:
        return False
    
    if not task.temperature_reports:
        if task.assigned_at:
            return datetime.now() - task.assigned_at > timedelta(hours=24)
        return False
    
    latest_report_time = task.temperature_reports[-1].timestamp
    return datetime.now() - latest_report_time > timedelta(hours=24)


def update_task_statuses() -> None:
    for task in store.list_tasks():
        if task.status == DeliveryTaskStatus.COMPLETED:
            continue
        
        if check_communication_lost(task):
            task.status = DeliveryTaskStatus.COMMUNICATION_LOST
            store.update_task(task)
        elif check_temperature_abnormal(task):
            if task.status != DeliveryTaskStatus.COMMUNICATION_LOST:
                task.status = DeliveryTaskStatus.TEMPERATURE_ABNORMAL
                store.update_task(task)
