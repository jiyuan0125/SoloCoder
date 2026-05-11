from typing import Dict, List, Optional
from .models import Origin, Vehicle, Order, DeliveryTask


class DataStore:
    def __init__(self):
        self.origins: Dict[str, Origin] = {}
        self.vehicles: Dict[str, Vehicle] = {}
        self.orders: Dict[str, Order] = {}
        self.tasks: Dict[str, DeliveryTask] = {}
    
    def add_origin(self, origin: Origin) -> Origin:
        self.origins[origin.id] = origin
        return origin
    
    def get_origin(self, origin_id: str) -> Optional[Origin]:
        return self.origins.get(origin_id)
    
    def list_origins(self) -> List[Origin]:
        return list(self.origins.values())
    
    def add_vehicle(self, vehicle: Vehicle) -> Vehicle:
        self.vehicles[vehicle.id] = vehicle
        return vehicle
    
    def get_vehicle(self, vehicle_id: str) -> Optional[Vehicle]:
        return self.vehicles.get(vehicle_id)
    
    def list_vehicles(self) -> List[Vehicle]:
        return list(self.vehicles.values())
    
    def update_vehicle(self, vehicle: Vehicle) -> Optional[Vehicle]:
        if vehicle.id in self.vehicles:
            self.vehicles[vehicle.id] = vehicle
            return vehicle
        return None
    
    def add_order(self, order: Order) -> Order:
        self.orders[order.id] = order
        return order
    
    def get_order(self, order_id: str) -> Optional[Order]:
        return self.orders.get(order_id)
    
    def list_orders(self) -> List[Order]:
        return list(self.orders.values())
    
    def add_task(self, task: DeliveryTask) -> DeliveryTask:
        self.tasks[task.id] = task
        return task
    
    def get_task(self, task_id: str) -> Optional[DeliveryTask]:
        return self.tasks.get(task_id)
    
    def list_tasks(self) -> List[DeliveryTask]:
        return list(self.tasks.values())
    
    def update_task(self, task: DeliveryTask) -> Optional[DeliveryTask]:
        if task.id in self.tasks:
            self.tasks[task.id] = task
            return task
        return None


store = DataStore()
