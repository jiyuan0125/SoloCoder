from typing import List, Optional
from datetime import datetime, timedelta
from .models import (
    Furnace, BatchingOrder, Material, AlertType,
    FurnaceStatus, BatchingOrderStatus
)
from .storage import storage


class ServiceException(Exception):
    pass


class FurnaceService:
    @staticmethod
    def create_furnace(name: str, design_capacity: float,
                       min_temperature: float, max_temperature: float) -> Furnace:
        if design_capacity <= 0:
            raise ServiceException("炉子设计容量必须大于0")
        if min_temperature >= max_temperature:
            raise ServiceException("最低安全温度必须低于最高安全温度")
        return storage.create_furnace(name, design_capacity, min_temperature, max_temperature)
    
    @staticmethod
    def get_all_furnaces() -> List[Furnace]:
        return storage.get_all_furnaces()
    
    @staticmethod
    def get_furnace(furnace_id: str) -> Optional[Furnace]:
        return storage.get_furnace(furnace_id)
    
    @staticmethod
    def get_idle_furnaces() -> List[Furnace]:
        return [f for f in storage.get_all_furnaces() if f.status == FurnaceStatus.IDLE]


class BatchingService:
    @staticmethod
    def create_order(materials: List[Material]) -> BatchingOrder:
        if not materials:
            raise ServiceException("配料单必须至少包含一种原料")
        for m in materials:
            if m.weight <= 0:
                raise ServiceException(f"原料 {m.name} 的重量必须大于0")
        return storage.create_batching_order(materials)
    
    @staticmethod
    def get_all_orders() -> List[BatchingOrder]:
        return storage.get_all_batching_orders()
    
    @staticmethod
    def get_order(order_id: str) -> Optional[BatchingOrder]:
        return storage.get_batching_order(order_id)
    
    @staticmethod
    def check_capacity(order: BatchingOrder, furnace: Furnace) -> bool:
        total_weight = order.get_total_weight()
        return total_weight <= furnace.design_capacity
    
    @staticmethod
    def assign_order_to_furnace(order_id: str, furnace_id: str) -> BatchingOrder:
        order = storage.get_batching_order(order_id)
        if not order:
            raise ServiceException("配料单不存在")
        if order.status != BatchingOrderStatus.PENDING:
            raise ServiceException("只有待处理的配料单才能分配")
        
        furnace = storage.get_furnace(furnace_id)
        if not furnace:
            raise ServiceException("炉子不存在")
        if furnace.status != FurnaceStatus.IDLE:
            raise ServiceException("炉子当前不可用")
        if furnace.current_smelting_id is not None:
            raise ServiceException("炉子已有正在进行的冶炼任务")
        
        if not BatchingService.check_capacity(order, furnace):
            total_weight = order.get_total_weight()
            raise ServiceException(
                f"配料单总重量 {total_weight} 超过炉子设计容量 {furnace.design_capacity}"
            )
        
        order.furnace_id = furnace_id
        order.status = BatchingOrderStatus.ASSIGNED
        order.assigned_at = datetime.now()
        storage.update_batching_order(order)
        
        furnace.current_smelting_id = order_id
        furnace.status = FurnaceStatus.HEATING
        storage.update_furnace(furnace)
        
        return order
    
    @staticmethod
    def charge_materials(furnace_id: str, materials: List[Material]) -> BatchingOrder:
        furnace = storage.get_furnace(furnace_id)
        if not furnace:
            raise ServiceException("炉子不存在")
        if not furnace.current_smelting_id:
            raise ServiceException("炉子没有正在进行的配料单")
        
        order = storage.get_batching_order(furnace.current_smelting_id)
        if not order:
            raise ServiceException("配料单不存在")
        
        if order.status not in [BatchingOrderStatus.ASSIGNED, BatchingOrderStatus.SMELTING]:
            raise ServiceException("当前状态不允许投料")
        
        for m in materials:
            for original in order.materials:
                if original.name == m.name:
                    original.actual_weight = m.actual_weight
                    break
        
        if order.check_deviation():
            order.is_abnormal = True
            order.abnormal_reason = "投料重量偏差超过10%"
            order.status = BatchingOrderStatus.ABNORMAL
        
        order.status = BatchingOrderStatus.SMELTING
        order.started_at = datetime.now()
        storage.update_batching_order(order)
        
        furnace.status = FurnaceStatus.MELTING
        storage.update_furnace(furnace)
        
        return order
    
    @staticmethod
    def complete_smelting(furnace_id: str, output_weight: float, 
                          energy_consumed: float) -> BatchingOrder:
        furnace = storage.get_furnace(furnace_id)
        if not furnace:
            raise ServiceException("炉子不存在")
        if not furnace.current_smelting_id:
            raise ServiceException("炉子没有正在进行的配料单")
        
        order = storage.get_batching_order(furnace.current_smelting_id)
        if not order:
            raise ServiceException("配料单不存在")
        
        order.output_weight = output_weight
        order.energy_consumed = energy_consumed
        order.status = BatchingOrderStatus.COMPLETED
        order.completed_at = datetime.now()
        storage.update_batching_order(order)
        
        furnace.status = FurnaceStatus.COOLING
        storage.update_furnace(furnace)
        
        return order
    
    @staticmethod
    def set_furnace_idle(furnace_id: str) -> Furnace:
        furnace = storage.get_furnace(furnace_id)
        if not furnace:
            raise ServiceException("炉子不存在")
        
        furnace.status = FurnaceStatus.IDLE
        furnace.current_smelting_id = None
        storage.update_furnace(furnace)
        
        return furnace


class TemperatureService:
    @staticmethod
    def report_temperature(furnace_id: str, temperature: float) -> tuple:
        furnace = storage.get_furnace(furnace_id)
        if not furnace:
            raise ServiceException("炉子不存在")
        
        reading = storage.create_temperature_reading(furnace_id, temperature)
        
        is_alert = False
        alert_type = None
        message = ""
        
        if temperature > furnace.max_temperature:
            reading.is_alert = True
            is_alert = True
            alert_type = AlertType.TEMPERATURE_HIGH
            message = f"炉温 {temperature} 超过最高安全温度 {furnace.max_temperature}"
        elif temperature < furnace.min_temperature:
            reading.is_alert = True
            is_alert = True
            alert_type = AlertType.TEMPERATURE_LOW
            message = f"炉温 {temperature} 低于最低安全温度 {furnace.min_temperature}"
        
        alert = None
        if is_alert:
            alert = storage.create_alert(furnace_id, alert_type, message)
            
            if TemperatureService._check_continuous_alert(furnace_id, 5):
                storage.create_alert(
                    furnace_id,
                    AlertType.EMERGENCY,
                    f"炉温连续5分钟超出安全范围，请立即处理"
                )
        
        return reading, alert
    
    @staticmethod
    def _check_continuous_alert(furnace_id: str, minutes: int) -> bool:
        readings = storage.get_furnace_temperatures(furnace_id, limit=100)
        if len(readings) < 2:
            return False
        
        now = datetime.now()
        cutoff_time = now - timedelta(minutes=minutes)
        
        recent_readings = [r for r in readings if r.timestamp >= cutoff_time]
        if len(recent_readings) < 2:
            return False
        
        first_alert_time = None
        for reading in recent_readings:
            if reading.is_alert:
                if first_alert_time is None:
                    first_alert_time = reading.timestamp
            else:
                first_alert_time = None
        
        if first_alert_time:
            return (now - first_alert_time).total_seconds() >= minutes * 60
        return False
    
    @staticmethod
    def get_furnace_temperatures(furnace_id: str, limit: int = 100) -> List:
        return storage.get_furnace_temperatures(furnace_id, limit)


class AlertService:
    @staticmethod
    def get_all_alerts(unresolved_only: bool = False):
        return storage.get_all_alerts(unresolved_only)
    
    @staticmethod
    def resolve_alert(alert_id: str):
        return storage.resolve_alert(alert_id)
