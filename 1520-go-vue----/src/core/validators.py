from datetime import date, datetime
from typing import List, Tuple
from .models import (
    BreedingCycle, SalesOrder, DeliveryTask, 
    DeliveryStatus, InventoryItem
)


class ValidationError(Exception):
    """验证错误"""
    pass


def validate_dates(spawning_date: date, emergence_date: date) -> Tuple[bool, str]:
    """
    验证出苗日期不能早于产卵日期
    """
    if emergence_date < spawning_date:
        return False, "出苗日期不能早于产卵日期"
    return True, "OK"


def validate_sales_quantity(
    requested: int, 
    inventory: List[InventoryItem],
    species_id: str
) -> Tuple[bool, str, int]:
    """
    验证销售数量并计算可销售数量
    返回: (是否可以完全满足, 消息, 可销售数量)
    """
    total_available = sum(
        item.quantity 
        for item in inventory 
        if item.species_id == species_id
    )
    
    if total_available >= requested:
        return True, "库存充足", requested
    elif total_available > 0:
        return False, f"库存不足，仅可销售 {total_available} 尾（请求 {requested} 尾）", total_available
    else:
        return False, "库存为零，无法销售", 0


def validate_vehicle_availability(
    vehicle_id: str,
    scheduled_start: datetime,
    scheduled_end: datetime,
    existing_tasks: List[DeliveryTask],
    task_id: str = None
) -> Tuple[bool, str]:
    """
    验证车辆在指定时间段内是否可用
    同一车辆同一时间只能执行一个配送任务
    """
    for task in existing_tasks:
        if task.vehicle_id != vehicle_id:
            continue
        
        if task_id and task.id == task_id:
            continue
            
        if task.status == DeliveryStatus.CANCELLED or task.status == DeliveryStatus.COMPLETED:
            continue
            
        if task.actual_end:
            continue
            
        task_end = task.actual_end or task.scheduled_end or datetime.max
        
        if scheduled_start < task_end and scheduled_end > task.scheduled_start:
            return False, (
                f"车辆在时间段 {scheduled_start} 到 {scheduled_end} 不可用。"
                f"已有任务: {task.id} (时间段: {task.scheduled_start} - {task_end})"
            )
    
    return True, "车辆可用"
