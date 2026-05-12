from collections import defaultdict
from datetime import datetime
from typing import List, Dict
from sqlalchemy.orm import Session
from ..models import MaintenanceTask, GreenZone, TaskStatus


def schedule_tasks_by_zone(db: Session, limit: int = 50) -> List[Dict]:
    """
    调度器：优先安排同一区域内的多个任务，减少移动距离
    
    算法：
    1. 按区域分组待处理任务
    2. 对每个区域，按调度日期排序
    3. 优先返回任务数最多的区域的任务
    """
    pending_tasks = db.query(MaintenanceTask).join(GreenZone).filter(
        MaintenanceTask.status.in_([TaskStatus.PENDING, TaskStatus.OVERDUE])
    ).order_by(MaintenanceTask.scheduled_date.asc()).all()
    
    zone_task_count = defaultdict(int)
    zone_tasks = defaultdict(list)
    
    for task in pending_tasks:
        zone_task_count[task.green_zone_id] += 1
        zone_tasks[task.green_zone_id].append(task)
    
    sorted_zones = sorted(
        zone_tasks.keys(),
        key=lambda z: (-zone_task_count[z], min(t.scheduled_date for t in zone_tasks[z]))
    )
    
    scheduled_result = []
    remaining = limit
    
    for zone_id in sorted_zones:
        if remaining <= 0:
            break
        
        tasks_in_zone = sorted(
            zone_tasks[zone_id],
            key=lambda t: t.scheduled_date
        )
        
        tasks_to_take = min(len(tasks_in_zone), remaining)
        
        for task in tasks_in_zone[:tasks_to_take]:
            scheduled_result.append({
                "task_id": task.id,
                "task_code": task.task_code,
                "green_zone_id": task.green_zone_id,
                "green_zone_name": task.green_zone.name if task.green_zone else "Unknown",
                "scheduled_date": task.scheduled_date,
                "maintenance_grade": task.maintenance_grade,
                "status": task.status,
                "tasks_in_zone": zone_task_count[zone_id]
            })
        
        remaining -= tasks_to_take
    
    return scheduled_result


def get_overdue_tasks(db: Session) -> List[Dict]:
    """获取所有逾期任务"""
    now = datetime.utcnow()
    
    overdue_tasks = db.query(MaintenanceTask).join(GreenZone).filter(
        MaintenanceTask.status.in_([TaskStatus.PENDING, TaskStatus.IN_PROGRESS]),
        MaintenanceTask.scheduled_date < now
    ).order_by(MaintenanceTask.scheduled_date.asc()).all()
    
    result = []
    for task in overdue_tasks:
        result.append({
            "task_id": task.id,
            "task_code": task.task_code,
            "green_zone_id": task.green_zone_id,
            "green_zone_name": task.green_zone.name if task.green_zone else "Unknown",
            "scheduled_date": task.scheduled_date,
            "maintenance_grade": task.maintenance_grade,
            "days_overdue": (now - task.scheduled_date).days
        })
    
    return result


def get_zone_task_summary(db: Session) -> List[Dict]:
    """按区域统计任务情况"""
    zones = db.query(GreenZone).all()
    
    summary = []
    for zone in zones:
        pending_tasks = db.query(MaintenanceTask).filter(
            MaintenanceTask.green_zone_id == zone.id,
            MaintenanceTask.status == TaskStatus.PENDING
        ).count()
        
        overdue_tasks = db.query(MaintenanceTask).filter(
            MaintenanceTask.green_zone_id == zone.id,
            MaintenanceTask.status.in_([TaskStatus.PENDING, TaskStatus.OVERDUE]),
            MaintenanceTask.scheduled_date < datetime.utcnow()
        ).count()
        
        completed_tasks = db.query(MaintenanceTask).filter(
            MaintenanceTask.green_zone_id == zone.id,
            MaintenanceTask.status == TaskStatus.COMPLETED
        ).count()
        
        summary.append({
            "green_zone_id": zone.id,
            "zone_code": zone.zone_code,
            "name": zone.name,
            "current_grade": zone.current_grade,
            "consecutive_missed": zone.consecutive_missed_tasks,
            "pending_tasks": pending_tasks,
            "overdue_tasks": overdue_tasks,
            "completed_tasks": completed_tasks
        })
    
    return sorted(summary, key=lambda x: (-x["overdue_tasks"], -x["pending_tasks"]))
