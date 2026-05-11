from datetime import datetime
from typing import List
from .models import DeliveryTask, DeliveryTaskStatus
from .store import store


def get_tasks_by_date_range(start_date: datetime, end_date: datetime) -> List[DeliveryTask]:
    tasks = []
    for task in store.list_tasks():
        task_time = task.assigned_at or task.created_at
        if start_date <= task_time <= end_date:
            tasks.append(task)
    return tasks


def format_tasks_as_table(tasks: List[DeliveryTask]) -> str:
    if not tasks:
        return "No delivery records found in the specified date range."
    
    headers = [
        "Task ID",
        "Order ID",
        "Vehicle ID",
        "Origin",
        "Destination",
        "Weight (kg)",
        "Temp Range (°C)",
        "Status",
        "Created At",
        "Completed At"
    ]
    
    rows = []
    for task in tasks:
        origin = store.get_origin(task.origin_id)
        origin_name = origin.name if origin else "Unknown"
        
        temp_range = f"{task.min_temperature:.1f} - {task.max_temperature:.1f}"
        
        created_at = task.created_at.strftime("%Y-%m-%d %H:%M:%S")
        completed_at = task.completed_at.strftime("%Y-%m-%d %H:%M:%S") if task.completed_at else "N/A"
        
        rows.append([
            task.id,
            task.order_id,
            task.vehicle_id or "N/A",
            origin_name,
            task.destination_address,
            f"{task.weight:.1f}",
            temp_range,
            task.status.value,
            created_at,
            completed_at
        ])
    
    col_widths = [len(header) for header in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if len(cell) > col_widths[i]:
                col_widths[i] = len(cell)
    
    def format_row(row):
        return " | ".join(cell.ljust(width) for cell, width in zip(row, col_widths))
    
    separator = "-+-".join("-" * width for width in col_widths)
    
    lines = []
    lines.append(format_row(headers))
    lines.append(separator)
    for row in rows:
        lines.append(format_row(row))
    
    return "\n".join(lines)


def export_delivery_records(start_date: datetime, end_date: datetime) -> str:
    tasks = get_tasks_by_date_range(start_date, end_date)
    return format_tasks_as_table(tasks)
