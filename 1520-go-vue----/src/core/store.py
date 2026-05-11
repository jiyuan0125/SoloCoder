from typing import Dict, List, Optional, Any
from threading import Lock
import uuid
from datetime import datetime, date

from .models import (
    Pond, BreedingCycle, BreedingRecord, PondTemperatureLog,
    Species, SalesOrder, DeliveryVehicle, DeliveryTask,
    TodoItem, TodoPriority, TodoStatus, InventoryItem,
    WaterType
)
from .validators import validate_sales_quantity, validate_vehicle_availability
from .rules import check_temperature_range, update_todo_priority, find_existing_temperature_todo


def generate_id() -> str:
    return str(uuid.uuid4())


class Store:
    """内存数据存储"""
    
    _instance = None
    _lock = Lock()
    
    def __new__(cls):
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._init_store()
        return cls._instance
    
    def _init_store(self):
        self.ponds: Dict[str, Pond] = {}
        self.species: Dict[str, Species] = {}
        self.breeding_cycles: Dict[str, BreedingCycle] = {}
        self.breeding_records: Dict[str, BreedingRecord] = {}
        self.temperature_logs: Dict[str, PondTemperatureLog] = {}
        self.inventory: Dict[str, InventoryItem] = {}
        self.sales_orders: Dict[str, SalesOrder] = {}
        self.delivery_vehicles: Dict[str, DeliveryVehicle] = {}
        self.delivery_tasks: Dict[str, DeliveryTask] = {}
        self.todos: Dict[str, TodoItem] = {}
        self._init_sample_data()
    
    def _init_sample_data(self):
        self.create_species(Species(
            id=generate_id(),
            name="南美白对虾",
            min_temperature=25.0,
            max_temperature=32.0,
            description="适应温水环境"
        ))
        self.create_species(Species(
            id=generate_id(),
            name="罗非鱼",
            min_temperature=18.0,
            max_temperature=30.0,
            description="热带淡水鱼"
        ))
        
        self.create_pond(Pond(
            id=generate_id(),
            name="1号池",
            water_type=WaterType.FRESH,
            capacity=5000.0
        ))
        self.create_pond(Pond(
            id=generate_id(),
            name="2号池",
            water_type=WaterType.SALT,
            capacity=8000.0
        ))
        
        self.create_delivery_vehicle(DeliveryVehicle(
            id=generate_id(),
            license_plate="粤A12345",
            name="冷链车1号",
            capacity=100000
        ))
        self.create_delivery_vehicle(DeliveryVehicle(
            id=generate_id(),
            license_plate="粤A67890",
            name="冷链车2号",
            capacity=150000
        ))
    
    def create_species(self, species: Species) -> Species:
        self.species[species.id] = species
        return species
    
    def get_all_species(self) -> List[Species]:
        return list(self.species.values())
    
    def get_species(self, species_id: str) -> Optional[Species]:
        return self.species.get(species_id)
    
    def create_pond(self, pond: Pond) -> Pond:
        self.ponds[pond.id] = pond
        return pond
    
    def get_all_ponds(self) -> List[Pond]:
        return list(self.ponds.values())
    
    def get_pond(self, pond_id: str) -> Optional[Pond]:
        return self.ponds.get(pond_id)
    
    def update_pond(self, pond_id: str, data: Dict[str, Any]) -> Optional[Pond]:
        if pond_id not in self.ponds:
            return None
        existing = self.ponds[pond_id]
        update_data = existing.dict(exclude_unset=True)
        update_data.update(data)
        update_data['updated_at'] = datetime.utcnow()
        updated = Pond(**update_data)
        self.ponds[pond_id] = updated
        return updated
    
    def record_temperature(self, pond_id: str, temperature: float) -> Optional[PondTemperatureLog]:
        pond = self.get_pond(pond_id)
        if not pond:
            return None
        
        log = PondTemperatureLog(
            id=generate_id(),
            pond_id=pond_id,
            temperature=temperature
        )
        self.temperature_logs[log.id] = log
        
        self.update_pond(pond_id, {'current_temperature': temperature})
        
        species = None
        if pond.species_id:
            species = self.get_species(pond.species_id)
        
        is_ok, message = check_temperature_range(temperature, species)
        
        if not is_ok and species:
            existing_todo = find_existing_temperature_todo(
                list(self.todos.values()), 
                pond_id
            )
            
            if existing_todo:
                existing_todo.consecutive_violations += 1
                existing_todo, upgraded = update_todo_priority(
                    existing_todo, 
                    existing_todo.consecutive_violations
                )
                existing_todo.description = message
                existing_todo.updated_at = datetime.utcnow()
                self.todos[existing_todo.id] = existing_todo
            else:
                todo = TodoItem(
                    id=generate_id(),
                    title=f"{pond.name} 水温异常",
                    description=message,
                    priority=TodoPriority.HIGH,
                    status=TodoStatus.PENDING,
                    category="temperature",
                    related_entity_id=pond_id,
                    related_entity_type="pond",
                    consecutive_violations=1
                )
                self.todos[todo.id] = todo
        else:
            existing_todo = find_existing_temperature_todo(
                list(self.todos.values()), 
                pond_id
            )
            if existing_todo:
                existing_todo.status = TodoStatus.RESOLVED
                existing_todo.resolved_at = datetime.utcnow()
                existing_todo.updated_at = datetime.utcnow()
                self.todos[existing_todo.id] = existing_todo
        
        return log
    
    def get_temperature_logs(self, pond_id: Optional[str] = None) -> List[PondTemperatureLog]:
        logs = list(self.temperature_logs.values())
        if pond_id:
            logs = [l for l in logs if l.pond_id == pond_id]
        logs.sort(key=lambda x: x.recorded_at, reverse=True)
        return logs
    
    def create_breeding_cycle(self, cycle: BreedingCycle) -> BreedingCycle:
        self.breeding_cycles[cycle.id] = cycle
        return cycle
    
    def get_all_breeding_cycles(self) -> List[BreedingCycle]:
        return list(self.breeding_cycles.values())
    
    def get_breeding_cycle(self, cycle_id: str) -> Optional[BreedingCycle]:
        return self.breeding_cycles.get(cycle_id)
    
    def update_breeding_cycle(self, cycle_id: str, data: Dict[str, Any]) -> Optional[BreedingCycle]:
        if cycle_id not in self.breeding_cycles:
            return None
        existing = self.breeding_cycles[cycle_id]
        update_data = existing.dict(exclude_unset=True)
        update_data.update(data)
        update_data['updated_at'] = datetime.utcnow()
        updated = BreedingCycle(**update_data)
        self.breeding_cycles[cycle_id] = updated
        
        if 'actual_quantity' in data and data['actual_quantity']:
            self._update_inventory_from_emergence(updated)
        
        return updated
    
    def _update_inventory_from_emergence(self, cycle: BreedingCycle):
        quantity = cycle.actual_quantity or 0
        if quantity <= 0:
            return
        
        existing_item = None
        for item in self.inventory.values():
            if item.species_id == cycle.species_id and item.pond_id == cycle.pond_id:
                existing_item = item
                break
        
        if existing_item:
            existing_item.quantity += quantity
            existing_item.last_updated = datetime.utcnow()
        else:
            item = InventoryItem(
                id=generate_id(),
                species_id=cycle.species_id,
                quantity=quantity,
                pond_id=cycle.pond_id
            )
            self.inventory[item.id] = item
    
    def get_all_inventory(self) -> List[InventoryItem]:
        return list(self.inventory.values())
    
    def get_inventory(self, inventory_id: str) -> Optional[InventoryItem]:
        return self.inventory.get(inventory_id)
    
    def create_sales_order(self, order: SalesOrder) -> Dict:
        inventory_list = self.get_all_inventory()
        
        can_fullfill, message, approved_qty = validate_sales_quantity(
            order.requested_quantity,
            inventory_list,
            order.species_id
        )
        
        order_data = order.dict()
        order_data['approved_quantity'] = approved_qty
        if approved_qty:
            order_data['total_amount'] = approved_qty * order.unit_price
        
        new_order = SalesOrder(**order_data)
        self.sales_orders[new_order.id] = new_order
        
        if approved_qty > 0:
            remaining = approved_qty
            for item in self.inventory.values():
                if item.species_id == order.species_id and remaining > 0:
                    if item.quantity <= remaining:
                        remaining -= item.quantity
                        item.quantity = 0
                    else:
                        item.quantity -= remaining
                        remaining = 0
                    item.last_updated = datetime.utcnow()
        
        return {
            'order': new_order,
            'can_fullfill': can_fullfill,
            'message': message,
            'approved_quantity': approved_qty
        }
    
    def get_all_sales_orders(self) -> List[SalesOrder]:
        return list(self.sales_orders.values())
    
    def get_sales_order(self, order_id: str) -> Optional[SalesOrder]:
        return self.sales_orders.get(order_id)
    
    def create_delivery_vehicle(self, vehicle: DeliveryVehicle) -> DeliveryVehicle:
        self.delivery_vehicles[vehicle.id] = vehicle
        return vehicle
    
    def get_all_delivery_vehicles(self) -> List[DeliveryVehicle]:
        return list(self.delivery_vehicles.values())
    
    def get_delivery_vehicle(self, vehicle_id: str) -> Optional[DeliveryVehicle]:
        return self.delivery_vehicles.get(vehicle_id)
    
    def create_delivery_task(self, task: DeliveryTask) -> Dict:
        all_tasks = list(self.delivery_tasks.values())
        
        is_available, message = validate_vehicle_availability(
            task.vehicle_id,
            task.scheduled_start,
            task.scheduled_end or datetime.max,
            all_tasks
        )
        
        if not is_available:
            return {
                'success': False,
                'message': message,
                'task': None
            }
        
        self.delivery_tasks[task.id] = task
        
        order = self.sales_orders.get(task.order_id)
        if order:
            order.delivery_task_id = task.id
            self.sales_orders[task.order_id] = order
        
        return {
            'success': True,
            'message': '配送任务创建成功',
            'task': task
        }
    
    def get_all_delivery_tasks(self) -> List[DeliveryTask]:
        return list(self.delivery_tasks.values())
    
    def get_delivery_task(self, task_id: str) -> Optional[DeliveryTask]:
        return self.delivery_tasks.get(task_id)
    
    def update_delivery_task(self, task_id: str, data: Dict[str, Any]) -> Optional[DeliveryTask]:
        if task_id not in self.delivery_tasks:
            return None
        existing = self.delivery_tasks[task_id]
        update_data = existing.dict(exclude_unset=True)
        update_data.update(data)
        updated = DeliveryTask(**update_data)
        self.delivery_tasks[task_id] = updated
        return updated
    
    def create_todo(self, todo: TodoItem) -> TodoItem:
        self.todos[todo.id] = todo
        return todo
    
    def get_all_todos(self) -> List[TodoItem]:
        todos = list(self.todos.values())
        todos.sort(key=lambda x: (
            {
                TodoPriority.URGENT: 0,
                TodoPriority.HIGH: 1,
                TodoPriority.MEDIUM: 2,
                TodoPriority.LOW: 3
            }[x.priority],
            x.created_at
        ))
        return todos
    
    def get_todo(self, todo_id: str) -> Optional[TodoItem]:
        return self.todos.get(todo_id)
    
    def update_todo(self, todo_id: str, data: Dict[str, Any]) -> Optional[TodoItem]:
        if todo_id not in self.todos:
            return None
        existing = self.todos[todo_id]
        update_data = existing.dict(exclude_unset=True)
        update_data.update(data)
        update_data['updated_at'] = datetime.utcnow()
        updated = TodoItem(**update_data)
        self.todos[todo_id] = updated
        return updated
    
    def check_inventory_and_generate_todo(self, species_id: str) -> Optional[TodoItem]:
        inventory_list = self.get_all_inventory()
        total = sum(
            item.quantity 
            for item in inventory_list 
            if item.species_id == species_id
        )
        
        if total < 5000:
            species = self.get_species(species_id)
            species_name = species.name if species else species_id
            
            existing_todo = None
            for todo in self.todos.values():
                if (
                    todo.category == "inventory" 
                    and todo.related_entity_id == species_id
                    and todo.status != TodoStatus.RESOLVED
                ):
                    existing_todo = todo
                    break
            
            if not existing_todo:
                todo = TodoItem(
                    id=generate_id(),
                    title=f"{species_name} 库存不足",
                    description=f"当前库存仅 {total} 尾，建议安排新一轮繁育",
                    priority=TodoPriority.HIGH,
                    status=TodoStatus.PENDING,
                    category="inventory",
                    related_entity_id=species_id,
                    related_entity_type="species"
                )
                self.todos[todo.id] = todo
                return todo
        
        return None
