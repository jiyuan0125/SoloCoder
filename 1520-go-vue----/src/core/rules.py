from typing import Tuple, Optional, List
from .models import (
    Species, Pond, PondTemperatureLog, 
    TodoItem, TodoPriority, TodoStatus
)


def check_temperature_range(
    temperature: float,
    species: Optional[Species]
) -> Tuple[bool, str]:
    """
    检查水温是否在品种适宜范围内
    """
    if not species:
        return True, "未指定品种，跳过温度检查"
    
    if temperature < species.min_temperature:
        return False, f"水温过低: {temperature}°C < {species.min_temperature}°C"
    
    if temperature > species.max_temperature:
        return False, f"水温过高: {temperature}°C > {species.max_temperature}°C"
    
    return True, "水温在适宜范围内"


def update_todo_priority(
    todo: TodoItem,
    consecutive_violations: int
) -> Tuple[TodoItem, bool]:
    """
    根据连续违规次数更新待办优先级
    - 连续3次及以上: 升级为紧急(URGENT)
    - 返回: (更新后的待办, 是否升级了优先级)
    """
    old_priority = todo.priority
    upgraded = False
    
    if consecutive_violations >= 3 and todo.priority != TodoPriority.URGENT:
        todo.priority = TodoPriority.URGENT
        todo.description = (
            f"{todo.description}\n【升级提醒】已连续{consecutive_violations}次超范围，"
            f"优先级已升级为紧急！"
        )
        upgraded = True
    elif consecutive_violations >= 2 and todo.priority == TodoPriority.MEDIUM:
        todo.priority = TodoPriority.HIGH
        upgraded = True
    
    return todo, upgraded


def find_existing_temperature_todo(
    todos: List[TodoItem],
    pond_id: str
) -> Optional[TodoItem]:
    """
    查找指定池塘的未解决温度待办
    """
    for todo in todos:
        if (
            todo.category == "temperature" 
            and todo.related_entity_id == pond_id
            and todo.status != TodoStatus.RESOLVED
        ):
            return todo
    return None
