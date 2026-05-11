from typing import Dict, List, Optional
from datetime import datetime
from .models import (
    Furnace, BatchingOrder, TemperatureReading, Alert,
    FurnaceStatus, BatchingOrderStatus, AlertType
)


class Storage:
    def __init__(self):
        self.furnaces: Dict[str, Furnace] = {}
        self.batching_orders: Dict[str, BatchingOrder] = {}
        self.temperature_readings: Dict[str, List[TemperatureReading]] = {}
        self.alerts: List[Alert] = []
        self._furnace_counter = 0
        self._order_counter = 0
        self._reading_counter = 0
        self._alert_counter = 0
    
    def generate_furnace_id(self) -> str:
        self._furnace_counter += 1
        return f"FUR-{self._furnace_counter:04d}"
    
    def generate_order_id(self) -> str:
        self._order_counter += 1
        return f"ORD-{self._order_counter:04d}"
    
    def generate_reading_id(self) -> str:
        self._reading_counter += 1
        return f"TRD-{self._reading_counter:06d}"
    
    def generate_alert_id(self) -> str:
        self._alert_counter += 1
        return f"ALT-{self._alert_counter:06d}"
    
    def create_furnace(self, name: str, design_capacity: float, 
                       min_temperature: float, max_temperature: float) -> Furnace:
        furnace_id = self.generate_furnace_id()
        furnace = Furnace(
            id=furnace_id,
            name=name,
            design_capacity=design_capacity,
            min_temperature=min_temperature,
            max_temperature=max_temperature
        )
        self.furnaces[furnace_id] = furnace
        return furnace
    
    def get_furnace(self, furnace_id: str) -> Optional[Furnace]:
        return self.furnaces.get(furnace_id)
    
    def get_all_furnaces(self) -> List[Furnace]:
        return list(self.furnaces.values())
    
    def update_furnace(self, furnace: Furnace) -> bool:
        if furnace.id in self.furnaces:
            self.furnaces[furnace.id] = furnace
            return True
        return False
    
    def create_batching_order(self, materials: list) -> BatchingOrder:
        order_id = self.generate_order_id()
        order = BatchingOrder(
            id=order_id,
            materials=materials
        )
        self.batching_orders[order_id] = order
        return order
    
    def get_batching_order(self, order_id: str) -> Optional[BatchingOrder]:
        return self.batching_orders.get(order_id)
    
    def get_all_batching_orders(self) -> List[BatchingOrder]:
        return list(self.batching_orders.values())
    
    def update_batching_order(self, order: BatchingOrder) -> bool:
        if order.id in self.batching_orders:
            self.batching_orders[order.id] = order
            return True
        return False
    
    def get_pending_orders(self) -> List[BatchingOrder]:
        return [o for o in self.batching_orders.values() 
                if o.status == BatchingOrderStatus.PENDING]
    
    def get_assigned_orders(self) -> List[BatchingOrder]:
        return [o for o in self.batching_orders.values() 
                if o.status == BatchingOrderStatus.ASSIGNED]
    
    def create_temperature_reading(self, furnace_id: str, temperature: float) -> TemperatureReading:
        reading_id = self.generate_reading_id()
        reading = TemperatureReading(
            id=reading_id,
            furnace_id=furnace_id,
            temperature=temperature
        )
        if furnace_id not in self.temperature_readings:
            self.temperature_readings[furnace_id] = []
        self.temperature_readings[furnace_id].append(reading)
        return reading
    
    def get_furnace_temperatures(self, furnace_id: str, limit: int = 100) -> List[TemperatureReading]:
        readings = self.temperature_readings.get(furnace_id, [])
        return readings[-limit:]
    
    def create_alert(self, furnace_id: str, alert_type: AlertType, message: str) -> Alert:
        alert_id = self.generate_alert_id()
        alert = Alert(
            id=alert_id,
            furnace_id=furnace_id,
            type=alert_type,
            message=message
        )
        self.alerts.append(alert)
        return alert
    
    def get_all_alerts(self, unresolved_only: bool = False) -> List[Alert]:
        if unresolved_only:
            return [a for a in self.alerts if not a.is_resolved]
        return self.alerts
    
    def resolve_alert(self, alert_id: str) -> Optional[Alert]:
        for alert in self.alerts:
            if alert.id == alert_id:
                alert.is_resolved = True
                return alert
        return None


storage = Storage()
